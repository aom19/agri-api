-- Reface tabelul operators din conturile cu rolul operator (și din cele asignate pe lucrări).
-- Conturile create de migrarea up rămân în users.
CREATE TABLE operators (
    id                    BIGSERIAL PRIMARY KEY,
    name                  VARCHAR(50) NOT NULL,
    status                VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ,
    phone                 TEXT,
    email                 TEXT,
    notes                 TEXT,
    allowed_machine_types TEXT[] NOT NULL DEFAULT '{}',
    user_id               BIGINT REFERENCES users(id) ON DELETE SET NULL
);
CREATE UNIQUE INDEX idx_operators_user_id_unique ON operators(user_id) WHERE user_id IS NOT NULL;

INSERT INTO operators (name, status, phone, email, notes, allowed_machine_types, user_id, created_at, updated_at)
SELECT
    LEFT(COALESCE(NULLIF(BTRIM(CONCAT_WS(' ', up.first_name, up.last_name)), ''), u.email), 50),
    CASE WHEN u.deleted_at IS NULL THEN 'active' ELSE 'inactive' END,
    up.phone,
    CASE WHEN u.email LIKE '%@fara-email.local' THEN NULL ELSE u.email END,
    up.notes,
    COALESCE(up.allowed_machine_types, '{}'),
    u.id,
    u.created_at,
    u.updated_at
FROM users u
LEFT JOIN user_profiles up ON up.user_id = u.id
WHERE u.role_id = (SELECT id FROM roles WHERE code = 'operator')
   OR u.id IN (SELECT operator_id FROM field_operations WHERE operator_id IS NOT NULL)
ORDER BY u.id;

ALTER TABLE field_operations DROP CONSTRAINT IF EXISTS field_operations_operator_id_fkey;

UPDATE field_operations fo
SET operator_id = op.id
FROM operators op
WHERE fo.operator_id = op.user_id;

ALTER TABLE field_operations
    ADD CONSTRAINT field_operations_operator_id_fkey FOREIGN KEY (operator_id) REFERENCES operators(id);

UPDATE audit_log al
SET entity_id = op.id::text
FROM operators op
WHERE al.entity_type = 'operator' AND al.entity_id = op.user_id::text;

UPDATE notifications n
SET entity_id = op.id::text
FROM operators op
WHERE n.entity_type = 'operator' AND n.entity_id = op.user_id::text;

ALTER TABLE user_profiles
    DROP COLUMN IF EXISTS phone,
    DROP COLUMN IF EXISTS notes,
    DROP COLUMN IF EXISTS allowed_machine_types;
