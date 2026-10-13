package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"todo_api/internal/models"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const databaseUrlTest = "postgres://postgres:password@localhost:5432/todo_api_test?sslmode=disable"

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	//dbUrl := database_url_test
	dbUrl := os.Getenv("TEST_DATABASE_URL")
	if dbUrl == "" {
		t.Skip("Specify the URL for the database")
	}

	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	query := "TRUNCATE tasks, users RESTART IDENTITY CASCADE"
	_, err = db.ExecContext(context.Background(), query)

	if err != nil {
		t.Fatalf("error for truncate: %v", err)
	}
	return db
}

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
