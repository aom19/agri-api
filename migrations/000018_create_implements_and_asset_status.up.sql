ALTER TABLE machines
    ADD COLUMN IF NOT EXISTS operating_hours DOUBLE PRECISION;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'machines' AND column_name = 'status'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'machines' AND column_name = 'asset_status'
    ) THEN
        EXECUTE 'ALTER TABLE machines RENAME COLUMN status TO asset_status';
    END IF;
END $$;

UPDATE machines
SET asset_status = 'active'
WHERE asset_status IS NULL OR asset_status NOT IN ('active', 'maintenance', 'inactive');

ALTER TABLE machines
    ALTER COLUMN asset_status SET NOT NULL,
    ALTER COLUMN asset_status SET DEFAULT 'active';

ALTER TABLE machines
    DROP CONSTRAINT IF EXISTS machines_status_check,
    DROP CONSTRAINT IF EXISTS machines_asset_status_check;

ALTER TABLE machines
    ADD CONSTRAINT machines_asset_status_check
    CHECK (asset_status IN ('active', 'maintenance', 'inactive'));

CREATE TABLE IF NOT EXISTS implements (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    type VARCHAR(50) NOT NULL,
    brand VARCHAR(100),
    model VARCHAR(100),
    year INT,
    working_width DOUBLE PRECISION,
    capacity DOUBLE PRECISION,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT implements_type_check CHECK (type IN ('plow', 'disc_harrow', 'cultivator', 'seeder', 'fertilizer_spreader', 'sprayer', 'trailer', 'header', 'other')),
    CONSTRAINT implements_status_check CHECK (status IN ('active', 'maintenance', 'inactive')),
    CONSTRAINT implements_year_check CHECK (year IS NULL OR year BETWEEN 1900 AND 2100)
);

CREATE UNIQUE INDEX IF NOT EXISTS implements_code_unique_idx
    ON implements (code);

CREATE TABLE IF NOT EXISTS implement_compatibilities (
    id BIGSERIAL PRIMARY KEY,
    machine_type VARCHAR(50) NOT NULL,
    implement_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT implement_compatibilities_machine_type_check CHECK (machine_type IN ('tractor', 'combine', 'drone', 'sprayer', 'car', 'small_truck', 'other')),
    CONSTRAINT implement_compatibilities_implement_type_check CHECK (implement_type IN ('plow', 'disc_harrow', 'cultivator', 'seeder', 'fertilizer_spreader', 'sprayer', 'trailer', 'header', 'other')),
    CONSTRAINT implement_compatibilities_machine_implement_unique UNIQUE (machine_type, implement_type)
);

CREATE INDEX IF NOT EXISTS idx_implement_compatibilities_machine_type
    ON implement_compatibilities (machine_type);

CREATE INDEX IF NOT EXISTS idx_implement_compatibilities_implement_type
    ON implement_compatibilities (implement_type);
