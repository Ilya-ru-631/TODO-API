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
	ErrDuplicateTitle = errors.New("a task with such a title already exists")
)

func (s *TaskService) Create(ctx context.Context, t models.Task) (models.Task, error) {
	tasks, err := s.repo.GetAll(ctx)
	if err != nil {
		return models.Task{}, err
	}

	for _, task := range tasks {
		if task.Title == t.Title && !task.Done{
			return models.Task{}, ErrDuplicateTitle
		}
	}

	res, err := s.repo.Create(ctx, t)
	if err != nil {
		return models.Task{}, err
	}
	return res, nil
}


func(s *TaskService) GetAll(ctx context.Context) ([]models.Task, error) {
	tasks, err := s.repo.GetAll(ctx)
	if err != nil {
		return []models.Task{}, err
	}
	return tasks, nil
}

func (s *TaskService) GetByID(ctx context.Context, id int) (models.Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return models.Task{}, err
	}
	return task, nil
}

func(s *TaskService) Delete(ctx context.Context, id int) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}	

func(s *TaskService) Update(ctx context.Context, id int, t models.Task) (models.Task, error) {
	//Заглушка, далее когда появится поле is-progress нужно будет поменять данную функцию
	task, err := s.repo.Update(ctx, id, t)
		if err != nil {
		return models.Task{}, err
	}
	return task, nil
}