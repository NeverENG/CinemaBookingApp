-- 积分流水（migrations005）

CREATE TABLE IF NOT EXISTS points_ledger (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT NOT NULL REFERENCES users(id),
    change_points  INT NOT NULL,
    balance_after  INT NOT NULL,
    biz_type       VARCHAR(32) NOT NULL,
    biz_no         VARCHAR(64) NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (biz_type, biz_no)
);

CREATE INDEX IF NOT EXISTS idx_points_ledger_user ON points_ledger (user_id, id DESC);
