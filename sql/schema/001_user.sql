-- +goose Up
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE user (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username CITEXT UNIQUE NOT NULL,
    email CITEXT UNIQUE NOT NULL,
    password_hash VARCHAR (250) NOT NULL,
    password_changed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ,
    CHECK (length(username) BETWEEN 3 AND 50),
    CHECK (position('@' in email) > 1)
);

CREATE INDEX users_not_deleted_idx ON users (deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX users_created_at_id_desc_idx ON users (created_at DESC, id DESC);

-- +goose Down
DROP TABLE  IF EXISTS user;