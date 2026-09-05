CREATE TABLE IF NOT EXISTS stock_movements (
    id                  BIGSERIAL PRIMARY KEY,
    stock_id            BIGINT NOT NULL REFERENCES stocks(id) ON DELETE CASCADE,
    resource_id         BIGINT NOT NULL REFERENCES resources(id),
    field_operation_id  BIGINT REFERENCES field_operations(id) ON DELETE SET NULL,
    movement_type       VARCHAR(20) NOT NULL CHECK (movement_type IN ('in', 'out', 'adjustment')),
    -- Variația de stoc (pozitivă la intrare, negativă la ieșire, oricare la ajustare).
    quantity_delta      NUMERIC(14,3) NOT NULL,
    -- Cantitatea rămasă în stoc după aplicarea mișcării.
    resulting_quantity  NUMERIC(14,3) NOT NULL,
    unit_cost           NUMERIC(14,4),
    -- Valoarea mișcării (|quantity_delta| × unit_cost), pozitivă.
    total_cost          NUMERIC(16,2),
    notes               TEXT NOT NULL DEFAULT '',
    actor_id            BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_stock_movements_stock_created
    ON stock_movements (stock_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_stock_movements_field_operation
    ON stock_movements (field_operation_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_created_at
    ON stock_movements (created_at);
