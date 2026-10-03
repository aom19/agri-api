ALTER TABLE field_operations ADD COLUMN IF NOT EXISTS fuel_used_l DOUBLE PRECISION;

-- Coloana e refăcută din ieșirile de combustibil legate de operațiune.
UPDATE field_operations fo
SET fuel_used_l = fuel.total
FROM (
    SELECT sm.field_operation_id, SUM(-sm.quantity_delta) AS total
    FROM stock_movements sm
    JOIN resources r ON r.id = sm.resource_id
    JOIN resource_types rt ON rt.id = r.resource_type_id
    WHERE sm.movement_type = 'out' AND rt.category = 'fuel' AND sm.field_operation_id IS NOT NULL
    GROUP BY sm.field_operation_id
) fuel
WHERE fuel.field_operation_id = fo.id;
