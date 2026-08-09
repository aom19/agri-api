DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE code = 'operator')
  AND permission_id IN (
      SELECT id
      FROM permissions
      WHERE name IN (
          'dashboard:read',
          'assignments:read',
          'field_operations:read',
          'fields:read',
          'machines:read',
          'implements:read',
          'operators:read',
          'operations:read'
      )
  );
