-- Rollback: recreează coloana legacy users.role pe baza users.role_id -> roles.name.

ALTER TABLE users
ADD COLUMN IF NOT EXISTS role VARCHAR(50) NOT NULL DEFAULT 'viewer';

UPDATE users u
SET role = r.name
FROM roles r
WHERE r.id = u.role_id;

-- role_id revine opțional (stare apropiată de migrarea anterioară).
ALTER TABLE users
ALTER COLUMN role_id DROP NOT NULL;
