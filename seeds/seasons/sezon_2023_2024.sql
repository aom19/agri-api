-- Seed sezon Sezon 2023/2024 — generat de seeds/seasons/generate.py; nu edita manual, modifică generatorul.
-- Rulează după seeds/seed.sql. Poate fi rulat de mai multe ori: își șterge întâi propriile date.
BEGIN;

-- ─── Funcție temporară: o mișcare de stoc care păstrează invariantul stoc = Σ mișcări ───
CREATE OR REPLACE FUNCTION pg_temp.seed_move(p_resource text, p_type text, p_qty numeric, p_at timestamptz, p_op bigint, p_fc bigint, p_notes text)
RETURNS void LANGUAGE plpgsql AS $f$
DECLARE v_res bigint; v_price numeric; v_qty numeric; v_delta numeric; v_new numeric;
BEGIN
    SELECT r.id, r.price_per_unit, r.quantity INTO v_res, v_price, v_qty
    FROM resources r WHERE r.name = p_resource FOR UPDATE;
    IF v_res IS NULL THEN RAISE EXCEPTION 'resursa % nu există', p_resource; END IF;
    v_delta := CASE WHEN p_type = 'in' THEN ROUND(p_qty, 4) ELSE -ROUND(p_qty, 4) END;
    v_new := v_qty + v_delta;
    IF v_new < 0 THEN RAISE EXCEPTION 'stoc negativ pentru % (% %)', p_resource, v_qty, v_delta; END IF;
    INSERT INTO stock_movements (resource_id, field_operation_id, movement_type, quantity_delta, resulting_quantity, unit_cost, total_cost, notes, field_crop_id, created_at)
    VALUES (v_res, p_op, p_type, v_delta, v_new, v_price, ABS(v_delta) * v_price, p_notes, p_fc, p_at);
    UPDATE resources SET quantity = v_new, updated_at = NOW() WHERE id = v_res;
END $f$;

-- ─── Curățenie: datele acestui sezon, dacă a mai fost rulat ───
DO $clean$
DECLARE v_season bigint;
BEGIN
    SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2023/2024';
    IF v_season IS NOT NULL THEN
        DELETE FROM stock_movements WHERE field_operation_id IN (SELECT id FROM field_operations WHERE field_crop_id IN (SELECT id FROM field_crops WHERE season_id = v_season));
        DELETE FROM stock_movements WHERE field_crop_id IN (SELECT id FROM field_crops WHERE season_id = v_season);
        DELETE FROM field_operations WHERE field_crop_id IN (SELECT id FROM field_crops WHERE season_id = v_season);
        DELETE FROM field_crops WHERE season_id = v_season;
        DELETE FROM seasons WHERE id = v_season;
    END IF;
    -- aprovizionările și lucrările fără cultură ale sezonului sunt etichetate în notes
    DELETE FROM stock_movements WHERE field_operation_id IN (SELECT id FROM field_operations WHERE notes LIKE '[Sezon 2023/2024]%');
    DELETE FROM field_operations WHERE notes LIKE '[Sezon 2023/2024]%';
    DELETE FROM stock_movements WHERE notes LIKE '[Sezon 2023/2024]%';
    UPDATE resources r SET quantity = COALESCE((SELECT SUM(sm.quantity_delta) FROM stock_movements sm WHERE sm.resource_id = r.id), 0);
END $clean$;

