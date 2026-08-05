DROP INDEX IF EXISTS idx_fields_cadastral_number_unique;

ALTER TABLE fields
DROP COLUMN IF EXISTS cadastral_number;