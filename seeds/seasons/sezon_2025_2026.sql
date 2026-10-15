-- Seed sezon Sezon 2025/2026 — generat de seeds/seasons/generate.py; nu edita manual, modifică generatorul.
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
    SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2025/2026';
    IF v_season IS NOT NULL THEN
        DELETE FROM stock_movements WHERE field_operation_id IN (SELECT id FROM field_operations WHERE field_crop_id IN (SELECT id FROM field_crops WHERE season_id = v_season));
        DELETE FROM stock_movements WHERE field_crop_id IN (SELECT id FROM field_crops WHERE season_id = v_season);
        DELETE FROM field_operations WHERE field_crop_id IN (SELECT id FROM field_crops WHERE season_id = v_season);
        DELETE FROM field_crops WHERE season_id = v_season;
        DELETE FROM seasons WHERE id = v_season;
    END IF;
    -- aprovizionările și lucrările fără cultură ale sezonului sunt etichetate în notes
    DELETE FROM stock_movements WHERE field_operation_id IN (SELECT id FROM field_operations WHERE notes LIKE '[Sezon 2025/2026]%');
    DELETE FROM field_operations WHERE notes LIKE '[Sezon 2025/2026]%';
    DELETE FROM stock_movements WHERE notes LIKE '[Sezon 2025/2026]%';
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
UPDATE seasons SET is_active = FALSE WHERE is_active;
INSERT INTO seasons (name, start_date, end_date, is_active, notes) VALUES ('Sezon 2025/2026', '2025-08-01', NULL, true, 'În curs. Grâul de pe Tartaul - Tarla Mica a fost compromis de îngheț și reînsămânțat cu porumb.');

-- ─── Aprovizionări (dimensionate după consumul sezonului, +5%) ───
DO $sup$ BEGIN
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'in', 55000, '2026-04-09 07:16:50+00', NULL, NULL, '[Sezon 2025/2026] Alimentare bazin pivot');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'in', 1450, '2026-04-09 07:16:50+00', NULL, NULL, '[Sezon 2025/2026] Recepție furnizor');
  PERFORM pg_temp.seed_move('Fungicid lot rapita', 'in', 255, '2026-04-09 20:30:51+00', NULL, NULL, '[Sezon 2025/2026] Recepție furnizor');
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'in', 24500, '2025-08-03 14:55:34+00', NULL, NULL, '[Sezon 2025/2026] Recepție furnizor — tranșa 1');
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'in', 24500, '2025-12-31 14:55:34+00', NULL, NULL, '[Sezon 2025/2026] Recepție furnizor — tranșa 2');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'in', 144150, '2025-08-19 16:18:35+00', NULL, NULL, '[Sezon 2025/2026] Recepție furnizor');
  PERFORM pg_temp.seed_move('Samanta floarea-soarelui hibrid', 'in', 330, '2026-04-12 15:32:15+00', NULL, NULL, '[Sezon 2025/2026] Recepție furnizor');
  PERFORM pg_temp.seed_move('Samanta grau lot A', 'in', 34950, '2025-09-30 08:11:15+00', NULL, NULL, '[Sezon 2025/2026] Recepție furnizor');
  PERFORM pg_temp.seed_move('Samanta orz', 'in', 1550, '2025-09-25 09:19:51+00', NULL, NULL, '[Sezon 2025/2026] Recepție furnizor');
  PERFORM pg_temp.seed_move('Samanta porumb lot B', 'in', 235, '2026-04-14 16:01:48+00', NULL, NULL, '[Sezon 2025/2026] Recepție furnizor');
  PERFORM pg_temp.seed_move('Samanta rapita hibrid', 'in', 550, '2025-08-24 22:21:24+00', NULL, NULL, '[Sezon 2025/2026] Recepție furnizor');
  PERFORM pg_temp.seed_move('Samanta sfecla de zahar', 'in', 140, '2026-03-27 23:46:39+00', NULL, NULL, '[Sezon 2025/2026] Recepție furnizor');
  PERFORM pg_temp.seed_move('Samanta soia', 'in', 205, '2026-04-22 08:57:50+00', NULL, NULL, '[Sezon 2025/2026] Recepție furnizor');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'in', 46450, '2026-02-26 01:13:21+00', NULL, NULL, '[Sezon 2025/2026] Recepție furnizor');
END $sup$;

