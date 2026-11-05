-- +goose Up
CREATE UNIQUE INDEX users_username_unique_idx ON users (username);

-- +goose Down
DROP INDEX users_username_unique_idx;
