package repository

import (
	"context"
	"fmt"
	"sync"
	"todo_api/internal/models"
)

type MemoryRepo struct {
	notes  map[int]models.Task
	nextID int
	mu     sync.RWMutex
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		notes:  make(map[int]models.Task),
		nextID: 1,
	}
}

func (m *MemoryRepo) GetAll(ctx context.Context) ([]models.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]models.Task, 0, len(m.notes))
	for _, v := range m.notes {
		tasks = append(tasks, v)
	}

	return tasks, nil
}

func (m *MemoryRepo) GetByID(ctx context.Context, id int) (models.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.notes[id]
	if !ok {
		return models.Task{}, fmt.Errorf("ID must be a number: %w", ErrTaskNotFound)
	}

	return t, nil
}

func (m *MemoryRepo) Create(ctx context.Context, t models.Task) (models.Task, error) {
	if t.Title == "" {
		return models.Task{}, fmt.Errorf("Text should not be empty: %w", ErrValidation)
	}

	m.mu.Lock()
	t.ID = m.nextID
	m.notes[t.ID] = t
	m.nextID++
	m.mu.Unlock()

	return t, nil

}

func (m *MemoryRepo) Update(ctx context.Context, id int, t models.Task) (models.Task, error) {
	m.mu.RLock()
	value, ok := m.notes[id]
	m.mu.RUnlock()
	if !ok {
		return models.Task{}, fmt.Errorf("Tasks not found: %w", ErrTaskNotFound)
	}

	m.mu.Lock()
	t.ID = value.ID
	m.notes[value.ID] = t
	m.mu.Unlock()
	return t, nil
}

func (m *MemoryRepo) Delete(ctx context.Context, id int) error {
	m.mu.RLock()
	_, ok := m.notes[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("Tasks not found: %w", ErrTaskNotFound)
	}

	m.mu.Lock()
	delete(m.notes, id)
	m.mu.Unlock()

	return nil
}
