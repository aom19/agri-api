-- T24 „cultura deschisă pe teren”: lucrarea se leagă de cultură explicit, la creare, nu după
-- data planificată. Regula de business, impusă de bază: pe un teren e cel mult o cultură
-- deschisă; lucrările noi merg implicit la ea. Sezonul rămâne doar o etichetă de raportare.

-- Coloanele și backfill-ul rulează o singură dată: pe o bază care are deja statusul (de ex. pusă
-- la punct cu această migrare înainte de renumerotare), backfill-ul ar închide culturi deschise.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'field_crops' AND column_name = 'status'
    ) THEN
        ALTER TABLE field_crops
            ADD COLUMN status    TEXT NOT NULL DEFAULT 'open'
                CHECK (status IN ('open', 'closed', 'compromised')),
            ADD COLUMN closed_at DATE;

        -- Backfill: culturile recoltate sau din sezoane încheiate sunt închise.
        UPDATE field_crops fc
        SET status    = 'closed',
            closed_at = COALESCE(fc.harvested_at, s.end_date)
        FROM seasons s
        WHERE s.id = fc.season_id
          AND (fc.harvested_at IS NOT NULL OR s.end_date < CURRENT_DATE);

        -- Dacă un teren rămâne cu mai multe culturi deschise, rămâne deschisă cea din sezonul cel
        -- mai recent; celelalte se închid la sfârșitul sezonului lor.
        UPDATE field_crops fc
        SET status    = 'closed',
            closed_at = s.end_date
        FROM seasons s
        WHERE s.id = fc.season_id
          AND fc.status = 'open'
          AND fc.id <> (
              SELECT other.id
              FROM field_crops other
              JOIN seasons os ON os.id = other.season_id
              WHERE other.field_id = fc.field_id AND other.status = 'open'
              ORDER BY os.start_date DESC, other.id DESC
              LIMIT 1
          );
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS field_crops_one_open_per_field
    ON field_crops (field_id)
    WHERE status = 'open';

-- Un teren poate avea două culturi în același sezon (ex. grâu compromis, apoi porumb).
ALTER TABLE field_crops DROP CONSTRAINT IF EXISTS field_crops_field_id_season_id_key;

CREATE INDEX IF NOT EXISTS idx_field_crops_field_status ON field_crops (field_id, status);
