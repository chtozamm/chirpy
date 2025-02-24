-- +goose Up
ALTER TABLE users ADD COLUMN is_chirpy_red BOOL NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE users DELETE COLUMN is_chirpy_red;
