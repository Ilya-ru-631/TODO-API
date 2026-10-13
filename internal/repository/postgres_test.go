package repository

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"todo_api/internal/models"
)

func strPtr(s string) *string {
	return &s
}

func Test_GetByID(t *testing.T) {
	db := setupTestDB(t)
	realPostgresRepo := NewPostgresRepo(db)

	username := "first"
	userID := createTestUser(t, db, username)

	usernameDif := "different_user"
	userIdDif := createTestUser(t, db, usernameDif)

	homework := models.Task{
		UserID:      userID,
		Title:       "Homework",
		Description: strPtr("deadline at 19:00"),
		Done:        false,
	}
	homework.ID = createTestTask(t, db, homework)

	withoutDescription := models.Task{
		UserID:      userID,
		Title:       "Buy milk",
		Description: nil,
		Done:        false,
	}
	withoutDescription.ID = createTestTask(t, db, withoutDescription)

	doneTask := models.Task{
		UserID:      userID,
		Title:       "Call mom",
		Description: strPtr("done yesterday"),
		Done:        true,
	}
	doneTask.ID = createTestTask(t, db, doneTask)

	tests := []struct {
		name        string
		inputUserID int
		inputTaskID int
		wantTask    models.Task
		wantErr     error
	}{
		{
			name:        "good_task",
			inputUserID: userID,
			inputTaskID: homework.ID,
			wantTask:    homework,
			wantErr:     nil,
		},
		{
			name:        "without_description",
			inputUserID: userID,
			inputTaskID: withoutDescription.ID,
			wantTask:    withoutDescription,
		},
		{
			name:        "done_task",
			inputUserID: userID,
			inputTaskID: doneTask.ID,
			wantTask:    doneTask,
		},
		{
			name:        "different_users",
			inputUserID: userIdDif,
			inputTaskID: homework.ID,
			wantErr:     ErrTaskNotFound,
		},
		{
			name:        "wrong_task_id",
			inputUserID: userID,
			inputTaskID: 100,
			wantErr:     ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := realPostgresRepo.GetByID(context.Background(), tt.inputTaskID, tt.inputUserID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got error: %v, want error: %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(task, tt.wantTask) {
				t.Errorf("got task: %+v, want task: %+v", task, tt.wantTask)
			}
		})
	}

}
