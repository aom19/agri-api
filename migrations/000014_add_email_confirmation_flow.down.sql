DROP TABLE IF EXISTS email_confirmation_tokens;

ALTER TABLE users
    DROP COLUMN IF EXISTS email_confirmed;
