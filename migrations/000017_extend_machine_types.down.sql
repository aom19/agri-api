ALTER TABLE machines
    DROP CONSTRAINT IF EXISTS machines_type_check;

ALTER TABLE machines
    ADD CONSTRAINT machines_type_check
    CHECK (type IN ('tractor', 'combine', 'drone', 'sprayer', 'other'));

UPDATE machines
SET type = 'other'
WHERE type IN ('car', 'small_truck');
