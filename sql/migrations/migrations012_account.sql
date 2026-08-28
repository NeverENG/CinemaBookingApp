-- 邮箱账号 + 密码重置 + 会员等级（migrations012）

ALTER TABLE users ADD COLUMN IF NOT EXISTS email VARCHAR(128);
UPDATE users SET email = username WHERE email IS NULL;
ALTER TABLE users ALTER COLUMN email SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_users_email ON users (email);

CREATE TABLE IF NOT EXISTS password_reset_codes (
    id          BIGSERIAL PRIMARY KEY,
    email       VARCHAR(128) NOT NULL,
    code_hash   VARCHAR(255) NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_password_reset_email ON password_reset_codes (email, id DESC);

CREATE TABLE IF NOT EXISTS membership_levels (
    id                BIGSERIAL PRIMARY KEY,
    level_code        VARCHAR(32) NOT NULL UNIQUE,
    name              VARCHAR(32) NOT NULL,
    min_total_points  INT NOT NULL DEFAULT 0,
    discount_bp       INT NOT NULL DEFAULT 10000,
    status            VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO membership_levels (level_code, name, min_total_points, discount_bp) VALUES
    ('BRONZE', '青铜', 0, 10000),
    ('SILVER', '白银', 1000, 9500),
    ('GOLD', '黄金', 5000, 9000),
    ('DIAMOND', '钻石', 20000, 8500)
ON CONFLICT (level_code) DO NOTHING;

CREATE TABLE IF NOT EXISTS membership_level_logs (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT NOT NULL REFERENCES users(id),
    from_level_id  BIGINT REFERENCES membership_levels(id),
    to_level_id    BIGINT NOT NULL REFERENCES membership_levels(id),
    change_type    VARCHAR(32) NOT NULL DEFAULT 'UPGRADE',
    reason         VARCHAR(255),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
