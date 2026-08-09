INSERT INTO permissions (name, description)
VALUES ('dashboard:read', 'Vizualizare carduri dashboard')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code IN ('admin', 'manager', 'viewer')
  AND p.name = 'dashboard:read'
ON CONFLICT DO NOTHING;