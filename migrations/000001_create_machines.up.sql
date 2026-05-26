CREATE TABLE IF NOT EXISTS machines (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    type        VARCHAR(255) NOT NULL,
    status      VARCHAR(50)  NOT NULL DEFAULT 'available',
    description TEXT
);
