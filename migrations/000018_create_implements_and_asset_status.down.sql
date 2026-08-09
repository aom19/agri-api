DROP INDEX IF EXISTS idx_implement_compatibilities_implement_type;
DROP INDEX IF EXISTS idx_implement_compatibilities_machine_type;

DROP TABLE IF EXISTS implement_compatibilities;

DROP INDEX IF EXISTS implements_code_unique_idx;
DROP TABLE IF EXISTS implements;

ALTER TABLE machines
    DROP CONSTRAINT IF EXISTS machines_asset_status_check,
    DROP CONSTRAINT IF EXISTS machines_status_check;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'machines' AND column_name = 'asset_status'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'machines' AND column_name = 'status'
    ) THEN
        EXECUTE 'ALTER TABLE machines RENAME COLUMN asset_status TO status';
    END IF;
END $$;

UPDATE machines
SET status = 'active'
WHERE status IS NULL OR status NOT IN ('active', 'maintenance', 'inactive');

ALTER TABLE machines
    ALTER COLUMN status SET NOT NULL,
    ALTER COLUMN status SET DEFAULT 'active';

ALTER TABLE machines
    ADD CONSTRAINT machines_status_check
    CHECK (status IN ('active', 'maintenance', 'inactive'));

ALTER TABLE machines
    DROP COLUMN IF EXISTS operating_hours;
