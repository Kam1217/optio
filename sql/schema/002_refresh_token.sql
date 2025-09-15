-- +goose Up
CREATE TABLE refresh_token(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  issued_at TIMESTAMPTZ  DEFAULT NOW() NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  user_agent TEXT NOT NULL,
  ip TEXT NOT NULL
);

-- +goose Down
DROP TABLE refresh_token;