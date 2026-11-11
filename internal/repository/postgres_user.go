package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"todo_api/internal/models"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrUsernameTaken = errors.New("username is already taken")
)

type UserPostgresRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserPostgresRepository {
	return &UserPostgresRepository{db: db}
}

func (up *UserPostgresRepository) Create(ctx context.Context, u models.User) (models.User, error) {
	if u.Username == "" {
		return models.User{}, fmt.Errorf("username should not be empty: %w", ErrValidation)
	}

	query := "INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id"
	row := up.db.QueryRowContext(ctx, query, u.Username, u.PasswordHash)

	if err := row.Scan(&u.ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.User{}, errors.Join(ErrUsernameTaken, pgErr)
		}
		return models.User{}, err
	}

	return u, nil
}

func (up *UserPostgresRepository) GetByUsername(ctx context.Context, username string) (models.User, error) {
	if username == "" {
		return models.User{}, ErrUserNotFound
	}

	query := "SELECT id, username, password_hash, created_at FROM users WHERE username = $1"
	row := up.db.QueryRowContext(ctx, query, username)
	var u models.User

	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return models.User{}, ErrUserNotFound
	case err != nil:
		return models.User{}, fmt.Errorf("error for scan user: %w", err)
	}

	return u, nil
}
