-- T6: combustibilul consumat are o singură sursă: ieșirile din stoc pe resurse de combustibil
-- legate de operațiune. Câmpul de la finalizare generează o astfel de ieșire, iar rapoartele
-- citesc tot de acolo.

-- Valorile raportate înainte, pe care mișcările nu le acoperă, sunt păstrate în observații,
-- ca să nu se piardă la ștergerea coloanei.
UPDATE field_operations fo
SET completion_notes = BTRIM(CONCAT_WS(E'\n', NULLIF(fo.completion_notes, ''),
        FORMAT('Combustibil raportat (fără mișcare de stoc): %s l', fo.fuel_used_l)))
WHERE fo.fuel_used_l > 0
  AND fo.fuel_used_l <> COALESCE((
        SELECT SUM(-sm.quantity_delta)
        FROM stock_movements sm
        JOIN resources r ON r.id = sm.resource_id
        JOIN resource_types rt ON rt.id = r.resource_type_id
        WHERE sm.field_operation_id = fo.id AND sm.movement_type = 'out' AND rt.category = 'fuel'
      ), 0);

ALTER TABLE field_operations DROP COLUMN IF EXISTS fuel_used_l;
