-- 1. Operațiunile pe teren pot fi legate explicit de cultura din sezon.
ALTER TABLE field_operations
    ADD COLUMN IF NOT EXISTS field_crop_id BIGINT REFERENCES field_crops(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_field_operations_field_crop ON field_operations (field_crop_id);

-- 2. Mișcările de stoc pot proveni din recoltă (legate de cultura pe teren).
ALTER TABLE stock_movements
    ADD COLUMN IF NOT EXISTS field_crop_id BIGINT REFERENCES field_crops(id) ON DELETE SET NULL;

-- 3. Recolta înregistrată în stoc: cantitatea deja intrată și momentul.
ALTER TABLE field_crops
    ADD COLUMN IF NOT EXISTS harvest_recorded_quantity NUMERIC(14,3),
    ADD COLUMN IF NOT EXISTS harvest_recorded_at TIMESTAMPTZ;

-- 4. Fiecare cultură are (la nevoie) o resursă de stoc pentru recoltă.
ALTER TABLE crops
    ADD COLUMN IF NOT EXISTS harvest_resource_id BIGINT REFERENCES resources(id) ON DELETE SET NULL;

-- 5. Șabloanele de operațiuni pot fi recomandate pentru o cultură.
ALTER TABLE operation_templates
    ADD COLUMN IF NOT EXISTS crop_id BIGINT REFERENCES crops(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_operation_templates_crop ON operation_templates (crop_id);

-- 6. Categorie nouă de resurse: recoltă.
ALTER TABLE resource_types DROP CONSTRAINT IF EXISTS resource_types_category_check;
ALTER TABLE resource_types ADD CONSTRAINT resource_types_category_check
    CHECK (category IN ('fuel', 'fertilizer', 'seed', 'pesticide', 'water', 'harvest', 'other'));

-- 7. Backfill: leagă operațiunile existente de cultura terenului din sezonul în care se încadrează.
UPDATE field_operations fo
SET field_crop_id = fc.id
FROM field_crops fc
JOIN seasons s ON s.id = fc.season_id
WHERE fo.field_crop_id IS NULL
  AND fo.field_id = fc.field_id
  AND COALESCE(fo.planned_start_at, fo.created_at) >= s.start_date
  AND COALESCE(fo.planned_start_at, fo.created_at) < s.end_date + INTERVAL '1 day';
