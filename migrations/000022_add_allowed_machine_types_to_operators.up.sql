ALTER TABLE operators
    ADD COLUMN IF NOT EXISTS allowed_machine_types text[] NOT NULL DEFAULT '{}';
