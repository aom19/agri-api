ALTER TABLE resource_types DROP CONSTRAINT IF EXISTS resource_types_category_check;
ALTER TABLE resource_types ADD CONSTRAINT resource_types_category_check
    CHECK (category IN ('fuel', 'fertilizer', 'seed', 'pesticide', 'water', 'other'));

DROP INDEX IF EXISTS idx_operation_templates_crop;
ALTER TABLE operation_templates DROP COLUMN IF EXISTS crop_id;
ALTER TABLE crops DROP COLUMN IF EXISTS harvest_resource_id;
ALTER TABLE field_crops
    DROP COLUMN IF EXISTS harvest_recorded_quantity,
    DROP COLUMN IF EXISTS harvest_recorded_at;
ALTER TABLE stock_movements DROP COLUMN IF EXISTS field_crop_id;
DROP INDEX IF EXISTS idx_field_operations_field_crop;
ALTER TABLE field_operations DROP COLUMN IF EXISTS field_crop_id;
