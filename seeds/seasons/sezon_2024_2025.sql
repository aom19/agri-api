-- Seed sezon Sezon 2024/2025 — generat de seeds/seasons/generate.py; nu edita manual, modifică generatorul.
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
    SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2024/2025';
    IF v_season IS NOT NULL THEN
        DELETE FROM stock_movements WHERE field_operation_id IN (SELECT id FROM field_operations WHERE field_crop_id IN (SELECT id FROM field_crops WHERE season_id = v_season));
        DELETE FROM stock_movements WHERE field_crop_id IN (SELECT id FROM field_crops WHERE season_id = v_season);
        DELETE FROM field_operations WHERE field_crop_id IN (SELECT id FROM field_crops WHERE season_id = v_season);
        DELETE FROM field_crops WHERE season_id = v_season;
        DELETE FROM seasons WHERE id = v_season;
    END IF;
    -- aprovizionările și lucrările fără cultură ale sezonului sunt etichetate în notes
    DELETE FROM stock_movements WHERE field_operation_id IN (SELECT id FROM field_operations WHERE notes LIKE '[Sezon 2024/2025]%');
    DELETE FROM field_operations WHERE notes LIKE '[Sezon 2024/2025]%';
    DELETE FROM stock_movements WHERE notes LIKE '[Sezon 2024/2025]%';
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
INSERT INTO seasons (name, start_date, end_date, is_active, notes) VALUES ('Sezon 2024/2025', '2024-08-01', '2025-10-31', false, 'An bun, cu precipitații normale.');

-- ─── Aprovizionări (dimensionate după consumul sezonului, +5%) ───
DO $sup$ BEGIN
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'in', 39000, '2025-04-08 12:11:48+00', NULL, NULL, '[Sezon 2024/2025] Alimentare bazin pivot');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'in', 1600, '2025-04-11 07:36:41+00', NULL, NULL, '[Sezon 2024/2025] Recepție furnizor');
  PERFORM pg_temp.seed_move('Fungicid lot rapita', 'in', 130, '2025-04-08 12:11:48+00', NULL, NULL, '[Sezon 2024/2025] Recepție furnizor');
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'in', 24000, '2024-07-31 12:07:27+00', NULL, NULL, '[Sezon 2024/2025] Recepție furnizor — tranșa 1');
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'in', 24000, '2024-12-28 12:07:27+00', NULL, NULL, '[Sezon 2024/2025] Recepție furnizor — tranșa 2');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'in', 139000, '2024-08-12 12:42:38+00', NULL, NULL, '[Sezon 2024/2025] Recepție furnizor');
  PERFORM pg_temp.seed_move('Samanta floarea-soarelui hibrid', 'in', 800, '2025-04-14 03:02:34+00', NULL, NULL, '[Sezon 2024/2025] Recepție furnizor');
  PERFORM pg_temp.seed_move('Samanta grau lot A', 'in', 63150, '2024-09-30 13:52:48+00', NULL, NULL, '[Sezon 2024/2025] Recepție furnizor');
  PERFORM pg_temp.seed_move('Samanta porumb lot B', 'in', 130, '2025-04-15 07:34:25+00', NULL, NULL, '[Sezon 2024/2025] Recepție furnizor');
  PERFORM pg_temp.seed_move('Samanta rapita hibrid', 'in', 45, '2024-08-23 08:36:22+00', NULL, NULL, '[Sezon 2024/2025] Recepție furnizor');
  PERFORM pg_temp.seed_move('Samanta soia', 'in', 5650, '2025-04-25 22:50:37+00', NULL, NULL, '[Sezon 2024/2025] Recepție furnizor');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'in', 49500, '2025-02-27 08:20:00+00', NULL, NULL, '[Sezon 2024/2025] Recepție furnizor');
END $sup$;

