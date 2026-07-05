ALTER TABLE machines
    ADD COLUMN IF NOT EXISTS code VARCHAR(50),
    ADD COLUMN IF NOT EXISTS brand VARCHAR(100),
    ADD COLUMN IF NOT EXISTS model VARCHAR(100),
    ADD COLUMN IF NOT EXISTS year INT,
    ADD COLUMN IF NOT EXISTS registration_number VARCHAR(50),
    ADD COLUMN IF NOT EXISTS fuel_type VARCHAR(20),
    ADD COLUMN IF NOT EXISTS notes TEXT;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'machines' AND column_name = 'description'
    ) THEN
        EXECUTE 'UPDATE machines SET notes = COALESCE(notes, description) WHERE notes IS NULL';
        EXECUTE 'ALTER TABLE machines DROP COLUMN description';
    END IF;
END $$;

UPDATE machines
SET code = 'MCH-' || id
WHERE code IS NULL OR btrim(code) = '';

UPDATE machines
SET type = 'other'
WHERE type IS NULL OR type NOT IN ('tractor', 'combine', 'drone', 'sprayer', 'other');

UPDATE machines
SET status = CASE
    WHEN status IN ('active', 'inactive') THEN status
    WHEN status = 'in_service' THEN 'maintenance'
    ELSE 'active'
END
WHERE status IS NULL OR status NOT IN ('active', 'maintenance', 'inactive');

UPDATE machines
SET registration_number = NULL
WHERE registration_number IS NOT NULL AND btrim(registration_number) = '';

UPDATE machines
SET fuel_type = NULL
WHERE fuel_type IS NOT NULL AND fuel_type NOT IN ('diesel', 'gasoline', 'electric', 'hybrid');

ALTER TABLE machines
    ALTER COLUMN code SET NOT NULL,
    ALTER COLUMN type SET NOT NULL,
    ALTER COLUMN status SET NOT NULL,
    ALTER COLUMN status SET DEFAULT 'active';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'machines_type_check'
    ) THEN
        ALTER TABLE machines
            ADD CONSTRAINT machines_type_check
            CHECK (type IN ('tractor', 'combine', 'drone', 'sprayer', 'other'));
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'machines_fuel_type_check'
    ) THEN
        ALTER TABLE machines
            ADD CONSTRAINT machines_fuel_type_check
            CHECK (fuel_type IS NULL OR fuel_type IN ('diesel', 'gasoline', 'electric', 'hybrid'));
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'machines_status_check'
    ) THEN
        ALTER TABLE machines
            ADD CONSTRAINT machines_status_check
            CHECK (status IN ('active', 'maintenance', 'inactive'));
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'machines_year_check'
    ) THEN
        ALTER TABLE machines
            ADD CONSTRAINT machines_year_check
            CHECK (year IS NULL OR year BETWEEN 1900 AND 2100);
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS machines_code_unique_idx
    ON machines (code)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS machines_registration_number_unique_idx
    ON machines (registration_number)
    WHERE registration_number IS NOT NULL AND deleted_at IS NULL;
