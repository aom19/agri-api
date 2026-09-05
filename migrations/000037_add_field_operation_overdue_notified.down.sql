DROP INDEX IF EXISTS idx_field_operations_overdue_check;

ALTER TABLE field_operations
    DROP COLUMN IF EXISTS overdue_notified_at;
