package service

import (
	"context"
	"errors"
	"testing"
	"todo_api/internal/models"
)

type fakeRepo struct {
	GetAllFunc  func(ctx context.Context, userID int) ([]models.Task, error)
	GetByIDFunc func(ctx context.Context, id, userID int) (models.Task, error)
	CreateFunc  func(ctx context.Context, t models.Task) (models.Task, error)
	UpdateFunc  func(ctx context.Context, id, userID int, t models.Task) (models.Task, error)
	DeleteFunc  func(ctx context.Context, id, userID int) error
}

func (f *fakeRepo) GetAll(ctx context.Context, userID int) ([]models.Task, error) {
	return f.GetAllFunc(ctx, userID)
}

func (f *fakeRepo) GetByID(ctx context.Context, id, userID int) (models.Task, error) {
	return f.GetByIDFunc(ctx, id, userID)
}

func (f *fakeRepo) Create(ctx context.Context, t models.Task) (models.Task, error) {
	return f.CreateFunc(ctx, t)
}

func (f *fakeRepo) Update(ctx context.Context, id, userID int, t models.Task) (models.Task, error) {
	return f.UpdateFunc(ctx, id, userID, t)
}

func (f *fakeRepo) Delete(ctx context.Context, id, userID int) error {
	return f.DeleteFunc(ctx, id, userID)
}

func Test_Create_Duplicate(t *testing.T) {
	wantUserID := 5

	tests := []struct {
		name        string
		existing    []models.Task
		wantErr     error
		wantCreated bool
	}{
		{
			name:        "title_matched",
			existing:    []models.Task{{Title: "Dinner", Done: false}},
			wantErr:     ErrDuplicateTitle,
			wantCreated: false,
		},
		{
			name:        "title_match_done_true",
			existing:    []models.Task{{Title: "Dinner", Done: true}},
			wantErr:     nil,
			wantCreated: true,
		},
		{
			name:        "title_unmatched",
			existing:    []models.Task{{Title: "Homework"}},
			wantErr:     nil,
			wantCreated: true,
		},
		{
			name:        "empty",
			existing:    []models.Task{},
			wantErr:     nil,
			wantCreated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created := false
			repo := &fakeRepo{
				GetAllFunc: func(ctx context.Context, userID int) ([]models.Task, error) {
					if userID != wantUserID {
						t.Errorf("got: %v, want: %v", userID, wantUserID)
					}
					res := tt.existing
					return res, nil
				},
				CreateFunc: func(ctx context.Context, mt models.Task) (models.Task, error) {
					if mt.UserID != wantUserID {
						t.Errorf("got: %v, want: %v", mt.UserID, wantUserID)
					}
					created = true
					return mt, nil
				},
			}

			svc := NewTaskService(repo)
			taskForCreate := models.Task{Title: "Dinner", Done: false}

			_, err := svc.Create(context.Background(), wantUserID, taskForCreate)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got: %v, want: %v", err, tt.wantErr)
			}

			if created != tt.wantCreated {
				t.Errorf("repo.Create called: %v, want: %v", created, tt.wantCreated)
			}
		})
	}
}

func Test_Create_RepoErrors(t *testing.T) {
	errDB := errors.New("db down")
	tests := []struct {
		name        string
		getAllErr   error
		createErr   error
		wantErr     error
		wantCreated bool
	}{
		{
			name:        "getall_fails",
			getAllErr:   errDB,
			createErr:   nil,
			wantErr:     errDB,
			wantCreated: false,
		},
		{
			name:        "create_fails",
			getAllErr:   nil,
			createErr:   errDB,
			wantErr:     errDB,
			wantCreated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created := false
			repo := &fakeRepo{
				GetAllFunc: func(ctx context.Context, userID int) ([]models.Task, error) {
					return nil, tt.getAllErr
				},
				CreateFunc: func(ctx context.Context, mt models.Task) (models.Task, error) {
					created = true
					return models.Task{}, tt.createErr
				},
			}

			svc := NewTaskService(repo)

			_, err := svc.Create(context.Background(), 1, models.Task{Title: "Dinner"})

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got: %v, want: %v", err, tt.wantErr)
			}

			if created != tt.wantCreated {
				t.Errorf("repo.Create called: %v, want: %v", created, tt.wantCreated)
			}
		})
	}
}