-- ── Plopi - Tarla Mare (210.25 ha): Porumb ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2025/2026';
  SELECT id INTO v_field FROM fields WHERE name = 'Plopi - Tarla Mare' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Plopi - Tarla Mare lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Porumb'), 210.25, '2026-04-22', 9.0, '', 'open', NULL, '2025-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2025-10-25 07:00:00+00', '2025-10-30 18:24:22+00', 210.25, '', 'completed', v_fc,
      '2025-10-25 07:05:00+00', '2025-10-30 18:22:19+00', 210.25, 139.71, '',
      NULL, '2025-10-18 07:00:00+00', '2025-10-30 18:22:19+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 4073.596, '2025-10-30 18:22:19+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 1 LIMIT 1),
      '2026-04-08 07:00:00+00', '2026-04-13 18:24:22+00', 210.25, '', 'completed', v_fc,
      '2026-04-08 07:40:00+00', '2026-04-13 18:47:23+00', 208.14, 140.44, '',
      NULL, '2026-04-01 07:00:00+00', '2026-04-13 18:47:23+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 2218.064, '2026-04-13 18:47:23+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 37305.787, '2026-04-13 18:47:23+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat porumb standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2026-04-21 07:00:00+00', '2026-04-24 00:42:11+00', 210.25, '', 'completed', v_fc,
      '2026-04-21 06:55:00+00', '2026-04-24 00:02:38+00', 210.25, 60.59, '',
      NULL, '2026-04-14 07:00:00+00', '2026-04-24 00:02:38+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1960.415, '2026-04-24 00:02:38+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta porumb lot B', 'out', 174.805, '2026-04-24 00:02:38+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2026-05-19 07:00:00+00', '2026-05-20 06:21:40+00', 210.25, '', 'completed', v_fc,
      '2026-05-19 06:44:00+00', '2026-05-20 06:17:36+00', 210.25, 25.27, '',
      NULL, '2026-05-12 07:00:00+00', '2026-05-20 06:17:36+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1045.199, '2026-05-20 06:17:36+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 479.326, '2026-05-20 06:17:36+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 53.904, '2026-05-20 06:17:36+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare porumb' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-CLAAS-LEXION780-01'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2026-10-05 07:00:00+00', '2026-10-08 22:36:15+00', 210.25, '', 'completed', v_fc,
      '2026-10-05 07:56:00+00', '2026-10-08 23:04:05+00', 210.25, 81.29, '',
      NULL, '2026-09-28 07:00:00+00', '2026-10-08 23:04:05+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 3460.189, '2026-10-08 23:04:05+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Porumb', 'in', 1736.92, '2026-10-09 19:04:05+00', NULL, v_fc, 'Recoltă Porumb - Plopi - Tarla Mare (Sezon 2025/2026)');
  UPDATE field_crops SET production_total = 1736.92, harvest_recorded_quantity = 1736.92, harvest_recorded_at = '2026-10-09 19:04:05+00', harvested_at = '2026-10-09', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2026-10-13', harvested_at = COALESCE(harvested_at, '2026-10-13'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Gotesti - Tarla Mare (130.03 ha): Rapiță ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2025/2026';
  SELECT id INTO v_field FROM fields WHERE name = 'Gotesti - Tarla Mare' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Gotesti - Tarla Mare lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Rapiță'), 130.03, '2025-08-28', 3.2, '', 'open', NULL, '2025-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2025-08-06 07:00:00+00', '2025-08-09 16:16:07+00', 130.03, '', 'completed', v_fc,
      '2025-08-06 07:31:00+00', '2025-08-09 14:55:34+00', 126.52, 80.23, '',
      NULL, '2025-07-30 07:00:00+00', '2025-08-09 14:55:34+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 2519.823, '2025-08-09 14:55:34+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2025-08-22 07:00:00+00', '2025-08-25 16:16:07+00', 130.03, '', 'completed', v_fc,
      '2025-08-22 06:49:00+00', '2025-08-25 16:18:35+00', 130.03, 83.05, '',
      NULL, '2025-08-15 07:00:00+00', '2025-08-25 16:18:35+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1349.857, '2025-08-25 16:18:35+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 23599.874, '2025-08-25 16:18:35+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat rapita standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 8 LIMIT 1),
      '2025-08-29 07:00:00+00', '2025-08-30 23:38:03+00', 130.03, '', 'completed', v_fc,
      '2025-08-29 07:54:00+00', '2025-08-30 22:21:24+00', 130.03, 39.04, '',
      NULL, '2025-08-22 07:00:00+00', '2025-08-30 22:21:24+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 974.963, '2025-08-30 22:21:24+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta rapita hibrid', 'out', 522.032, '2025-08-30 22:21:24+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare faziala grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 9 LIMIT 1),
      '2026-03-03 07:00:00+00', '2026-03-04 01:34:32+00', 130.03, '', 'completed', v_fc,
      '2026-03-03 07:32:00+00', '2026-03-04 01:13:21+00', 130.03, 19.74, '',
      NULL, '2026-02-24 07:00:00+00', '2026-03-04 01:13:21+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 862.907, '2026-03-04 01:13:21+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'out', 21048.641, '2026-03-04 01:13:21+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid rapita' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 10 LIMIT 1),
      '2026-04-15 07:00:00+00', '2026-04-15 21:26:52+00', 130.03, '', 'completed', v_fc,
      '2026-04-15 07:35:00+00', '2026-04-15 20:30:51+00', 130.03, 15.44, '',
      NULL, '2026-04-08 07:00:00+00', '2026-04-15 20:30:51+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 689.471, '2026-04-15 20:30:51+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Fungicid lot rapita', 'out', 127.848, '2026-04-15 20:30:51+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 28.322, '2026-04-15 20:30:51+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare rapita' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-JOHNDEERE-S780-02'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 11 LIMIT 1),
      '2026-07-01 07:00:00+00', '2026-07-03 13:10:45+00', 130.03, '', 'completed', v_fc,
      '2026-07-01 07:22:00+00', '2026-07-03 11:49:53+00', 128.96, 57.21, '',
      NULL, '2026-06-24 07:00:00+00', '2026-07-03 11:49:53+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1784.523, '2026-07-03 11:49:53+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Rapiță', 'in', 380.25, '2026-07-04 07:49:53+00', NULL, v_fc, 'Recoltă Rapiță - Gotesti - Tarla Mare (Sezon 2025/2026)');
  UPDATE field_crops SET production_total = 380.25, harvest_recorded_quantity = 380.25, harvest_recorded_at = '2026-07-04 07:49:53+00', harvested_at = '2026-07-04', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2026-07-08', harvested_at = COALESCE(harvested_at, '2026-07-08'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Gotesti - Tarla Est (92.34 ha): Floarea-soarelui ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2025/2026';
  SELECT id INTO v_field FROM fields WHERE name = 'Gotesti - Tarla Est' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Gotesti - Tarla Est lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Floarea-soarelui'), 92.34, '2026-04-14', 2.6, '', 'open', NULL, '2025-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2025-10-27 07:00:00+00', '2025-10-29 16:42:45+00', 92.34, '', 'completed', v_fc,
      '2025-10-27 07:01:00+00', '2025-10-29 15:29:24+00', 92.34, 59.19, '',
      NULL, '2025-10-20 07:00:00+00', '2025-10-29 15:29:24+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1859.036, '2025-10-29 15:29:24+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 1 LIMIT 1),
      '2026-04-03 07:00:00+00', '2026-04-05 16:42:45+00', 92.34, '', 'completed', v_fc,
      '2026-04-03 07:50:00+00', '2026-04-05 17:03:38+00', 92.34, 55.31, '',
      NULL, '2026-03-27 07:00:00+00', '2026-04-05 17:03:38+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1010.248, '2026-04-05 17:03:38+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 16574.257, '2026-04-05 17:03:38+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat floarea-soarelui standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2026-04-17 07:00:00+00', '2026-04-18 11:51:22+00', 92.34, '', 'completed', v_fc,
      '2026-04-17 07:08:00+00', '2026-04-18 15:32:15+00', 89.79, 29.07, 'Întârziere: ploaie',
      NULL, '2026-04-10 07:00:00+00', '2026-04-18 15:32:15+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 769.046, '2026-04-18 15:32:15+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta floarea-soarelui hibrid', 'out', 309.536, '2026-04-18 15:32:15+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2026-05-15 07:00:00+00', '2026-05-15 17:15:36+00', 92.34, '', 'completed', v_fc,
      '2026-05-15 07:45:00+00', '2026-05-15 17:02:59+00', 92.34, 9.95, '',
      NULL, '2026-05-08 07:00:00+00', '2026-05-15 17:02:59+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 458.428, '2026-05-15 17:02:59+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 201.76, '2026-05-15 17:02:59+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 24.004, '2026-05-15 17:02:59+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare floarea-soarelui' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-JOHNDEERE-S780-02'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2026-09-08 07:00:00+00', '2026-09-09 21:28:30+00', 92.34, '', 'completed', v_fc,
      '2026-09-08 06:54:00+00', '2026-09-09 21:07:41+00', 92.34, 40.6, '',
      NULL, '2026-09-01 07:00:00+00', '2026-09-09 21:07:41+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1175.05, '2026-09-09 21:07:41+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Floarea-soarelui', 'in', 224.99, '2026-09-10 17:07:41+00', NULL, v_fc, 'Recoltă Floarea-soarelui - Gotesti - Tarla Est (Sezon 2025/2026)');
  UPDATE field_crops SET production_total = 224.99, harvest_recorded_quantity = 224.99, harvest_recorded_at = '2026-09-10 17:07:41+00', harvested_at = '2026-09-10', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2026-09-14', harvested_at = COALESCE(harvested_at, '2026-09-14'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Tartaul - Tarla Mare (84.02 ha): Grâu ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2025/2026';
  SELECT id INTO v_field FROM fields WHERE name = 'Tartaul - Tarla Mare' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Tartaul - Tarla Mare lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Grâu'), 84.02, '2025-10-08', 5.8, '', 'open', NULL, '2025-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2025-08-26 07:00:00+00', '2025-08-28 11:30:45+00', 84.02, '', 'completed', v_fc,
      '2025-08-26 07:14:00+00', '2025-08-28 10:35:20+00', 84.02, 55.72, '',
      NULL, '2025-08-19 07:00:00+00', '2025-08-28 10:35:20+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1609.214, '2025-08-28 10:35:20+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2025-09-30 07:00:00+00', '2025-10-02 11:30:45+00', 84.02, '', 'completed', v_fc,
      '2025-09-30 06:51:00+00', '2025-10-02 16:09:55+00', 84.02, 54.72, 'Întârziere: ploaie',
      NULL, '2025-09-23 07:00:00+00', '2025-10-02 16:09:55+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 886.902, '2025-10-02 16:09:55+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 15293.626, '2025-10-02 16:09:55+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat grau standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 8 LIMIT 1),
      '2025-10-07 07:00:00+00', '2025-10-08 09:15:22+00', 84.02, '', 'completed', v_fc,
      '2025-10-07 06:48:00+00', '2025-10-08 09:37:39+00', 84.02, 27.59, '',
      NULL, '2025-09-30 07:00:00+00', '2025-10-08 09:37:39+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 826.538, '2025-10-08 09:37:39+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta grau lot A', 'out', 18579.839, '2025-10-08 09:37:39+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare faziala grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 9 LIMIT 1),
      '2026-03-15 07:00:00+00', '2026-03-15 19:00:10+00', 84.02, '', 'completed', v_fc,
      '2026-03-15 07:48:00+00', '2026-03-15 17:44:54+00', 84.02, 13.01, '',
      NULL, '2026-03-08 07:00:00+00', '2026-03-15 17:44:54+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 518.247, '2026-03-15 17:44:54+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'out', 13121.844, '2026-03-15 17:44:54+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 10 LIMIT 1),
      '2026-04-21 07:00:00+00', '2026-04-21 16:20:08+00', 84.02, '', 'completed', v_fc,
      '2026-04-21 07:24:00+00', '2026-04-21 15:11:39+00', 84.02, 8.65, '',
      NULL, '2026-04-14 07:00:00+00', '2026-04-21 15:11:39+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 389.738, '2026-04-21 15:11:39+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 180.983, '2026-04-21 15:11:39+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 20.91, '2026-04-21 15:11:39+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 11 LIMIT 1),
      '2026-05-12 07:00:00+00', '2026-05-12 16:20:08+00', 84.02, '', 'completed', v_fc,
      '2026-05-12 07:48:00+00', '2026-05-12 14:53:17+00', 84.02, 9.82, '',
      NULL, '2026-05-05 07:00:00+00', '2026-05-12 14:53:17+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 448.354, '2026-05-12 14:53:17+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Fungicid lot rapita', 'out', 67.195, '2026-05-12 14:53:17+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 18.354, '2026-05-12 14:53:17+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-JOHNDEERE-S780-02'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 12 LIMIT 1),
      '2026-07-06 07:00:00+00', '2026-07-07 18:00:30+00', 84.02, '', 'completed', v_fc,
      '2026-07-06 07:53:00+00', '2026-07-07 23:55:51+00', 84.02, 37.82, 'Întârziere: teren umed',
      NULL, '2026-06-29 07:00:00+00', '2026-07-07 23:55:51+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1214.6, '2026-07-07 23:55:51+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'in', 455.27, '2026-07-08 19:55:51+00', NULL, v_fc, 'Recoltă Grâu - Tartaul - Tarla Mare (Sezon 2025/2026)');
  UPDATE field_crops SET production_total = 455.27, harvest_recorded_quantity = 455.27, harvest_recorded_at = '2026-07-08 19:55:51+00', harvested_at = '2026-07-08', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Balotat paie' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2026-07-11 07:00:00+00', '2026-07-12 18:00:30+00', 84.02, '', 'completed', v_fc,
      '2026-07-11 07:23:00+00', '2026-07-12 17:45:46+00', 84.02, 33.1, '',
      NULL, '2026-07-04 07:00:00+00', '2026-07-12 17:45:46+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 280.988, '2026-07-12 17:45:46+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Paie baloti', 'in', 207.71, '2026-07-12 23:45:46+00', NULL, v_fc, 'Paie Grâu - Tartaul - Tarla Mare, balotat');
  UPDATE field_crops SET status = 'closed', closed_at = '2026-07-17', harvested_at = COALESCE(harvested_at, '2026-07-17'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Gotesti - Tarla Vest (68.47 ha): Lucernă, anul 3 ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2025/2026';
  SELECT id INTO v_field FROM fields WHERE name = 'Gotesti - Tarla Vest' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Gotesti - Tarla Vest lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Lucernă'), 68.47, NULL, 9.5, 'Lucernă anul 3', 'open', NULL, '2025-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare lucerna' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2026-03-06 07:00:00+00', '2026-03-06 16:46:53+00', 68.47, '', 'completed', v_fc,
      '2026-03-06 07:53:00+00', '2026-03-06 17:02:07+00', 68.47, 10.25, '',
      NULL, '2026-02-27 07:00:00+00', '2026-03-06 17:02:07+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 440.828, '2026-03-06 17:02:07+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 10674.999, '2026-03-06 17:02:07+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Cosit si balotat lucerna' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2026-05-18 07:00:00+00', '2026-05-19 11:31:45+00', 68.47, '', 'completed', v_fc,
      '2026-05-18 07:25:00+00', '2026-05-19 10:30:50+00', 68.47, 30.02, '',
      NULL, '2026-05-11 07:00:00+00', '2026-05-19 10:30:50+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 580.951, '2026-05-19 10:30:50+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'in', 168.14, '2026-05-20 16:30:50+00', NULL, v_fc, 'Recoltă Lucernă - Gotesti - Tarla Vest (Sezon 2025/2026)');
  UPDATE field_crops SET production_total = 168.14, harvest_recorded_quantity = 168.14, harvest_recorded_at = '2026-05-20 16:30:50+00', harvested_at = '2026-05-20', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Cosit si balotat lucerna' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2026-06-28 07:00:00+00', '2026-06-29 11:31:45+00', 68.47, '', 'completed', v_fc,
      '2026-06-28 07:03:00+00', '2026-06-29 10:08:44+00', 68.47, 27.21, '',
      NULL, '2026-06-21 07:00:00+00', '2026-06-29 10:08:44+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 596.233, '2026-06-29 10:08:44+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'in', 194.01, '2026-06-30 16:08:44+00', NULL, v_fc, 'Recoltă Lucernă - Gotesti - Tarla Vest (Sezon 2025/2026)');
  UPDATE field_crops SET production_total = 362.15, harvest_recorded_quantity = 362.15, harvest_recorded_at = '2026-06-30 16:08:44+00', harvested_at = '2026-06-30', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Cosit si balotat lucerna' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 5 LIMIT 1),
      '2026-08-13 07:00:00+00', '2026-08-14 11:31:45+00', 68.47, '', 'completed', v_fc,
      '2026-08-13 07:19:00+00', '2026-08-14 17:01:29+00', 68.47, 30.04, 'Întârziere: teren umed',
      NULL, '2026-08-06 07:00:00+00', '2026-08-14 17:01:29+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 567.108, '2026-08-14 17:01:29+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'in', 155.21, '2026-08-15 23:01:29+00', NULL, v_fc, 'Recoltă Lucernă - Gotesti - Tarla Vest (Sezon 2025/2026)');
  UPDATE field_crops SET production_total = 517.36, harvest_recorded_quantity = 517.36, harvest_recorded_at = '2026-08-15 23:01:29+00', harvested_at = '2026-08-15', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Cosit si balotat lucerna' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2026-09-18 07:00:00+00', '2026-09-19 11:31:45+00', 68.47, '', 'completed', v_fc,
      '2026-09-18 06:56:00+00', '2026-09-19 10:05:26+00', 68.47, 29.9, '',
      NULL, '2026-09-11 07:00:00+00', '2026-09-19 10:05:26+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 588.061, '2026-09-19 10:05:26+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'in', 109.94, '2026-09-20 16:05:26+00', NULL, v_fc, 'Recoltă Lucernă - Gotesti - Tarla Vest (Sezon 2025/2026)');
  UPDATE field_crops SET production_total = 627.3, harvest_recorded_quantity = 627.3, harvest_recorded_at = '2026-09-20 16:05:26+00', harvested_at = '2026-09-20', updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Tartaul - Tarla Centrala (53.5 ha): Grâu ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2025/2026';
  SELECT id INTO v_field FROM fields WHERE name = 'Tartaul - Tarla Centrala' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Tartaul - Tarla Centrala lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Grâu'), 53.5, '2025-10-08', 5.8, '', 'open', NULL, '2025-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 8 LIMIT 1),
      '2025-08-26 07:00:00+00', '2025-08-27 16:26:15+00', 53.5, '', 'completed', v_fc,
      '2025-08-26 07:37:00+00', '2025-08-27 18:57:21+00', 53.5, 34.21, 'Întârziere: defecțiune minoră',
      NULL, '2025-08-19 07:00:00+00', '2025-08-27 18:57:21+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1061.004, '2025-08-27 18:57:21+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 9 LIMIT 1),
      '2025-09-26 07:00:00+00', '2025-09-27 16:26:15+00', 53.5, '', 'completed', v_fc,
      '2025-09-26 07:38:00+00', '2025-09-27 16:46:49+00', 53.5, 36.53, '',
      NULL, '2025-09-19 07:00:00+00', '2025-09-27 16:46:49+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 579.46, '2025-09-27 16:46:49+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 9804.286, '2025-09-27 16:46:49+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat grau standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 10 LIMIT 1),
      '2025-10-06 07:00:00+00', '2025-10-06 23:43:07+00', 53.5, '', 'completed', v_fc,
      '2025-10-06 07:14:00+00', '2025-10-06 22:34:24+00', 53.5, 16.64, '',
      NULL, '2025-09-29 07:00:00+00', '2025-10-06 22:34:24+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 502.804, '2025-10-06 22:34:24+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta grau lot A', 'out', 11928.871, '2025-10-06 22:34:24+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare faziala grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 11 LIMIT 1),
      '2026-03-11 07:00:00+00', '2026-03-11 14:38:34+00', 53.5, '', 'completed', v_fc,
      '2026-03-11 07:56:00+00', '2026-03-11 13:18:08+00', 53.5, 7.2, '',
      NULL, '2026-03-04 07:00:00+00', '2026-03-11 13:18:08+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 342.169, '2026-03-11 13:18:08+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'out', 8357.091, '2026-03-11 13:18:08+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 12 LIMIT 1),
      '2026-04-17 07:00:00+00', '2026-04-17 12:56:40+00', 53.5, '', 'completed', v_fc,
      '2026-04-17 06:47:00+00', '2026-04-17 15:03:12+00', 53.5, 6.39, 'Întârziere: teren umed',
      NULL, '2026-04-10 07:00:00+00', '2026-04-17 15:03:12+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 249.888, '2026-04-17 15:03:12+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 122.111, '2026-04-17 15:03:12+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 12.991, '2026-04-17 15:03:12+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2026-05-12 07:00:00+00', '2026-05-12 12:56:40+00', 53.5, '', 'completed', v_fc,
      '2026-05-12 07:50:00+00', '2026-05-12 11:29:00+00', 53.5, 5.58, '',
      NULL, '2026-05-05 07:00:00+00', '2026-05-12 11:29:00+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 282.654, '2026-05-12 11:29:00+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Fungicid lot rapita', 'out', 44.13, '2026-05-12 11:29:00+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 11.597, '2026-05-12 11:29:00+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-CLAAS-LEXION780-01'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 1 LIMIT 1),
      '2026-07-05 07:00:00+00', '2026-07-06 05:17:30+00', 53.5, '', 'completed', v_fc,
      '2026-07-05 07:44:00+00', '2026-07-06 04:14:20+00', 53.5, 23.37, '',
      NULL, '2026-06-28 07:00:00+00', '2026-07-06 04:14:20+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 822.869, '2026-07-06 04:14:20+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'in', 284.25, '2026-07-07 00:14:20+00', NULL, v_fc, 'Recoltă Grâu - Tartaul - Tarla Centrala (Sezon 2025/2026)');
  UPDATE field_crops SET production_total = 284.25, harvest_recorded_quantity = 284.25, harvest_recorded_at = '2026-07-07 00:14:20+00', harvested_at = '2026-07-07', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Balotat paie' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2026-07-12 07:00:00+00', '2026-07-13 05:17:30+00', 53.5, '', 'completed', v_fc,
      '2026-07-12 07:27:00+00', '2026-07-13 10:56:17+00', 50.94, 24.15, 'Întârziere: ploaie',
      NULL, '2026-07-05 07:00:00+00', '2026-07-13 10:56:17+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 160.681, '2026-07-13 10:56:17+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Paie baloti', 'in', 131.09, '2026-07-13 16:56:17+00', NULL, v_fc, 'Paie Grâu - Tartaul - Tarla Centrala, balotat');
  UPDATE field_crops SET status = 'closed', closed_at = '2026-07-18', harvested_at = COALESCE(harvested_at, '2026-07-18'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Plopi - Tarla Nord (52.41 ha): Sfeclă de zahăr ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2025/2026';
  SELECT id INTO v_field FROM fields WHERE name = 'Plopi - Tarla Nord' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Plopi - Tarla Nord lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Sfeclă de zahăr'), 52.41, '2026-03-30', 60.0, '', 'open', NULL, '2025-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2025-10-31 07:00:00+00', '2025-11-01 15:45:22+00', 52.41, '', 'completed', v_fc,
      '2025-10-31 07:21:00+00', '2025-11-01 15:03:21+00', 52.41, 32.57, '',
      NULL, '2025-10-24 07:00:00+00', '2025-11-01 15:03:21+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 1004.677, '2025-11-01 15:03:21+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 5 LIMIT 1),
      '2026-03-25 07:00:00+00', '2026-03-26 15:45:22+00', 52.41, '', 'completed', v_fc,
      '2026-03-25 06:51:00+00', '2026-03-26 14:34:15+00', 52.08, 32.38, '',
      NULL, '2026-03-18 07:00:00+00', '2026-03-26 14:34:15+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 541.891, '2026-03-26 14:34:15+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 9640.458, '2026-03-26 14:34:15+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat sfecla standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2026-04-02 07:00:00+00', '2026-04-02 23:22:41+00', 52.41, '', 'completed', v_fc,
      '2026-04-02 07:49:00+00', '2026-04-02 23:46:39+00', 52.41, 15.86, '',
      NULL, '2026-03-26 07:00:00+00', '2026-04-02 23:46:39+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 472.242, '2026-04-02 23:46:39+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta sfecla de zahar', 'out', 132.455, '2026-04-02 23:46:39+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2026-05-05 07:00:00+00', '2026-05-05 12:49:24+00', 52.41, '', 'completed', v_fc,
      '2026-05-05 07:40:00+00', '2026-05-05 11:44:15+00', 52.41, 6.2, '',
      NULL, '2026-04-28 07:00:00+00', '2026-05-05 11:44:15+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 242.43, '2026-05-05 11:44:15+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 118.013, '2026-05-05 11:44:15+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 13.329, '2026-05-05 11:44:15+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 8 LIMIT 1),
      '2026-05-31 07:00:00+00', '2026-05-31 12:49:24+00', 52.41, '', 'completed', v_fc,
      '2026-05-31 06:50:00+00', '2026-05-31 13:17:39+00', 52.41, 6.05, '',
      NULL, '2026-05-24 07:00:00+00', '2026-05-31 13:17:39+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 244.666, '2026-05-31 13:17:39+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 113.119, '2026-05-31 13:17:39+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 13.273, '2026-05-31 13:17:39+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Irigare pivot' AND deleted_at IS NULL),
      NULL, NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 9 LIMIT 1),
      '2026-07-10 07:00:00+00', '2026-07-11 03:00:00+00', 52.41, '', 'completed', v_fc,
      '2026-07-10 06:52:00+00', '2026-07-11 02:47:26+00', 52.41, NULL, '',
      NULL, '2026-07-03 07:00:00+00', '2026-07-11 02:47:26+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 18343.5, '2026-07-11 02:47:26+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare sfecla' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-JOHNDEERE-S780-02'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 10 LIMIT 1),
      '2026-10-28 07:00:00+00', '2026-10-29 04:50:15+00', 52.41, '', 'planned', v_fc,
      NULL, NULL, NULL, NULL, '',
      NULL, '2026-10-21 07:00:00+00', '2026-10-28 07:00:00+00')
  RETURNING id INTO v_op;
