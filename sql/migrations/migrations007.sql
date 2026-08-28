-- 票房看板（migrations007）

CREATE TABLE IF NOT EXISTS box_office_ledger (
    id            BIGSERIAL PRIMARY KEY,
    biz_type      VARCHAR(32) NOT NULL,
    biz_no        VARCHAR(64) NOT NULL,
    stat_date     DATE NOT NULL,
    cinema_id     BIGINT NOT NULL,
    movie_id      BIGINT NOT NULL,
    order_delta   INT NOT NULL DEFAULT 0,
    ticket_delta  INT NOT NULL DEFAULT 0,
    gross_delta   BIGINT NOT NULL DEFAULT 0,
    refund_delta  BIGINT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (biz_type, biz_no)
);

CREATE TABLE IF NOT EXISTS daily_box_office (
    id            BIGSERIAL PRIMARY KEY,
    stat_date     DATE NOT NULL,
    cinema_id     BIGINT NOT NULL,
    movie_id      BIGINT NOT NULL,
    order_count   INT NOT NULL DEFAULT 0,
    ticket_count  INT NOT NULL DEFAULT 0,
    gross_cents   BIGINT NOT NULL DEFAULT 0,
    refund_cents  BIGINT NOT NULL DEFAULT 0,
    net_cents     BIGINT NOT NULL DEFAULT 0,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (stat_date, cinema_id, movie_id)
);
