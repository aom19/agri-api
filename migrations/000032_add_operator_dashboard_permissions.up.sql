INSERT INTO roles (code, name, description)
VALUES ('operator', 'Operator', 'Operator de teren — acces la lucrările și alocările proprii')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    description = CASE
        WHEN roles.description = '' OR roles.description = 'Migrated from legacy users.role'
            THEN EXCLUDED.description
        ELSE roles.description
    END;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'operator'
  AND p.name IN (
      'dashboard:read',
      'assignments:read',
      'field_operations:read',
      'fields:read',
      'machines:read',
      'implements:read',
      'operators:read',
      'operations:read'
  )
ON CONFLICT DO NOTHING;
