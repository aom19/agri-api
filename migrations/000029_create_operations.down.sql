-- Remove permissions from roles
DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE name IN ('operations:read', 'operations:write', 'operations:delete'));

DELETE FROM permissions WHERE name IN ('operations:read', 'operations:write', 'operations:delete');

DROP TABLE IF EXISTS template_implement_types;
DROP TABLE IF EXISTS template_machine_types;
DROP TABLE IF EXISTS template_resources;
DROP TABLE IF EXISTS operation_templates;
DROP TABLE IF EXISTS operation_types;
