package repository

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"todo_api/internal/models"
)

func strPtr(s string) *string {
	return &s
}

func Test_CreateMemory(t *testing.T) {
	repo := NewMemoryRepo()

	if repo.tasks == nil || repo.nextID != 1 {
		t.Errorf("error for create memory")
	}
}

func Test_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(repo *MemoryRepo) (int, models.Task)
		wantErr error
	}{
		{
			name: "task_exists",
			setup: func(repo *MemoryRepo) (int, models.Task) {
				task, _ := repo.Create(context.Background(), models.Task{
					Title:       "Homework",
					Description: strPtr("Go it before tomorrow"),
					Done:        false,
				})
				return task.ID, task
			},
			wantErr: nil,
		},
		{
			name:    "task_not_found",
			setup:   func(repo *MemoryRepo) (int, models.Task) { return 10, models.Task{} },
			wantErr: ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMemoryRepo()
			id, wantTask := tt.setup(repo)
			got, err := repo.GetByID(context.Background(), id)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got: %v, want: %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !reflect.DeepEqual(got, wantTask) {
				t.Errorf("incorrect search by id")
			}
		})
	}
}

func Test_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   models.Task
		wantErr error
	}{
		{
			name:    "errValidate",
			input:   models.Task{Title: "", Done: false},
			wantErr: ErrValidation,
		},
		{
			name:    "good_validate",
			input:   models.Task{Title: "Homework", Done: false},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMemoryRepo()
			task, err := repo.Create(context.Background(), tt.input)

			if tt.wantErr == nil && task.ID == 0 {
				t.Errorf("expected non-zero ID after successful create, got 0")
			}

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got: %v, want: %v", err, tt.wantErr)
			}

		})
	}
}

func Test_GetAll(t *testing.T) {
	repo := NewMemoryRepo()

	tasks, err := repo.GetAll(context.Background())
	if err != nil {
		t.Errorf("error when displaying tasks")
	}

	if len(tasks) != 0 {
		t.Errorf("error in the number of tasks")
	}

	repo.Create(context.Background(), models.Task{
		Title:       "homework",
		Description: strPtr("Go it before tomorrow"),
		Done:        false,
	})

	repo.Create(context.Background(), models.Task{
		Title:       "dinner",
		Description: strPtr("Go it before 20:00"),
		Done:        false,
	})

	tasks2, _ := repo.GetAll(context.Background())
	if len(tasks2) != 2 {
		t.Errorf("error in the number of tasks")
	}

}

func Test_Update(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(repo *MemoryRepo) int
		updateWith models.Task
		wantErr    error
	}{
		{
			name: "task-exists",
			setup: func(repo *MemoryRepo) int {
				task, _ := repo.Create(context.Background(), models.Task{
					Title:       "homework",
					Description: strPtr("Go it before tomorrow"),
					Done:        false,
				})
				return task.ID
			},
			updateWith: models.Task{
				Title:       "homework",
				Description: strPtr("made homework at 19:00"),
				Done:        true,
			},
			wantErr: nil,
		},
		{
			name:       "task-no-exists",
			setup:      func(repo *MemoryRepo) int { return 10 },
			updateWith: models.Task{},
			wantErr:    ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMemoryRepo()
			id := tt.setup(repo)
			want := tt.updateWith
			want.ID = id
			task, err := repo.Update(context.Background(), id, tt.updateWith)

			if tt.wantErr == nil {
				if !reflect.DeepEqual(task, want) {
					t.Errorf("issue update error")
				}
			}

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got: %v, want: %v", err, tt.wantErr)
			}

		})
	}
}

func Test_Delete(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(repo *MemoryRepo) int
		wantErr error
	}{
		{
			name: "deletion-was-successful",
			setup: func(repo *MemoryRepo) int {
				task, _ := repo.Create(context.Background(), models.Task{
					Title:       "Homework",
					Description: strPtr("Go it before tomorrow"),
					Done:        false,
				})
				return task.ID
			},
			wantErr: nil,
		},
		{
			name:    "task not found",
			setup:   func(repo *MemoryRepo) int { return 10 },
			wantErr: ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMemoryRepo()
			id := tt.setup(repo)
			err := repo.Delete(context.Background(), id)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got: %v, want: %v", err, tt.wantErr)
			}

			if tt.wantErr == nil {
				_, err := repo.GetByID(context.Background(), id)
				if !errors.Is(err, ErrTaskNotFound) {
					t.Fatalf("got: %v, want: %v", err, ErrTaskNotFound)
				}
			}

		})
	}
}

func Test_concurrencyCreate(t *testing.T) {
	repo := NewMemoryRepo()
	var wg sync.WaitGroup

	wg.Add(50)
	for i := 0; i < 50; i++ {
		go func() {
			defer wg.Done()
			repo.Create(context.Background(), models.Task{
				Title:       "Homework",
				Description: strPtr("Go it before tomorrow"),
				Done:        false,
			})
		}()
	}
	wg.Wait()

	tasks, _ := repo.GetAll(context.Background())
	if len(tasks) != 50 {
		t.Errorf("error when creating multiple tasks")
	}
}