-- ── Plopi - Tarla Mare (210.25 ha): Floarea-soarelui ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2024/2025';
  SELECT id INTO v_field FROM fields WHERE name = 'Plopi - Tarla Mare' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Plopi - Tarla Mare lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Floarea-soarelui'), 210.25, '2025-04-14', 2.6, '', 'closed', '2025-07-15', '2024-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2024-10-26 07:00:00+00', '2024-10-31 18:24:22+00', 210.25, '', 'completed', v_fc,
      '2024-10-26 07:19:00+00', '2024-10-31 17:19:40+00', 210.25, 132.31, '',
      NULL, '2024-10-19 07:00:00+00', '2024-10-31 17:19:40+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 3808.106, '2024-10-31 17:19:40+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 1 LIMIT 1),
      '2025-04-07 07:00:00+00', '2025-04-12 18:24:22+00', 210.25, '', 'completed', v_fc,
      '2025-04-07 07:30:00+00', '2025-04-12 17:10:56+00', 210.25, 139.88, '',
      NULL, '2025-03-31 07:00:00+00', '2025-04-12 17:10:56+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 2139.642, '2025-04-12 17:10:56+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 36739.004, '2025-04-12 17:10:56+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat floarea-soarelui standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2025-04-17 07:00:00+00', '2025-04-20 00:42:11+00', 210.25, '', 'completed', v_fc,
      '2025-04-17 06:52:00+00', '2025-04-20 03:02:34+00', 210.25, 70.99, 'Întârziere: ploaie',
      NULL, '2025-04-10 07:00:00+00', '2025-04-20 03:02:34+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1852.134, '2025-04-20 03:02:34+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta floarea-soarelui hibrid', 'out', 743.678, '2025-04-20 03:02:34+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2025-05-11 07:00:00+00', '2025-05-12 06:21:40+00', 210.25, '', 'completed', v_fc,
      '2025-05-11 07:58:00+00', '2025-05-12 09:05:17+00', 201.24, 24.4, 'Întârziere: ploaie',
      NULL, '2025-05-04 07:00:00+00', '2025-05-12 09:05:17+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 960.224, '2025-05-12 09:05:17+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 440.214, '2025-05-12 09:05:17+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 49.721, '2025-05-12 09:05:17+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare floarea-soarelui' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-CLAAS-LEXION780-01'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2025-09-07 07:00:00+00', '2025-09-10 22:36:15+00', 210.25, '', 'completed', v_fc,
      '2025-09-07 07:39:00+00', '2025-09-10 22:18:47+00', 208.3, 88.34, '',
      NULL, '2025-08-31 07:00:00+00', '2025-09-10 22:18:47+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 2509.161, '2025-09-10 22:18:47+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Floarea-soarelui', 'in', 549.53, '2025-09-11 18:18:47+00', NULL, v_fc, 'Recoltă Floarea-soarelui - Plopi - Tarla Mare (Sezon 2024/2025)');
  UPDATE field_crops SET production_total = 549.53, harvest_recorded_quantity = 549.53, harvest_recorded_at = '2025-09-11 18:18:47+00', harvested_at = '2025-09-11', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2025-09-15', harvested_at = COALESCE(harvested_at, '2025-09-15'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Gotesti - Tarla Mare (130.03 ha): Grâu ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2024/2025';
  SELECT id INTO v_field FROM fields WHERE name = 'Gotesti - Tarla Mare' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Gotesti - Tarla Mare lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Grâu'), 130.03, '2024-10-08', 5.8, '', 'closed', '2025-07-15', '2024-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2024-08-25 07:00:00+00', '2024-08-28 16:16:07+00', 130.03, '', 'completed', v_fc,
      '2024-08-25 07:53:00+00', '2024-08-28 16:15:30+00', 127.01, 79.39, '',
      NULL, '2024-08-18 07:00:00+00', '2024-08-28 16:15:30+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 2317.259, '2024-08-28 16:15:30+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2024-09-29 07:00:00+00', '2024-10-02 16:16:07+00', 130.03, '', 'completed', v_fc,
      '2024-09-29 07:21:00+00', '2024-10-02 14:57:30+00', 125.69, 75.69, '',
      NULL, '2024-09-22 07:00:00+00', '2024-10-02 14:57:30+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1288.177, '2024-10-02 14:57:30+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 22331.963, '2024-10-02 14:57:30+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat grau standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 8 LIMIT 1),
      '2024-10-09 07:00:00+00', '2024-10-10 23:38:03+00', 130.03, '', 'completed', v_fc,
      '2024-10-09 07:53:00+00', '2024-10-10 23:32:27+00', 123.96, 42.62, '',
      NULL, '2024-10-02 07:00:00+00', '2024-10-10 23:32:27+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1142.574, '2024-10-10 23:32:27+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta grau lot A', 'out', 26985.574, '2024-10-10 23:32:27+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare faziala grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 9 LIMIT 1),
      '2025-03-14 07:00:00+00', '2025-03-15 01:34:32+00', 130.03, '', 'completed', v_fc,
      '2025-03-14 07:19:00+00', '2025-03-15 01:04:15+00', 130.03, 17.79, '',
      NULL, '2025-03-07 07:00:00+00', '2025-03-15 01:04:15+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 812.166, '2025-03-15 01:04:15+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'out', 20904.239, '2025-03-15 01:04:15+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 10 LIMIT 1),
      '2025-04-20 07:00:00+00', '2025-04-20 21:26:52+00', 130.03, '', 'completed', v_fc,
      '2025-04-20 07:57:00+00', '2025-04-21 01:16:22+00', 129.43, 14.26, 'Întârziere: teren umed',
      NULL, '2025-04-13 07:00:00+00', '2025-04-21 01:16:22+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 618.389, '2025-04-21 01:16:22+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 293.084, '2025-04-21 01:16:22+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 33.648, '2025-04-21 01:16:22+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 11 LIMIT 1),
      '2025-05-11 07:00:00+00', '2025-05-11 21:26:52+00', 130.03, '', 'completed', v_fc,
      '2025-05-11 07:33:00+00', '2025-05-11 23:04:55+00', 127.46, 15.58, 'Întârziere: ploaie',
      NULL, '2025-05-04 07:00:00+00', '2025-05-11 23:04:55+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 618.95, '2025-05-11 23:04:55+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Fungicid lot rapita', 'out', 103.671, '2025-05-11 23:04:55+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 29.087, '2025-05-11 23:04:55+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-CLAAS-LEXION780-01'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 12 LIMIT 1),
      '2025-07-04 07:00:00+00', '2025-07-06 13:10:45+00', 130.03, '', 'completed', v_fc,
      '2025-07-04 07:49:00+00', '2025-07-06 13:21:23+00', 130.03, 55.09, '',
      NULL, '2025-06-27 07:00:00+00', '2025-07-06 13:21:23+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 2028.784, '2025-07-06 13:21:23+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'in', 785.7, '2025-07-07 09:21:23+00', NULL, v_fc, 'Recoltă Grâu - Gotesti - Tarla Mare (Sezon 2024/2025)');
  UPDATE field_crops SET production_total = 785.7, harvest_recorded_quantity = 785.7, harvest_recorded_at = '2025-07-07 09:21:23+00', harvested_at = '2025-07-07', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Balotat paie' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2025-07-10 07:00:00+00', '2025-07-12 13:10:45+00', 130.03, '', 'completed', v_fc,
      '2025-07-10 07:28:00+00', '2025-07-12 12:16:47+00', 123.57, 51.92, '',
      NULL, '2025-07-03 07:00:00+00', '2025-07-12 12:16:47+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 387.93, '2025-07-12 12:16:47+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Paie baloti', 'in', 299.59, '2025-07-12 18:16:47+00', NULL, v_fc, 'Paie Grâu - Gotesti - Tarla Mare, balotat');
  UPDATE field_crops SET status = 'closed', closed_at = '2025-07-17', harvested_at = COALESCE(harvested_at, '2025-07-17'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Gotesti - Tarla Est (92.34 ha): Grâu ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2024/2025';
  SELECT id INTO v_field FROM fields WHERE name = 'Gotesti - Tarla Est' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Gotesti - Tarla Est lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Grâu'), 92.34, '2024-10-08', 5.8, '', 'closed', '2025-07-15', '2024-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2024-08-25 07:00:00+00', '2024-08-27 16:42:45+00', 92.34, '', 'completed', v_fc,
      '2024-08-25 07:45:00+00', '2024-08-27 15:56:43+00', 90.18, 53.95, '',
      NULL, '2024-08-18 07:00:00+00', '2024-08-27 15:56:43+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1759.878, '2024-08-27 15:56:43+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2024-09-30 07:00:00+00', '2024-10-02 16:42:45+00', 92.34, '', 'completed', v_fc,
      '2024-09-30 07:26:00+00', '2024-10-02 22:15:21+00', 92.34, 54.5, 'Întârziere: ploaie',
      NULL, '2024-09-23 07:00:00+00', '2024-10-02 22:15:21+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 935.221, '2024-10-02 22:15:21+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 16436.316, '2024-10-02 22:15:21+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat grau standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2024-10-10 07:00:00+00', '2024-10-11 11:51:22+00', 92.34, '', 'completed', v_fc,
      '2024-10-10 06:48:00+00', '2024-10-11 11:26:12+00', 92.34, 30.91, '',
      NULL, '2024-10-03 07:00:00+00', '2024-10-11 11:26:12+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 867.545, '2024-10-11 11:26:12+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta grau lot A', 'out', 19886.298, '2024-10-11 11:26:12+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare faziala grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 5 LIMIT 1),
      '2025-03-14 07:00:00+00', '2025-03-14 20:11:29+00', 92.34, '', 'completed', v_fc,
      '2025-03-14 06:48:00+00', '2025-03-14 19:03:25+00', 92.34, 13.24, '',
      NULL, '2025-03-07 07:00:00+00', '2025-03-14 19:03:25+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 583.771, '2025-03-14 19:03:25+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'out', 14985.217, '2025-03-14 19:03:25+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2025-04-17 07:00:00+00', '2025-04-17 17:15:36+00', 92.34, '', 'completed', v_fc,
      '2025-04-17 07:47:00+00', '2025-04-17 15:50:06+00', 88.84, 9.81, '',
      NULL, '2025-04-10 07:00:00+00', '2025-04-17 15:50:06+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 416.19, '2025-04-17 15:50:06+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 203.106, '2025-04-17 15:50:06+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 22.636, '2025-04-17 15:50:06+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2025-05-08 07:00:00+00', '2025-05-08 17:15:36+00', 92.34, '', 'canceled', v_fc,
      NULL, NULL, NULL, NULL, 'Anulată: ploi persistente în fereastra de tratament',
      NULL, '2025-05-01 07:00:00+00', '2025-05-08 07:00:00+00')
  RETURNING id INTO v_op;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-JOHNDEERE-S780-02'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 8 LIMIT 1),
      '2025-07-06 07:00:00+00', '2025-07-07 21:28:30+00', 92.34, '', 'completed', v_fc,
      '2025-07-06 07:01:00+00', '2025-07-07 20:00:09+00', 92.34, 37.41, '',
      NULL, '2025-06-29 07:00:00+00', '2025-07-07 20:00:09+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1332.922, '2025-07-07 20:00:09+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'in', 588.34, '2025-07-08 16:00:09+00', NULL, v_fc, 'Recoltă Grâu - Gotesti - Tarla Est (Sezon 2024/2025)');
  UPDATE field_crops SET production_total = 588.34, harvest_recorded_quantity = 588.34, harvest_recorded_at = '2025-07-08 16:00:09+00', harvested_at = '2025-07-08', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Balotat paie' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 9 LIMIT 1),
      '2025-07-14 07:00:00+00', '2025-07-15 21:28:30+00', 92.34, '', 'completed', v_fc,
      '2025-07-14 07:34:00+00', '2025-07-15 20:32:09+00', 92.34, 39.95, '',
      NULL, '2025-07-07 07:00:00+00', '2025-07-15 20:32:09+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 284.545, '2025-07-15 20:32:09+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Paie baloti', 'in', 233.62, '2025-07-16 02:32:09+00', NULL, v_fc, 'Paie Grâu - Gotesti - Tarla Est, balotat');
  UPDATE field_crops SET status = 'closed', closed_at = '2025-07-20', harvested_at = COALESCE(harvested_at, '2025-07-20'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Tartaul - Tarla Mare (84.02 ha): Porumb ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2024/2025';
  SELECT id INTO v_field FROM fields WHERE name = 'Tartaul - Tarla Mare' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Tartaul - Tarla Mare lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Porumb'), 84.02, '2025-04-22', 9.0, '', 'closed', '2025-07-15', '2024-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 11 LIMIT 1),
      '2024-10-27 07:00:00+00', '2024-10-29 11:30:45+00', 84.02, '', 'completed', v_fc,
      '2024-10-27 06:53:00+00', '2024-10-29 17:06:14+00', 83.52, 51.65, 'Întârziere: defecțiune minoră',
      NULL, '2024-10-20 07:00:00+00', '2024-10-29 17:06:14+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1533.79, '2024-10-29 17:06:14+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 12 LIMIT 1),
      '2025-04-12 07:00:00+00', '2025-04-14 11:30:45+00', 84.02, '', 'completed', v_fc,
      '2025-04-12 07:35:00+00', '2025-04-14 11:34:47+00', 84.02, 53.05, '',
      NULL, '2025-04-05 07:00:00+00', '2025-04-14 11:34:47+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 925.703, '2025-04-14 11:34:47+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 14806.313, '2025-04-14 11:34:47+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat porumb standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2025-04-20 07:00:00+00', '2025-04-21 09:15:22+00', 84.02, '', 'completed', v_fc,
      '2025-04-20 06:41:00+00', '2025-04-21 07:54:32+00', 84.02, 27.61, '',
      NULL, '2025-04-13 07:00:00+00', '2025-04-21 07:54:32+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 752.878, '2025-04-21 07:54:32+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta porumb lot B', 'out', 72.744, '2025-04-21 07:54:32+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 1 LIMIT 1),
      '2025-05-18 07:00:00+00', '2025-05-18 16:20:08+00', 84.02, '', 'completed', v_fc,
      '2025-05-18 07:57:00+00', '2025-05-18 16:40:23+00', 84.02, 8.7, '',
      NULL, '2025-05-11 07:00:00+00', '2025-05-18 16:40:23+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 408.397, '2025-05-18 16:40:23+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 179.698, '2025-05-18 16:40:23+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 21.331, '2025-05-18 16:40:23+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare porumb' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-CLAAS-LEXION780-01'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2025-10-01 07:00:00+00', '2025-10-02 18:00:30+00', 84.02, '', 'completed', v_fc,
      '2025-10-01 07:40:00+00', '2025-10-02 16:43:00+00', 84.02, 34.93, '',
      NULL, '2025-09-24 07:00:00+00', '2025-10-02 16:43:00+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1425.687, '2025-10-02 16:43:00+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Porumb', 'in', 830.02, '2025-10-03 12:43:00+00', NULL, v_fc, 'Recoltă Porumb - Tartaul - Tarla Mare (Sezon 2024/2025)');
  UPDATE field_crops SET production_total = 830.02, harvest_recorded_quantity = 830.02, harvest_recorded_at = '2025-10-03 12:43:00+00', harvested_at = '2025-10-03', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2025-10-07', harvested_at = COALESCE(harvested_at, '2025-10-07'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Gotesti - Tarla Vest (68.47 ha): Lucernă, anul 2 ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2024/2025';
  SELECT id INTO v_field FROM fields WHERE name = 'Gotesti - Tarla Vest' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Gotesti - Tarla Vest lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Lucernă'), 68.47, NULL, 9.5, 'Lucernă anul 2', 'closed', '2025-07-15', '2024-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare lucerna' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2025-03-08 07:00:00+00', '2025-03-08 16:46:53+00', 68.47, '', 'completed', v_fc,
      '2025-03-08 06:43:00+00', '2025-03-08 16:20:32+00', 68.47, 9.0, '',
      NULL, '2025-03-01 07:00:00+00', '2025-03-08 16:20:32+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 416.487, '2025-03-08 16:20:32+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 10316.569, '2025-03-08 16:20:32+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Cosit si balotat lucerna' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 5 LIMIT 1),
      '2025-05-15 07:00:00+00', '2025-05-16 11:31:45+00', 68.47, '', 'completed', v_fc,
      '2025-05-15 07:48:00+00', '2025-05-16 11:29:13+00', 68.47, 27.15, '',
      NULL, '2025-05-08 07:00:00+00', '2025-05-16 11:29:13+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 581.717, '2025-05-16 11:29:13+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'in', 194.79, '2025-05-17 17:29:13+00', NULL, v_fc, 'Recoltă Lucernă - Gotesti - Tarla Vest (Sezon 2024/2025)');
  UPDATE field_crops SET production_total = 194.79, harvest_recorded_quantity = 194.79, harvest_recorded_at = '2025-05-17 17:29:13+00', harvested_at = '2025-05-17', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Cosit si balotat lucerna' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2025-06-30 07:00:00+00', '2025-07-01 11:31:45+00', 68.47, '', 'completed', v_fc,
      '2025-06-30 07:21:00+00', '2025-07-01 11:30:22+00', 68.47, 28.21, '',
      NULL, '2025-06-23 07:00:00+00', '2025-07-01 11:30:22+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 570.82, '2025-07-01 11:30:22+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'in', 222.62, '2025-07-02 17:30:22+00', NULL, v_fc, 'Recoltă Lucernă - Gotesti - Tarla Vest (Sezon 2024/2025)');
  UPDATE field_crops SET production_total = 417.41, harvest_recorded_quantity = 417.41, harvest_recorded_at = '2025-07-02 17:30:22+00', harvested_at = '2025-07-02', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Cosit si balotat lucerna' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2025-08-10 07:00:00+00', '2025-08-11 11:31:45+00', 68.47, '', 'completed', v_fc,
      '2025-08-10 07:15:00+00', '2025-08-11 10:17:10+00', 68.47, 29.4, '',
      NULL, '2025-08-03 07:00:00+00', '2025-08-11 10:17:10+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 556.834, '2025-08-11 10:17:10+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'in', 180.88, '2025-08-12 16:17:10+00', NULL, v_fc, 'Recoltă Lucernă - Gotesti - Tarla Vest (Sezon 2024/2025)');
  UPDATE field_crops SET production_total = 598.29, harvest_recorded_quantity = 598.29, harvest_recorded_at = '2025-08-12 16:17:10+00', harvested_at = '2025-08-12', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Cosit si balotat lucerna' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 8 LIMIT 1),
      '2025-09-18 07:00:00+00', '2025-09-19 11:31:45+00', 68.47, '', 'completed', v_fc,
      '2025-09-18 07:31:00+00', '2025-09-19 16:18:24+00', 68.47, 29.76, 'Întârziere: ploaie',
      NULL, '2025-09-11 07:00:00+00', '2025-09-19 16:18:24+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 556.309, '2025-09-19 16:18:24+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'in', 132.18, '2025-09-20 22:18:24+00', NULL, v_fc, 'Recoltă Lucernă - Gotesti - Tarla Vest (Sezon 2024/2025)');
  UPDATE field_crops SET production_total = 730.47, harvest_recorded_quantity = 730.47, harvest_recorded_at = '2025-09-20 22:18:24+00', harvested_at = '2025-09-20', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2025-09-24', harvested_at = COALESCE(harvested_at, '2025-09-24'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Tartaul - Tarla Centrala (53.5 ha): Soia ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2024/2025';
  SELECT id INTO v_field FROM fields WHERE name = 'Tartaul - Tarla Centrala' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Tartaul - Tarla Centrala lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Soia'), 53.5, '2025-04-28', 2.8, '', 'closed', '2025-07-15', '2024-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 10 LIMIT 1),
      '2024-11-02 07:00:00+00', '2024-11-03 16:26:15+00', 53.5, '', 'completed', v_fc,
      '2024-11-02 07:18:00+00', '2024-11-03 15:41:21+00', 53.5, 33.31, '',
      NULL, '2024-10-26 07:00:00+00', '2024-11-03 15:41:21+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 967.553, '2024-11-03 15:41:21+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 11 LIMIT 1),
      '2025-04-19 07:00:00+00', '2025-04-20 16:26:15+00', 53.5, '', 'completed', v_fc,
      '2025-04-19 07:50:00+00', '2025-04-20 15:37:19+00', 53.5, 35.22, '',
      NULL, '2025-04-12 07:00:00+00', '2025-04-20 15:37:19+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 550.382, '2025-04-20 15:37:19+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 10012.501, '2025-04-20 15:37:19+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat soia standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 12 LIMIT 1),
      '2025-05-01 07:00:00+00', '2025-05-01 23:43:07+00', 53.5, '', 'completed', v_fc,
      '2025-05-01 06:44:00+00', '2025-05-01 22:50:37+00', 51.96, 15.85, '',
      NULL, '2025-04-24 07:00:00+00', '2025-05-01 22:50:37+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 447.014, '2025-05-01 22:50:37+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta soia', 'out', 5354.812, '2025-05-01 22:50:37+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2025-05-31 07:00:00+00', '2025-05-31 12:56:40+00', 53.5, '', 'completed', v_fc,
      '2025-05-31 06:46:00+00', '2025-05-31 12:11:24+00', 53.5, 6.19, '',
      NULL, '2025-05-24 07:00:00+00', '2025-05-31 12:11:24+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 248.126, '2025-05-31 12:11:24+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 115.451, '2025-05-31 12:11:24+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 13.356, '2025-05-31 12:11:24+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare soia' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-CLAAS-LEXION780-01'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 1 LIMIT 1),
      '2025-09-23 07:00:00+00', '2025-09-24 05:17:30+00', 53.5, '', 'completed', v_fc,
      '2025-09-23 06:47:00+00', '2025-09-24 04:54:40+00', 53.5, 23.99, '',
      NULL, '2025-09-16 07:00:00+00', '2025-09-24 04:54:40+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 673.27, '2025-09-24 04:54:40+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Soia', 'in', 166.94, '2025-09-25 00:54:40+00', NULL, v_fc, 'Recoltă Soia - Tartaul - Tarla Centrala (Sezon 2024/2025)');
  UPDATE field_crops SET production_total = 166.94, harvest_recorded_quantity = 166.94, harvest_recorded_at = '2025-09-25 00:54:40+00', harvested_at = '2025-09-25', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2025-09-29', harvested_at = COALESCE(harvested_at, '2025-09-29'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Plopi - Tarla Nord (52.41 ha): Porumb ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2024/2025';
  SELECT id INTO v_field FROM fields WHERE name = 'Plopi - Tarla Nord' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Plopi - Tarla Nord lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Porumb'), 52.41, '2025-04-22', 10.35, '', 'closed', '2025-07-15', '2024-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2024-10-26 07:00:00+00', '2024-10-27 15:45:22+00', 52.41, '', 'completed', v_fc,
      '2024-10-26 07:07:00+00', '2024-10-27 17:23:57+00', 52.41, 30.15, 'Întârziere: ploaie',
      NULL, '2024-10-19 07:00:00+00', '2024-10-27 17:23:57+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 998.742, '2024-10-27 17:23:57+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2025-04-11 07:00:00+00', '2025-04-12 15:45:22+00', 52.41, '', 'completed', v_fc,
      '2025-04-11 07:27:00+00', '2025-04-12 15:15:36+00', 52.41, 35.6, '',
      NULL, '2025-04-04 07:00:00+00', '2025-04-12 15:15:36+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 531.817, '2025-04-12 15:15:36+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 9260.492, '2025-04-12 15:15:36+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat porumb standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 5 LIMIT 1),
      '2025-04-24 07:00:00+00', '2025-04-24 23:22:41+00', 52.41, '', 'completed', v_fc,
      '2025-04-24 07:06:00+00', '2025-04-24 22:26:42+00', 52.41, 17.1, '',
      NULL, '2025-04-17 07:00:00+00', '2025-04-24 22:26:42+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 463.917, '2025-04-24 22:26:42+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta porumb lot B', 'out', 45.475, '2025-04-24 22:26:42+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2025-05-23 07:00:00+00', '2025-05-23 12:49:24+00', 52.41, '', 'completed', v_fc,
      '2025-05-23 07:18:00+00', '2025-05-23 12:15:19+00', 51.81, 6.01, '',
      NULL, '2025-05-16 07:00:00+00', '2025-05-23 12:15:19+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 238.277, '2025-05-23 12:15:19+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 111.806, '2025-05-23 12:15:19+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 12.976, '2025-05-23 12:15:19+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Irigare pivot' AND deleted_at IS NULL),
      NULL, NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2025-06-23 07:00:00+00', '2025-06-24 03:00:00+00', 52.41, '', 'completed', v_fc,
      '2025-06-23 07:53:00+00', '2025-06-24 03:09:30+00', 52.41, NULL, '',
      NULL, '2025-06-16 07:00:00+00', '2025-06-24 03:09:30+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 18343.5, '2025-06-24 03:09:30+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Irigare pivot' AND deleted_at IS NULL),
      NULL, NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 8 LIMIT 1),
      '2025-07-19 07:00:00+00', '2025-07-20 03:00:00+00', 52.41, '', 'completed', v_fc,
      '2025-07-19 07:49:00+00', '2025-07-20 01:39:35+00', 52.41, NULL, '',
      NULL, '2025-07-12 07:00:00+00', '2025-07-20 01:39:35+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 18343.5, '2025-07-20 01:39:35+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare porumb' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-CLAAS-LEXION780-01'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 9 LIMIT 1),
      '2025-10-01 07:00:00+00', '2025-10-02 04:50:15+00', 52.41, '', 'completed', v_fc,
      '2025-10-01 07:41:00+00', '2025-10-02 04:25:29+00', 52.41, 23.66, '',
      NULL, '2025-09-24 07:00:00+00', '2025-10-02 04:25:29+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 887.797, '2025-10-02 04:25:29+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Porumb', 'in', 579.89, '2025-10-03 00:25:29+00', NULL, v_fc, 'Recoltă Porumb - Plopi - Tarla Nord (Sezon 2024/2025)');
  UPDATE field_crops SET production_total = 579.89, harvest_recorded_quantity = 579.89, harvest_recorded_at = '2025-10-03 00:25:29+00', harvested_at = '2025-10-03', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2025-10-07', harvested_at = COALESCE(harvested_at, '2025-10-07'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Plopi - Tarla Sud (48.28 ha): Grâu ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2024/2025';
  SELECT id INTO v_field FROM fields WHERE name = 'Plopi - Tarla Sud' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Plopi - Tarla Sud lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Grâu'), 48.28, '2024-10-08', 5.8, '', 'closed', '2025-07-15', '2024-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 11 LIMIT 1),
      '2024-08-28 07:00:00+00', '2024-08-29 13:10:30+00', 48.28, '', 'completed', v_fc,
      '2024-08-28 07:46:00+00', '2024-08-29 12:59:44+00', 48.28, 29.98, '',
      NULL, '2024-08-21 07:00:00+00', '2024-08-29 12:59:44+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 962.95, '2024-08-29 12:59:44+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 12 LIMIT 1),
      '2024-09-29 07:00:00+00', '2024-09-30 13:10:30+00', 48.28, '', 'completed', v_fc,
      '2024-09-29 07:33:00+00', '2024-09-30 12:31:08+00', 45.91, 27.96, '',
      NULL, '2024-09-22 07:00:00+00', '2024-09-30 12:31:08+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 465.818, '2024-09-30 12:31:08+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 8187.504, '2024-09-30 12:31:08+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat grau standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2024-10-10 07:00:00+00', '2024-10-10 22:05:15+00', 48.28, '', 'completed', v_fc,
      '2024-10-10 07:40:00+00', '2024-10-10 21:26:20+00', 48.28, 15.95, '',
      NULL, '2024-10-03 07:00:00+00', '2024-10-10 21:26:20+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 467.365, '2024-10-10 21:26:20+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta grau lot A', 'out', 10806.937, '2024-10-10 21:26:20+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare faziala grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 1 LIMIT 1),
      '2025-03-12 07:00:00+00', '2025-03-12 13:53:49+00', 48.28, '', 'completed', v_fc,
      '2025-03-12 07:23:00+00', '2025-03-12 13:26:06+00', 48.28, 6.93, '',
      NULL, '2025-03-05 07:00:00+00', '2025-03-12 13:26:06+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 309.853, '2025-03-12 13:26:06+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'out', 7783.424, '2025-03-12 13:26:06+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2025-04-20 07:00:00+00', '2025-04-20 12:21:52+00', 48.28, '', 'completed', v_fc,
      '2025-04-20 07:20:00+00', '2025-04-20 16:41:13+00', 48.28, 5.17, 'Întârziere: teren umed',
      NULL, '2025-04-13 07:00:00+00', '2025-04-20 16:41:13+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 233.821, '2025-04-20 16:41:13+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 107.491, '2025-04-20 16:41:13+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 11.884, '2025-04-20 16:41:13+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2025-05-09 07:00:00+00', '2025-05-09 12:21:52+00', 48.28, '', 'canceled', v_fc,
      NULL, NULL, NULL, NULL, 'Anulată: ploi persistente în fereastra de tratament',
      NULL, '2025-05-02 07:00:00+00', '2025-05-09 07:00:00+00')
  RETURNING id INTO v_op;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-CLAAS-LEXION780-01'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2025-07-04 07:00:00+00', '2025-07-05 03:07:00+00', 48.28, '', 'completed', v_fc,
      '2025-07-04 07:57:00+00', '2025-07-05 05:38:09+00', 48.28, 19.93, 'Întârziere: defecțiune minoră',
      NULL, '2025-06-27 07:00:00+00', '2025-07-05 05:38:09+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 712.585, '2025-07-05 05:38:09+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'in', 283.17, '2025-07-06 01:38:09+00', NULL, v_fc, 'Recoltă Grâu - Plopi - Tarla Sud (Sezon 2024/2025)');
  UPDATE field_crops SET production_total = 283.17, harvest_recorded_quantity = 283.17, harvest_recorded_at = '2025-07-06 01:38:09+00', harvested_at = '2025-07-06', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Balotat paie' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 5 LIMIT 1),
      '2025-07-13 07:00:00+00', '2025-07-14 03:07:00+00', 48.28, '', 'completed', v_fc,
      '2025-07-13 07:13:00+00', '2025-07-14 03:01:14+00', 48.28, 21.53, '',
      NULL, '2025-07-06 07:00:00+00', '2025-07-14 03:01:14+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 146.081, '2025-07-14 03:01:14+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Paie baloti', 'in', 121.11, '2025-07-14 09:01:14+00', NULL, v_fc, 'Paie Grâu - Plopi - Tarla Sud, balotat');
  UPDATE field_crops SET status = 'closed', closed_at = '2025-07-19', harvested_at = COALESCE(harvested_at, '2025-07-19'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Tartaul - Tarla Mica (9.84 ha): Rapiță ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2024/2025';
  SELECT id INTO v_field FROM fields WHERE name = 'Tartaul - Tarla Mica' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Tartaul - Tarla Mica lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Rapiță'), 9.84, '2024-08-28', 3.2, '', 'closed', '2025-07-15', '2024-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2024-08-06 07:00:00+00', '2024-08-06 13:09:00+00', 9.84, '', 'completed', v_fc,
      '2024-08-06 06:58:00+00', '2024-08-06 12:07:27+00', 9.84, 5.69, '',
      NULL, '2024-07-30 07:00:00+00', '2024-08-06 12:07:27+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 188.912, '2024-08-06 12:07:27+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 8 LIMIT 1),
      '2024-08-18 07:00:00+00', '2024-08-18 13:09:00+00', 9.84, '', 'completed', v_fc,
      '2024-08-18 07:13:00+00', '2024-08-18 12:42:38+00', 9.84, 6.22, '',
      NULL, '2024-08-11 07:00:00+00', '2024-08-18 12:42:38+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 98.65, '2024-08-18 12:42:38+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 1785.03, '2024-08-18 12:42:38+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat rapita standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 9 LIMIT 1),
      '2024-08-29 07:00:00+00', '2024-08-29 10:04:30+00', 9.84, '', 'completed', v_fc,
      '2024-08-29 07:57:00+00', '2024-08-29 08:36:22+00', 9.84, 3.13, '',
      NULL, '2024-08-22 07:00:00+00', '2024-08-29 08:36:22+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 74.355, '2024-08-29 08:36:22+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta rapita hibrid', 'out', 38.641, '2024-08-29 08:36:22+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare faziala grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 10 LIMIT 1),
      '2025-03-05 07:00:00+00', '2025-03-05 08:30:00+00', 9.84, '', 'completed', v_fc,
      '2025-03-05 07:20:00+00', '2025-03-05 08:20:00+00', 9.84, 1.62, '',
      NULL, '2025-02-26 07:00:00+00', '2025-03-05 08:20:00+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 59.643, '2025-03-05 08:20:00+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'out', 1601.489, '2025-03-05 08:20:00+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid rapita' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 11 LIMIT 1),
      '2025-04-14 07:00:00+00', '2025-04-14 08:30:00+00', 9.84, '', 'completed', v_fc,
      '2025-04-14 06:40:00+00', '2025-04-14 12:11:48+00', 9.84, 1.39, 'Întârziere: teren umed',
      NULL, '2025-04-07 07:00:00+00', '2025-04-14 12:11:48+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 51.553, '2025-04-14 12:11:48+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Fungicid lot rapita', 'out', 9.972, '2025-04-14 12:11:48+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 2.109, '2025-04-14 12:11:48+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare rapita' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-CLAAS-LEXION780-01'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 12 LIMIT 1),
      '2025-06-26 07:00:00+00', '2025-06-26 11:06:00+00', 9.84, '', 'completed', v_fc,
      '2025-06-26 07:51:00+00', '2025-06-26 10:40:07+00', 9.84, 4.09, '',
      NULL, '2025-06-19 07:00:00+00', '2025-06-26 10:40:07+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 128.541, '2025-06-26 10:40:07+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Rapiță', 'in', 34.53, '2025-06-27 06:40:07+00', NULL, v_fc, 'Recoltă Rapiță - Tartaul - Tarla Mica (Sezon 2024/2025)');
  UPDATE field_crops SET production_total = 34.53, harvest_recorded_quantity = 34.53, harvest_recorded_at = '2025-06-27 06:40:07+00', harvested_at = '2025-06-27', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2025-07-01', harvested_at = COALESCE(harvested_at, '2025-07-01'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Lingura - Teren agricol Nord (7.98 ha): Grâu ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2024/2025';
  SELECT id INTO v_field FROM fields WHERE name = 'Lingura - Teren agricol Nord' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Lingura - Teren agricol Nord lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Grâu'), 7.98, '2024-10-08', 5.8, '', 'closed', '2025-07-15', '2024-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 1 LIMIT 1),
      '2024-08-26 07:00:00+00', '2024-08-26 11:59:15+00', 7.98, '', 'completed', v_fc,
      '2024-08-26 07:03:00+00', '2024-08-26 11:47:51+00', 7.81, 5.01, '',
      NULL, '2024-08-19 07:00:00+00', '2024-08-26 11:47:51+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 155.843, '2024-08-26 11:47:51+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2024-10-01 07:00:00+00', '2024-10-01 11:59:15+00', 7.98, '', 'completed', v_fc,
      '2024-10-01 07:03:00+00', '2024-10-01 16:50:20+00', 7.98, 5.37, 'Întârziere: ploaie',
      NULL, '2024-09-24 07:00:00+00', '2024-10-01 16:50:20+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 83.574, '2024-10-01 16:50:20+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 1421.021, '2024-10-01 16:50:20+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat grau standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2024-10-10 07:00:00+00', '2024-10-10 09:29:37+00', 7.98, '', 'completed', v_fc,
      '2024-10-10 06:57:00+00', '2024-10-10 08:20:49+00', 7.77, 2.72, '',
      NULL, '2024-10-03 07:00:00+00', '2024-10-10 08:20:49+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 76.01, '2024-10-10 08:20:49+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta grau lot A', 'out', 1701.405, '2024-10-10 08:20:49+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare faziala grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2025-03-15 07:00:00+00', '2025-03-15 08:30:00+00', 7.98, '', 'completed', v_fc,
      '2025-03-15 07:34:00+00', '2025-03-15 08:34:00+00', 7.98, 1.61, '',
      NULL, '2025-03-08 07:00:00+00', '2025-03-15 08:34:00+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 47.912, '2025-03-15 08:34:00+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'out', 1273.491, '2025-03-15 08:34:00+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 5 LIMIT 1),
      '2025-04-17 07:00:00+00', '2025-04-17 08:30:00+00', 7.98, '', 'completed', v_fc,
      '2025-04-17 07:40:00+00', '2025-04-17 08:58:00+00', 7.98, 1.6, '',
      NULL, '2025-04-10 07:00:00+00', '2025-04-17 08:58:00+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 39.585, '2025-04-17 08:58:00+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 17.228, '2025-04-17 08:58:00+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 2.025, '2025-04-17 08:58:00+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2025-05-11 07:00:00+00', '2025-05-11 08:30:00+00', 7.98, '', 'completed', v_fc,
      '2025-05-11 07:03:00+00', '2025-05-11 08:15:15+00', 7.86, 1.41, '',
      NULL, '2025-05-04 07:00:00+00', '2025-05-11 08:15:15+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 41.038, '2025-05-11 08:15:15+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Fungicid lot rapita', 'out', 6.445, '2025-05-11 08:15:15+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 1.735, '2025-05-11 08:15:15+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-CLAAS-LEXION780-01'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2025-07-07 07:00:00+00', '2025-07-07 10:19:30+00', 7.98, '', 'completed', v_fc,
      '2025-07-07 07:27:00+00', '2025-07-07 10:15:29+00', 7.98, 3.33, '',
      NULL, '2025-06-30 07:00:00+00', '2025-07-07 10:15:29+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 120.832, '2025-07-07 10:15:29+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'in', 47.37, '2025-07-08 06:15:29+00', NULL, v_fc, 'Recoltă Grâu - Lingura - Teren agricol Nord (Sezon 2024/2025)');
  UPDATE field_crops SET production_total = 47.37, harvest_recorded_quantity = 47.37, harvest_recorded_at = '2025-07-08 06:15:29+00', harvested_at = '2025-07-08', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Balotat paie' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 8 LIMIT 1),
      '2025-07-15 07:00:00+00', '2025-07-15 10:19:30+00', 7.98, '', 'completed', v_fc,
      '2025-07-15 07:07:00+00', '2025-07-15 09:38:39+00', 7.98, 3.29, '',
      NULL, '2025-07-08 07:00:00+00', '2025-07-15 09:38:39+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 24.187, '2025-07-15 09:38:39+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Paie baloti', 'in', 17.3, '2025-07-15 15:38:39+00', NULL, v_fc, 'Paie Grâu - Lingura - Teren agricol Nord, balotat');
  UPDATE field_crops SET status = 'closed', closed_at = '2025-07-20', harvested_at = COALESCE(harvested_at, '2025-07-20'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Cirpesti - Lot Nord (2.38 ha): Porumb ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2024/2025';
  SELECT id INTO v_field FROM fields WHERE name = 'Cirpesti - Lot Nord' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Cirpesti - Lot Nord lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Porumb'), 2.38, '2025-04-22', 9.0, '', 'closed', '2025-07-15', '2024-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 10 LIMIT 1),
      '2024-10-27 07:00:00+00', '2024-10-27 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2024-10-27 07:14:00+00', '2024-10-27 08:41:31+00', 2.38, 1.6, '',
      NULL, '2024-10-20 07:00:00+00', '2024-10-27 08:41:31+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 44.574, '2024-10-27 08:41:31+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 11 LIMIT 1),
      '2025-04-12 07:00:00+00', '2025-04-12 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2025-04-12 07:38:00+00', '2025-04-12 08:38:00+00', 2.38, 1.63, '',
      NULL, '2025-04-05 07:00:00+00', '2025-04-12 08:38:00+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 24.494, '2025-04-12 08:38:00+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 441.543, '2025-04-12 08:38:00+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat porumb standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 12 LIMIT 1),
      '2025-04-21 07:00:00+00', '2025-04-21 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2025-04-21 07:28:00+00', '2025-04-21 07:34:25+00', 2.38, 1.59, '',
      NULL, '2025-04-14 07:00:00+00', '2025-04-21 07:34:25+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 20.584, '2025-04-21 07:34:25+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta porumb lot B', 'out', 2.097, '2025-04-21 07:34:25+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2025-05-23 07:00:00+00', '2025-05-23 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2025-05-23 06:47:00+00', '2025-05-23 07:57:37+00', 2.38, 1.5, '',
      NULL, '2025-05-16 07:00:00+00', '2025-05-23 07:57:37+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 11.624, '2025-05-23 07:57:37+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 5.202, '2025-05-23 07:57:37+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 0.6, '2025-05-23 07:57:37+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare porumb' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-JOHNDEERE-S780-02'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 1 LIMIT 1),
      '2025-10-02 07:00:00+00', '2025-10-02 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2025-10-02 06:40:00+00', '2025-10-02 09:54:00+00', 2.38, 1.46, 'Întârziere: teren umed',
      NULL, '2025-09-25 07:00:00+00', '2025-10-02 09:54:00+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 39.114, '2025-10-02 09:54:00+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Porumb', 'in', 22.12, '2025-10-03 05:54:00+00', NULL, v_fc, 'Recoltă Porumb - Cirpesti - Lot Nord (Sezon 2024/2025)');
  UPDATE field_crops SET production_total = 22.12, harvest_recorded_quantity = 22.12, harvest_recorded_at = '2025-10-03 05:54:00+00', harvested_at = '2025-10-03', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2025-10-07', harvested_at = COALESCE(harvested_at, '2025-10-07'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Cirpesti - Lot Central (1.91 ha): Grâu ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2024/2025';
  SELECT id INTO v_field FROM fields WHERE name = 'Cirpesti - Lot Central' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Cirpesti - Lot Central lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Grâu'), 1.91, '2024-10-08', 5.8, '', 'closed', '2025-07-15', '2024-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2024-08-26 07:00:00+00', '2024-08-26 08:30:00+00', 1.91, '', 'completed', v_fc,
      '2024-08-26 07:50:00+00', '2024-08-26 08:50:00+00', 1.91, 1.44, '',
      NULL, '2024-08-19 07:00:00+00', '2024-08-26 08:50:00+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 37.151, '2024-08-26 08:50:00+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2024-09-27 07:00:00+00', '2024-09-27 08:30:00+00', 1.91, '', 'completed', v_fc,
      '2024-09-27 07:58:00+00', '2024-09-27 08:58:00+00', 1.91, 1.47, '',
      NULL, '2024-09-20 07:00:00+00', '2024-09-27 08:58:00+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 20.622, '2024-09-27 08:58:00+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 354.546, '2024-09-27 08:58:00+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat grau standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 5 LIMIT 1),
      '2024-10-06 07:00:00+00', '2024-10-06 08:30:00+00', 1.91, '', 'completed', v_fc,
      '2024-10-06 06:52:00+00', '2024-10-06 13:52:48+00', 1.84, 1.41, 'Întârziere: ploaie',
      NULL, '2024-09-29 07:00:00+00', '2024-10-06 13:52:48+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 17.687, '2024-10-06 13:52:48+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta grau lot A', 'out', 406.717, '2024-10-06 13:52:48+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare faziala grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2025-03-14 07:00:00+00', '2025-03-14 08:30:00+00', 1.91, '', 'completed', v_fc,
      '2025-03-14 07:49:00+00', '2025-03-14 08:49:00+00', 1.91, 1.5, '',
      NULL, '2025-03-07 07:00:00+00', '2025-03-14 08:49:00+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 12.699, '2025-03-14 08:49:00+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'out', 310.6, '2025-03-14 08:49:00+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2025-04-18 07:00:00+00', '2025-04-18 08:30:00+00', 1.91, '', 'completed', v_fc,
      '2025-04-18 07:40:00+00', '2025-04-18 08:48:14+00', 1.84, 1.62, '',
      NULL, '2025-04-11 07:00:00+00', '2025-04-18 08:48:14+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 9.153, '2025-04-18 08:48:14+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 4.157, '2025-04-18 08:48:14+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 0.459, '2025-04-18 08:48:14+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 8 LIMIT 1),
      '2025-05-13 07:00:00+00', '2025-05-13 08:30:00+00', 1.91, '', 'completed', v_fc,
      '2025-05-13 07:12:00+00', '2025-05-13 08:18:53+00', 1.91, 1.48, '',
      NULL, '2025-05-06 07:00:00+00', '2025-05-13 08:18:53+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 9.901, '2025-05-13 08:18:53+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Fungicid lot rapita', 'out', 1.495, '2025-05-13 08:18:53+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 0.411, '2025-05-13 08:18:53+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-JOHNDEERE-S780-02'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 9 LIMIT 1),
      '2025-07-07 07:00:00+00', '2025-07-07 08:30:00+00', 1.91, '', 'completed', v_fc,
      '2025-07-07 06:44:00+00', '2025-07-07 07:27:54+00', 1.91, 1.62, '',
      NULL, '2025-06-30 07:00:00+00', '2025-07-07 07:27:54+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 29.786, '2025-07-07 07:27:54+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'in', 11.84, '2025-07-08 03:27:54+00', NULL, v_fc, 'Recoltă Grâu - Cirpesti - Lot Central (Sezon 2024/2025)');
  UPDATE field_crops SET production_total = 11.84, harvest_recorded_quantity = 11.84, harvest_recorded_at = '2025-07-08 03:27:54+00', harvested_at = '2025-07-08', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Balotat paie' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 10 LIMIT 1),
      '2025-07-10 07:00:00+00', '2025-07-10 08:30:00+00', 1.91, '', 'completed', v_fc,
      '2025-07-10 08:00:00+00', '2025-07-10 09:00:00+00', 1.91, 1.62, '',
      NULL, '2025-07-03 07:00:00+00', '2025-07-10 09:00:00+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 6.373, '2025-07-10 09:00:00+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Paie baloti', 'in', 4.21, '2025-07-10 15:00:00+00', NULL, v_fc, 'Paie Grâu - Cirpesti - Lot Central, balotat');
  UPDATE field_crops SET status = 'closed', closed_at = '2025-07-15', harvested_at = COALESCE(harvested_at, '2025-07-15'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Cirpesti - Lot Sud (1.56 ha): Grâu ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2024/2025';
  SELECT id INTO v_field FROM fields WHERE name = 'Cirpesti - Lot Sud' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Cirpesti - Lot Sud lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Grâu'), 1.56, '2024-10-08', 5.8, '', 'closed', '2025-07-15', '2024-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 12 LIMIT 1),
      '2024-08-25 07:00:00+00', '2024-08-25 08:30:00+00', 1.56, '', 'completed', v_fc,
      '2024-08-25 07:00:00+00', '2024-08-25 07:29:30+00', 1.56, 1.62, '',
      NULL, '2024-08-18 07:00:00+00', '2024-08-25 07:29:30+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 29.153, '2024-08-25 07:29:30+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2024-09-28 07:00:00+00', '2024-09-28 08:30:00+00', 1.56, '', 'completed', v_fc,
      '2024-09-28 07:47:00+00', '2024-09-28 11:08:34+00', 1.56, 1.42, 'Întârziere: ploaie',
      NULL, '2024-09-21 07:00:00+00', '2024-09-28 11:08:34+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 16.141, '2024-09-28 11:08:34+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 286.272, '2024-09-28 11:08:34+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat grau standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 1 LIMIT 1),
      '2024-10-08 07:00:00+00', '2024-10-08 08:30:00+00', 1.56, '', 'completed', v_fc,
      '2024-10-08 06:54:00+00', '2024-10-08 08:17:10+00', 1.56, 1.52, '',
      NULL, '2024-10-01 07:00:00+00', '2024-10-08 08:17:10+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 14.347, '2024-10-08 08:17:10+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta grau lot A', 'out', 354.32, '2024-10-08 08:17:10+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare faziala grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2025-03-15 07:00:00+00', '2025-03-15 08:30:00+00', 1.56, '', 'completed', v_fc,
      '2025-03-15 07:18:00+00', '2025-03-15 09:44:08+00', 1.56, 1.5, 'Întârziere: teren umed',
      NULL, '2025-03-08 07:00:00+00', '2025-03-15 09:44:08+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 9.515, '2025-03-15 09:44:08+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'out', 251.517, '2025-03-15 09:44:08+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2025-04-17 07:00:00+00', '2025-04-17 08:30:00+00', 1.56, '', 'completed', v_fc,
      '2025-04-17 07:17:00+00', '2025-04-17 07:36:41+00', 1.54, 1.4, '',
      NULL, '2025-04-10 07:00:00+00', '2025-04-17 07:36:41+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 6.989, '2025-04-17 07:36:41+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 3.36, '2025-04-17 07:36:41+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 0.384, '2025-04-17 07:36:41+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2025-05-09 07:00:00+00', '2025-05-09 08:30:00+00', 1.56, '', 'canceled', v_fc,
      NULL, NULL, NULL, NULL, 'Anulată: ploi persistente în fereastra de tratament',
      NULL, '2025-05-02 07:00:00+00', '2025-05-09 07:00:00+00')
  RETURNING id INTO v_op;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-CLAAS-LEXION780-01'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 5 LIMIT 1),
      '2025-07-04 07:00:00+00', '2025-07-04 08:30:00+00', 1.56, '', 'completed', v_fc,
      '2025-07-04 07:21:00+00', '2025-07-04 07:59:08+00', 1.48, 1.5, '',
      NULL, '2025-06-27 07:00:00+00', '2025-07-04 07:59:08+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 21.376, '2025-07-04 07:59:08+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'in', 9.21, '2025-07-05 03:59:08+00', NULL, v_fc, 'Recoltă Grâu - Cirpesti - Lot Sud (Sezon 2024/2025)');
  UPDATE field_crops SET production_total = 9.21, harvest_recorded_quantity = 9.21, harvest_recorded_at = '2025-07-05 03:59:08+00', harvested_at = '2025-07-05', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Balotat paie' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2025-07-14 07:00:00+00', '2025-07-14 08:30:00+00', 1.56, '', 'completed', v_fc,
      '2025-07-14 07:03:00+00', '2025-07-14 08:22:56+00', 1.56, 1.41, '',
      NULL, '2025-07-07 07:00:00+00', '2025-07-14 08:22:56+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 5.173, '2025-07-14 08:22:56+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Paie baloti', 'in', 3.53, '2025-07-14 14:22:56+00', NULL, v_fc, 'Paie Grâu - Cirpesti - Lot Sud, balotat');
  UPDATE field_crops SET status = 'closed', closed_at = '2025-07-19', harvested_at = COALESCE(harvested_at, '2025-07-19'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ─── Vânzări din stoc, după recoltă (ieșiri fără lucrare) ───
DO $sale$ BEGIN
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'out', 155.83, '2025-06-23 20:29:13+00', NULL, NULL, '[Sezon 2024/2025] Vânzare Lucernă — contract, Gotesti - Tarla Vest');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'out', 40.26, '2025-08-10 09:15:29+00', NULL, NULL, '[Sezon 2024/2025] Vânzare Grâu — contract, Lingura - Teren agricol Nord');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'out', 178.1, '2025-08-11 20:30:22+00', NULL, NULL, '[Sezon 2024/2025] Vânzare Lucernă — contract, Gotesti - Tarla Vest');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'out', 10.06, '2025-08-16 06:27:54+00', NULL, NULL, '[Sezon 2024/2025] Vânzare Grâu — contract, Cirpesti - Lot Central');
  PERFORM pg_temp.seed_move('Paie baloti', 'out', 10.38, '2025-08-17 18:38:39+00', NULL, NULL, '[Sezon 2024/2025] Vânzare paie — Lingura - Teren agricol Nord');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'out', 500.09, '2025-08-20 19:00:09+00', NULL, NULL, '[Sezon 2024/2025] Vânzare Grâu — contract, Gotesti - Tarla Est');
  PERFORM pg_temp.seed_move('Recoltă Rapiță', 'out', 29.35, '2025-08-22 09:40:07+00', NULL, NULL, '[Sezon 2024/2025] Vânzare Rapiță — contract, Tartaul - Tarla Mica');
  PERFORM pg_temp.seed_move('Paie baloti', 'out', 140.17, '2025-08-26 05:32:09+00', NULL, NULL, '[Sezon 2024/2025] Vânzare paie — Gotesti - Tarla Est');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'out', 667.85, '2025-08-26 12:21:23+00', NULL, NULL, '[Sezon 2024/2025] Vânzare Grâu — contract, Gotesti - Tarla Mare');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'out', 240.69, '2025-08-30 04:38:09+00', NULL, NULL, '[Sezon 2024/2025] Vânzare Grâu — contract, Plopi - Tarla Sud');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'out', 7.83, '2025-08-30 06:59:08+00', NULL, NULL, '[Sezon 2024/2025] Vânzare Grâu — contract, Cirpesti - Lot Sud');
  PERFORM pg_temp.seed_move('Paie baloti', 'out', 72.67, '2025-09-03 12:01:14+00', NULL, NULL, '[Sezon 2024/2025] Vânzare paie — Plopi - Tarla Sud');
  PERFORM pg_temp.seed_move('Paie baloti', 'out', 179.75, '2025-09-03 21:16:47+00', NULL, NULL, '[Sezon 2024/2025] Vânzare paie — Gotesti - Tarla Mare');
  PERFORM pg_temp.seed_move('Paie baloti', 'out', 2.53, '2025-09-05 18:00:00+00', NULL, NULL, '[Sezon 2024/2025] Vânzare paie — Cirpesti - Lot Central');
  PERFORM pg_temp.seed_move('Paie baloti', 'out', 2.12, '2025-09-09 17:22:56+00', NULL, NULL, '[Sezon 2024/2025] Vânzare paie — Cirpesti - Lot Sud');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'out', 144.7, '2025-09-17 19:17:10+00', NULL, NULL, '[Sezon 2024/2025] Vânzare Lucernă — contract, Gotesti - Tarla Vest');
  PERFORM pg_temp.seed_move('Recoltă Floarea-soarelui', 'out', 467.1, '2025-10-16 21:18:47+00', NULL, NULL, '[Sezon 2024/2025] Vânzare Floarea-soarelui — contract, Plopi - Tarla Mare');
  PERFORM pg_temp.seed_move('Recoltă Porumb', 'out', 18.8, '2025-11-04 08:54:00+00', NULL, NULL, '[Sezon 2024/2025] Vânzare Porumb — contract, Cirpesti - Lot Nord');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'out', 105.74, '2025-11-07 01:18:24+00', NULL, NULL, '[Sezon 2024/2025] Vânzare Lucernă — contract, Gotesti - Tarla Vest');
  PERFORM pg_temp.seed_move('Recoltă Porumb', 'out', 705.52, '2025-11-09 15:43:00+00', NULL, NULL, '[Sezon 2024/2025] Vânzare Porumb — contract, Tartaul - Tarla Mare');
  PERFORM pg_temp.seed_move('Recoltă Soia', 'out', 141.9, '2025-11-24 03:54:40+00', NULL, NULL, '[Sezon 2024/2025] Vânzare Soia — contract, Tartaul - Tarla Centrala');
  PERFORM pg_temp.seed_move('Recoltă Porumb', 'out', 492.91, '2025-11-30 03:25:29+00', NULL, NULL, '[Sezon 2024/2025] Vânzare Porumb — contract, Plopi - Tarla Nord');
END $sale$;

-- ─── Istoricul stocului în ordine cronologică (resulting_quantity ca sold curent la acea dată) ───
UPDATE stock_movements sm SET resulting_quantity = t.run FROM (SELECT id, SUM(quantity_delta) OVER (PARTITION BY resource_id ORDER BY created_at, id) AS run FROM stock_movements) t WHERE t.id = sm.id;

COMMIT;
