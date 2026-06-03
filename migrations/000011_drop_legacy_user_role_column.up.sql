-- Migrare de tranziție: mută valorile legacy users.role -> users.role_id,
-- apoi șterge coloana veche users.role.

-- Asigură existența rolului implicit pentru fallback.
INSERT INTO roles (name, description)
VALUES ('viewer', 'Fallback role for migrated users')
ON CONFLICT (name) DO NOTHING;

-- Creează în roles orice nume de rol existent în users.role (date legacy).
INSERT INTO roles (name, description)
SELECT DISTINCT u.role, 'Migrated from legacy users.role'
FROM users u
WHERE u.role IS NOT NULL AND btrim(u.role) <> ''
ON CONFLICT (name) DO NOTHING;

-- Mapează users.role_id din users.role când role_id este încă NULL.
UPDATE users u
SET role_id = r.id
FROM roles r
WHERE u.role_id IS NULL
  AND u.role IS NOT NULL
  AND btrim(u.role) <> ''
  AND r.name = u.role;

-- Fallback final: orice user rămas fără role_id devine viewer.
UPDATE users
SET role_id = (SELECT id FROM roles WHERE name = 'viewer')
WHERE role_id IS NULL;

-- După backfill, role_id devine obligatoriu.
ALTER TABLE users
ALTER COLUMN role_id SET NOT NULL;

-- Șterge coloana legacy.
ALTER TABLE users
DROP COLUMN IF EXISTS role;
