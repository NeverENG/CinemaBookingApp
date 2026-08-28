-- 首页 banner（migrations004）

CREATE TABLE IF NOT EXISTS banners (
    id          BIGSERIAL PRIMARY KEY,
    title       VARCHAR(128) NOT NULL,
    image_url   VARCHAR(512) NOT NULL,
    sort        INT NOT NULL DEFAULT 0,
    enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
