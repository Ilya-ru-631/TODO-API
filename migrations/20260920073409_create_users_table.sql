-- +goose Up
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ
);

-- +goose Down
DROP TABLE users;
