-- +goose Up

CREATE TABLE session_item (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL, 
    item_title VARCHAR(250) NOT NULL, 
    item_description TEXT, 
    image_url VARCHAR(250),
    source_type VARCHAR(50) NOT NULL,
    source_id VARCHAR(250),
    metadata JSONB, 
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    added_by_user_id UUID NOT NULL,
    FOREIGN KEY (session_id) REFERENCES session(id),
    FOREIGN KEY (added_by_user_id) REFERENCES users(id)
);

-- +goose Down
DROP TABLE IF EXISTS session_item;