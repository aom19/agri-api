-- Sezoanele în curs primesc ca sfârșit un an după început, ca să poată reveni la NOT NULL.
UPDATE seasons SET end_date = start_date + INTERVAL '1 year' - INTERVAL '1 day' WHERE end_date IS NULL;
ALTER TABLE seasons ALTER COLUMN end_date SET NOT NULL;
