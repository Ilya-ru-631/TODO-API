package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"todo_api/internal/models"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func nullStringToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func (p *PostgresRepo) GetAll(ctx context.Context) ([]models.Task, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT id, title, description, done FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		var desc sql.NullString

		if err := rows.Scan(&t.ID, &t.Title, &desc, &t.Done); err != nil {
			return nil, err
		}

		t.Description = nullStringToPtr(desc)
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (p *PostgresRepo) GetByID(ctx context.Context, id int) (models.Task, error) {
	query := "SELECT id, title, description, done FROM tasks WHERE id = $1"

	row := p.db.QueryRowContext(ctx, query, id)
	var t models.Task
	var desc sql.NullString

	err := row.Scan(&t.ID, &t.Title, &desc, &t.Done)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Task{}, ErrTaskNotFound
		}
		return models.Task{}, err
	}

	t.Description = nullStringToPtr(desc)
	return t, nil
}

func (p *PostgresRepo) Create(ctx context.Context, t models.Task) (models.Task, error) {
	if t.Title == "" {
		return models.Task{}, fmt.Errorf("text should not be empty: %w", ErrValidation)
	}

	query := "INSERT INTO tasks (title, description, done) VALUES ($1, $2, $3) RETURNING id"

	row := p.db.QueryRowContext(ctx, query, t.Title, t.Description, t.Done)

	if err := row.Scan(&t.ID); err != nil {
		return models.Task{}, err
	}

	return t, nil
}

func (p *PostgresRepo) Update(ctx context.Context, id int, t models.Task) (models.Task, error) {
	if t.Title == "" {
		return models.Task{}, fmt.Errorf("text should not be empty: %w", ErrValidation)
	}

	query := "UPDATE tasks SET title = $1, description = $2, done = $3 WHERE id = $4"

	result, err := p.db.ExecContext(ctx, query, t.Title, t.Description, t.Done, id)

	if err != nil {
		return models.Task{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Task{}, err
	}

	if rowsAffected == 0 {
		return models.Task{}, ErrTaskNotFound
	}

	t.ID = id

	return t, nil
}

func (p *PostgresRepo) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM tasks WHERE id = $1"

	result, err := p.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil

}