-- ─── Resurse, resurse de recoltă și șabloane care lipsesc (idempotent) ───
INSERT INTO resource_types (name, category, default_unit) SELECT 'Seminte', 'seed', 'kg' WHERE NOT EXISTS (SELECT 1 FROM resource_types WHERE category = 'seed' AND default_unit = 'kg');
INSERT INTO resources (name, resource_type_id, price_per_unit, minimum_quantity, notes) SELECT 'Samanta rapita hibrid', (SELECT id FROM resource_types WHERE category = 'seed' AND default_unit = 'kg' ORDER BY id LIMIT 1), 95.0, 20, 'Creată de seed-ul pe sezoane' WHERE NOT EXISTS (SELECT 1 FROM resources WHERE name = 'Samanta rapita hibrid');
INSERT INTO resource_types (name, category, default_unit) SELECT 'Seminte', 'seed', 'kg' WHERE NOT EXISTS (SELECT 1 FROM resource_types WHERE category = 'seed' AND default_unit = 'kg');
INSERT INTO resources (name, resource_type_id, price_per_unit, minimum_quantity, notes) SELECT 'Samanta floarea-soarelui hibrid', (SELECT id FROM resource_types WHERE category = 'seed' AND default_unit = 'kg' ORDER BY id LIMIT 1), 320.0, 15, 'Creată de seed-ul pe sezoane' WHERE NOT EXISTS (SELECT 1 FROM resources WHERE name = 'Samanta floarea-soarelui hibrid');
INSERT INTO resource_types (name, category, default_unit) SELECT 'Seminte', 'seed', 'kg' WHERE NOT EXISTS (SELECT 1 FROM resource_types WHERE category = 'seed' AND default_unit = 'kg');
INSERT INTO resources (name, resource_type_id, price_per_unit, minimum_quantity, notes) SELECT 'Samanta orz', (SELECT id FROM resource_types WHERE category = 'seed' AND default_unit = 'kg' ORDER BY id LIMIT 1), 4.2, 500, 'Creată de seed-ul pe sezoane' WHERE NOT EXISTS (SELECT 1 FROM resources WHERE name = 'Samanta orz');
INSERT INTO resource_types (name, category, default_unit) SELECT 'Seminte', 'seed', 'kg' WHERE NOT EXISTS (SELECT 1 FROM resource_types WHERE category = 'seed' AND default_unit = 'kg');
INSERT INTO resources (name, resource_type_id, price_per_unit, minimum_quantity, notes) SELECT 'Samanta soia', (SELECT id FROM resource_types WHERE category = 'seed' AND default_unit = 'kg' ORDER BY id LIMIT 1), 6.5, 400, 'Creată de seed-ul pe sezoane' WHERE NOT EXISTS (SELECT 1 FROM resources WHERE name = 'Samanta soia');
INSERT INTO resource_types (name, category, default_unit) SELECT 'Seminte', 'seed', 'kg' WHERE NOT EXISTS (SELECT 1 FROM resource_types WHERE category = 'seed' AND default_unit = 'kg');
INSERT INTO resources (name, resource_type_id, price_per_unit, minimum_quantity, notes) SELECT 'Samanta sfecla de zahar', (SELECT id FROM resource_types WHERE category = 'seed' AND default_unit = 'kg' ORDER BY id LIMIT 1), 900.0, 5, 'Creată de seed-ul pe sezoane' WHERE NOT EXISTS (SELECT 1 FROM resources WHERE name = 'Samanta sfecla de zahar');
INSERT INTO resource_types (name, category, default_unit) SELECT 'Seminte', 'seed', 'kg' WHERE NOT EXISTS (SELECT 1 FROM resource_types WHERE category = 'seed' AND default_unit = 'kg');
INSERT INTO resources (name, resource_type_id, price_per_unit, minimum_quantity, notes) SELECT 'Samanta lucerna', (SELECT id FROM resource_types WHERE category = 'seed' AND default_unit = 'kg' ORDER BY id LIMIT 1), 38.0, 50, 'Creată de seed-ul pe sezoane' WHERE NOT EXISTS (SELECT 1 FROM resources WHERE name = 'Samanta lucerna');
INSERT INTO resource_types (name, category, default_unit) SELECT 'Recoltă', 'harvest', 't' WHERE NOT EXISTS (SELECT 1 FROM resource_types WHERE category = 'harvest' AND default_unit = 't');
INSERT INTO resources (name, resource_type_id, price_per_unit, minimum_quantity, notes) SELECT 'Paie baloti', (SELECT id FROM resource_types WHERE category = 'harvest' AND default_unit = 't' ORDER BY id LIMIT 1), 180.0, 0, 'Creată de seed-ul pe sezoane' WHERE NOT EXISTS (SELECT 1 FROM resources WHERE name = 'Paie baloti');
INSERT INTO resource_types (name, category, default_unit) SELECT 'Recoltă', 'harvest', 't' WHERE NOT EXISTS (SELECT 1 FROM resource_types WHERE category = 'harvest' AND default_unit = 't');
INSERT INTO resources (name, resource_type_id, price_per_unit, notes) SELECT 'Recoltă Grâu', (SELECT id FROM resource_types WHERE category = 'harvest' AND default_unit = 't' ORDER BY id LIMIT 1), 900, 'Resursa de recoltă a culturii (seed pe sezoane)' WHERE NOT EXISTS (SELECT 1 FROM resources WHERE name = 'Recoltă Grâu');
UPDATE resources SET price_per_unit = 900 WHERE name = 'Recoltă Grâu' AND price_per_unit = 0;
UPDATE crops SET harvest_resource_id = (SELECT id FROM resources WHERE name = 'Recoltă Grâu') WHERE name = 'Grâu' AND harvest_resource_id IS NULL;
INSERT INTO resources (name, resource_type_id, price_per_unit, notes) SELECT 'Recoltă Porumb', (SELECT id FROM resource_types WHERE category = 'harvest' AND default_unit = 't' ORDER BY id LIMIT 1), 850, 'Resursa de recoltă a culturii (seed pe sezoane)' WHERE NOT EXISTS (SELECT 1 FROM resources WHERE name = 'Recoltă Porumb');
UPDATE resources SET price_per_unit = 850 WHERE name = 'Recoltă Porumb' AND price_per_unit = 0;
UPDATE crops SET harvest_resource_id = (SELECT id FROM resources WHERE name = 'Recoltă Porumb') WHERE name = 'Porumb' AND harvest_resource_id IS NULL;
INSERT INTO resources (name, resource_type_id, price_per_unit, notes) SELECT 'Recoltă Orz', (SELECT id FROM resource_types WHERE category = 'harvest' AND default_unit = 't' ORDER BY id LIMIT 1), 800, 'Resursa de recoltă a culturii (seed pe sezoane)' WHERE NOT EXISTS (SELECT 1 FROM resources WHERE name = 'Recoltă Orz');
UPDATE resources SET price_per_unit = 800 WHERE name = 'Recoltă Orz' AND price_per_unit = 0;
UPDATE crops SET harvest_resource_id = (SELECT id FROM resources WHERE name = 'Recoltă Orz') WHERE name = 'Orz' AND harvest_resource_id IS NULL;
INSERT INTO resources (name, resource_type_id, price_per_unit, notes) SELECT 'Recoltă Floarea-soarelui', (SELECT id FROM resource_types WHERE category = 'harvest' AND default_unit = 't' ORDER BY id LIMIT 1), 1900, 'Resursa de recoltă a culturii (seed pe sezoane)' WHERE NOT EXISTS (SELECT 1 FROM resources WHERE name = 'Recoltă Floarea-soarelui');
UPDATE resources SET price_per_unit = 1900 WHERE name = 'Recoltă Floarea-soarelui' AND price_per_unit = 0;
UPDATE crops SET harvest_resource_id = (SELECT id FROM resources WHERE name = 'Recoltă Floarea-soarelui') WHERE name = 'Floarea-soarelui' AND harvest_resource_id IS NULL;
INSERT INTO resources (name, resource_type_id, price_per_unit, notes) SELECT 'Recoltă Rapiță', (SELECT id FROM resource_types WHERE category = 'harvest' AND default_unit = 't' ORDER BY id LIMIT 1), 2100, 'Resursa de recoltă a culturii (seed pe sezoane)' WHERE NOT EXISTS (SELECT 1 FROM resources WHERE name = 'Recoltă Rapiță');
UPDATE resources SET price_per_unit = 2100 WHERE name = 'Recoltă Rapiță' AND price_per_unit = 0;
UPDATE crops SET harvest_resource_id = (SELECT id FROM resources WHERE name = 'Recoltă Rapiță') WHERE name = 'Rapiță' AND harvest_resource_id IS NULL;
INSERT INTO resources (name, resource_type_id, price_per_unit, notes) SELECT 'Recoltă Soia', (SELECT id FROM resource_types WHERE category = 'harvest' AND default_unit = 't' ORDER BY id LIMIT 1), 1950, 'Resursa de recoltă a culturii (seed pe sezoane)' WHERE NOT EXISTS (SELECT 1 FROM resources WHERE name = 'Recoltă Soia');
UPDATE resources SET price_per_unit = 1950 WHERE name = 'Recoltă Soia' AND price_per_unit = 0;
UPDATE crops SET harvest_resource_id = (SELECT id FROM resources WHERE name = 'Recoltă Soia') WHERE name = 'Soia' AND harvest_resource_id IS NULL;
INSERT INTO resources (name, resource_type_id, price_per_unit, notes) SELECT 'Recoltă Sfeclă de zahăr', (SELECT id FROM resource_types WHERE category = 'harvest' AND default_unit = 't' ORDER BY id LIMIT 1), 280, 'Resursa de recoltă a culturii (seed pe sezoane)' WHERE NOT EXISTS (SELECT 1 FROM resources WHERE name = 'Recoltă Sfeclă de zahăr');
UPDATE resources SET price_per_unit = 280 WHERE name = 'Recoltă Sfeclă de zahăr' AND price_per_unit = 0;
UPDATE crops SET harvest_resource_id = (SELECT id FROM resources WHERE name = 'Recoltă Sfeclă de zahăr') WHERE name = 'Sfeclă de zahăr' AND harvest_resource_id IS NULL;
INSERT INTO resources (name, resource_type_id, price_per_unit, notes) SELECT 'Recoltă Lucernă', (SELECT id FROM resource_types WHERE category = 'harvest' AND default_unit = 't' ORDER BY id LIMIT 1), 650, 'Resursa de recoltă a culturii (seed pe sezoane)' WHERE NOT EXISTS (SELECT 1 FROM resources WHERE name = 'Recoltă Lucernă');
UPDATE resources SET price_per_unit = 650 WHERE name = 'Recoltă Lucernă' AND price_per_unit = 0;
UPDATE crops SET harvest_resource_id = (SELECT id FROM resources WHERE name = 'Recoltă Lucernă') WHERE name = 'Lucernă' AND harvest_resource_id IS NULL;
INSERT INTO operation_templates (operation_type, name, description, unit, crop_id) SELECT 'seeding', 'Semanat rapita standard', 'Șablon creat de seed-ul pe sezoane', 'ha', (SELECT id FROM crops WHERE name = 'Rapiță') WHERE NOT EXISTS (SELECT 1 FROM operation_templates WHERE name = 'Semanat rapita standard' AND deleted_at IS NULL);
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Semanat rapita standard' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'), 7 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Semanat rapita standard' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'));
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Semanat rapita standard' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Samanta rapita hibrid'), 4 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Semanat rapita standard' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Samanta rapita hibrid'));
INSERT INTO operation_templates (operation_type, name, description, unit, crop_id) SELECT 'seeding', 'Semanat floarea-soarelui standard', 'Șablon creat de seed-ul pe sezoane', 'ha', (SELECT id FROM crops WHERE name = 'Floarea-soarelui') WHERE NOT EXISTS (SELECT 1 FROM operation_templates WHERE name = 'Semanat floarea-soarelui standard' AND deleted_at IS NULL);
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Semanat floarea-soarelui standard' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'), 8 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Semanat floarea-soarelui standard' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'));
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Semanat floarea-soarelui standard' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Samanta floarea-soarelui hibrid'), 3.5 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Semanat floarea-soarelui standard' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Samanta floarea-soarelui hibrid'));
INSERT INTO operation_templates (operation_type, name, description, unit, crop_id) SELECT 'seeding', 'Semanat orz standard', 'Șablon creat de seed-ul pe sezoane', 'ha', (SELECT id FROM crops WHERE name = 'Orz') WHERE NOT EXISTS (SELECT 1 FROM operation_templates WHERE name = 'Semanat orz standard' AND deleted_at IS NULL);
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Semanat orz standard' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'), 9 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Semanat orz standard' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'));
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Semanat orz standard' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Samanta orz'), 180 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Semanat orz standard' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Samanta orz'));
INSERT INTO operation_templates (operation_type, name, description, unit, crop_id) SELECT 'seeding', 'Semanat soia standard', 'Șablon creat de seed-ul pe sezoane', 'ha', (SELECT id FROM crops WHERE name = 'Soia') WHERE NOT EXISTS (SELECT 1 FROM operation_templates WHERE name = 'Semanat soia standard' AND deleted_at IS NULL);
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Semanat soia standard' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'), 8.5 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Semanat soia standard' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'));
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Semanat soia standard' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Samanta soia'), 100 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Semanat soia standard' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Samanta soia'));
INSERT INTO operation_templates (operation_type, name, description, unit, crop_id) SELECT 'seeding', 'Semanat sfecla standard', 'Șablon creat de seed-ul pe sezoane', 'ha', (SELECT id FROM crops WHERE name = 'Sfeclă de zahăr') WHERE NOT EXISTS (SELECT 1 FROM operation_templates WHERE name = 'Semanat sfecla standard' AND deleted_at IS NULL);
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Semanat sfecla standard' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'), 9 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Semanat sfecla standard' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'));
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Semanat sfecla standard' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Samanta sfecla de zahar'), 2.5 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Semanat sfecla standard' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Samanta sfecla de zahar'));
INSERT INTO operation_templates (operation_type, name, description, unit, crop_id) SELECT 'seeding', 'Semanat lucerna standard', 'Șablon creat de seed-ul pe sezoane', 'ha', (SELECT id FROM crops WHERE name = 'Lucernă') WHERE NOT EXISTS (SELECT 1 FROM operation_templates WHERE name = 'Semanat lucerna standard' AND deleted_at IS NULL);
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Semanat lucerna standard' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'), 8 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Semanat lucerna standard' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'));
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Semanat lucerna standard' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Samanta lucerna'), 22 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Semanat lucerna standard' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Samanta lucerna'));
INSERT INTO operation_templates (operation_type, name, description, unit, crop_id) SELECT 'fertilization', 'Fertilizare lucerna', 'Șablon creat de seed-ul pe sezoane', 'ha', (SELECT id FROM crops WHERE name = 'Lucernă') WHERE NOT EXISTS (SELECT 1 FROM operation_templates WHERE name = 'Fertilizare lucerna' AND deleted_at IS NULL);
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Fertilizare lucerna' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'), 6 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Fertilizare lucerna' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'));
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Fertilizare lucerna' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'NPK pentru pregatire teren'), 150 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Fertilizare lucerna' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'NPK pentru pregatire teren'));
INSERT INTO operation_templates (operation_type, name, description, unit, crop_id) SELECT 'spraying', 'Tratament fungicid grau', 'Șablon creat de seed-ul pe sezoane', 'ha', (SELECT id FROM crops WHERE name = 'Grâu') WHERE NOT EXISTS (SELECT 1 FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL);
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'), 4.8 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'));
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Fungicid lot rapita'), 0.8 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Fungicid lot rapita'));
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Apa sistem pivot 1'), 0.22 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Apa sistem pivot 1'));
INSERT INTO operation_templates (operation_type, name, description, unit, crop_id) SELECT 'harvesting', 'Recoltare porumb', 'Șablon creat de seed-ul pe sezoane', 'ha', (SELECT id FROM crops WHERE name = 'Porumb') WHERE NOT EXISTS (SELECT 1 FROM operation_templates WHERE name = 'Recoltare porumb' AND deleted_at IS NULL);
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Recoltare porumb' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'), 16 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Recoltare porumb' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'));
INSERT INTO operation_templates (operation_type, name, description, unit, crop_id) SELECT 'harvesting', 'Recoltare rapita', 'Șablon creat de seed-ul pe sezoane', 'ha', (SELECT id FROM crops WHERE name = 'Rapiță') WHERE NOT EXISTS (SELECT 1 FROM operation_templates WHERE name = 'Recoltare rapita' AND deleted_at IS NULL);
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Recoltare rapita' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'), 13 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Recoltare rapita' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'));
INSERT INTO operation_templates (operation_type, name, description, unit, crop_id) SELECT 'harvesting', 'Recoltare floarea-soarelui', 'Șablon creat de seed-ul pe sezoane', 'ha', (SELECT id FROM crops WHERE name = 'Floarea-soarelui') WHERE NOT EXISTS (SELECT 1 FROM operation_templates WHERE name = 'Recoltare floarea-soarelui' AND deleted_at IS NULL);
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Recoltare floarea-soarelui' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'), 12 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Recoltare floarea-soarelui' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'));
INSERT INTO operation_templates (operation_type, name, description, unit, crop_id) SELECT 'harvesting', 'Recoltare orz', 'Șablon creat de seed-ul pe sezoane', 'ha', (SELECT id FROM crops WHERE name = 'Orz') WHERE NOT EXISTS (SELECT 1 FROM operation_templates WHERE name = 'Recoltare orz' AND deleted_at IS NULL);
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Recoltare orz' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'), 13 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Recoltare orz' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'));
INSERT INTO operation_templates (operation_type, name, description, unit, crop_id) SELECT 'harvesting', 'Recoltare soia', 'Șablon creat de seed-ul pe sezoane', 'ha', (SELECT id FROM crops WHERE name = 'Soia') WHERE NOT EXISTS (SELECT 1 FROM operation_templates WHERE name = 'Recoltare soia' AND deleted_at IS NULL);
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Recoltare soia' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'), 12 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Recoltare soia' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'));
INSERT INTO operation_templates (operation_type, name, description, unit, crop_id) SELECT 'harvesting', 'Recoltare sfecla', 'Șablon creat de seed-ul pe sezoane', 'ha', (SELECT id FROM crops WHERE name = 'Sfeclă de zahăr') WHERE NOT EXISTS (SELECT 1 FROM operation_templates WHERE name = 'Recoltare sfecla' AND deleted_at IS NULL);
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Recoltare sfecla' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'), 22 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Recoltare sfecla' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'));
INSERT INTO operation_templates (operation_type, name, description, unit, crop_id) SELECT 'harvesting', 'Cosit si balotat lucerna', 'Șablon creat de seed-ul pe sezoane', 'ha', (SELECT id FROM crops WHERE name = 'Lucernă') WHERE NOT EXISTS (SELECT 1 FROM operation_templates WHERE name = 'Cosit si balotat lucerna' AND deleted_at IS NULL);
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Cosit si balotat lucerna' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'), 8 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Cosit si balotat lucerna' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'));
INSERT INTO operation_templates (operation_type, name, description, unit, crop_id) SELECT 'harvesting', 'Balotat paie', 'Șablon creat de seed-ul pe sezoane', 'ha', NULL WHERE NOT EXISTS (SELECT 1 FROM operation_templates WHERE name = 'Balotat paie' AND deleted_at IS NULL);
INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = 'Balotat paie' AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'), 3 WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = 'Balotat paie' AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = 'Motorina flota utilaje'));

