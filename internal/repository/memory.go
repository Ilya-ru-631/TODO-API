package repository

import (
	"context"
	"fmt"
	"sync"
	"todo_api/internal/models"
)

type MemoryRepo struct {
	tasks  map[int]models.Task
	nextID int
	mu     sync.RWMutex
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		tasks:  make(map[int]models.Task),
		nextID: 1,
	}
}

func (m *MemoryRepo) GetAll(ctx context.Context, userID int) ([]models.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]models.Task, 0, len(m.tasks))
	for _, v := range m.tasks {
		if v.UserID == userID {
			tasks = append(tasks, v)
		}
	}

	return tasks, nil
}

func (m *MemoryRepo) GetByID(ctx context.Context, id, userID int) (models.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	t, ok := m.tasks[id]
	if !ok {
		return models.Task{}, fmt.Errorf("task not a found: %w", ErrTaskNotFound)
	}

	if t.UserID != userID {
		return models.Task{}, fmt.Errorf("task not a found: %w", ErrTaskNotFound)
	}

	return t, nil
}

func (m *MemoryRepo) Create(ctx context.Context, t models.Task) (models.Task, error) {
	if t.Title == "" {
		return models.Task{}, fmt.Errorf("text should not be empty: %w", ErrValidation)
	}

	m.mu.Lock()
	t.ID = m.nextID
	m.tasks[t.ID] = t
	m.nextID++
	m.mu.Unlock()

	return t, nil

}

func (m *MemoryRepo) Update(ctx context.Context, id, userID int, t models.Task) (models.Task, error) {
	if t.Title == "" {
		return models.Task{}, fmt.Errorf("text should not be empty: %w", ErrValidation)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	value, ok := m.tasks[id]
	if !ok {
		return models.Task{}, fmt.Errorf("tasks not found: %w", ErrTaskNotFound)
	}

	if t.UserID != userID {
		return models.Task{}, fmt.Errorf("tasks not found: %w", ErrTaskNotFound)
	}		

	t.ID = value.ID
	m.tasks[value.ID] = t
	return t, nil
}

func (m *MemoryRepo) Delete(ctx context.Context, id, userID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[id]
	if !ok {
		return fmt.Errorf("Tasks not found: %w", ErrTaskNotFound)
	}

	if t.UserID != userID {
		return fmt.Errorf("tasks not found: %w", ErrTaskNotFound)
	}

	delete(m.tasks, id)
	return nil
}
