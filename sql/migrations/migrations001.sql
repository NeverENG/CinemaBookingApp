-- 下单链路建表（按 docs/database/数据库表.md）

CREATE TABLE IF NOT EXISTS users (
    id                   BIGSERIAL PRIMARY KEY,
    username             VARCHAR(64) NOT NULL UNIQUE,
    password_hash        VARCHAR(255) NOT NULL,
    nickname             VARCHAR(64) NOT NULL,
    avatar_url           VARCHAR(512),
    membership_level_id  BIGINT,
    points_balance       INT NOT NULL DEFAULT 0,
    total_earned_points  INT NOT NULL DEFAULT 0,
    status               VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at           TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS cinemas (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    city        VARCHAR(64) NOT NULL,
    address     VARCHAR(255) NOT NULL,
    longitude   DOUBLE PRECISION,
    latitude    DOUBLE PRECISION,
    status      VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS halls (
    id                BIGSERIAL PRIMARY KEY,
    cinema_id         BIGINT NOT NULL REFERENCES cinemas(id),
    name              VARCHAR(64) NOT NULL,
    seat_layout_json  JSONB NOT NULL DEFAULT '{}',
    status            VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS movies (
    id                BIGSERIAL PRIMARY KEY,
    title             VARCHAR(128) NOT NULL,
    cover_url         VARCHAR(512) NOT NULL,
    trailer_url       VARCHAR(512),
    description       TEXT,
    duration_minutes  INT,
    status            VARCHAR(16) NOT NULL DEFAULT 'ON_SALE',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS seats (
    id          BIGSERIAL PRIMARY KEY,
    hall_id     BIGINT NOT NULL REFERENCES halls(id),
    row_no      INT NOT NULL,
    col_no      INT NOT NULL,
    seat_no     VARCHAR(16) NOT NULL,
    type        VARCHAR(16) NOT NULL DEFAULT 'STANDARD',
    status      VARCHAR(16) NOT NULL DEFAULT 'ENABLED',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (hall_id, row_no, col_no)
);

CREATE TABLE IF NOT EXISTS show_sessions (
    id                BIGSERIAL PRIMARY KEY,
    cinema_id         BIGINT NOT NULL REFERENCES cinemas(id),
    hall_id           BIGINT NOT NULL REFERENCES halls(id),
    movie_id          BIGINT NOT NULL REFERENCES movies(id),
    start_time        TIMESTAMPTZ NOT NULL,
    end_time          TIMESTAMPTZ NOT NULL,
    base_price_cents  BIGINT NOT NULL CHECK (base_price_cents > 0),
    status            VARCHAR(16) NOT NULL DEFAULT 'OPEN',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_time > start_time)
);

CREATE INDEX IF NOT EXISTS idx_show_sessions_movie ON show_sessions (movie_id, start_time);
CREATE INDEX IF NOT EXISTS idx_show_sessions_cinema ON show_sessions (cinema_id, start_time);

CREATE TABLE IF NOT EXISTS seat_locks (
    id           BIGSERIAL PRIMARY KEY,
    session_id   BIGINT NOT NULL REFERENCES show_sessions(id),
    seat_id      BIGINT NOT NULL REFERENCES seats(id),
    user_id      BIGINT NOT NULL REFERENCES users(id),
    order_no     VARCHAR(32),
    lock_token   VARCHAR(64) NOT NULL,
    status       VARCHAR(16) NOT NULL DEFAULT 'LOCKED',
    expires_at   TIMESTAMPTZ NOT NULL,
    released_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 防超卖核心：同一场次同一座位同时只允许一个有效锁（LOCKED/BOOKED）
CREATE UNIQUE INDEX IF NOT EXISTS uq_seat_locks_active
    ON seat_locks (session_id, seat_id)
    WHERE status IN ('LOCKED', 'BOOKED');

CREATE TABLE IF NOT EXISTS orders (
    order_no           VARCHAR(32) PRIMARY KEY,
    user_id            BIGINT NOT NULL REFERENCES users(id),
    session_id         BIGINT NOT NULL REFERENCES show_sessions(id),
    cinema_id          BIGINT NOT NULL,
    movie_id           BIGINT NOT NULL,
    status             VARCHAR(24) NOT NULL DEFAULT 'PENDING_PAYMENT',
    total_cents        BIGINT NOT NULL DEFAULT 0,
    discount_cents     BIGINT NOT NULL DEFAULT 0,
    coupon_cents       BIGINT NOT NULL DEFAULT 0,
    paid_cents         BIGINT NOT NULL DEFAULT 0,
    coupon_instance_id BIGINT,
    expire_at          TIMESTAMPTZ NOT NULL,
    version            INT NOT NULL DEFAULT 1,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    paid_at            TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_orders_user ON orders (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_session ON orders (session_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders (status);

CREATE TABLE IF NOT EXISTS order_items (
    id           BIGSERIAL PRIMARY KEY,
    order_no     VARCHAR(32) NOT NULL REFERENCES orders(order_no),
    session_id   BIGINT NOT NULL,
    seat_id      BIGINT NOT NULL REFERENCES seats(id),
    seat_no      VARCHAR(16) NOT NULL,
    price_cents  BIGINT NOT NULL,
    ticket_no    VARCHAR(32) UNIQUE,
    used_at      TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS coupon_templates (
    id                  BIGSERIAL PRIMARY KEY,
    name                VARCHAR(64) NOT NULL,
    type                VARCHAR(16) NOT NULL,
    value_cents         BIGINT NOT NULL DEFAULT 0,
    percent_bp          INT NOT NULL DEFAULT 0,
    min_spend_cents     BIGINT NOT NULL DEFAULT 0,
    max_discount_cents  BIGINT NOT NULL DEFAULT 0,
    total_qty           INT NOT NULL DEFAULT 0,
    per_user_limit      INT NOT NULL DEFAULT 1,
    status              VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_coupons (
    id           BIGSERIAL PRIMARY KEY,
    coupon_no    VARCHAR(32) NOT NULL UNIQUE,
    template_id  BIGINT NOT NULL REFERENCES coupon_templates(id),
    user_id      BIGINT NOT NULL REFERENCES users(id),
    status       VARCHAR(16) NOT NULL DEFAULT 'UNUSED',
    order_no     VARCHAR(32),
    expire_at    TIMESTAMPTZ NOT NULL,
    used_at      TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_coupons_user_status ON user_coupons (user_id, status);

-- 支付交易：一单一付，biz_type + biz_no 唯一
CREATE TABLE IF NOT EXISTS payment_transactions (
    transaction_no    VARCHAR(32) PRIMARY KEY,
    biz_type          VARCHAR(16) NOT NULL DEFAULT 'ORDER_PAY',
    biz_no            VARCHAR(64) NOT NULL,
    user_id           BIGINT NOT NULL REFERENCES users(id),
    amount_cents      BIGINT NOT NULL CHECK (amount_cents > 0),
    channel           VARCHAR(32) NOT NULL DEFAULT 'MOCK_PAY',
    status            VARCHAR(16) NOT NULL DEFAULT 'PENDING',
    external_trade_no VARCHAR(64) UNIQUE,
    version           INT NOT NULL DEFAULT 1,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    paid_at           TIMESTAMPTZ,
    closed_at         TIMESTAMPTZ,
    UNIQUE (biz_type, biz_no)
);

-- 支付回调：event_id 唯一是幂等键，只增不改
CREATE TABLE IF NOT EXISTS payment_callbacks (
    id              BIGSERIAL PRIMARY KEY,
    event_id        VARCHAR(64) NOT NULL UNIQUE,
    transaction_no  VARCHAR(32) NOT NULL,
    amount_cents    BIGINT NOT NULL,
    payload         TEXT NOT NULL DEFAULT '',
    status          VARCHAR(16) NOT NULL DEFAULT 'RECEIVED',
    process_result  VARCHAR(512),
    retry_count     INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at    TIMESTAMPTZ
);

-- 兼容已建库：老表补 used_at
ALTER TABLE user_coupons ADD COLUMN IF NOT EXISTS used_at TIMESTAMPTZ;
