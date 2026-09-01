ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ;

UPDATE users
SET email_verified_at = created_at
WHERE email_verified_at IS NULL;

CREATE TABLE IF NOT EXISTS email_verification_codes (
    id          BIGSERIAL PRIMARY KEY,
    email       VARCHAR(128) NOT NULL,
    code_hash   VARCHAR(255) NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_email_verification_email
    ON email_verification_codes (email, id DESC);
