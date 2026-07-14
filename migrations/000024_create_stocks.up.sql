CREATE TABLE IF NOT EXISTS stocks (
    id               BIGSERIAL      PRIMARY KEY,
    resource_id      BIGINT         NOT NULL REFERENCES resources(id) ON DELETE RESTRICT,
    quantity         NUMERIC(14, 4) NOT NULL DEFAULT 0,
    minimum_quantity NUMERIC(14, 4) NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    CONSTRAINT stocks_quantity_check         CHECK (quantity >= 0),
    CONSTRAINT stocks_minimum_quantity_check CHECK (minimum_quantity >= 0),
    CONSTRAINT stocks_resource_id_unique     UNIQUE (resource_id)
);