END $seed$;

-- ── Plopi - Tarla Sud (48.28 ha): Porumb ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2025/2026';
  SELECT id INTO v_field FROM fields WHERE name = 'Plopi - Tarla Sud' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Plopi - Tarla Sud lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Porumb'), 48.28, '2026-04-22', 10.35, '', 'open', NULL, '2025-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 12 LIMIT 1),
      '2025-10-27 07:00:00+00', '2025-10-28 13:10:30+00', 48.28, '', 'completed', v_fc,
      '2025-10-27 07:23:00+00', '2025-10-28 12:21:27+00', 48.28, 28.66, '',
      NULL, '2025-10-20 07:00:00+00', '2025-10-28 12:21:27+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 939.477, '2025-10-28 12:21:27+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2026-04-12 07:00:00+00', '2026-04-13 13:10:30+00', 48.28, '', 'completed', v_fc,
      '2026-04-12 06:40:00+00', '2026-04-13 12:21:29+00', 48.28, 29.36, '',
      NULL, '2026-04-05 07:00:00+00', '2026-04-13 12:21:29+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 503.961, '2026-04-13 12:21:29+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 8746.64, '2026-04-13 12:21:29+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat porumb standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 1 LIMIT 1),
      '2026-04-22 07:00:00+00', '2026-04-22 22:05:15+00', 48.28, '', 'completed', v_fc,
      '2026-04-22 07:27:00+00', '2026-04-22 22:02:02+00', 48.28, 15.15, '',
      NULL, '2026-04-15 07:00:00+00', '2026-04-22 22:02:02+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 452.351, '2026-04-22 22:02:02+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta porumb lot B', 'out', 40.464, '2026-04-22 22:02:02+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2026-05-23 07:00:00+00', '2026-05-23 12:21:52+00', 48.28, '', 'completed', v_fc,
      '2026-05-23 07:58:00+00', '2026-05-23 10:54:12+00', 48.28, 5.08, '',
      NULL, '2026-05-16 07:00:00+00', '2026-05-23 10:54:12+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 230.196, '2026-05-23 10:54:12+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 105.717, '2026-05-23 10:54:12+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 11.82, '2026-05-23 10:54:12+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Irigare pivot' AND deleted_at IS NULL),
      NULL, NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2026-06-28 07:00:00+00', '2026-06-29 03:00:00+00', 48.28, '', 'completed', v_fc,
      '2026-06-28 07:53:00+00', '2026-06-29 08:46:13+00', 48.28, NULL, 'Întârziere: ploaie',
      NULL, '2026-06-21 07:00:00+00', '2026-06-29 08:46:13+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 16898.0, '2026-06-29 08:46:13+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Irigare pivot' AND deleted_at IS NULL),
      NULL, NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2026-07-23 07:00:00+00', '2026-07-24 03:00:00+00', 48.28, '', 'completed', v_fc,
      '2026-07-23 07:52:00+00', '2026-07-24 03:17:33+00', 48.28, NULL, '',
      NULL, '2026-07-16 07:00:00+00', '2026-07-24 03:17:33+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 16898.0, '2026-07-24 03:17:33+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare porumb' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-JOHNDEERE-S780-02'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 5 LIMIT 1),
      '2026-10-04 07:00:00+00', '2026-10-05 03:07:00+00', 48.28, '', 'completed', v_fc,
      '2026-10-04 06:56:00+00', '2026-10-05 03:16:03+00', 48.28, 20.83, '',
      NULL, '2026-09-27 07:00:00+00', '2026-10-05 03:16:03+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 864.136, '2026-10-05 03:16:03+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Porumb', 'in', 512.74, '2026-10-05 23:16:03+00', NULL, v_fc, 'Recoltă Porumb - Plopi - Tarla Sud (Sezon 2025/2026)');
  UPDATE field_crops SET production_total = 512.74, harvest_recorded_quantity = 512.74, harvest_recorded_at = '2026-10-05 23:16:03+00', harvested_at = '2026-10-05', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2026-10-10', harvested_at = COALESCE(harvested_at, '2026-10-10'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Tartaul - Tarla Mica (9.84 ha): Grâu compromis (îngheț), apoi Porumb în același sezon ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2025/2026';
  SELECT id INTO v_field FROM fields WHERE name = 'Tartaul - Tarla Mica' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, production_total, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Grâu'), 9.84, '2025-10-09', 5.8, 0, 'Compromisă: îngheț 12–18.02.2026, plante distruse pe 90% din suprafață. Reînsămânțat cu porumb.', 'compromised', '2026-02-28', '2025-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2025-08-24 07:00:00+00', '2025-08-24 13:09:00+00', 9.84, '', 'completed', v_fc,
      '2025-08-24 06:40:00+00', '2025-08-24 12:39:10+00', 9.36, 6.54, '',
      NULL, '2025-08-17 07:00:00+00', '2025-08-24 12:39:10+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 172.97, '2025-08-24 12:39:10+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 8 LIMIT 1),
      '2025-09-29 07:00:00+00', '2025-09-29 13:09:00+00', 9.84, '', 'completed', v_fc,
      '2025-09-29 07:02:00+00', '2025-09-29 13:32:25+00', 9.58, 6.43, '',
      NULL, '2025-09-22 07:00:00+00', '2025-09-29 13:32:25+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 102.477, '2025-09-29 13:32:25+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 1702.734, '2025-09-29 13:32:25+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat grau standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 9 LIMIT 1),
      '2025-10-09 07:00:00+00', '2025-10-09 10:04:30+00', 9.84, '', 'completed', v_fc,
      '2025-10-09 07:26:00+00', '2025-10-09 10:29:38+00', 9.84, 3.07, '',
      NULL, '2025-10-02 07:00:00+00', '2025-10-09 10:29:38+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 89.347, '2025-10-09 10:29:38+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta grau lot A', 'out', 2236.593, '2025-10-09 10:29:38+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  -- porumbul, pe același teren, în același sezon
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Porumb'), 9.84, '2026-04-20', 8.5, 'Reînsămânțare după grâul compromis.', 'open', '2026-03-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 10 LIMIT 1),
      '2026-04-08 07:00:00+00', '2026-04-08 13:09:00+00', 9.84, '', 'completed', v_fc,
      '2026-04-08 06:57:00+00', '2026-04-08 12:39:10+00', 9.84, 6.5, '',
      NULL, '2026-04-01 07:00:00+00', '2026-04-08 12:39:10+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 100.488, '2026-04-08 12:39:10+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 1719.488, '2026-04-08 12:39:10+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat porumb standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 11 LIMIT 1),
      '2026-04-20 07:00:00+00', '2026-04-20 10:04:30+00', 9.84, '', 'completed', v_fc,
      '2026-04-20 07:09:00+00', '2026-04-20 16:01:48+00', 9.84, 2.99, 'Întârziere: teren umed',
      NULL, '2026-04-13 07:00:00+00', '2026-04-20 16:01:48+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 87.586, '2026-04-20 16:01:48+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta porumb lot B', 'out', 8.307, '2026-04-20 16:01:48+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 12 LIMIT 1),
      '2026-05-18 07:00:00+00', '2026-05-18 08:30:00+00', 9.84, '', 'completed', v_fc,
      '2026-05-18 06:46:00+00', '2026-05-18 07:13:12+00', 9.84, 1.6, '',
      NULL, '2026-05-11 07:00:00+00', '2026-05-18 07:13:12+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 48.024, '2026-05-18 07:13:12+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 21.654, '2026-05-18 07:13:12+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 2.53, '2026-05-18 07:13:12+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare porumb' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-JOHNDEERE-S780-02'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2026-10-05 07:00:00+00', '2026-10-05 11:06:00+00', 9.84, '', 'completed', v_fc,
      '2026-10-05 07:16:00+00', '2026-10-05 12:13:37+00', 9.84, 4.12, 'Întârziere: defecțiune minoră',
      NULL, '2026-09-28 07:00:00+00', '2026-10-05 12:13:37+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 158.731, '2026-10-05 12:13:37+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Porumb', 'in', 79.73, '2026-10-06 08:13:37+00', NULL, v_fc, 'Recoltă Porumb - Tartaul - Tarla Mica (Sezon 2025/2026)');
  UPDATE field_crops SET production_total = 79.73, harvest_recorded_quantity = 79.73, harvest_recorded_at = '2026-10-06 08:13:37+00', harvested_at = '2026-10-06', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2026-10-09', updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Lingura - Teren agricol Nord (7.98 ha): Orz ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2025/2026';
  SELECT id INTO v_field FROM fields WHERE name = 'Lingura - Teren agricol Nord' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Lingura - Teren agricol Nord lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Orz'), 7.98, '2025-09-30', 5.2, '', 'open', NULL, '2025-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2025-08-20 07:00:00+00', '2025-08-20 11:59:15+00', 7.98, '', 'completed', v_fc,
      '2025-08-20 07:13:00+00', '2025-08-20 11:15:18+00', 7.76, 5.42, '',
      NULL, '2025-08-13 07:00:00+00', '2025-08-20 11:15:18+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 143.007, '2025-08-20 11:15:18+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2025-09-20 07:00:00+00', '2025-09-20 11:59:15+00', 7.98, '', 'completed', v_fc,
      '2025-09-20 07:42:00+00', '2025-09-20 11:23:34+00', 7.98, 5.42, '',
      NULL, '2025-09-13 07:00:00+00', '2025-09-20 11:23:34+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 83.418, '2025-09-20 11:23:34+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 1449.143, '2025-09-20 11:23:34+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat orz standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2025-10-01 07:00:00+00', '2025-10-01 09:29:37+00', 7.98, '', 'completed', v_fc,
      '2025-10-01 06:47:00+00', '2025-10-01 09:19:51+00', 7.92, 2.74, '',
      NULL, '2025-09-24 07:00:00+00', '2025-10-01 09:19:51+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 75.615, '2025-10-01 09:19:51+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta orz', 'out', 1442.184, '2025-10-01 09:19:51+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare faziala grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 5 LIMIT 1),
      '2026-03-08 07:00:00+00', '2026-03-08 08:30:00+00', 7.98, '', 'completed', v_fc,
      '2026-03-08 07:09:00+00', '2026-03-08 08:37:55+00', 7.98, 1.62, '',
      NULL, '2026-03-01 07:00:00+00', '2026-03-08 08:37:55+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 50.592, '2026-03-08 08:37:55+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'out', 1286.835, '2026-03-08 08:37:55+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2026-04-15 07:00:00+00', '2026-04-15 08:30:00+00', 7.98, '', 'completed', v_fc,
      '2026-04-15 07:12:00+00', '2026-04-15 07:16:50+00', 7.8, 1.46, '',
      NULL, '2026-04-08 07:00:00+00', '2026-04-15 07:16:50+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 37.809, '2026-04-15 07:16:50+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 16.708, '2026-04-15 07:16:50+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 2.016, '2026-04-15 07:16:50+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare orz' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-CLAAS-LEXION780-01'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2026-06-30 07:00:00+00', '2026-06-30 10:19:30+00', 7.98, '', 'completed', v_fc,
      '2026-06-30 07:27:00+00', '2026-06-30 08:53:23+00', 7.98, 3.49, '',
      NULL, '2026-06-23 07:00:00+00', '2026-06-30 08:53:23+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 114.265, '2026-06-30 08:53:23+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Orz', 'in', 39.43, '2026-07-01 04:53:23+00', NULL, v_fc, 'Recoltă Orz - Lingura - Teren agricol Nord (Sezon 2025/2026)');
  UPDATE field_crops SET production_total = 39.43, harvest_recorded_quantity = 39.43, harvest_recorded_at = '2026-07-01 04:53:23+00', harvested_at = '2026-07-01', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Balotat paie' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 8 LIMIT 1),
      '2026-07-03 07:00:00+00', '2026-07-03 10:19:30+00', 7.98, '', 'completed', v_fc,
      '2026-07-03 07:29:00+00', '2026-07-03 09:19:12+00', 7.98, 3.59, '',
      NULL, '2026-06-26 07:00:00+00', '2026-07-03 09:19:12+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 25.904, '2026-07-03 09:19:12+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Paie baloti', 'in', 15.44, '2026-07-03 15:19:12+00', NULL, v_fc, 'Paie Orz - Lingura - Teren agricol Nord, balotat');
  UPDATE field_crops SET status = 'closed', closed_at = '2026-07-08', harvested_at = COALESCE(harvested_at, '2026-07-08'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Cirpesti - Lot Nord (2.38 ha): Grâu ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2025/2026';
  SELECT id INTO v_field FROM fields WHERE name = 'Cirpesti - Lot Nord' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Cirpesti - Lot Nord lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Grâu'), 2.38, '2025-10-08', 5.8, '', 'open', NULL, '2025-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 10 LIMIT 1),
      '2025-08-28 07:00:00+00', '2025-08-28 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2025-08-28 06:42:00+00', '2025-08-28 07:12:55+00', 2.29, 1.61, '',
      NULL, '2025-08-21 07:00:00+00', '2025-08-28 07:12:55+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 41.729, '2025-08-28 07:12:55+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 11 LIMIT 1),
      '2025-09-26 07:00:00+00', '2025-09-26 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2025-09-26 06:44:00+00', '2025-09-26 09:58:23+00', 2.38, 1.45, 'Întârziere: ploaie',
      NULL, '2025-09-19 07:00:00+00', '2025-09-26 09:58:23+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 24.403, '2025-09-26 09:58:23+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 426.356, '2025-09-26 09:58:23+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat grau standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 12 LIMIT 1),
      '2025-10-06 07:00:00+00', '2025-10-06 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2025-10-06 07:32:00+00', '2025-10-06 08:11:15+00', 2.37, 1.54, '',
      NULL, '2025-09-29 07:00:00+00', '2025-10-06 08:11:15+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 23.625, '2025-10-06 08:11:15+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta grau lot A', 'out', 512.191, '2025-10-06 08:11:15+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Fertilizare faziala grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'FS-AMAZONE-ZA-TS4200-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 0 LIMIT 1),
      '2026-03-14 07:00:00+00', '2026-03-14 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2026-03-14 06:50:00+00', '2026-03-14 12:12:51+00', 2.38, 1.57, 'Întârziere: defecțiune minoră',
      NULL, '2026-03-07 07:00:00+00', '2026-03-14 12:12:51+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 15.76, '2026-03-14 12:12:51+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Uree pentru fertilizare faziala', 'out', 388.011, '2026-03-14 12:12:51+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 1 LIMIT 1),
      '2026-04-21 07:00:00+00', '2026-04-21 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2026-04-21 07:40:00+00', '2026-04-21 08:40:00+00', 2.38, 1.56, '',
      NULL, '2026-04-14 07:00:00+00', '2026-04-21 08:40:00+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 11.804, '2026-04-21 08:40:00+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 5.226, '2026-04-21 08:40:00+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 0.578, '2026-04-21 08:40:00+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Tratament fungicid grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 2 LIMIT 1),
      '2026-05-13 07:00:00+00', '2026-05-13 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2026-05-13 07:04:00+00', '2026-05-13 08:42:01+00', 2.38, 1.57, '',
      NULL, '2026-05-06 07:00:00+00', '2026-05-13 08:42:01+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 12.777, '2026-05-13 08:42:01+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Fungicid lot rapita', 'out', 1.916, '2026-05-13 08:42:01+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 0.531, '2026-05-13 08:42:01+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare grau' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-JOHNDEERE-S780-02'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2026-07-03 07:00:00+00', '2026-07-03 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2026-07-03 07:17:00+00', '2026-07-03 12:45:22+00', 2.38, 1.56, 'Întârziere: defecțiune minoră',
      NULL, '2026-06-26 07:00:00+00', '2026-07-03 12:45:22+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 34.271, '2026-07-03 12:45:22+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'in', 14.02, '2026-07-04 08:45:22+00', NULL, v_fc, 'Recoltă Grâu - Cirpesti - Lot Nord (Sezon 2025/2026)');
  UPDATE field_crops SET production_total = 14.02, harvest_recorded_quantity = 14.02, harvest_recorded_at = '2026-07-04 08:45:22+00', harvested_at = '2026-07-04', updated_at = NOW() WHERE id = v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Balotat paie' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-CASEIH-PUMA240-04'), (SELECT id FROM implements WHERE name = 'TR-KRAMPE-HALFPIPE-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 4 LIMIT 1),
      '2026-07-12 07:00:00+00', '2026-07-12 08:30:00+00', 2.38, '', 'completed', v_fc,
      '2026-07-12 06:46:00+00', '2026-07-12 08:34:31+00', 2.38, 1.53, '',
      NULL, '2026-07-05 07:00:00+00', '2026-07-12 08:34:31+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 7.589, '2026-07-12 08:34:31+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Paie baloti', 'in', 5.38, '2026-07-12 14:34:31+00', NULL, v_fc, 'Paie Grâu - Cirpesti - Lot Nord, balotat');
  UPDATE field_crops SET status = 'closed', closed_at = '2026-07-17', harvested_at = COALESCE(harvested_at, '2026-07-17'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Cirpesti - Lot Central (1.91 ha): Soia ──
DO $seed$
DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_season FROM seasons WHERE name = 'Sezon 2025/2026';
  SELECT id INTO v_field FROM fields WHERE name = 'Cirpesti - Lot Central' AND deleted_at IS NULL;
  IF v_field IS NULL THEN RAISE NOTICE 'terenul Cirpesti - Lot Central lipsește; sărit'; RETURN; END IF;
  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)
  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Soia'), 1.91, '2026-04-28', 2.8, '', 'open', NULL, '2025-08-01 08:00:00+00', NOW())
  RETURNING id INTO v_fc;
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2025-11-03 07:00:00+00', '2025-11-03 08:30:00+00', 1.91, '', 'completed', v_fc,
      '2025-11-03 08:00:00+00', '2025-11-03 09:00:00+00', 1.91, 1.44, '',
      NULL, '2025-10-27 07:00:00+00', '2025-11-03 09:00:00+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 34.986, '2025-11-03 09:00:00+00', v_op, NULL, 'Combustibil raportat la finalizare');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2026-04-18 07:00:00+00', '2026-04-18 08:30:00+00', 1.91, '', 'completed', v_fc,
      '2026-04-18 07:29:00+00', '2026-04-18 12:15:38+00', 1.91, 1.54, 'Întârziere: teren umed',
      NULL, '2026-04-11 07:00:00+00', '2026-04-18 12:15:38+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 20.065, '2026-04-18 12:15:38+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('NPK pentru pregatire teren', 'out', 345.527, '2026-04-18 12:15:38+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Semanat soia standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'SE-HORSCH-PRONTO6DC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 8 LIMIT 1),
      '2026-04-28 07:00:00+00', '2026-04-28 08:30:00+00', 1.91, '', 'completed', v_fc,
      '2026-04-28 06:46:00+00', '2026-04-28 08:57:50+00', 1.91, 1.59, '',
      NULL, '2026-04-21 07:00:00+00', '2026-04-28 08:57:50+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 16.625, '2026-04-28 08:57:50+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Samanta soia', 'out', 191.61, '2026-04-28 08:57:50+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Erbicidare camp est' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'SP-HARDI-ALPHAEVO-01'), NULL,
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 9 LIMIT 1),
      '2026-05-30 07:00:00+00', '2026-05-30 08:30:00+00', 1.91, '', 'completed', v_fc,
      '2026-05-30 07:07:00+00', '2026-05-30 07:10:03+00', 1.91, 1.64, '',
      NULL, '2026-05-23 07:00:00+00', '2026-05-30 07:10:03+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 9.116, '2026-05-30 07:10:03+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Erbicid camp est', 'out', 4.173, '2026-05-30 07:10:03+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  PERFORM pg_temp.seed_move('Apa sistem pivot 1', 'out', 0.493, '2026-05-30 07:10:03+00', v_op, NULL, 'Consum la finalizarea operațiunii');
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Recoltare soia' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'CB-CLAAS-LEXION780-01'), (SELECT id FROM implements WHERE name = 'HD-CLAAS-VARIO1080-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 10 LIMIT 1),
      '2026-09-22 07:00:00+00', '2026-09-22 08:30:00+00', 1.91, '', 'completed', v_fc,
      '2026-09-22 07:13:00+00', '2026-09-22 08:58:46+00', 1.91, 1.47, '',
      NULL, '2026-09-15 07:00:00+00', '2026-09-22 08:58:46+00')
  RETURNING id INTO v_op;
  PERFORM pg_temp.seed_move('Motorina flota utilaje', 'out', 24.091, '2026-09-22 08:58:46+00', v_op, NULL, 'Combustibil raportat la finalizare');
  PERFORM pg_temp.seed_move('Recoltă Soia', 'in', 5.01, '2026-09-23 04:58:46+00', NULL, v_fc, 'Recoltă Soia - Cirpesti - Lot Central (Sezon 2025/2026)');
  UPDATE field_crops SET production_total = 5.01, harvest_recorded_quantity = 5.01, harvest_recorded_at = '2026-09-23 04:58:46+00', harvested_at = '2026-09-23', updated_at = NOW() WHERE id = v_fc;
  UPDATE field_crops SET status = 'closed', closed_at = '2026-09-27', harvested_at = COALESCE(harvested_at, '2026-09-27'), updated_at = NOW() WHERE id = v_fc;
