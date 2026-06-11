-- 1. Roles
CREATE TABLE IF NOT EXISTS roles (
    id          BIGSERIAL    PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- 2. Permissions
CREATE TABLE IF NOT EXISTS permissions (
    id          BIGSERIAL    PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT         NOT NULL DEFAULT ''
);

-- 3. Role ↔ Permissions (M2M)
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       BIGINT NOT NULL REFERENCES roles(id)       ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- 4. Adaugă role_id pe users (nullable în prima fază pentru migrare sigură)
ALTER TABLE users ADD COLUMN IF NOT EXISTS role_id BIGINT REFERENCES roles(id) ON DELETE SET NULL;

-- Index pentru query-ul RequirePermission (role_id + permission name)
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_users_role_id            ON users(role_id);
