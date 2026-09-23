-- +goose Up
TRUNCATE TABLE tasks;

ALTER TABLE tasks
    ADD COLUMN user_id INT NOT NULL REFERENCES users(id);

-- +goose Down
ALTER TABLE tasks
    DROP COLUMN user_id;