END $seed$;

-- ── Cirpesti - Lot Sud: fără cultură în acest sezon (teren liber) ──

-- ── Toamna 2026: lucrări fără cultură (sezonul următor nu e deschis încă) ──
DO $seed$
DECLARE v_field uuid; v_fc bigint; v_op bigint;
BEGIN
  SELECT id INTO v_field FROM fields WHERE name = 'Gotesti - Tarla Est' AND deleted_at IS NULL;
  IF v_field IS NOT NULL THEN
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2026-10-14 07:00:00+00', '2026-10-16 16:42:45+00', 92.34, '[Sezon 2025/2026] Arat de toamnă pentru grâu 2026/2027', 'in_progress', NULL,
      '2026-10-14 07:15:00+00', NULL, NULL, NULL, '',
      NULL, '2026-10-07 07:00:00+00', '2026-10-14 07:00:00+00')
  RETURNING id INTO v_op;
  END IF;
  SELECT id INTO v_field FROM fields WHERE name = 'Plopi - Tarla Nord' AND deleted_at IS NULL;
  IF v_field IS NOT NULL THEN
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-FENDT-724VARIO-02'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 3 LIMIT 1),
      '2026-10-16 08:00:00+00', '2026-10-17 16:45:22+00', 52.41, '[Sezon 2025/2026] Arat de toamnă; același echipaj ca pe Gotesti - Tarla Est (suprapunere)', 'planned', NULL,
      NULL, NULL, NULL, NULL, '',
      NULL, '2026-10-09 08:00:00+00', '2026-10-16 08:00:00+00')
  RETURNING id INTO v_op;
  END IF;
  SELECT id INTO v_field FROM fields WHERE name = 'Tartaul - Tarla Mare' AND deleted_at IS NULL;
  IF v_field IS NOT NULL THEN
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Pregatire pat germinativ' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-VALTRA-T254-06'), (SELECT id FROM implements WHERE name = 'DH-KVERNELAND-QUALIDISC-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 5 LIMIT 1),
      '2026-10-08 07:00:00+00', '2026-10-10 11:30:45+00', 84.02, '[Sezon 2025/2026] Pregătire pentru rapiță; a depășit termenul', 'in_progress', NULL,
      '2026-10-08 07:32:00+00', NULL, NULL, NULL, '',
      '2026-10-10 11:31:45+00', '2026-10-01 07:00:00+00', '2026-10-08 07:00:00+00')
  RETURNING id INTO v_op;
  END IF;
  SELECT id INTO v_field FROM fields WHERE name = 'Plopi - Tarla Sud' AND deleted_at IS NULL;
  IF v_field IS NOT NULL THEN
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-SAME-EXPLORER120-07'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 6 LIMIT 1),
      '2026-10-21 07:00:00+00', '2026-10-22 13:10:30+00', 48.28, '[Sezon 2025/2026] Arat de toamnă', 'planned', NULL,
      NULL, NULL, NULL, NULL, '',
      NULL, '2026-10-14 07:00:00+00', '2026-10-21 07:00:00+00')
  RETURNING id INTO v_op;
  END IF;
  SELECT id INTO v_field FROM fields WHERE name = 'Cirpesti - Lot Sud' AND deleted_at IS NULL;
  IF v_field IS NOT NULL THEN
  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,
      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,
      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,
      overdue_notified_at, created_at, updated_at)
  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = 'Arat adanc standard' AND deleted_at IS NULL),
      (SELECT id FROM machines WHERE name = 'TR-JOHNDEERE-6R185-01'), (SELECT id FROM implements WHERE name = 'PL-LEMKEN-JUWEL8-01'),
      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET 7 LIMIT 1),
      '2026-10-27 07:00:00+00', '2026-10-27 08:30:00+00', 1.56, '[Sezon 2025/2026] Arat de toamnă pe lotul liber', 'planned', NULL,
      NULL, NULL, NULL, NULL, '',
      NULL, '2026-10-20 07:00:00+00', '2026-10-27 07:00:00+00')
  RETURNING id INTO v_op;
  END IF;
