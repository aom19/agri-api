DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions
    WHERE name IN ('field_operations:read', 'field_operations:write', 'field_operations:delete')
);

DELETE FROM permissions
WHERE name IN ('field_operations:read', 'field_operations:write', 'field_operations:delete');

DROP TABLE IF EXISTS field_operations;