-- ─── Sezonul ───
INSERT INTO seasons (name, start_date, end_date, is_active, notes) VALUES ('Sezon 2023/2024', '2023-08-01', '2024-10-31', false, 'An secetos: randamente sub așteptări la porumb și floarea-soarelui.');

-- ─── Aprovizionări (dimensionate după consumul sezonului, +5%) ───
DO $sup$ BEGIN
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'in', 500, '2024-04-06 11:57:17+00', NULL, NULL, '[Sezon 2023/2024] Alimentare bazin pivot');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'in', 1400, '2024-04-06 11:57:17+00', NULL, NULL, '[Sezon 2023/2024] Recepție furnizor');
  PERFORM pg_temp.seed_move('Fungicid lot rapita', 'in', 235, '2024-05-03 08:26:01+00', NULL, NULL, '[Sezon 2023/2024] Recepție furnizor');
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'in', 26000, '2023-08-01 16:55:54+00', NULL, NULL, '[Sezon 2023/2024] Recepție furnizor — tranșa 1');
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'in', 26000, '2023-12-29 16:55:54+00', NULL, NULL, '[Sezon 2023/2024] Recepție furnizor — tranșa 2');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'in', 140850, '2023-08-16 15:25:26+00', NULL, NULL, '[Sezon 2023/2024] Recepție furnizor');
  PERFORM pg_temp.seed_move('Samanta floarea-soarelui hibrid', 'in', 305, '2024-04-07 08:02:52+00', NULL, NULL, '[Sezon 2023/2024] Recepție furnizor');
  PERFORM pg_temp.seed_move('Samanta grau lot A', 'in', 64650, '2023-09-30 22:43:15+00', NULL, NULL, '[Sezon 2023/2024] Recepție furnizor');
  PERFORM pg_temp.seed_move('Samanta lucerna', 'in', 1600, '2024-03-20 03:32:37+00', NULL, NULL, '[Sezon 2023/2024] Recepție furnizor');
  PERFORM pg_temp.seed_move('Samanta orz', 'in', 10300, '2023-09-24 22:16:25+00', NULL, NULL, '[Sezon 2023/2024] Recepție furnizor');
  PERFORM pg_temp.seed_move('Samanta porumb lot B', 'in', 125, '2024-04-16 08:54:05+00', NULL, NULL, '[Sezon 2023/2024] Recepție furnizor');
  PERFORM pg_temp.seed_move('Samanta rapita hibrid', 'in', 400, '2023-08-22 11:21:17+00', NULL, NULL, '[Sezon 2023/2024] Recepție furnizor');
  PERFORM pg_temp.seed_move('Samanta soia', 'in', 5100, '2024-04-23 22:10:18+00', NULL, NULL, '[Sezon 2023/2024] Recepție furnizor');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'in', 69500, '2024-02-27 13:10:37+00', NULL, NULL, '[Sezon 2023/2024] Recepție furnizor');
END $sup$;

