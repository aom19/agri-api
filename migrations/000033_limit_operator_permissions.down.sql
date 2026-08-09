INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'operator'
  AND p.name IN (
      'assignments:read',
      'fields:read',
      'machines:read',
      'implements:read',
      'operators:read',
      'operations:read'
  )
ON CONFLICT DO NOTHING;
