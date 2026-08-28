-- 认证与管理员（migrations002）

CREATE TABLE IF NOT EXISTS roles (
    id                BIGSERIAL PRIMARY KEY,
    code              VARCHAR(32) NOT NULL UNIQUE,
    name              VARCHAR(32) NOT NULL,
    permissions_json  JSONB NOT NULL DEFAULT '[]',
    status            VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO roles (code, name) VALUES
    ('SUPER_ADMIN', '超级管理员'),
    ('CINEMA_ADMIN', '影院管理员'),
    ('FINANCE', '财务')
ON CONFLICT (code) DO NOTHING;

CREATE TABLE IF NOT EXISTS admins (
    id             BIGSERIAL PRIMARY KEY,
    username       VARCHAR(64) NOT NULL UNIQUE,
    password_hash  VARCHAR(255) NOT NULL,
    nickname       VARCHAR(64) NOT NULL,
    avatar_url     VARCHAR(512),
    role_id        BIGINT NOT NULL REFERENCES roles(id),
    cinema_id      BIGINT REFERENCES cinemas(id),
    status         VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    last_login_at  TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS operation_logs (
    id           BIGSERIAL PRIMARY KEY,
    admin_id     BIGINT NOT NULL REFERENCES admins(id),
    action       VARCHAR(64) NOT NULL,
    target_type  VARCHAR(32) NOT NULL,
    target_id    VARCHAR(64) NOT NULL,
    detail_json  JSONB NOT NULL DEFAULT '{}',
    ip           VARCHAR(45),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_operation_logs_admin ON operation_logs (admin_id, created_at DESC);
