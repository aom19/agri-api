-- Reface constrângerea „o cultură pe teren per sezon”. Eșuează, și trebuie să eșueze, dacă
-- între timp au apărut două culturi în același sezon pe același teren.
DROP INDEX IF EXISTS idx_field_crops_field_status;
DROP INDEX IF EXISTS field_crops_one_open_per_field;
ALTER TABLE field_crops ADD CONSTRAINT field_crops_field_id_season_id_key UNIQUE (field_id, season_id);
ALTER TABLE field_crops
    DROP COLUMN IF EXISTS closed_at,
    DROP COLUMN IF EXISTS status;
