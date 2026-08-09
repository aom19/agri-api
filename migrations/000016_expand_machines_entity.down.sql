DROP INDEX IF EXISTS machines_registration_number_unique_idx;
DROP INDEX IF EXISTS machines_code_unique_idx;

ALTER TABLE machines
    DROP CONSTRAINT IF EXISTS machines_year_check,
    DROP CONSTRAINT IF EXISTS machines_status_check,
    DROP CONSTRAINT IF EXISTS machines_fuel_type_check,
    DROP CONSTRAINT IF EXISTS machines_type_check;

ALTER TABLE machines
    ALTER COLUMN status SET DEFAULT 'available';

ALTER TABLE machines
    DROP COLUMN IF EXISTS code,
    DROP COLUMN IF EXISTS brand,
    DROP COLUMN IF EXISTS model,
    DROP COLUMN IF EXISTS year,
    DROP COLUMN IF EXISTS registration_number,
    DROP COLUMN IF EXISTS fuel_type;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'machines' AND column_name = 'notes'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'machines' AND column_name = 'description'
    ) THEN
        EXECUTE 'ALTER TABLE machines RENAME COLUMN notes TO description';
    END IF;
END $$;
