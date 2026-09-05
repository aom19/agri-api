CREATE TABLE IF NOT EXISTS report_subscriptions (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    frequency     VARCHAR(10) NOT NULL CHECK (frequency IN ('daily', 'weekly', 'monthly')),
    send_hour     INT NOT NULL DEFAULT 7 CHECK (send_hour BETWEEN 0 AND 23),
    weekday       INT NOT NULL DEFAULT 1 CHECK (weekday BETWEEN 1 AND 7),
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    last_sent_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
