-- Recreează tabelul operation_types (000029) cu cele 6 tipuri. Fiecare operațiune pe teren
-- primește din nou un tip propriu: al template-ului când are unul.
CREATE TABLE IF NOT EXISTS operation_types (
    id          BIGSERIAL PRIMARY KEY,
    code        VARCHAR(50) NOT NULL UNIQUE,
    name        VARCHAR(100) NOT NULL,
    description TEXT DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

INSERT INTO operation_types (code, name) VALUES
    ('soil_preparation', 'Pregătire sol'),
    ('seeding', 'Semănat'),
    ('fertilization', 'Fertilizare'),
    ('spraying', 'Stropire'),
    ('harvesting', 'Recoltare'),
    ('irrigation', 'Irigare')
ON CONFLICT (code) DO NOTHING;

ALTER TABLE field_operations
    DROP CONSTRAINT IF EXISTS field_operations_operation_type_source_check,
    ADD COLUMN operation_type_id BIGINT REFERENCES operation_types(id);

UPDATE field_operations fo
SET operation_type_id = ot.id
FROM operation_types ot
WHERE ot.code = COALESCE(
    (SELECT t.operation_type FROM operation_templates t WHERE t.id = fo.operation_template_id),
    fo.operation_type
);

ALTER TABLE field_operations
    ALTER COLUMN operation_type_id SET NOT NULL,
    DROP COLUMN operation_type;

CREATE INDEX IF NOT EXISTS idx_field_operations_operation_type_id ON field_operations(operation_type_id);

ALTER TABLE operation_templates ADD COLUMN operation_type_id BIGINT REFERENCES operation_types(id);

UPDATE operation_templates t
SET operation_type_id = ot.id
FROM operation_types ot
WHERE ot.code = t.operation_type;

ALTER TABLE operation_templates
    ALTER COLUMN operation_type_id SET NOT NULL,
    DROP COLUMN operation_type;

UPDATE permissions SET description = 'Vizualizare tipuri operațiuni și template-uri' WHERE name = 'operations:read';
UPDATE permissions SET description = 'Creare și editare tipuri operațiuni și template-uri' WHERE name = 'operations:write';
UPDATE permissions SET description = 'Ștergere tipuri operațiuni și template-uri' WHERE name = 'operations:delete';
