DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE name = 'field_operations:complete');
DELETE FROM permissions WHERE name = 'field_operations:complete';

DROP INDEX IF EXISTS idx_field_operations_actual_end_at;

ALTER TABLE field_operations
    DROP COLUMN IF EXISTS actual_start_at,
    DROP COLUMN IF EXISTS actual_end_at,
    DROP COLUMN IF EXISTS area_completed_ha,
    DROP COLUMN IF EXISTS fuel_used_l,
    DROP COLUMN IF EXISTS machine_hours,
    DROP COLUMN IF EXISTS completion_notes;
