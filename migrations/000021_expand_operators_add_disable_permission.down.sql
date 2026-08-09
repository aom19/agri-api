-- Remove operators:disable permission
DELETE FROM role_permissions
WHERE permission_id = (SELECT id FROM permissions WHERE name = 'operators:disable');

DELETE FROM permissions WHERE name = 'operators:disable';

-- Remove expanded columns from operators table
ALTER TABLE operators
    DROP COLUMN IF EXISTS phone,
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS notes;
