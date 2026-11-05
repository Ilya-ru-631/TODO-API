-- +goose Up
-- дата у старых строк приблизительная (время миграции)
UPDATE users SET created_at = now() WHERE created_at IS NULL;
ALTER TABLE users ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE users ALTER COLUMN created_at SET DEFAULT now();

-- +goose Down
ALTER TABLE users ALTER COLUMN created_at DROP DEFAULT;
ALTER TABLE users ALTER COLUMN created_at DROP NOT NULL;
-- старые NULL заменены текущим временем, а настоящее время создания неизвестно, поэтому вернуть их нельзя