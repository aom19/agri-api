CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    area_ha DOUBLE PRECISION,
    geometry JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_fields_created_at ON fields(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_fields_geometry_gin ON fields USING GIN (geometry);

INSERT INTO permissions (name, description)
VALUES
    ('fields:read', 'Vizualizare terenuri'),
    ('fields:write', 'Creare si editare terenuri'),
    ('fields:delete', 'Stergere terenuri')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN ('fields:read', 'fields:write', 'fields:delete')
WHERE r.code = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN ('fields:read', 'fields:write')
WHERE r.code = 'manager'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name = 'fields:read'
WHERE r.code = 'viewer'
ON CONFLICT DO NOTHING;
