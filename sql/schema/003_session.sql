-- +goose Up

CREATE TABLE session (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), 
    session_code VARCHAR(100) UNIQUE NOT NULL,
    session_name VARCHAR(250) NOT NULL,
    creator_user_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    FOREIGN KEY (creator_user_id) REFERENCES users(id)
);

-- +goose Down
DROP TABLE IF EXISTS session;