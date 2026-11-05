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
	repo := NewPostgresRepo(db)

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
			task, err := repo.GetByID(context.Background(), tt.inputTaskID, tt.inputUserID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got error: %v, want error: %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(task, tt.wantTask) {
				t.Errorf("got task: %+v, want task: %+v", task, tt.wantTask)
			}
		})
	}
}

func Test_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgresRepo(db)

	username := "first"
	userID := createTestUser(t, db, username)

	ownerUsername := "owner"
	ownerUserID := createTestUser(t, db, ownerUsername)

	emptyArrUsername := "empty"
	emptyArrUserID := createTestUser(t, db, emptyArrUsername)

	var tasksFirst []models.Task
	var tasksOwner []models.Task

	firstTask := models.Task{
		UserID:      userID,
		Title:       "Homework",
		Description: strPtr("deadline at 19:00"),
		Done:        false,
	}
	firstTask.ID = createTestTask(t, db, firstTask)

	firstTask2 := models.Task{
		UserID: userID,
		Title:  "Dinner",
		Done:   false,
	}
	firstTask2.ID = createTestTask(t, db, firstTask2)

	secondTask := models.Task{
		UserID:      ownerUserID,
		Title:       "Go to school",
		Description: strPtr("start at 9.00"),
		Done:        true,
	}
	secondTask.ID = createTestTask(t, db, secondTask)

	tasksFirst = append(tasksFirst, firstTask, firstTask2)
	tasksOwner = append(tasksOwner, secondTask)

	cases := []struct {
		name     string
		userID   int
		arrTasks []models.Task
		wantErr  error
	}{
		{
			name:     "good_test",
			userID:   userID,
			arrTasks: tasksFirst,
			wantErr:  nil,
		},
		{
			name:     "owner_user",
			userID:   ownerUserID,
			arrTasks: tasksOwner,
			wantErr:  nil,
		},
		{
			name:     "empty_array",
			userID:   emptyArrUserID,
			arrTasks: []models.Task{},
			wantErr:  nil,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tasks, err := repo.GetAll(context.Background(), tt.userID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got error: %v, want error: %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tasks, tt.arrTasks) {
				t.Fatalf("got array for tasks: %v, want array for tasks: %v", tasks, tt.arrTasks)
			}
		})
	}
}

func Test_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgresRepo(db)

	username := "fitst"
	userID := createTestUser(t, db, username)

	firstTask := models.Task{
		UserID:      userID,
		Title:       "Homework",
		Description: strPtr("deadline at 19:00"),
		Done:        false,
	}

	cases := []struct {
		name    string
		taskCr  models.Task
		wantErr error
	}{
		{
			name: "good_test",
			taskCr: firstTask,
			wantErr: nil,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			task, err := repo.Create(context.Background(), tt.taskCr)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got error: %v, want error: %v", err, tt.wantErr)
			}

			dbTask := readTestTask(t, db, task.ID)

			if !reflect.DeepEqual(task, dbTask) {
				t.Errorf("got task: %v, want task: %v", task, dbTask)
			}
		})
	}
}
