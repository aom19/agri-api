ALTER TABLE resources
ADD COLUMN IF NOT EXISTS unit VARCHAR(50);

UPDATE resources r
SET unit = rt.default_unit
FROM resource_types rt
WHERE rt.id = r.resource_type_id
  AND r.unit IS NULL;

ALTER TABLE resources
ALTER COLUMN unit SET NOT NULL;
