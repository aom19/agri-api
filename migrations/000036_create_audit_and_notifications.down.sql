DROP TABLE IF EXISTS user_notifications;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS audit_log;

DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE name = 'notifications:read');

DELETE FROM permissions WHERE name = 'notifications:read';
