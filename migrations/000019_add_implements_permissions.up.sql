-- Add implements permissions and grant them to admin and manager roles.

INSERT INTO permissions (name, description)
VALUES
    ('implements:read', 'Vizualizare implementuri'),
    ('implements:write', 'Creare și editare implementuri'),
    ('implements:delete', 'Ștergere implementuri')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN ('implements:read', 'implements:write', 'implements:delete')
WHERE r.code IN ('admin', 'manager')
ON CONFLICT DO NOTHING;