-- ── Plopi - Tarla Mare (210.25 ha): Grâu ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2023/2024';
  SELECT id INTO v_field FROM fields WHERE name = 'Plopi - Tarla Mare' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Plopi - Tarla Mare lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Grâu'), 210.25, '2023-10-08', 5.8, '', 'closed', '2024-07-15', '2023-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2023-08-26 07:00:00+00', '2023-08-31 18:24:22+00', 210.25, '', 'completed', v_fc,
      '2023-08-26 07:40:00+00', '2023-09-01 00:22:37+00', 210.25, 125.83, 'Întârziere: defecțiune minoră',
      NULL, '2023-08-19 07:00:00+00', '2023-09-01 00:22:37+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 3934.106, '2023-09-01 00:22:37+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 1 LIMIT 1),
      '2023-09-27 07:00:00+00', '2023-10-02 18:24:22+00', 210.25, '', 'completed', v_fc,
      '2023-09-27 06:56:00+00', '2023-10-02 18:42:56+00', 201.93, 124.35, '',
      NULL, '2023-09-20 07:00:00+00', '2023-10-02 18:42:56+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 2212.106, '2023-10-02 18:42:56+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 35628.69, '2023-10-02 18:42:56+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat grau standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2023-10-06 07:00:00+00', '2023-10-09 00:42:11+00', 210.25, '', 'completed', v_fc,
      '2023-10-06 07:02:00+00', '2023-10-09 01:09:23+00', 210.25, 65.36, '',
      NULL, '2023-09-29 07:00:00+00', '2023-10-09 01:09:23+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1897.52, '2023-10-09 01:09:23+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta grau lot A', 'out', 47005.684, '2023-10-09 01:09:23+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare faziala grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2024-03-12 07:00:00+00', '2024-03-13 13:02:08+00', 210.25, '', 'completed', v_fc,
      '2024-03-12 07:40:00+00', '2024-03-13 13:08:50+00', 210.25, 32.87, '',
      NULL, '2024-03-05 07:00:00+00', '2024-03-13 13:08:50+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1301.273, '2024-03-13 13:08:50+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'out', 32708.965, '2024-03-13 13:08:50+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2024-04-17 07:00:00+00', '2024-04-18 06:21:40+00', 210.25, '', 'completed', v_fc,
      '2024-04-17 07:17:00+00', '2024-04-18 06:49:48+00', 207.02, 22.15, '',
      NULL, '2024-04-10 07:00:00+00', '2024-04-18 06:49:48+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 951.148, '2024-04-18 06:49:48+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 468.064, '2024-04-18 06:49:48+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 52.93, '2024-04-18 06:49:48+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 5 LIMIT 1),
      '2024-05-08 07:00:00+00', '2024-05-09 06:21:40+00', 210.25, '', 'completed', v_fc,
      '2024-05-08 07:36:00+00', '2024-05-09 11:54:06+00', 210.25, 25.01, 'Întârziere: defecțiune minoră',
      NULL, '2024-05-01 07:00:00+00', '2024-05-09 11:54:06+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1074.915, '2024-05-09 11:54:06+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Fungicid lot rapita', 'out', 168.634, '2024-05-09 11:54:06+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 47.786, '2024-05-09 11:54:06+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-CLAAS-LEXION780-01'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2024-07-07 07:00:00+00', '2024-07-10 22:36:15+00', 210.25, '', 'completed', v_fc,
      '2024-07-07 07:43:00+00', '2024-07-10 22:09:45+00', 210.25, 83.58, '',
      NULL, '2024-06-30 07:00:00+00', '2024-07-10 22:09:45+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 3098.452, '2024-07-10 22:09:45+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'in', 1083.16, '2024-07-11 18:09:45+00', NULL, v_fc, 'Recoltă Grâu - Plopi - Tarla Mare (Sezon 2023/2024)');
  UPDATE field_crops SET production_total = 1083.16, harvest_recorded_quantity = 1083.16, harvest_recorded_at = '2024-07-11 18:09:45+00', harvested_at = '2024-07-11', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Balotat paie' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2024-07-12 07:00:00+00', '2024-07-15 22:36:15+00', 210.25, '', 'completed', v_fc,
      '2024-07-12 07:30:00+00', '2024-07-15 22:00:49+00', 210.25, 90.14, '',
      NULL, '2024-07-05 07:00:00+00', '2024-07-15 22:00:49+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 675.804, '2024-07-15 22:00:49+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Paie baloti', 'in', 553.46, '2024-07-16 04:00:49+00', NULL, v_fc, 'Paie Grâu - Plopi - Tarla Mare, balotat');
  UPDATE field_crops SET status = 'closed', closed_at = '2024-07-20', harvested_at = COALESCE(harvested_at, '2024-07-20'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Gotesti - Tarla Mare (130.03 ha): Porumb ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2023/2024';
  SELECT id INTO v_field FROM fields WHERE name = 'Gotesti - Tarla Mare' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Gotesti - Tarla Mare lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Porumb'), 130.03, '2024-04-22', 9.0, '', 'closed', '2024-07-15', '2023-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 9 LIMIT 1),
      '2023-10-28 07:00:00+00', '2023-10-31 16:16:07+00', 130.03, '', 'completed', v_fc,
      '2023-10-28 07:23:00+00', '2023-10-31 19:29:15+00', 126.42, 76.31, 'Întârziere: teren umed',
      NULL, '2023-10-21 07:00:00+00', '2023-10-31 19:29:15+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 2334.503, '2023-10-31 19:29:15+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 10 LIMIT 1),
      '2024-04-08 07:00:00+00', '2024-04-11 16:16:07+00', 130.03, '', 'completed', v_fc,
      '2024-04-08 06:40:00+00', '2024-04-11 15:31:30+00', 130.03, 75.48, '',
      NULL, '2024-04-01 07:00:00+00', '2024-04-11 15:31:30+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1431.622, '2024-04-11 15:31:30+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 23522.675, '2024-04-11 15:31:30+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat porumb standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 11 LIMIT 1),
      '2024-04-25 07:00:00+00', '2024-04-26 23:38:03+00', 130.03, '', 'completed', v_fc,
      '2024-04-25 07:57:00+00', '2024-04-26 22:52:53+00', 127.17, 42.55, '',
      NULL, '2024-04-18 07:00:00+00', '2024-04-26 22:52:53+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1143.213, '2024-04-26 22:52:53+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta porumb lot B', 'out', 108.029, '2024-04-26 22:52:53+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 12 LIMIT 1),
      '2024-05-18 07:00:00+00', '2024-05-18 21:26:52+00', 130.03, '', 'completed', v_fc,
      '2024-05-18 07:59:00+00', '2024-05-18 20:42:17+00', 130.03, 15.04, '',
      NULL, '2024-05-11 07:00:00+00', '2024-05-18 20:42:17+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 640.834, '2024-05-18 20:42:17+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 283.264, '2024-05-18 20:42:17+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 33.484, '2024-05-18 20:42:17+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare porumb' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-JOHNDEERE-S780-02'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2024-10-04 07:00:00+00', '2024-10-06 13:10:45+00', 130.03, '', 'completed', v_fc,
      '2024-10-04 07:25:00+00', '2024-10-06 12:20:00+00', 130.03, 53.58, '',
      NULL, '2024-09-27 07:00:00+00', '2024-10-06 12:20:00+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 2239.285, '2024-10-06 12:20:00+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Porumb', 'in', 994.02, '2024-10-07 08:20:00+00', NULL, v_fc, 'Recoltă Porumb - Gotesti - Tarla Mare (Sezon 2023/2024)');
  UPDATE field_crops SET production_total = 994.02, harvest_recorded_quantity = 994.02, harvest_recorded_at = '2024-10-07 08:20:00+00', harvested_at = '2024-10-07', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2024-10-11', harvested_at = COALESCE(harvested_at, '2024-10-11'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Gotesti - Tarla Est (92.34 ha): Rapiță ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2023/2024';
  SELECT id INTO v_field FROM fields WHERE name = 'Gotesti - Tarla Est' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Gotesti - Tarla Est lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Rapiță'), 92.34, '2023-08-28', 3.2, '', 'closed', '2024-07-15', '2023-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2023-08-05 07:00:00+00', '2023-08-07 16:42:45+00', 92.34, '', 'completed', v_fc,
      '2023-08-05 07:52:00+00', '2023-08-07 16:55:54+00', 92.34, 55.66, '',
      NULL, '2023-07-29 07:00:00+00', '2023-08-07 16:55:54+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1762.626, '2023-08-07 16:55:54+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2023-08-20 07:00:00+00', '2023-08-22 16:42:45+00', 92.34, '', 'completed', v_fc,
      '2023-08-20 07:51:00+00', '2023-08-22 15:25:26+00', 92.34, 58.5, '',
      NULL, '2023-08-13 07:00:00+00', '2023-08-22 15:25:26+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 970.907, '2023-08-22 15:25:26+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 16210.755, '2023-08-22 15:25:26+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat rapita standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2023-08-27 07:00:00+00', '2023-08-28 11:51:22+00', 92.34, '', 'completed', v_fc,
      '2023-08-27 07:56:00+00', '2023-08-28 11:21:17+00', 92.34, 26.65, '',
      NULL, '2023-08-20 07:00:00+00', '2023-08-28 11:21:17+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 692.681, '2023-08-28 11:21:17+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta rapita hibrid', 'out', 377.367, '2023-08-28 11:21:17+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare faziala grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 5 LIMIT 1),
      '2024-03-04 07:00:00+00', '2024-03-04 20:11:29+00', 92.34, '', 'completed', v_fc,
      '2024-03-04 07:20:00+00', '2024-03-04 19:32:51+00', 92.34, 14.0, '',
      NULL, '2024-02-26 07:00:00+00', '2024-03-04 19:32:51+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 569.692, '2024-03-04 19:32:51+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'out', 14735.085, '2024-03-04 19:32:51+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid rapita' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2024-04-16 07:00:00+00', '2024-04-16 17:15:36+00', 92.34, '', 'canceled', v_fc,
      NULL, NULL, NULL, NULL, 'Anulată: ploi persistente în fereastra de tratament',
      NULL, '2024-04-09 07:00:00+00', '2024-04-16 07:00:00+00')
  RETURNING id INTO v_op;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare rapita' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-CLAAS-LEXION780-01'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2024-06-29 07:00:00+00', '2024-06-30 21:28:30+00', 92.34, '', 'completed', v_fc,
      '2024-06-29 07:12:00+00', '2024-06-30 20:58:15+00', 92.34, 40.84, '',
      NULL, '2024-06-22 07:00:00+00', '2024-06-30 20:58:15+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1278.072, '2024-06-30 20:58:15+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Rapiță', 'in', 266.51, '2024-07-01 16:58:15+00', NULL, v_fc, 'Recoltă Rapiță - Gotesti - Tarla Est (Sezon 2023/2024)');
  UPDATE field_crops SET production_total = 266.51, harvest_recorded_quantity = 266.51, harvest_recorded_at = '2024-07-01 16:58:15+00', harvested_at = '2024-07-01', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2024-07-05', harvested_at = COALESCE(harvested_at, '2024-07-05'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Tartaul - Tarla Mare (84.02 ha): Floarea-soarelui ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2023/2024';
  SELECT id INTO v_field FROM fields WHERE name = 'Tartaul - Tarla Mare' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Tartaul - Tarla Mare lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Floarea-soarelui'), 84.02, '2024-04-14', 2.6, '', 'closed', '2024-07-15', '2023-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 9 LIMIT 1),
      '2023-10-26 07:00:00+00', '2023-10-28 11:30:45+00', 84.02, '', 'completed', v_fc,
      '2023-10-26 07:08:00+00', '2023-10-28 10:35:26+00', 80.72, 52.5, '',
      NULL, '2023-10-19 07:00:00+00', '2023-10-28 10:35:26+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1544.632, '2023-10-28 10:35:26+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 10 LIMIT 1),
      '2024-04-07 07:00:00+00', '2024-04-09 11:30:45+00', 84.02, '', 'completed', v_fc,
      '2024-04-07 07:59:00+00', '2024-04-09 16:29:01+00', 84.02, 51.03, 'Întârziere: defecțiune minoră',
      NULL, '2024-03-31 07:00:00+00', '2024-04-09 16:29:01+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 852.406, '2024-04-09 16:29:01+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 14837.691, '2024-04-09 16:29:01+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat floarea-soarelui standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 11 LIMIT 1),
      '2024-04-15 07:00:00+00', '2024-04-16 09:15:22+00', 84.02, '', 'completed', v_fc,
      '2024-04-15 06:48:00+00', '2024-04-16 14:15:57+00', 82.71, 27.62, 'Întârziere: defecțiune minoră',
      NULL, '2024-04-08 07:00:00+00', '2024-04-16 14:15:57+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 705.05, '2024-04-16 14:15:57+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta floarea-soarelui hibrid', 'out', 281.174, '2024-04-16 14:15:57+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 12 LIMIT 1),
      '2024-05-15 07:00:00+00', '2024-05-15 16:20:08+00', 84.02, '', 'completed', v_fc,
      '2024-05-15 07:48:00+00', '2024-05-15 16:02:46+00', 84.02, 9.41, '',
      NULL, '2024-05-08 07:00:00+00', '2024-05-15 16:02:46+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 413.97, '2024-05-15 16:02:46+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 183.902, '2024-05-15 16:02:46+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 20.489, '2024-05-15 16:02:46+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare floarea-soarelui' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-CLAAS-LEXION780-01'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2024-09-11 07:00:00+00', '2024-09-12 18:00:30+00', 84.02, '', 'completed', v_fc,
      '2024-09-11 07:31:00+00', '2024-09-12 18:05:42+00', 80.38, 38.48, '',
      NULL, '2024-09-04 07:00:00+00', '2024-09-12 18:05:42+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1048.322, '2024-09-12 18:05:42+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Floarea-soarelui', 'in', 190.06, '2024-09-13 14:05:42+00', NULL, v_fc, 'Recoltă Floarea-soarelui - Tartaul - Tarla Mare (Sezon 2023/2024)');
  UPDATE field_crops SET production_total = 190.06, harvest_recorded_quantity = 190.06, harvest_recorded_at = '2024-09-13 14:05:42+00', harvested_at = '2024-09-13', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2024-09-17', harvested_at = COALESCE(harvested_at, '2024-09-17'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Gotesti - Tarla Vest (68.47 ha): Lucernă, anul 1 ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2023/2024';
  SELECT id INTO v_field FROM fields WHERE name = 'Gotesti - Tarla Vest' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Gotesti - Tarla Vest lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Lucernă'), 68.47, '2024-03-22', 9.5, 'Lucernă anul 1', 'closed', '2024-07-15', '2023-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2023-08-19 07:00:00+00', '2023-08-21 01:47:37+00', 68.47, '', 'completed', v_fc,
      '2023-08-19 07:07:00+00', '2023-08-21 00:51:50+00', 68.47, 41.61, '',
      NULL, '2023-08-12 07:00:00+00', '2023-08-21 00:51:50+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1314.166, '2023-08-21 00:51:50+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2024-03-15 07:00:00+00', '2024-03-17 01:47:37+00', 68.47, '', 'completed', v_fc,
      '2024-03-15 07:41:00+00', '2024-03-17 00:42:43+00', 66.83, 46.81, '',
      NULL, '2024-03-08 07:00:00+00', '2024-03-17 00:42:43+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 678.691, '2024-03-17 00:42:43+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 11961.903, '2024-03-17 00:42:43+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat lucerna standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2024-03-25 07:00:00+00', '2024-03-26 04:23:48+00', 68.47, '', 'completed', v_fc,
      '2024-03-25 07:25:00+00', '2024-03-26 03:32:37+00', 68.47, 21.36, '',
      NULL, '2024-03-18 07:00:00+00', '2024-03-26 03:32:37+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 567.381, '2024-03-26 03:32:37+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta lucerna', 'out', 1476.275, '2024-03-26 03:32:37+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Cosit si balotat lucerna' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 5 LIMIT 1),
      '2024-05-26 07:00:00+00', '2024-05-27 11:31:45+00', 68.47, '', 'completed', v_fc,
      '2024-05-26 07:40:00+00', '2024-05-27 11:38:16+00', 66.25, 29.7, '',
      NULL, '2024-05-19 07:00:00+00', '2024-05-27 11:38:16+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 565.051, '2024-05-27 11:38:16+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'in', 102.36, '2024-05-28 17:38:16+00', NULL, v_fc, 'Recoltă Lucernă - Gotesti - Tarla Vest (Sezon 2023/2024)');
  UPDATE field_crops SET production_total = 102.36, harvest_recorded_quantity = 102.36, harvest_recorded_at = '2024-05-28 17:38:16+00', harvested_at = '2024-05-28', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Cosit si balotat lucerna' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2024-07-06 07:00:00+00', '2024-07-07 11:31:45+00', 68.47, '', 'completed', v_fc,
      '2024-07-06 07:09:00+00', '2024-07-07 10:51:33+00', 68.47, 27.28, '',
      NULL, '2024-06-29 07:00:00+00', '2024-07-07 10:51:33+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 598.628, '2024-07-07 10:51:33+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'in', 152.81, '2024-07-08 16:51:33+00', NULL, v_fc, 'Recoltă Lucernă - Gotesti - Tarla Vest (Sezon 2023/2024)');
  UPDATE field_crops SET production_total = 255.17, harvest_recorded_quantity = 255.17, harvest_recorded_at = '2024-07-08 16:51:33+00', harvested_at = '2024-07-08', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Cosit si balotat lucerna' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2024-08-21 07:00:00+00', '2024-08-22 11:31:45+00', 68.47, '', 'completed', v_fc,
      '2024-08-21 06:42:00+00', '2024-08-22 10:20:48+00', 68.47, 30.56, '',
      NULL, '2024-08-14 07:00:00+00', '2024-08-22 10:20:48+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 577.982, '2024-08-22 10:20:48+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'in', 129.3, '2024-08-23 16:20:48+00', NULL, v_fc, 'Recoltă Lucernă - Gotesti - Tarla Vest (Sezon 2023/2024)');
  UPDATE field_crops SET production_total = 384.47, harvest_recorded_quantity = 384.47, harvest_recorded_at = '2024-08-23 16:20:48+00', harvested_at = '2024-08-23', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Cosit si balotat lucerna' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 8 LIMIT 1),
      '2024-09-27 07:00:00+00', '2024-09-28 11:31:45+00', 68.47, '', 'completed', v_fc,
      '2024-09-27 07:48:00+00', '2024-09-28 10:05:12+00', 68.47, 30.36, '',
      NULL, '2024-09-20 07:00:00+00', '2024-09-28 10:05:12+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 579.365, '2024-09-28 10:05:12+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'in', 94.04, '2024-09-29 16:05:12+00', NULL, v_fc, 'Recoltă Lucernă - Gotesti - Tarla Vest (Sezon 2023/2024)');
  UPDATE field_crops SET production_total = 478.51, harvest_recorded_quantity = 478.51, harvest_recorded_at = '2024-09-29 16:05:12+00', harvested_at = '2024-09-29', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2024-10-03', harvested_at = COALESCE(harvested_at, '2024-10-03'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Tartaul - Tarla Centrala (53.5 ha): Orz ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2023/2024';
  SELECT id INTO v_field FROM fields WHERE name = 'Tartaul - Tarla Centrala' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Tartaul - Tarla Centrala lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Orz'), 53.5, '2023-09-30', 5.2, '', 'closed', '2024-07-15', '2023-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 10 LIMIT 1),
      '2023-08-20 07:00:00+00', '2023-08-21 16:26:15+00', 53.5, '', 'completed', v_fc,
      '2023-08-20 07:44:00+00', '2023-08-21 18:14:25+00', 51.68, 35.9, 'Întârziere: ploaie',
      NULL, '2023-08-13 07:00:00+00', '2023-08-21 18:14:25+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1027.632, '2023-08-21 18:14:25+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 11 LIMIT 1),
      '2023-09-19 07:00:00+00', '2023-09-20 16:26:15+00', 53.5, '', 'completed', v_fc,
      '2023-09-19 06:50:00+00', '2023-09-20 20:10:06+00', 53.5, 31.64, 'Întârziere: ploaie',
      NULL, '2023-09-12 07:00:00+00', '2023-09-20 20:10:06+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 543.992, '2023-09-20 20:10:06+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 9686.374, '2023-09-20 20:10:06+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat orz standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 12 LIMIT 1),
      '2023-09-30 07:00:00+00', '2023-09-30 23:43:07+00', 53.5, '', 'completed', v_fc,
      '2023-09-30 06:55:00+00', '2023-09-30 22:16:25+00', 53.5, 18.14, '',
      NULL, '2023-09-23 07:00:00+00', '2023-09-30 22:16:25+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 527.64, '2023-09-30 22:16:25+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta orz', 'out', 9785.554, '2023-09-30 22:16:25+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare faziala grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2024-03-04 07:00:00+00', '2024-03-04 14:38:34+00', 53.5, '', 'completed', v_fc,
      '2024-03-04 06:55:00+00', '2024-03-04 13:10:37+00', 53.5, 8.05, '',
      NULL, '2024-02-26 07:00:00+00', '2024-03-04 13:10:37+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 354.305, '2024-03-04 13:10:37+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'out', 8359.135, '2024-03-04 13:10:37+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 1 LIMIT 1),
      '2024-04-12 07:00:00+00', '2024-04-12 12:56:40+00', 53.5, '', 'completed', v_fc,
      '2024-04-12 07:49:00+00', '2024-04-12 11:57:17+00', 53.5, 6.28, '',
      NULL, '2024-04-05 07:00:00+00', '2024-04-12 11:57:17+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 260.095, '2024-04-12 11:57:17+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 118.156, '2024-04-12 11:57:17+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 13.379, '2024-04-12 11:57:17+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare orz' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-JOHNDEERE-S780-02'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2024-06-26 07:00:00+00', '2024-06-27 05:17:30+00', 53.5, '', 'completed', v_fc,
      '2024-06-26 06:50:00+00', '2024-06-27 09:59:28+00', 53.5, 21.97, 'Întârziere: ploaie',
      NULL, '2024-06-19 07:00:00+00', '2024-06-27 09:59:28+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 721.304, '2024-06-27 09:59:28+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Orz', 'in', 254.27, '2024-06-28 05:59:28+00', NULL, v_fc, 'Recoltă Orz - Tartaul - Tarla Centrala (Sezon 2023/2024)');
  UPDATE field_crops SET production_total = 254.27, harvest_recorded_quantity = 254.27, harvest_recorded_at = '2024-06-28 05:59:28+00', harvested_at = '2024-06-28', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Balotat paie' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2024-07-05 07:00:00+00', '2024-07-06 05:17:30+00', 53.5, '', 'completed', v_fc,
      '2024-07-05 07:44:00+00', '2024-07-06 05:36:06+00', 51.09, 22.87, '',
      NULL, '2024-06-28 07:00:00+00', '2024-07-06 05:36:06+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 153.839, '2024-07-06 05:36:06+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Paie baloti', 'in', 92.25, '2024-07-06 11:36:06+00', NULL, v_fc, 'Paie Orz - Tartaul - Tarla Centrala, balotat');
  UPDATE field_crops SET status = 'closed', closed_at = '2024-07-11', harvested_at = COALESCE(harvested_at, '2024-07-11'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Plopi - Tarla Nord (52.41 ha): Grâu ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2023/2024';
  SELECT id INTO v_field FROM fields WHERE name = 'Plopi - Tarla Nord' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Plopi - Tarla Nord lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Grâu'), 52.41, '2023-10-08', 5.8, '', 'closed', '2024-07-15', '2023-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 5 LIMIT 1),
      '2023-08-23 07:00:00+00', '2023-08-24 15:45:22+00', 52.41, '', 'completed', v_fc,
      '2023-08-23 07:45:00+00', '2023-08-24 15:03:32+00', 52.41, 33.99, '',
      NULL, '2023-08-16 07:00:00+00', '2023-08-24 15:03:32+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1034.041, '2023-08-24 15:03:32+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2023-09-30 07:00:00+00', '2023-10-01 15:45:22+00', 52.41, '', 'completed', v_fc,
      '2023-09-30 07:36:00+00', '2023-10-01 15:08:03+00', 52.41, 34.91, '',
      NULL, '2023-09-23 07:00:00+00', '2023-10-01 15:08:03+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 574.371, '2023-10-01 15:08:03+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 9363.887, '2023-10-01 15:08:03+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat grau standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2023-10-06 07:00:00+00', '2023-10-06 23:22:41+00', 52.41, '', 'completed', v_fc,
      '2023-10-06 07:35:00+00', '2023-10-06 22:43:15+00', 52.41, 17.61, '',
      NULL, '2023-09-29 07:00:00+00', '2023-10-06 22:43:15+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 480.865, '2023-10-06 22:43:15+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta grau lot A', 'out', 11839.079, '2023-10-06 22:43:15+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare faziala grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 8 LIMIT 1),
      '2024-03-10 07:00:00+00', '2024-03-10 14:29:13+00', 52.41, '', 'completed', v_fc,
      '2024-03-10 06:42:00+00', '2024-03-10 14:02:45+00', 52.41, 7.94, '',
      NULL, '2024-03-03 07:00:00+00', '2024-03-10 14:02:45+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 339.55, '2024-03-10 14:02:45+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'out', 8421.766, '2024-03-10 14:02:45+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 9 LIMIT 1),
      '2024-04-21 07:00:00+00', '2024-04-21 12:49:24+00', 52.41, '', 'completed', v_fc,
      '2024-04-21 06:54:00+00', '2024-04-21 13:04:18+00', 52.41, 5.72, '',
      NULL, '2024-04-14 07:00:00+00', '2024-04-21 13:04:18+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 248.869, '2024-04-21 13:04:18+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 116.472, '2024-04-21 13:04:18+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 13.034, '2024-04-21 13:04:18+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 10 LIMIT 1),
      '2024-05-13 07:00:00+00', '2024-05-13 12:49:24+00', 52.41, '', 'completed', v_fc,
      '2024-05-13 07:25:00+00', '2024-05-13 12:55:02+00', 52.41, 6.24, '',
      NULL, '2024-05-06 07:00:00+00', '2024-05-13 12:55:02+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 272.076, '2024-05-13 12:55:02+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Fungicid lot rapita', 'out', 41.824, '2024-05-13 12:55:02+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 11.652, '2024-05-13 12:55:02+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-CLAAS-LEXION780-01'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 11 LIMIT 1),
      '2024-07-03 07:00:00+00', '2024-07-04 04:50:15+00', 52.41, '', 'completed', v_fc,
      '2024-07-03 07:43:00+00', '2024-07-04 04:34:16+00', 52.41, 21.65, '',
      NULL, '2024-06-26 07:00:00+00', '2024-07-04 04:34:16+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 768.312, '2024-07-04 04:34:16+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'in', 262.96, '2024-07-05 00:34:16+00', NULL, v_fc, 'Recoltă Grâu - Plopi - Tarla Nord (Sezon 2023/2024)');
  UPDATE field_crops SET production_total = 262.96, harvest_recorded_quantity = 262.96, harvest_recorded_at = '2024-07-05 00:34:16+00', harvested_at = '2024-07-05', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Balotat paie' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 12 LIMIT 1),
      '2024-07-13 07:00:00+00', '2024-07-14 04:50:15+00', 52.41, '', 'completed', v_fc,
      '2024-07-13 07:22:00+00', '2024-07-14 09:53:28+00', 52.41, 21.5, 'Întârziere: defecțiune minoră',
      NULL, '2024-07-06 07:00:00+00', '2024-07-14 09:53:28+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 169.822, '2024-07-14 09:53:28+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Paie baloti', 'in', 119.24, '2024-07-14 15:53:28+00', NULL, v_fc, 'Paie Grâu - Plopi - Tarla Nord, balotat');
  UPDATE field_crops SET status = 'closed', closed_at = '2024-07-19', harvested_at = COALESCE(harvested_at, '2024-07-19'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Plopi - Tarla Sud (48.28 ha): Soia ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2023/2024';
  SELECT id INTO v_field FROM fields WHERE name = 'Plopi - Tarla Sud' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Plopi - Tarla Sud lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Soia'), 48.28, '2024-04-28', 2.8, '', 'closed', '2024-07-15', '2023-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 1 LIMIT 1),
      '2023-11-05 07:00:00+00', '2023-11-06 13:10:30+00', 48.28, '', 'completed', v_fc,
      '2023-11-05 06:57:00+00', '2023-11-06 13:34:00+00', 46.67, 29.07, '',
      NULL, '2023-10-29 07:00:00+00', '2023-11-06 13:34:00+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 935.859, '2023-11-06 13:34:00+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2024-04-21 07:00:00+00', '2024-04-22 13:10:30+00', 48.28, '', 'completed', v_fc,
      '2024-04-21 07:12:00+00', '2024-04-22 15:09:19+00', 48.28, 31.84, 'Întârziere: ploaie',
      NULL, '2024-04-14 07:00:00+00', '2024-04-22 15:09:19+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 517.335, '2024-04-22 15:09:19+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 8599.35, '2024-04-22 15:09:19+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat soia standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2024-04-29 07:00:00+00', '2024-04-29 22:05:15+00', 48.28, '', 'completed', v_fc,
      '2024-04-29 07:32:00+00', '2024-04-29 22:10:18+00', 48.28, 14.12, '',
      NULL, '2024-04-22 07:00:00+00', '2024-04-29 22:10:18+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 433.733, '2024-04-29 22:10:18+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta soia', 'out', 4838.019, '2024-04-29 22:10:18+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2024-05-29 07:00:00+00', '2024-05-29 12:21:52+00', 48.28, '', 'completed', v_fc,
      '2024-05-29 06:53:00+00', '2024-05-29 12:09:55+00', 47.56, 5.79, '',
      NULL, '2024-05-22 07:00:00+00', '2024-05-29 12:09:55+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 229.363, '2024-05-29 12:09:55+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 108.573, '2024-05-29 12:09:55+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 11.788, '2024-05-29 12:09:55+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare soia' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-JOHNDEERE-S780-02'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 5 LIMIT 1),
      '2024-09-24 07:00:00+00', '2024-09-25 03:07:00+00', 48.28, '', 'completed', v_fc,
      '2024-09-24 07:39:00+00', '2024-09-25 02:45:20+00', 48.28, 19.16, '',
      NULL, '2024-09-17 07:00:00+00', '2024-09-25 02:45:20+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 633.206, '2024-09-25 02:45:20+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Soia', 'in', 120.4, '2024-09-25 22:45:20+00', NULL, v_fc, 'Recoltă Soia - Plopi - Tarla Sud (Sezon 2023/2024)');
  UPDATE field_crops SET production_total = 120.4, harvest_recorded_quantity = 120.4, harvest_recorded_at = '2024-09-25 22:45:20+00', harvested_at = '2024-09-25', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2024-09-30', harvested_at = COALESCE(harvested_at, '2024-09-30'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Tartaul - Tarla Mica (9.84 ha): Grâu ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2023/2024';
  SELECT id INTO v_field FROM fields WHERE name = 'Tartaul - Tarla Mica' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Tartaul - Tarla Mica lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Grâu'), 9.84, '2023-10-08', 5.8, '', 'closed', '2024-07-15', '2023-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2023-08-23 07:00:00+00', '2023-08-23 13:09:00+00', 9.84, '', 'completed', v_fc,
      '2023-08-23 07:57:00+00', '2023-08-23 15:59:38+00', 9.84, 6.5, 'Întârziere: teren umed',
      NULL, '2023-08-16 07:00:00+00', '2023-08-23 15:59:38+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 189.187, '2023-08-23 15:59:38+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 8 LIMIT 1),
      '2023-09-26 07:00:00+00', '2023-09-26 13:09:00+00', 9.84, '', 'completed', v_fc,
      '2023-09-26 06:58:00+00', '2023-09-26 13:22:32+00', 9.84, 5.87, '',
      NULL, '2023-09-19 07:00:00+00', '2023-09-26 13:22:32+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 104.2, '2023-09-26 13:22:32+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 1738.323, '2023-09-26 13:22:32+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat grau standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 9 LIMIT 1),
      '2023-10-11 07:00:00+00', '2023-10-11 10:04:30+00', 9.84, '', 'completed', v_fc,
      '2023-10-11 06:57:00+00', '2023-10-11 09:04:17+00', 9.58, 3.13, '',
      NULL, '2023-10-04 07:00:00+00', '2023-10-11 09:04:17+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 86.892, '2023-10-11 09:04:17+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta grau lot A', 'out', 2156.0, '2023-10-11 09:04:17+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare faziala grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 10 LIMIT 1),
      '2024-03-10 07:00:00+00', '2024-03-10 08:30:00+00', 9.84, '', 'completed', v_fc,
      '2024-03-10 06:50:00+00', '2024-03-10 07:13:58+00', 9.53, 1.41, '',
      NULL, '2024-03-03 07:00:00+00', '2024-03-10 07:13:58+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 62.933, '2024-03-10 07:13:58+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'out', 1555.342, '2024-03-10 07:13:58+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 11 LIMIT 1),
      '2024-04-21 07:00:00+00', '2024-04-21 08:30:00+00', 9.84, '', 'completed', v_fc,
      '2024-04-21 06:40:00+00', '2024-04-21 07:31:39+00', 9.76, 1.61, '',
      NULL, '2024-04-14 07:00:00+00', '2024-04-21 07:31:39+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 48.535, '2024-04-21 07:31:39+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 21.329, '2024-04-21 07:31:39+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 2.462, '2024-04-21 07:31:39+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 12 LIMIT 1),
      '2024-05-12 07:00:00+00', '2024-05-12 08:30:00+00', 9.84, '', 'completed', v_fc,
      '2024-05-12 07:08:00+00', '2024-05-12 07:30:03+00', 9.65, 1.47, '',
      NULL, '2024-05-05 07:00:00+00', '2024-05-12 07:30:03+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 47.398, '2024-05-12 07:30:03+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Fungicid lot rapita', 'out', 7.97, '2024-05-12 07:30:03+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 2.142, '2024-05-12 07:30:03+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-JOHNDEERE-S780-02'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2024-07-03 07:00:00+00', '2024-07-03 11:06:00+00', 9.84, '', 'completed', v_fc,
      '2024-07-03 06:44:00+00', '2024-07-03 09:45:30+00', 9.48, 3.8, '',
      NULL, '2024-06-26 07:00:00+00', '2024-07-03 09:45:30+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 136.689, '2024-07-03 09:45:30+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'in', 49.13, '2024-07-04 05:45:30+00', NULL, v_fc, 'Recoltă Grâu - Tartaul - Tarla Mica (Sezon 2023/2024)');
  UPDATE field_crops SET production_total = 49.13, harvest_recorded_quantity = 49.13, harvest_recorded_at = '2024-07-04 05:45:30+00', harvested_at = '2024-07-04', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Balotat paie' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 1 LIMIT 1),
      '2024-07-14 07:00:00+00', '2024-07-14 11:06:00+00', 9.84, '', 'completed', v_fc,
      '2024-07-14 07:11:00+00', '2024-07-14 10:05:00+00', 9.84, 4.19, '',
      NULL, '2024-07-07 07:00:00+00', '2024-07-14 10:05:00+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 30.553, '2024-07-14 10:05:00+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Paie baloti', 'in', 23.05, '2024-07-14 16:05:00+00', NULL, v_fc, 'Paie Grâu - Tartaul - Tarla Mica, balotat');
  UPDATE field_crops SET status = 'closed', closed_at = '2024-07-19', harvested_at = COALESCE(harvested_at, '2024-07-19'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Lingura - Teren agricol Nord (7.98 ha): Porumb ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2023/2024';
  SELECT id INTO v_field FROM fields WHERE name = 'Lingura - Teren agricol Nord' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Lingura - Teren agricol Nord lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Porumb'), 7.98, '2024-04-22', 9.0, '', 'closed', '2024-07-15', '2023-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2023-10-26 07:00:00+00', '2023-10-26 11:59:15+00', 7.98, '', 'completed', v_fc,
      '2023-10-26 07:40:00+00', '2023-10-26 10:44:05+00', 7.98, 4.94, '',
      NULL, '2023-10-19 07:00:00+00', '2023-10-26 10:44:05+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 154.225, '2023-10-26 10:44:05+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2024-04-09 07:00:00+00', '2024-04-09 11:59:15+00', 7.98, '', 'completed', v_fc,
      '2024-04-09 07:38:00+00', '2024-04-09 11:18:26+00', 7.93, 4.72, '',
      NULL, '2024-04-02 07:00:00+00', '2024-04-09 11:18:26+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 86.813, '2024-04-09 11:18:26+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 1483.285, '2024-04-09 11:18:26+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat porumb standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 5 LIMIT 1),
      '2024-04-22 07:00:00+00', '2024-04-22 09:29:37+00', 7.98, '', 'completed', v_fc,
      '2024-04-22 07:48:00+00', '2024-04-22 09:55:15+00', 7.62, 2.68, '',
      NULL, '2024-04-15 07:00:00+00', '2024-04-22 09:55:15+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 68.704, '2024-04-22 09:55:15+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta porumb lot B', 'out', 6.522, '2024-04-22 09:55:15+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2024-05-21 07:00:00+00', '2024-05-21 08:30:00+00', 7.98, '', 'completed', v_fc,
      '2024-05-21 07:18:00+00', '2024-05-21 07:39:30+00', 7.98, 1.41, '',
      NULL, '2024-05-14 07:00:00+00', '2024-05-21 07:39:30+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 37.945, '2024-05-21 07:39:30+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 17.199, '2024-05-21 07:39:30+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 2.019, '2024-05-21 07:39:30+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare porumb' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-CLAAS-LEXION780-01'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2024-10-04 07:00:00+00', '2024-10-04 10:19:30+00', 7.98, '', 'completed', v_fc,
      '2024-10-04 07:48:00+00', '2024-10-04 12:49:08+00', 7.98, 3.59, 'Întârziere: teren umed',
      NULL, '2024-09-27 07:00:00+00', '2024-10-04 12:49:08+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 139.717, '2024-10-04 12:49:08+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Porumb', 'in', 63.42, '2024-10-05 08:49:08+00', NULL, v_fc, 'Recoltă Porumb - Lingura - Teren agricol Nord (Sezon 2023/2024)');
  UPDATE field_crops SET production_total = 63.42, harvest_recorded_quantity = 63.42, harvest_recorded_at = '2024-10-05 08:49:08+00', harvested_at = '2024-10-05', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2024-10-09', harvested_at = COALESCE(harvested_at, '2024-10-09'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Cirpesti - Lot Nord (2.38 ha): Grâu ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2023/2024';
  SELECT id INTO v_field FROM fields WHERE name = 'Cirpesti - Lot Nord' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Cirpesti - Lot Nord lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Grâu'), 2.38, '2023-10-08', 5.8, '', 'closed', '2024-07-15', '2023-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 9 LIMIT 1),
      '2023-08-27 07:00:00+00', '2023-08-27 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2023-08-27 07:53:00+00', '2023-08-27 08:53:00+00', 2.38, 1.42, '',
      NULL, '2023-08-20 07:00:00+00', '2023-08-27 08:53:00+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 43.386, '2023-08-27 08:53:00+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 10 LIMIT 1),
      '2023-10-01 07:00:00+00', '2023-10-01 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2023-10-01 06:43:00+00', '2023-10-01 07:17:16+00', 2.38, 1.57, '',
      NULL, '2023-09-24 07:00:00+00', '2023-10-01 07:17:16+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 25.188, '2023-10-01 07:17:16+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 434.122, '2023-10-01 07:17:16+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat grau standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 11 LIMIT 1),
      '2023-10-09 07:00:00+00', '2023-10-09 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2023-10-09 06:43:00+00', '2023-10-09 08:19:34+00', 2.38, 1.58, '',
      NULL, '2023-10-02 07:00:00+00', '2023-10-09 08:19:34+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 23.558, '2023-10-09 08:19:34+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta grau lot A', 'out', 532.207, '2023-10-09 08:19:34+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare faziala grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 12 LIMIT 1),
      '2024-03-12 07:00:00+00', '2024-03-12 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2024-03-12 07:12:00+00', '2024-03-12 14:13:54+00', 2.38, 1.46, 'Întârziere: defecțiune minoră',
      NULL, '2024-03-05 07:00:00+00', '2024-03-12 14:13:54+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 15.972, '2024-03-12 14:13:54+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'out', 392.026, '2024-03-12 14:13:54+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2024-04-20 07:00:00+00', '2024-04-20 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2024-04-20 07:26:00+00', '2024-04-20 13:08:47+00', 2.38, 1.58, 'Întârziere: teren umed',
      NULL, '2024-04-13 07:00:00+00', '2024-04-20 13:08:47+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 11.572, '2024-04-20 13:08:47+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 5.369, '2024-04-20 13:08:47+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 0.596, '2024-04-20 13:08:47+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 1 LIMIT 1),
      '2024-05-09 07:00:00+00', '2024-05-09 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2024-05-09 06:41:00+00', '2024-05-09 08:26:01+00', 2.37, 1.62, '',
      NULL, '2024-05-02 07:00:00+00', '2024-05-09 08:26:01+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 12.228, '2024-05-09 08:26:01+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Fungicid lot rapita', 'out', 1.957, '2024-05-09 08:26:01+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 0.525, '2024-05-09 08:26:01+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-CLAAS-LEXION780-01'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2024-07-05 07:00:00+00', '2024-07-05 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2024-07-05 07:01:00+00', '2024-07-05 07:20:34+00', 2.38, 1.41, '',
      NULL, '2024-06-28 07:00:00+00', '2024-07-05 07:20:34+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 35.783, '2024-07-05 07:20:34+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'in', 11.43, '2024-07-06 03:20:34+00', NULL, v_fc, 'Recoltă Grâu - Cirpesti - Lot Nord (Sezon 2023/2024)');
  UPDATE field_crops SET production_total = 11.43, harvest_recorded_quantity = 11.43, harvest_recorded_at = '2024-07-06 03:20:34+00', harvested_at = '2024-07-06', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Balotat paie' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2024-07-12 07:00:00+00', '2024-07-12 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2024-07-12 06:50:00+00', '2024-07-12 08:01:38+00', 2.38, 1.39, '',
      NULL, '2024-07-05 07:00:00+00', '2024-07-12 08:01:38+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 7.351, '2024-07-12 08:01:38+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Paie baloti', 'in', 5.83, '2024-07-12 14:01:38+00', NULL, v_fc, 'Paie Grâu - Cirpesti - Lot Nord, balotat');
  UPDATE field_crops SET status = 'closed', closed_at = '2024-07-17', harvested_at = COALESCE(harvested_at, '2024-07-17'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Cirpesti - Lot Central (1.91 ha): Porumb ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2023/2024';
  SELECT id INTO v_field FROM fields WHERE name = 'Cirpesti - Lot Central' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Cirpesti - Lot Central lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Porumb'), 1.91, '2024-04-22', 9.0, '', 'closed', '2024-07-15', '2023-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 5 LIMIT 1),
      '2023-10-23 07:00:00+00', '2023-10-23 08:30:00+00', 1.91, '', 'completed', v_fc,
      '2023-10-23 07:39:00+00', '2023-10-23 07:52:12+00', 1.91, 1.63, '',
      NULL, '2023-10-16 07:00:00+00', '2023-10-23 07:52:12+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 37.9, '2023-10-23 07:52:12+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2024-04-12 07:00:00+00', '2024-04-12 08:30:00+00', 1.91, '', 'completed', v_fc,
      '2024-04-12 07:13:00+00', '2024-04-12 08:11:15+00', 1.91, 1.46, '',
      NULL, '2024-04-05 07:00:00+00', '2024-04-12 08:11:15+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 20.328, '2024-04-12 08:11:15+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 346.405, '2024-04-12 08:11:15+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat porumb standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2024-04-22 07:00:00+00', '2024-04-22 08:30:00+00', 1.91, '', 'completed', v_fc,
      '2024-04-22 07:22:00+00', '2024-04-22 08:54:05+00', 1.91, 1.52, '',
      NULL, '2024-04-15 07:00:00+00', '2024-04-22 08:54:05+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 18.05, '2024-04-22 08:54:05+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta porumb lot B', 'out', 1.601, '2024-04-22 08:54:05+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 8 LIMIT 1),
      '2024-05-23 07:00:00+00', '2024-05-23 08:30:00+00', 1.91, '', 'completed', v_fc,
      '2024-05-23 07:44:00+00', '2024-05-23 08:59:06+00', 1.91, 1.62, '',
      NULL, '2024-05-16 07:00:00+00', '2024-05-23 08:59:06+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 8.706, '2024-05-23 08:59:06+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 4.14, '2024-05-23 08:59:06+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 0.464, '2024-05-23 08:59:06+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare porumb' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-JOHNDEERE-S780-02'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 9 LIMIT 1),
      '2024-10-03 07:00:00+00', '2024-10-03 08:30:00+00', 1.91, '', 'completed', v_fc,
      '2024-10-03 07:20:00+00', '2024-10-03 07:45:41+00', 1.91, 1.4, '',
      NULL, '2024-09-26 07:00:00+00', '2024-10-03 07:45:41+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 32.54, '2024-10-03 07:45:41+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Porumb', 'in', 15.15, '2024-10-04 03:45:41+00', NULL, v_fc, 'Recoltă Porumb - Cirpesti - Lot Central (Sezon 2023/2024)');
  UPDATE field_crops SET production_total = 15.15, harvest_recorded_quantity = 15.15, harvest_recorded_at = '2024-10-04 03:45:41+00', harvested_at = '2024-10-04', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2024-10-08', harvested_at = COALESCE(harvested_at, '2024-10-08'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Cirpesti - Lot Sud (1.56 ha): Floarea-soarelui ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2023/2024';
  SELECT id INTO v_field FROM fields WHERE name = 'Cirpesti - Lot Sud' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Cirpesti - Lot Sud lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Floarea-soarelui'), 1.56, '2024-04-14', 2.6, '', 'closed', '2024-07-15', '2023-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 11 LIMIT 1),
      '2023-10-26 07:00:00+00', '2023-10-26 08:30:00+00', 1.56, '', 'completed', v_fc,
      '2023-10-26 06:57:00+00', '2023-10-26 07:57:58+00', 1.56, 1.54, '',
      NULL, '2023-10-19 07:00:00+00', '2023-10-26 07:57:58+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 29.295, '2023-10-26 07:57:58+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 12 LIMIT 1),
      '2024-04-03 07:00:00+00', '2024-04-03 08:30:00+00', 1.56, '', 'completed', v_fc,
      '2024-04-03 07:19:00+00', '2024-04-03 07:23:50+00', 1.56, 1.39, '',
      NULL, '2024-03-27 07:00:00+00', '2024-04-03 07:23:50+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 16.076, '2024-04-03 07:23:50+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 285.902, '2024-04-03 07:23:50+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat floarea-soarelui standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2024-04-13 07:00:00+00', '2024-04-13 08:30:00+00', 1.56, '', 'completed', v_fc,
      '2024-04-13 07:24:00+00', '2024-04-13 08:02:52+00', 1.56, 1.63, '',
      NULL, '2024-04-06 07:00:00+00', '2024-04-13 08:02:52+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 13.05, '2024-04-13 08:02:52+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta floarea-soarelui hibrid', 'out', 5.55, '2024-04-13 08:02:52+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 1 LIMIT 1),
      '2024-05-15 07:00:00+00', '2024-05-15 08:30:00+00', 1.56, '', 'completed', v_fc,
      '2024-05-15 06:59:00+00', '2024-05-15 09:52:56+00', 1.56, 1.55, 'Întârziere: teren umed',
      NULL, '2024-05-08 07:00:00+00', '2024-05-15 09:52:56+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 7.145, '2024-05-15 09:52:56+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 3.441, '2024-05-15 09:52:56+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 0.382, '2024-05-15 09:52:56+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare floarea-soarelui' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-JOHNDEERE-S780-02'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2024-09-11 07:00:00+00', '2024-09-11 08:30:00+00', 1.56, '', 'completed', v_fc,
      '2024-09-11 07:35:00+00', '2024-09-11 08:35:00+00', 1.56, 1.44, '',
      NULL, '2024-09-04 07:00:00+00', '2024-09-11 08:35:00+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 20.25, '2024-09-11 08:35:00+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Floarea-soarelui', 'in', 3.4, '2024-09-12 04:35:00+00', NULL, v_fc, 'Recoltă Floarea-soarelui - Cirpesti - Lot Sud (Sezon 2023/2024)');
  UPDATE field_crops SET production_total = 3.4, harvest_recorded_quantity = 3.4, harvest_recorded_at = '2024-09-12 04:35:00+00', harvested_at = '2024-09-12', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2024-09-16', harvested_at = COALESCE(harvested_at, '2024-09-16'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ─── Vânzări din stoc, după recoltă (ieșiri fără lucrare) ───
