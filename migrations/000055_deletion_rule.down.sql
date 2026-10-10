-- Readuce cascadele și unicitatea pe toate rândurile, inclusiv cele arhivate. Eșuează dacă un cod
-- de echipament sau un număr cadastral a fost refolosit după arhivare.
ALTER TABLE stock_movements
    DROP CONSTRAINT stock_movements_field_crop_id_fkey,
    ADD CONSTRAINT stock_movements_field_crop_id_fkey FOREIGN KEY (field_crop_id) REFERENCES field_crops(id) ON DELETE SET NULL;

ALTER TABLE field_crops
    DROP CONSTRAINT field_crops_season_id_fkey,
    ADD CONSTRAINT field_crops_season_id_fkey FOREIGN KEY (season_id) REFERENCES seasons(id) ON DELETE CASCADE;

DROP INDEX IF EXISTS idx_fields_cadastral_number_unique;
CREATE UNIQUE INDEX idx_fields_cadastral_number_unique ON fields (cadastral_number) WHERE cadastral_number IS NOT NULL;

DROP INDEX IF EXISTS implements_code_unique_idx;
CREATE UNIQUE INDEX implements_code_unique_idx ON implements (code);
