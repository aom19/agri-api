DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE code = 'operator')
  AND permission_id IN (
      SELECT id
      FROM permissions
      WHERE name NOT IN ('dashboard:read', 'field_operations:read')
  );

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'operator'
  AND p.name IN ('dashboard:read', 'field_operations:read')
ON CONFLICT DO NOTHING;