DO $sale$ BEGIN
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'out', 81.89, '2024-07-23 20:38:16+00', NULL, NULL, '[Sezon 2023/2024] Vânzare Lucernă — contract, Gotesti - Tarla Vest');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'out', 41.76, '2024-08-10 08:45:30+00', NULL, NULL, '[Sezon 2023/2024] Vânzare Grâu — contract, Tartaul - Tarla Mica');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'out', 920.69, '2024-08-10 21:09:45+00', NULL, NULL, '[Sezon 2023/2024] Vânzare Grâu — contract, Plopi - Tarla Mare');
  PERFORM pg_temp.seed_move('Paie baloti', 'out', 3.5, '2024-08-11 17:01:38+00', NULL, NULL, '[Sezon 2023/2024] Vânzare paie — Cirpesti - Lot Nord');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'out', 122.25, '2024-08-11 19:51:33+00', NULL, NULL, '[Sezon 2023/2024] Vânzare Lucernă — contract, Gotesti - Tarla Vest');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'out', 223.52, '2024-08-13 03:34:16+00', NULL, NULL, '[Sezon 2023/2024] Vânzare Grâu — contract, Plopi - Tarla Nord');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'out', 9.72, '2024-08-23 06:20:34+00', NULL, NULL, '[Sezon 2023/2024] Vânzare Grâu — contract, Cirpesti - Lot Nord');
  PERFORM pg_temp.seed_move('Recoltă Orz', 'out', 216.13, '2024-08-25 08:59:28+00', NULL, NULL, '[Sezon 2023/2024] Vânzare Orz — contract, Tartaul - Tarla Centrala');
  PERFORM pg_temp.seed_move('Paie baloti', 'out', 55.35, '2024-08-28 14:36:06+00', NULL, NULL, '[Sezon 2023/2024] Vânzare paie — Tartaul - Tarla Centrala');
  PERFORM pg_temp.seed_move('Paie baloti', 'out', 332.08, '2024-08-29 07:00:49+00', NULL, NULL, '[Sezon 2023/2024] Vânzare paie — Plopi - Tarla Mare');
  PERFORM pg_temp.seed_move('Recoltă Rapiță', 'out', 226.53, '2024-08-29 19:58:15+00', NULL, NULL, '[Sezon 2023/2024] Vânzare Rapiță — contract, Gotesti - Tarla Est');
  PERFORM pg_temp.seed_move('Paie baloti', 'out', 71.54, '2024-08-31 18:53:28+00', NULL, NULL, '[Sezon 2023/2024] Vânzare paie — Plopi - Tarla Nord');
  PERFORM pg_temp.seed_move('Paie baloti', 'out', 13.83, '2024-09-10 19:05:00+00', NULL, NULL, '[Sezon 2023/2024] Vânzare paie — Tartaul - Tarla Mica');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'out', 103.44, '2024-10-11 19:20:48+00', NULL, NULL, '[Sezon 2023/2024] Vânzare Lucernă — contract, Gotesti - Tarla Vest');
  PERFORM pg_temp.seed_move('Recoltă Floarea-soarelui', 'out', 161.55, '2024-10-27 17:05:42+00', NULL, NULL, '[Sezon 2023/2024] Vânzare Floarea-soarelui — contract, Tartaul - Tarla Mare');
  PERFORM pg_temp.seed_move('Recoltă Soia', 'out', 102.34, '2024-11-04 01:45:20+00', NULL, NULL, '[Sezon 2023/2024] Vânzare Soia — contract, Plopi - Tarla Sud');
  PERFORM pg_temp.seed_move('Recoltă Floarea-soarelui', 'out', 2.89, '2024-11-04 07:35:00+00', NULL, NULL, '[Sezon 2023/2024] Vânzare Floarea-soarelui — contract, Cirpesti - Lot Sud');
  PERFORM pg_temp.seed_move('Recoltă Porumb', 'out', 53.91, '2024-11-05 11:49:08+00', NULL, NULL, '[Sezon 2023/2024] Vânzare Porumb — contract, Lingura - Teren agricol Nord');
  PERFORM pg_temp.seed_move('Recoltă Porumb', 'out', 844.92, '2024-11-10 11:20:00+00', NULL, NULL, '[Sezon 2023/2024] Vânzare Porumb — contract, Gotesti - Tarla Mare');
  PERFORM pg_temp.seed_move('Recoltă Porumb', 'out', 12.88, '2024-11-14 06:45:41+00', NULL, NULL, '[Sezon 2023/2024] Vânzare Porumb — contract, Cirpesti - Lot Central');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'out', 75.23, '2024-11-23 19:05:12+00', NULL, NULL, '[Sezon 2023/2024] Vânzare Lucernă — contract, Gotesti - Tarla Vest');
END $sale$;

-- ─── Istoricul stocului în ordine cronologică (resulting_quantity ca sold curent la acea dată) ───
UPDATE stock_movements sm SET resulting_quantity = t.run FROM (SELECT id, SUM(quantity_delta) OVER (PARTITION BY resource_id ORDER BY created_at, id) AS run FROM stock_movements) t WHERE t.id = sm.id;

COMMIT;
