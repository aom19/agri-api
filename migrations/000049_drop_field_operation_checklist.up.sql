-- T10: checklistul de start avea 4 bife salvate, dar pornirea verifică oricum automat că mașina
-- și echipamentul sunt active. Rămâne o singură confirmare în pagină, care nu se salvează.
-- role_permissions se curăță prin ON DELETE CASCADE.
DELETE FROM permissions WHERE name = 'field_operations:checklist';

ALTER TABLE field_operations
    DROP COLUMN IF EXISTS check_machine_status,
    DROP COLUMN IF EXISTS check_implement_status,
    DROP COLUMN IF EXISTS check_field_area,
    DROP COLUMN IF EXISTS check_notes_confirmed,
    DROP COLUMN IF EXISTS checklist_updated_at;
