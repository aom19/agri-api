CREATE TABLE IF NOT EXISTS user_profiles (
    user_id         BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    first_name      VARCHAR(100),
    last_name       VARCHAR(100),
    date_of_birth   DATE,
    profile_photo   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
