package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"todo_api/internal/models"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// databaseUrlTest - не используется, оставлена ради примера
const databaseUrlTest = "postgres://postgres:password@localhost:5432/todo_api_test?sslmode=disable"

// setupTestDB подключается к тестовой БД и очищает таблицы, чтобы каждый тест
// начинал с чистого состояния. Если TEST_DATABASE_URL не задана, тест
// пропускается (t.Skip), а не падает: так go test ./... работает и на машине
// без Postgres.
//
// ВНИМАНИЕ: функция делает TRUNCATE. TEST_DATABASE_URL должна указывать только
// на отдельную тестовую базу, никогда на рабочую.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbUrl := os.Getenv("TEST_DATABASE_URL")
	if dbUrl == "" {
		t.Skip("Specify the URL for the database")
	}

	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	// Cleanup выполнится после завершения теста, даже если он упал.
	// Подключение закрывается здесь, а не через defer в тесте.
	t.Cleanup(func() {
		err := db.Close()
		if err != nil {
			t.Errorf("error for closing db: %v", err)
		}
	})

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	// TRUNCATE быстро удаляет все строки. RESTART IDENTITY сбрасывает счётчики
	// id (SERIAL) к 1, чтобы id в тестах были предсказуемыми. CASCADE нужен,
	// потому что tasks ссылается на users внешним ключом: без него Postgres
	// откажется очищать users.
	query := "TRUNCATE tasks, users RESTART IDENTITY CASCADE"
	_, err = db.ExecContext(context.Background(), query)

	if err != nil {
		t.Fatalf("error for truncate: %v", err)
	}
	return db
}

// createTestUser вставляет пользователя напрямую SQL-запросом, минуя
// репозиторий: хелпер не должен зависеть от кода, который тестируется.
func createTestUser(t *testing.T, db *sql.DB, username string) int {
	t.Helper()
	var userID int

	query := "INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id"

	row := db.QueryRowContext(context.Background(), query, username, "123")

	if err := row.Scan(&userID); err != nil {
		t.Fatalf("error for scan user id: %v", err)
	}

	return userID
}

// createTestTask вставляет задачу напрямую SQL-запросом и возвращает её id.
// Нужна тестам, которым требуется готовая задача в базе (Get, Update, Delete).
func createTestTask(t *testing.T, db *sql.DB, mt models.Task) int {
	t.Helper()
	var taskID int
	query := "INSERT INTO tasks (title, description, done, user_id) VALUES ($1, $2, $3, $4) RETURNING id"

	row := db.QueryRowContext(context.Background(), query, mt.Title, mt.Description, mt.Done, mt.UserID)

	if err := row.Scan(&taskID); err != nil {
		t.Fatalf("error for scan task id: %v", err)
	}

	return taskID
}

// readTestTask читает задачу из базы напрямую SQL-запросом (а не через
// репозиторий), чтобы проверить фактическое состояние таблицы после операции.
func readTestTask(t *testing.T, db *sql.DB, taskID int) models.Task {
	t.Helper()

	var task models.Task
	var desc sql.NullString

	query := "SELECT id, title, description, done, user_id FROM tasks WHERE id = $1"
	row := db.QueryRowContext(context.Background(), query, taskID)

	if err := row.Scan(&task.ID, &task.Title, &desc, &task.Done, &task.UserID); err != nil {
		t.Fatalf("failed to read task %d: %v", taskID, err)
	}

	task.Description = nullStringToPtr(desc)
	return task
}

// достаёт юзера из бд
func readTestUser(t *testing.T, db *sql.DB, userID int) models.User {
	t.Helper()

	var user models.User

	query := "SELECT id, username, password_hash, created_at FROM users WHERE id = $1"
	row := db.QueryRowContext(context.Background(), query, userID)

	if err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt); err != nil {
		t.Fatalf("failed to read user %d: %v", userID, err)
	}

	return user
}

// считает количество пользователей с одиннаковым именем
func countUserInDB(t *testing.T, db *sql.DB, username string) int {
	t.Helper()

	count := 0

	query := "SELECT COUNT(*) FROM users WHERE username = $1"
	row := db.QueryRowContext(context.Background(), query, username)

	if err := row.Scan(&count); err != nil {
		t.Fatalf("failed to count users %q: %v", username, err)
	}

	return count
}
