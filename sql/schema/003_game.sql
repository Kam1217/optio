-- +goose Up

CREATE TABLE game (
    game_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_code VARCHAR(10) UNIQUE NOT NULL,
    game_name VARCHAR(250) NOT NULL,
    creator_user_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    FOREIGN KEY (creator_user_id) REFERENCES users(id)
);

-- +goose Down
DROP TABLE IF EXISTS game;