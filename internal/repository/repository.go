package repository

import (
	"context"
	"errors"
	"todo_api/internal/models"
)

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrValidation   = errors.New("validation failed")
	ErrUserNotFound = errors.New("user not found")
)

type TaskRepository interface {
	GetAll(ctx context.Context) ([]models.Task, error)
	GetByID(ctx context.Context, id int) (models.Task, error)
	Create(ctx context.Context, t models.Task) (models.Task, error)
	Update(ctx context.Context, id int, t models.Task) (models.Task, error)
	Delete(ctx context.Context, id int) error
}

type UserRepository interface {
	Create(ctx context.Context, u models.User) (models.User, error)
	GetByUsername(ctx context.Context, username string) (models.User, error)
}