END $seed$;

-- ─── Vânzări din stoc, după recoltă (ieșiri fără lucrare) ───
DO $sale$ BEGIN
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'out', 134.51, '2026-07-01 19:30:50+00', NULL, NULL, '[Sezon 2025/2026] Vânzare Lucernă — contract, Gotesti - Tarla Vest');
  PERFORM pg_temp.seed_move('Recoltă Orz', 'out', 33.52, '2026-08-02 07:53:23+00', NULL, NULL, '[Sezon 2025/2026] Vânzare Orz — contract, Lingura - Teren agricol Nord');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'out', 11.92, '2026-08-09 11:45:22+00', NULL, NULL, '[Sezon 2025/2026] Vânzare Grâu — contract, Cirpesti - Lot Nord');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'out', 241.61, '2026-08-13 03:14:20+00', NULL, NULL, '[Sezon 2025/2026] Vânzare Grâu — contract, Tartaul - Tarla Centrala');
  PERFORM pg_temp.seed_move('Paie baloti', 'out', 3.23, '2026-08-14 17:34:31+00', NULL, NULL, '[Sezon 2025/2026] Vânzare paie — Cirpesti - Lot Nord');
  PERFORM pg_temp.seed_move('Paie baloti', 'out', 9.26, '2026-08-20 18:19:12+00', NULL, NULL, '[Sezon 2025/2026] Vânzare paie — Lingura - Teren agricol Nord');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'out', 155.21, '2026-08-25 19:08:44+00', NULL, NULL, '[Sezon 2025/2026] Vânzare Lucernă — contract, Gotesti - Tarla Vest');
  PERFORM pg_temp.seed_move('Recoltă Rapiță', 'out', 323.21, '2026-08-31 10:49:53+00', NULL, NULL, '[Sezon 2025/2026] Vânzare Rapiță — contract, Gotesti - Tarla Mare');
  PERFORM pg_temp.seed_move('Recoltă Grâu', 'out', 386.98, '2026-09-03 22:55:51+00', NULL, NULL, '[Sezon 2025/2026] Vânzare Grâu — contract, Tartaul - Tarla Mare');
  PERFORM pg_temp.seed_move('Paie baloti', 'out', 78.65, '2026-09-07 19:56:17+00', NULL, NULL, '[Sezon 2025/2026] Vânzare paie — Tartaul - Tarla Centrala');
  PERFORM pg_temp.seed_move('Paie baloti', 'out', 124.63, '2026-09-11 02:45:46+00', NULL, NULL, '[Sezon 2025/2026] Vânzare paie — Tartaul - Tarla Mare');
  PERFORM pg_temp.seed_move('Recoltă Lucernă', 'out', 124.17, '2026-10-13 02:01:29+00', NULL, NULL, '[Sezon 2025/2026] Vânzare Lucernă — contract, Gotesti - Tarla Vest');
END $sale$;

-- ─── Istoricul stocului în ordine cronologică (resulting_quantity ca sold curent la acea dată) ───
UPDATE stock_movements sm SET resulting_quantity = t.run FROM (SELECT id, SUM(quantity_delta) OVER (PARTITION BY resource_id ORDER BY created_at, id) AS run FROM stock_movements) t WHERE t.id = sm.id;

COMMIT;
