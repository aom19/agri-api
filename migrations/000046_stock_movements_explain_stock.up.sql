-- T5: stocul se modifică doar prin mișcări, deci suma mișcărilor fiecărui stoc trebuie să fie
-- egală cu cantitatea lui curentă.

-- Mișcările folosesc aceeași precizie ca stocul (4 zecimale), altfel rotunjirea strică egalitatea.
ALTER TABLE stock_movements
    ALTER COLUMN quantity_delta TYPE NUMERIC(14,4),
    ALTER COLUMN resulting_quantity TYPE NUMERIC(14,4);

-- Editarea directă a cantității (PATCH /stocks/:id) a lăsat stocuri pe care istoricul nu le
-- explică. Pentru fiecare, adaugă o ajustare de inventar cu diferența. Când soldul de deschidere
-- rezultat e pozitiv, ajustarea e datată înaintea primei mișcări, ca istoricul să rămână coerent;
-- altfel e datată acum și aduce stocul la cantitatea curentă.
WITH totals AS (
    SELECT
        s.id AS stock_id,
        s.resource_id,
        s.quantity,
        s.created_at,
        r.price_per_unit,
        s.quantity - COALESCE(SUM(sm.quantity_delta), 0) AS difference,
        MIN(sm.created_at) AS first_movement_at
    FROM stocks s
    JOIN resources r ON r.id = s.resource_id
    LEFT JOIN stock_movements sm ON sm.stock_id = s.id
    GROUP BY s.id, s.resource_id, s.quantity, s.created_at, r.price_per_unit
)
INSERT INTO stock_movements (
    stock_id, resource_id, movement_type, quantity_delta, resulting_quantity,
    unit_cost, total_cost, notes, created_at
)
SELECT
    stock_id,
    resource_id,
    'adjustment',
    difference,
    CASE WHEN difference >= 0 THEN difference ELSE quantity END,
    price_per_unit,
    ABS(difference) * price_per_unit,
    'Sold inițial (reconciliere istoric)',
    CASE
        WHEN difference >= 0 THEN LEAST(created_at, first_movement_at - INTERVAL '1 millisecond')
        ELSE NOW()
    END
FROM totals
WHERE difference <> 0;
