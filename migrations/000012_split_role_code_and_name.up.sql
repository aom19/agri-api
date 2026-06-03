-- roles.name a fost folosit până acum ca identificator tehnic.
-- Îl transformăm în roles.code, iar noul roles.name devine eticheta afișabilă în frontend.

ALTER TABLE roles RENAME COLUMN name TO code;

ALTER TABLE roles
ADD COLUMN IF NOT EXISTS name VARCHAR(100) NOT NULL DEFAULT '';

UPDATE roles
SET name = CASE code
	WHEN 'admin' THEN 'Administrator'
	WHEN 'manager' THEN 'Manager'
	WHEN 'viewer' THEN 'Vizualizator'
	ELSE initcap(replace(code, '_', ' '))
END
WHERE name = '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_code_unique ON roles(code);
