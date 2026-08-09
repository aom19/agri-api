DROP INDEX IF EXISTS idx_operators_user_id_unique;

ALTER TABLE operators
DROP COLUMN IF EXISTS user_id;
