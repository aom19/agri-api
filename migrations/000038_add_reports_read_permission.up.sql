INSERT INTO permissions (name, description)
VALUES ('reports:read', 'Vizualizare rapoarte')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code IN ('admin', 'manager', 'viewer')
  AND p.name = 'reports:read'
ON CONFLICT DO NOTHING;
