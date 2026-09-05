CREATE TABLE IF NOT EXISTS seasons (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(120) NOT NULL UNIQUE,
    start_date  DATE NOT NULL,
    end_date    DATE NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT FALSE,
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (end_date >= start_date)
);

CREATE TABLE IF NOT EXISTS crops (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(120) NOT NULL UNIQUE,
    code        VARCHAR(40) UNIQUE,
    category    VARCHAR(60) NOT NULL DEFAULT '',
    yield_unit  VARCHAR(20) NOT NULL DEFAULT 't',
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS field_crops (
    id                    BIGSERIAL PRIMARY KEY,
    field_id              UUID NOT NULL REFERENCES fields(id) ON DELETE CASCADE,
    season_id             BIGINT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
    crop_id               BIGINT NOT NULL REFERENCES crops(id),
    planted_area_ha       DOUBLE PRECISION,
    planted_at            DATE,
    harvested_at          DATE,
    production_total      NUMERIC(14,3),
    expected_yield_per_ha NUMERIC(10,3),
    notes                 TEXT NOT NULL DEFAULT '',
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (field_id, season_id)
);

CREATE INDEX IF NOT EXISTS idx_field_crops_season ON field_crops (season_id);
CREATE INDEX IF NOT EXISTS idx_field_crops_crop ON field_crops (crop_id);

-- Catalog inițial de culturi (poate fi completat din aplicație).
INSERT INTO crops (name, code, category, yield_unit) VALUES
    ('Grâu', 'wheat', 'cereale', 't'),
    ('Porumb', 'corn', 'cereale', 't'),
    ('Orz', 'barley', 'cereale', 't'),
    ('Floarea-soarelui', 'sunflower', 'oleaginoase', 't'),
    ('Rapiță', 'rapeseed', 'oleaginoase', 't'),
    ('Soia', 'soy', 'leguminoase', 't'),
    ('Sfeclă de zahăr', 'sugar_beet', 'rădăcinoase', 't'),
    ('Lucernă', 'alfalfa', 'furaje', 't')
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (name, description) VALUES
    ('crops:read',  'Vizualizare sezoane, culturi și culturi pe terenuri'),
    ('crops:write', 'Administrare sezoane, culturi și culturi pe terenuri')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE (r.code IN ('admin', 'manager') AND p.name IN ('crops:read', 'crops:write'))
   OR (r.code IN ('viewer', 'operator') AND p.name = 'crops:read')
ON CONFLICT DO NOTHING;
