-- Ensure users:disable and users:enable permissions exist and are attached to admin.

INSERT INTO permissions (name, description)
VALUES ('users:disable', 'Dezactivare utilizatori')
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (name, description)
VALUES ('users:enable', 'Reactivare utilizatori')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name = 'users:disable'
WHERE r.code = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name = 'users:enable'
WHERE r.code = 'admin'
ON CONFLICT DO NOTHING;
