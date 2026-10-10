-- T22: regula de ștergere, descrisă în README („Ștergere”).
--
-- Terenurile, mașinile, echipamentele, template-urile, conturile și operațiunile pe teren se
-- arhivează (deleted_at): istoricul le păstrează numele, dar nu mai pot fi alese în lucrări noi.
-- Codul sau numărul cadastral al unui element arhivat se poate refolosi, ca la mașini.
DROP INDEX IF EXISTS implements_code_unique_idx;
CREATE UNIQUE INDEX implements_code_unique_idx ON implements (code) WHERE deleted_at IS NULL;

DROP INDEX IF EXISTS idx_fields_cadastral_number_unique;
CREATE UNIQUE INDEX idx_fields_cadastral_number_unique ON fields (cadastral_number)
    WHERE cadastral_number IS NOT NULL AND deleted_at IS NULL;

-- Restul (sezoane, culturi, culturi pe terenuri, resurse) se șterge definitiv, dar doar fără
-- istoric. Ștergerea unui sezon ștergea în cascadă culturile lui pe terenuri, iar ștergerea unei
-- culturi pe teren desfăcea recolta din stoc de cultura din care venea. Acum baza le refuză.
ALTER TABLE field_crops
    DROP CONSTRAINT field_crops_season_id_fkey,
    ADD CONSTRAINT field_crops_season_id_fkey FOREIGN KEY (season_id) REFERENCES seasons(id);

ALTER TABLE stock_movements
    DROP CONSTRAINT stock_movements_field_crop_id_fkey,
    ADD CONSTRAINT stock_movements_field_crop_id_fkey FOREIGN KEY (field_crop_id) REFERENCES field_crops(id);
