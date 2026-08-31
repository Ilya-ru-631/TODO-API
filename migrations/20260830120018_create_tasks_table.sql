-- +goose Up
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    done BOOLEAN NOT NULL DEFAULT false
);


-- +goose Down
DROP TABLE tasks;
