-- T9 (varianta A): operatorul este un utilizator cu rolul `operator`, cu o singură înregistrare
-- (users) și un singur profil (user_profiles). Tabelul operators dispare.

-- Câmpurile proprii operatorului trec pe profil.
ALTER TABLE user_profiles
    ADD COLUMN IF NOT EXISTS phone                 TEXT,
    ADD COLUMN IF NOT EXISTS notes                 TEXT,
    ADD COLUMN IF NOT EXISTS allowed_machine_types TEXT[] NOT NULL DEFAULT '{}';

-- 1. Operatorii nelegați, dar cu e-mailul unui cont existent, sunt legați de acel cont.
UPDATE operators op
SET user_id = u.id
FROM users u
WHERE op.user_id IS NULL
  AND NULLIF(BTRIM(op.email), '') IS NOT NULL
  AND LOWER(BTRIM(op.email)) = LOWER(u.email)
  AND NOT EXISTS (SELECT 1 FROM operators other WHERE other.user_id = u.id);

-- 2. Ceilalți primesc un cont nou cu rolul operator, fără parolă utilizabilă („!” nu e un hash
--    bcrypt valid). Cei fără e-mail (sau cu un e-mail deja folosit) primesc o adresă tehnică
--    @fara-email.local, pe care adminul o înlocuiește cu cea reală. Operatorii inactivi sau
--    șterși devin conturi dezactivate.
DO $$
DECLARE
    op            RECORD;
    new_user_id   BIGINT;
    email_value   TEXT;
    operator_role BIGINT := (SELECT id FROM roles WHERE code = 'operator');
BEGIN
    -- conturile inserate cu id explicit (ex. din seed-uri) pot lăsa secvența în urmă
    PERFORM setval(pg_get_serial_sequence('users', 'id'), GREATEST((SELECT MAX(id) FROM users), 1));

    FOR op IN SELECT * FROM operators WHERE user_id IS NULL ORDER BY id LOOP
        email_value := NULLIF(BTRIM(op.email), '');
        IF email_value IS NULL OR EXISTS (SELECT 1 FROM users WHERE LOWER(email) = LOWER(email_value)) THEN
            email_value := 'operator-' || op.id || '@fara-email.local';
        END IF;

        INSERT INTO users (email, password_hash, role_id, email_confirmed, created_at, updated_at, deleted_at)
        VALUES (
            email_value, '!', operator_role, FALSE, op.created_at, op.updated_at,
            CASE WHEN op.deleted_at IS NOT NULL OR op.status <> 'active' THEN COALESCE(op.deleted_at, NOW()) END
        )
        RETURNING id INTO new_user_id;

        UPDATE operators SET user_id = new_user_id WHERE id = op.id;
    END LOOP;
END $$;

-- Un cont legat de un operator, dar cu rolul viewer (înregistrare publică), devine operator.
-- Rolurile admin și manager rămân neschimbate.
UPDATE users u
SET role_id = (SELECT id FROM roles WHERE code = 'operator'), updated_at = NOW()
FROM operators op
WHERE op.user_id = u.id
  AND u.role_id = (SELECT id FROM roles WHERE code = 'viewer');

-- 3. Profilul: numele operatorului (primul cuvânt = prenume, restul = nume) completează un profil
--    gol; telefonul, observațiile și tipurile de mașini vin de la operator.
INSERT INTO user_profiles (user_id, first_name, last_name, phone, notes, allowed_machine_types)
SELECT
    op.user_id,
    SPLIT_PART(BTRIM(op.name), ' ', 1),
    NULLIF(BTRIM(SUBSTRING(BTRIM(op.name) FROM LENGTH(SPLIT_PART(BTRIM(op.name), ' ', 1)) + 1)), ''),
    NULLIF(BTRIM(op.phone), ''),
    NULLIF(BTRIM(op.notes), ''),
    op.allowed_machine_types
FROM operators op
ON CONFLICT (user_id) DO UPDATE SET
    first_name            = COALESCE(NULLIF(user_profiles.first_name, ''), EXCLUDED.first_name),
    last_name             = COALESCE(NULLIF(user_profiles.last_name, ''), EXCLUDED.last_name),
    phone                 = EXCLUDED.phone,
    notes                 = EXCLUDED.notes,
    allowed_machine_types = EXCLUDED.allowed_machine_types,
    updated_at            = NOW();

-- 4. Lucrările și jurnalul de audit trec de la id-ul operatorului la id-ul contului.
ALTER TABLE field_operations DROP CONSTRAINT IF EXISTS field_operations_operator_id_fkey;

UPDATE field_operations fo
SET operator_id = op.user_id
FROM operators op
WHERE fo.operator_id = op.id;

ALTER TABLE field_operations
    ADD CONSTRAINT field_operations_operator_id_fkey FOREIGN KEY (operator_id) REFERENCES users(id);

UPDATE audit_log al
SET entity_id = op.user_id::text
FROM operators op
WHERE al.entity_type = 'operator' AND al.entity_id = op.id::text;

UPDATE notifications n
SET entity_id = op.user_id::text
FROM operators op
WHERE n.entity_type = 'operator' AND n.entity_id = op.id::text;

DROP TABLE operators;
