-- T12: fiecare resursă avea cel mult un stoc (UNIQUE pe stocks.resource_id), deci stocul era de
-- fapt două coloane ale resursei. Cantitatea și pragul minim trec pe resources, tabelul stocks
-- dispare, iar mișcările de stoc trimit doar la resursă.

ALTER TABLE resources
    ADD COLUMN quantity         NUMERIC(14, 4) NOT NULL DEFAULT 0,
    ADD COLUMN minimum_quantity NUMERIC(14, 4) NOT NULL DEFAULT 0,
    ADD CONSTRAINT resources_quantity_check         CHECK (quantity >= 0),
    ADD CONSTRAINT resources_minimum_quantity_check CHECK (minimum_quantity >= 0);

UPDATE resources r
SET quantity = s.quantity, minimum_quantity = s.minimum_quantity
FROM stocks s
WHERE s.resource_id = r.id;

-- Intrările de stoc din jurnalul de audit și din notificări trimit acum la resursă. Pentru un stoc
-- șters, resursa vine din intrarea lui de creare.
UPDATE audit_log al
SET entity_id = COALESCE(
    (SELECT s.resource_id::text FROM stocks s WHERE s.id::text = al.entity_id),
    (SELECT c.changes->>'resource_id' FROM audit_log c
     WHERE c.entity_type = 'stock' AND c.entity_id = al.entity_id AND c.action = 'create'
     ORDER BY c.created_at LIMIT 1),
    al.entity_id
)
WHERE al.entity_type = 'stock';

UPDATE notifications n
SET entity_id = s.resource_id::text
FROM stocks s
WHERE n.entity_type = 'stock' AND n.entity_id = s.id::text;

ALTER TABLE stock_movements DROP COLUMN stock_id;

CREATE INDEX IF NOT EXISTS idx_stock_movements_resource_created
    ON stock_movements (resource_id, created_at DESC);

DROP TABLE stocks;

-- Stocul se gestionează cu permisiunile resurselor: citirea stocului devine resources:read, iar
-- mișcările de stoc cer resources:write. Rolurile care aveau stock.view sau stock.update le
-- primesc. stock.create și stock.delete nu au echivalent: fișa de stoc nu se mai creează sau
-- șterge separat de resursă. role_permissions se curăță prin ON DELETE CASCADE.
INSERT INTO permissions (name, description) VALUES
    ('resources:read',   'Vizualizare resurse și stocuri'),
    ('resources:write',  'Creare și editare resurse, mișcări de stoc'),
    ('resources:delete', 'Ștergere resurse')
ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description;

INSERT INTO role_permissions (role_id, permission_id)
SELECT rp.role_id, np.id
FROM role_permissions rp
JOIN permissions op ON op.id = rp.permission_id
JOIN permissions np ON np.name = CASE op.name
    WHEN 'stock.view' THEN 'resources:read'
    WHEN 'stock.update' THEN 'resources:write'
END
WHERE op.name IN ('stock.view', 'stock.update')
ON CONFLICT DO NOTHING;

DELETE FROM permissions WHERE name IN ('stock.view', 'stock.create', 'stock.update', 'stock.delete');
