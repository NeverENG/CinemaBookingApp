-- 退款单（migrations003_refunds：补充丢失的建表）

CREATE TABLE IF NOT EXISTS refunds (
    id                 BIGSERIAL PRIMARY KEY,
    refund_no          VARCHAR(32) NOT NULL UNIQUE,
    order_no           VARCHAR(32) NOT NULL UNIQUE REFERENCES orders(order_no),
    user_id            BIGINT NOT NULL REFERENCES users(id),
    amount_cents       BIGINT NOT NULL CHECK (amount_cents > 0),
    reason             VARCHAR(255),
    status             VARCHAR(16) NOT NULL DEFAULT 'PENDING',
    external_refund_no VARCHAR(64) NOT NULL UNIQUE,
    refunded_at        TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
