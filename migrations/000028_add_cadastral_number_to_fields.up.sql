ALTER TABLE fields
ADD COLUMN IF NOT EXISTS cadastral_number VARCHAR(100);

CREATE UNIQUE INDEX IF NOT EXISTS idx_fields_cadastral_number_unique
ON fields(cadastral_number)
WHERE cadastral_number IS NOT NULL;