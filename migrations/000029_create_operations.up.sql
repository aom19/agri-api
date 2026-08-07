-- ============================================================
-- Operations: operation_types, operation_templates, template_resources,
--             template_machine_types, template_implement_types
-- ============================================================

CREATE TABLE IF NOT EXISTS operation_types (
    id          BIGSERIAL PRIMARY KEY,
    code        VARCHAR(50) NOT NULL UNIQUE,
    name        VARCHAR(100) NOT NULL,
    description TEXT DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS operation_templates (
    id                BIGSERIAL PRIMARY KEY,
    operation_type_id BIGINT NOT NULL REFERENCES operation_types(id),
    name              VARCHAR(200) NOT NULL,
    description       TEXT DEFAULT '',
    unit              VARCHAR(20) NOT NULL DEFAULT 'ha',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS template_resources (
    id                BIGSERIAL PRIMARY KEY,
    template_id       BIGINT NOT NULL REFERENCES operation_templates(id) ON DELETE CASCADE,
    resource_id       BIGINT NOT NULL REFERENCES resources(id),
    quantity_per_unit DOUBLE PRECISION NOT NULL DEFAULT 0,
    notes             TEXT DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (template_id, resource_id)
);

CREATE TABLE IF NOT EXISTS template_machine_types (
    id          BIGSERIAL PRIMARY KEY,
    template_id BIGINT NOT NULL REFERENCES operation_templates(id) ON DELETE CASCADE,
    machine_type VARCHAR(50) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (template_id, machine_type)
);

CREATE TABLE IF NOT EXISTS template_implement_types (
    id              BIGSERIAL PRIMARY KEY,
    template_id     BIGINT NOT NULL REFERENCES operation_templates(id) ON DELETE CASCADE,
    implement_type  VARCHAR(50) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (template_id, implement_type)
);

-- Permissions
INSERT INTO permissions (name, description) VALUES
    ('operations:read',  'Vizualizare tipuri operațiuni și template-uri'),
    ('operations:write', 'Creare și editare tipuri operațiuni și template-uri'),
    ('operations:delete','Ștergere tipuri operațiuni și template-uri')
ON CONFLICT (name) DO NOTHING;

-- Grant to admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'admin'
  AND p.name IN ('operations:read', 'operations:write', 'operations:delete')
ON CONFLICT DO NOTHING;

-- Grant read+write to manager
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'manager'
  AND p.name IN ('operations:read', 'operations:write')
ON CONFLICT DO NOTHING;

-- Grant read to viewer
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'viewer'
  AND p.name = 'operations:read'
ON CONFLICT DO NOTHING;
