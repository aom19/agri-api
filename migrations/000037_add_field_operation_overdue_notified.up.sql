-- Marchează momentul în care s-a trimis notificarea de depășire a timpului estimat de lucru
ALTER TABLE field_operations
    ADD COLUMN IF NOT EXISTS overdue_notified_at TIMESTAMPTZ;

-- Index parțial pentru verificarea periodică a operațiunilor în lucru care au depășit sfârșitul planificat
CREATE INDEX IF NOT EXISTS idx_field_operations_overdue_check
    ON field_operations(planned_end_at)
    WHERE status = 'in_progress'
      AND overdue_notified_at IS NULL
      AND deleted_at IS NULL;
