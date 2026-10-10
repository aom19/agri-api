-- Recreează tabelul stocks (000024), cu câte un stoc pentru fiecare resursă. Stocul primește
-- id-ul resursei, ca mișcările, jurnalul de audit și notificările să rămână valide.
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

INSERT INTO stocks (id, resource_id, quantity, minimum_quantity, created_at, updated_at)
SELECT id, id, quantity, minimum_quantity, created_at, updated_at FROM resources;

SELECT setval(pg_get_serial_sequence('stocks', 'id'), COALESCE((SELECT MAX(id) FROM stocks), 0) + 1, false);

DROP INDEX IF EXISTS idx_stock_movements_resource_created;

ALTER TABLE stock_movements ADD COLUMN stock_id BIGINT REFERENCES stocks(id) ON DELETE CASCADE;
UPDATE stock_movements SET stock_id = resource_id;
ALTER TABLE stock_movements ALTER COLUMN stock_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_stock_movements_stock_created
    ON stock_movements (stock_id, created_at DESC);

ALTER TABLE resources
    DROP CONSTRAINT IF EXISTS resources_quantity_check,
    DROP CONSTRAINT IF EXISTS resources_minimum_quantity_check,
    DROP COLUMN quantity,
    DROP COLUMN minimum_quantity;

-- Permisiunile stock.* revin pentru rolurile care au echivalentul pe resurse.
INSERT INTO permissions (name, description) VALUES
    ('stock.view',   'Vizualizare stocuri'),
    ('stock.create', 'Creare stocuri'),
    ('stock.update', 'Editare stocuri'),
    ('stock.delete', 'Ștergere stocuri')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT rp.role_id, np.id
FROM role_permissions rp
JOIN permissions op ON op.id = rp.permission_id
JOIN permissions np ON
    (op.name = 'resources:read' AND np.name = 'stock.view')
    OR (op.name = 'resources:write' AND np.name IN ('stock.create', 'stock.update'))
    OR (op.name = 'resources:delete' AND np.name = 'stock.delete')
ON CONFLICT DO NOTHING;

UPDATE permissions SET description = 'Vizualizare resurse' WHERE name = 'resources:read';
UPDATE permissions SET description = 'Creare și editare resurse' WHERE name = 'resources:write';
