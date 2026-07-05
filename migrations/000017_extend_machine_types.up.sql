ALTER TABLE machines
    DROP CONSTRAINT IF EXISTS machines_type_check;

UPDATE machines
SET type = 'car'
WHERE brand = 'Dacia' AND model = 'Duster' AND type = 'other';

UPDATE machines
SET type = 'small_truck'
WHERE brand = 'Mercedes' AND model = 'Sprinter' AND type = 'other';

ALTER TABLE machines
    ADD CONSTRAINT machines_type_check
    CHECK (type IN ('tractor', 'combine', 'drone', 'sprayer', 'car', 'small_truck', 'other'));
