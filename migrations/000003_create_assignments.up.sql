CREATE TABLE IF NOT EXISTS assignments (
    id          BIGSERIAL       PRIMARY KEY,
    machine_id  BIGINT          NOT NULL REFERENCES machines(id),
    operator_id BIGINT          NOT NULL REFERENCES operators(id),
    start_date  TIMESTAMPTZ     NOT NULL,
    end_date    TIMESTAMPTZ,
    status      VARCHAR(50)     NOT NULL DEFAULT 'active'
);
