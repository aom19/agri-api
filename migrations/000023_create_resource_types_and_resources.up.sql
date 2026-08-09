CREATE TABLE IF NOT EXISTS resource_types (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    category    VARCHAR(50)  NOT NULL,
    default_unit VARCHAR(50) NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT resource_types_category_check
        CHECK (category IN ('fuel', 'fertilizer', 'seed', 'pesticide', 'water', 'other'))
);

CREATE TABLE IF NOT EXISTS resources (
    id               BIGSERIAL PRIMARY KEY,
    name             VARCHAR(255)   NOT NULL,
    resource_type_id BIGINT         NOT NULL REFERENCES resource_types(id) ON DELETE RESTRICT,
    price_per_unit   NUMERIC(14, 4) NOT NULL DEFAULT 0,
    notes            TEXT,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    CONSTRAINT resources_price_check CHECK (price_per_unit >= 0)
);

CREATE INDEX IF NOT EXISTS resources_resource_type_id_idx ON resources (resource_type_id);
