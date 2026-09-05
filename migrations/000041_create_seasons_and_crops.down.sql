DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE name IN ('crops:read', 'crops:write'));
DELETE FROM permissions WHERE name IN ('crops:read', 'crops:write');

DROP TABLE IF EXISTS field_crops;
DROP TABLE IF EXISTS crops;
DROP TABLE IF EXISTS seasons;
