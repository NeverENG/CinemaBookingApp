-- 增量迁移：登录防爆破状态表

CREATE TABLE IF NOT EXISTS login_guards (
    id           BIGSERIAL PRIMARY KEY,
    scope        VARCHAR(16) NOT NULL,
    username     VARCHAR(64) NOT NULL,
    failed_count INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (scope, username)
);
