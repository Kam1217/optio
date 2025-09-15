-- +goose Up

CREATE TABLE session_participant (
    user_id UUID NOT NULL,
    session_id UUID NOT NULL,
    PRIMARY KEY (user_id, session_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (session_id) REFERENCES session(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    status VARCHAR(50) DEFAULT 'invited'
);

-- +goose Down
DROP TABLE IF EXISTS session_participant;