DELETE FROM role_permissions
WHERE permission_id = (SELECT id FROM permissions WHERE name = 'dashboard:read');

DELETE FROM permissions WHERE name = 'dashboard:read';