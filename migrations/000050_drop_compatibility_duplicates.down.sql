-- Recreează structura (000018, 000048). Compatibilitățile primesc din nou cele 8 rânduri din seed;
-- tipurile de mașini permise operatorilor nu se recuperează.
ALTER TABLE user_profiles
    ADD COLUMN IF NOT EXISTS allowed_machine_types TEXT[] NOT NULL DEFAULT '{}';

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

INSERT INTO implement_compatibilities (machine_type, implement_type) VALUES
    ('tractor', 'plow'),
    ('tractor', 'seeder'),
    ('tractor', 'fertilizer_spreader'),
    ('tractor', 'sprayer'),
    ('tractor', 'trailer'),
    ('combine', 'header'),
    ('combine', 'trailer'),
    ('drone', 'sprayer')
ON CONFLICT DO NOTHING;
