DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE name IN ('field_operations:checklist', 'field_operations:start')
);

DELETE FROM permissions
WHERE name IN ('field_operations:checklist', 'field_operations:start');

ALTER TABLE field_operations
    DROP COLUMN IF EXISTS checklist_updated_at,
    DROP COLUMN IF EXISTS check_notes_confirmed,
    DROP COLUMN IF EXISTS check_field_area,
    DROP COLUMN IF EXISTS check_implement_status,
    DROP COLUMN IF EXISTS check_machine_status;
