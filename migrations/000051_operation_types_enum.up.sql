-- T14: tipul operațiunii e un enum fix în cod (domain.OperationType), nu un tabel editabil.
-- Template-ul are tipul; operațiunea pe teren îl moștenește din template și are tip propriu
-- doar când nu are template, deci cele două nu se mai pot contrazice.

-- Un tip folosit care nu e unul din cele 6 nu are corespondent: migrarea se oprește.
DO $$
DECLARE
    unknown TEXT;
BEGIN
    SELECT string_agg(DISTINCT ot.code, ', ')
    INTO unknown
    FROM operation_types ot
    WHERE ot.code NOT IN ('soil_preparation', 'seeding', 'fertilization', 'spraying', 'harvesting', 'irrigation')
      AND (
          EXISTS (SELECT 1 FROM operation_templates t WHERE t.operation_type_id = ot.id)
          OR EXISTS (SELECT 1 FROM field_operations fo WHERE fo.operation_type_id = ot.id)
      );
    IF unknown IS NOT NULL THEN
        RAISE EXCEPTION 'Tipuri de operațiuni folosite care nu sunt în enum: %. Mută template-urile și operațiunile lor pe unul din cele 6 tipuri, apoi rulează din nou migrarea.', unknown;
    END IF;
END $$;

ALTER TABLE operation_templates ADD COLUMN operation_type VARCHAR(50);

UPDATE operation_templates t
SET operation_type = ot.code
FROM operation_types ot
WHERE ot.id = t.operation_type_id;

ALTER TABLE operation_templates
    ALTER COLUMN operation_type SET NOT NULL,
    ADD CONSTRAINT operation_templates_operation_type_check
        CHECK (operation_type IN ('soil_preparation', 'seeding', 'fertilization', 'spraying', 'harvesting', 'irrigation')),
    DROP COLUMN operation_type_id;

-- Operațiunile cu template iau tipul template-ului, chiar dacă aveau altul reținut.
ALTER TABLE field_operations ADD COLUMN operation_type VARCHAR(50);

UPDATE field_operations fo
SET operation_type = ot.code
FROM operation_types ot
WHERE ot.id = fo.operation_type_id
  AND fo.operation_template_id IS NULL;

ALTER TABLE field_operations
    ADD CONSTRAINT field_operations_operation_type_check
        CHECK (operation_type IN ('soil_preparation', 'seeding', 'fertilization', 'spraying', 'harvesting', 'irrigation')),
    ADD CONSTRAINT field_operations_operation_type_source_check
        CHECK ((operation_template_id IS NULL) = (operation_type IS NOT NULL)),
    DROP COLUMN operation_type_id;

DROP TABLE operation_types;

UPDATE permissions SET description = 'Vizualizare template-uri de operațiuni' WHERE name = 'operations:read';
UPDATE permissions SET description = 'Creare și editare template-uri de operațiuni' WHERE name = 'operations:write';
UPDATE permissions SET description = 'Ștergere template-uri de operațiuni' WHERE name = 'operations:delete';
