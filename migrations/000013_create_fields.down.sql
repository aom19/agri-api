DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE name IN ('fields:read', 'fields:write', 'fields:delete')
);

DELETE FROM permissions
WHERE name IN ('fields:read', 'fields:write', 'fields:delete');

DROP INDEX IF EXISTS idx_fields_geometry_gin;
DROP INDEX IF EXISTS idx_fields_created_at;
DROP TABLE IF EXISTS fields;
