-- +goose Up

CREATE TABLE custom_game (

);

-- +goose Down
DROP TABLE IF EXISTS custom_game;