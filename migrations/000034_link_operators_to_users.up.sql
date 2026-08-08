ALTER TABLE operators
ADD COLUMN IF NOT EXISTS user_id BIGINT REFERENCES users(id) ON DELETE SET NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_operators_user_id_unique
ON operators(user_id)
WHERE user_id IS NOT NULL;

UPDATE operators op
SET user_id = u.id
FROM users u
WHERE op.user_id IS NULL
  AND op.email IS NOT NULL
  AND btrim(op.email) <> ''
  AND LOWER(op.email) = LOWER(u.email);
