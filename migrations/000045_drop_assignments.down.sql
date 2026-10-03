-- Recreează structura (000003 + 000005) și permisiunile; datele șterse nu se recuperează.
CREATE TABLE IF NOT EXISTS assignments (
    id          BIGSERIAL       PRIMARY KEY,
    machine_id  BIGINT          NOT NULL REFERENCES machines(id),
    operator_id BIGINT          NOT NULL REFERENCES operators(id),
    start_date  TIMESTAMPTZ     NOT NULL,
    end_date    TIMESTAMPTZ,
    status      VARCHAR(50)     NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

INSERT INTO permissions (name, description) VALUES
    ('assignments:read',   'Vizualizare asignări'),
    ('assignments:write',  'Creare și editare asignări'),
    ('assignments:delete', 'Ștergere asignări')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE (r.code = 'admin' AND p.name IN ('assignments:read', 'assignments:write', 'assignments:delete'))
   OR (r.code = 'manager' AND p.name IN ('assignments:read', 'assignments:write'))
   OR (r.code = 'viewer' AND p.name = 'assignments:read')
ON CONFLICT DO NOTHING;
