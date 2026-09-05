ALTER TABLE field_operations
    ADD COLUMN IF NOT EXISTS actual_start_at   TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS actual_end_at     TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS area_completed_ha DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS fuel_used_l       DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS machine_hours     DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS completion_notes  TEXT NOT NULL DEFAULT '';

-- Backfill: momentul pornirii din jurnalul de audit (acțiunea "start"), unde există.
UPDATE field_operations fo
SET actual_start_at = started.started_at
FROM (
    SELECT entity_id, MIN(created_at) AS started_at
    FROM audit_log
    WHERE entity_type = 'field_operation' AND action = 'start'
    GROUP BY entity_id
) started
WHERE started.entity_id = fo.id::text
  AND fo.actual_start_at IS NULL
  AND fo.status IN ('in_progress', 'completed');

-- Backfill: operațiunile deja finalizate primesc ca sfârșit real ultima modificare.
UPDATE field_operations
SET actual_end_at = updated_at
WHERE status = 'completed' AND actual_end_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_field_operations_actual_end_at
    ON field_operations (actual_end_at)
    WHERE deleted_at IS NULL;

INSERT INTO permissions (name, description)
VALUES ('field_operations:complete', 'Finalizare operațiuni pe teren (timpi reali, consumuri)')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code IN ('admin', 'manager', 'operator')
  AND p.name = 'field_operations:complete'
ON CONFLICT DO NOTHING;
