package service

import (
	"context"
	"errors"
	"todo_api/internal/models"
	"todo_api/internal/repository"
)

type TaskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

var (
	ErrDuplicateTitle    = errors.New("a task with such a title already exists")
	ErrActiveTitleActive = errors.New("active tasks title already exists")
	ErrTaskNotFound      = errors.New("taks nopt found")
	ErrValidation        = errors.New("error validation")
)

func (s *TaskService) Create(ctx context.Context, userID int, t models.Task) (models.Task, error) {
	tasks, err := s.repo.GetAll(ctx, userID)
	if err != nil {
		return models.Task{}, err
	}

	for _, task := range tasks {
		if task.Title == t.Title && !task.Done {
			return models.Task{}, ErrDuplicateTitle
		}
	}

	t.UserID = userID
	return TaskResult(s.repo.Create(ctx, t))
}

func (s *TaskService) GetAll(ctx context.Context, userID int) ([]models.Task, error) {
	tasks, err := s.repo.GetAll(ctx, userID)
	if err != nil {
		return []models.Task{}, err
	}
	return tasks, nil
}

func (s *TaskService) GetByID(ctx context.Context, id, userID int) (models.Task, error) {
	task, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return models.Task{}, err
	}
	return task, nil
}

func (s *TaskService) Delete(ctx context.Context, id, userID int) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *TaskService) Update(ctx context.Context, id, userID int, t models.Task) (models.Task, error) {
	//Заглушка, далее когда появится поле is-progress нужно будет поменять данную функцию
	return TaskResult(s.repo.Update(ctx, id, userID, t))
}

func TaskResult(t models.Task, err error) (models.Task, error) {
	if errors.Is(err, ErrActiveTitleActive) {
		return models.Task{}, ErrDuplicateTitle
	}
	return t, nil
}
