package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"todo_api/internal/logger"
	"todo_api/internal/models"
	"todo_api/internal/repository"
	"todo_api/internal/service"
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

func Test_GetAll(t *testing.T) {
	wantUserID := 1
	want := []models.Task{
		{ID: 1, UserID: wantUserID, Title: "Homework", Done: false},
		{ID: 2, UserID: wantUserID, Title: "Dinner", Done: true},
	}

	repo := &fakeRepo{
		GetAllFunc: func(ctx context.Context, userID int) ([]models.Task, error) {
			if userID != wantUserID {
				t.Errorf("got: %v, want: %v", userID, wantUserID)
			}
			return want, nil
		},
	}

	svc := service.NewTaskService(repo)
	sloger := logger.New("dev")
	h := NewTaskHandler(svc, sloger)

	req := httptest.NewRequest("GET", "/tasks/", nil)
	rec := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), userIDKey, wantUserID)
	req = req.WithContext(ctx)
	h.GetAll(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusOK)
	}

	var data []models.Task
	err := json.Unmarshal(rec.Body.Bytes(), &data)
	if err != nil {
		t.Errorf("error for parsing body: %v", err)
	}

	if !reflect.DeepEqual(data, want) {
		t.Errorf("got: %+v, want: %+v", data, want)
	}
}

func Test_Unauthorized(t *testing.T) {
	repo := &fakeRepo{
		GetAllFunc: func(ctx context.Context, userID int) ([]models.Task, error) {
			t.Errorf("repository must not be called without userID")
			return nil, nil
		},
		GetByIDFunc: func(ctx context.Context, id, userID int) (models.Task, error) {
			t.Errorf("repository must not be called without userID")
			return models.Task{}, nil
		},
		CreateFunc: func(ctx context.Context, mt models.Task) (models.Task, error) {
			t.Errorf("repository must not be called without userID")
			return models.Task{}, nil
		},
		UpdateFunc: func(ctx context.Context, id, userID int, mt models.Task) (models.Task, error) {
			t.Errorf("repository must not be called without userID")
			return models.Task{}, nil
		},
		DeleteFunc: func(ctx context.Context, id, userID int) error {
			t.Errorf("repository must not be called without userID")
			return nil
		},
	}

	svc := service.NewTaskService(repo)
	sloger := logger.New("dev")

	h := NewTaskHandler(svc, sloger)

	tests := []struct {
		name    string
		handler http.HandlerFunc
		method  string
		path    string
		body    string
	}{
		{
			name:    "GetAll",
			handler: h.GetAll,
			method:  "GET",
			path:    "/tasks/",
			body:    "",
		},
		{
			name:    "GetByID",
			handler: h.GetByID,
			method:  "GET",
			path:    "/tasks/2",
			body:    "",
		},
		{
			name:    "Create",
			handler: h.Create,
			method:  "POST",
			path:    "/tasks/",
			body:    `{"title" : "Homework"}`,
		},
		{
			name:    "Update",
			handler: h.Update,
			method:  "PUT",
			path:    "/tasks/2",
			body:    `{"title": "Dinner"}`,
		},
		{
			name:    "Delete",
			handler: h.Delete,
			method:  "DELETE",
			path:    "/tasks/2",
			body:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.name == "GetByID" || tt.name == "Update" || tt.name == "Delete" {
				req.SetPathValue("id", "2")
			}

			rec := httptest.NewRecorder()
			tt.handler(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("got: %v, want: %v", rec.Code, http.StatusUnauthorized)
			}
		})
	}
}

func Test_GetByID(t *testing.T) {
	wantTaskId := 2
	wantUserID := 1
	want := models.Task{
		ID:     wantTaskId,
		UserID: wantUserID,
		Title:  "Homework",
		Done:   false,
	}

	repo := &fakeRepo{
		GetByIDFunc: func(ctx context.Context, id, userID int) (models.Task, error) {
			if userID != wantUserID {
				t.Errorf("got: %v, want: %v", userID, wantUserID)
			}

			if id != wantTaskId {
				t.Errorf("got: %v, want: %v", id, wantTaskId)
			}
			return want, nil
		},
	}

	svc := service.NewTaskService(repo)

	sloger := logger.New("dev")
	h := NewTaskHandler(svc, sloger)

	req := httptest.NewRequest("GET", "/tasks/2", nil)
	req.SetPathValue("id", "2")

	rec := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), userIDKey, wantUserID)
	req = req.WithContext(ctx)

	h.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusOK)
	}

	var data models.Task
	err := json.Unmarshal(rec.Body.Bytes(), &data)
	if err != nil {
		t.Errorf("error for parsing body: %v", err)
	}

	if !reflect.DeepEqual(data, want) {
		t.Errorf("got: %+v, want: %+v", data, want)
	}
}

func Test_GetByID_InvalidTaskID(t *testing.T) {
	wantUserID := 2
	repo := &fakeRepo{
		GetByIDFunc: func(ctx context.Context, id, userID int) (models.Task, error) {
			t.Errorf("repository must not be called on invalid task id")
			return models.Task{}, nil
		},
	}

	svc := service.NewTaskService(repo)
	sloger := logger.New("dev")

	h := NewTaskHandler(svc, sloger)
	req := httptest.NewRequest("GET", "/tasks/abc", nil)

	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), userIDKey, wantUserID)
	req = req.WithContext(ctx)

	h.GetByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusBadRequest)
	}
}

func Test_GetByID_NotFound(t *testing.T) {
	wantUserID := 1
	repo := &fakeRepo{
		GetByIDFunc: func(ctx context.Context, id, userID int) (models.Task, error) {
			if userID != wantUserID {
				t.Errorf("got: %v, want: %v", userID, wantUserID)
			}
			return models.Task{}, repository.ErrTaskNotFound
		},
	}

	svc := service.NewTaskService(repo)

	sloger := logger.New("dev")
	h := NewTaskHandler(svc, sloger)

	req := httptest.NewRequest("GET", "/tasks/2", nil)
	req.SetPathValue("id", "2")

	rec := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), userIDKey, wantUserID)
	req = req.WithContext(ctx)

	h.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusNotFound)
	}
}

func Test_Create(t *testing.T) {
	input := models.Task{
		Title: "Dinner",
		Done:  false,
	}

	body, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal input: %v", err)
	}

	want := input
	want.ID = 2
	want.UserID = 2

	repo := &fakeRepo{
		CreateFunc: func(ctx context.Context, mt models.Task) (models.Task, error) {
			if mt.UserID != want.UserID {
				t.Errorf("got: %v, want: %v", mt.UserID, want.UserID)
			}

			if mt.Title == "" {
				t.Errorf("empty title")
			}

			if mt.Title != input.Title {
				t.Errorf("got: %v, want: %v", mt.Title, input.Title)
			}
			return want, nil
		},
		GetAllFunc: func(ctx context.Context, userID int) ([]models.Task, error) {
			return []models.Task{}, nil
		},
	}

	svc := service.NewTaskService(repo)
	sloger := logger.New("dev")

	h := NewTaskHandler(svc, sloger)
	req := httptest.NewRequest("POST", "/tasks/", bytes.NewReader(body))

	rec := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), userIDKey, want.UserID)
	req = req.WithContext(ctx)

	h.Create(rec, req)

	var data models.Task
	errUnmarshal := json.Unmarshal(rec.Body.Bytes(), &data)
	if errUnmarshal != nil {
		t.Errorf("error for parsing body: %v", errUnmarshal)
	}

	if !reflect.DeepEqual(data, want) {
		t.Errorf("got: %+v, want: %+v", data, want)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusCreated)
	}

}

func Test_Create_InvalidJSON(t *testing.T) {
	wantUserID := 1
	repo := &fakeRepo{
		CreateFunc: func(ctx context.Context, mt models.Task) (models.Task, error) {
			t.Errorf("repository must not be called on invalid JSON")
			return mt, nil
		},
		GetAllFunc: func(ctx context.Context, userID int) ([]models.Task, error) {
			t.Errorf("repository must not be called on invalid JSON")
			return nil, nil
		},
	}

	req := httptest.NewRequest("POST", "/tasks/", strings.NewReader("{invalid json}"))
	rec := httptest.NewRecorder()

	svc := service.NewTaskService(repo)
	sloger := logger.New("dev")

	h := NewTaskHandler(svc, sloger)

	ctx := context.WithValue(req.Context(), userIDKey, wantUserID)
	req = req.WithContext(ctx)

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusBadRequest)
	}
}

func Test_Create_EmptyTitle(t *testing.T) {
	wantUserID := 32

	repo := &fakeRepo{
		CreateFunc: func(ctx context.Context, mt models.Task) (models.Task, error) {
			t.Errorf("repository must not be called on empty title")
			return models.Task{}, nil
		},
		GetAllFunc: func(ctx context.Context, userID int) ([]models.Task, error) {
			t.Errorf("repository must not be called on empty title")
			return nil, nil
		},
	}

	svc := service.NewTaskService(repo)
	sloger := logger.New("dev")

	h := NewTaskHandler(svc, sloger)

	tests := []struct {
		name string
		body string
	}{
		{
			name: "empty_string",
			body: `{"title" : ""}`,
		},
		{
			name: "only_spaces",
			body: `{"title" : "   "}`,
		},
		{
			name: "empty_title",
			body: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/tasks/", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			ctx := context.WithValue(req.Context(), userIDKey, wantUserID)
			req = req.WithContext(ctx)
			h.Create(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("got: %v, want: %v", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func Test_Create_Duplicate(t *testing.T) {
	wantUserID := 2

	repo := &fakeRepo{
		GetAllFunc: func(ctx context.Context, userID int) ([]models.Task, error) {
			return []models.Task{{Title: "Dinner", Done: false}}, nil
		},
		CreateFunc: func(ctx context.Context, mt models.Task) (models.Task, error) {
			t.Errorf("repository must not be called on duplicate title")
			return models.Task{}, nil
		},
	}

	svc := service.NewTaskService(repo)
	sloger := logger.New("dev")

	h := NewTaskHandler(svc, sloger)

	req := httptest.NewRequest("POST", "/tasks/", strings.NewReader(`{"title" : "Dinner"}`))
	rec := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), userIDKey, wantUserID)
	req = req.WithContext(ctx)

	h.Create(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusConflict)
	}

}

func Test_Create_BodyTooLarge(t *testing.T) {
	wantUserID := 2

	repo := &fakeRepo{
		GetAllFunc: func(ctx context.Context, userID int) ([]models.Task, error) {
			t.Errorf("repository must not be called on too large body")
			return nil, nil
		},
		CreateFunc: func(ctx context.Context, mt models.Task) (models.Task, error) {
			t.Errorf("repository must not be called on too large body")
			return models.Task{}, nil
		},
	}

	svc := service.NewTaskService(repo)
	sloger := logger.New("dev")

	h := NewTaskHandler(svc, sloger)

	body := `{"title":"` + strings.Repeat("a", 600) + `"}`
	req := httptest.NewRequest("POST", "/tasks/", strings.NewReader(body))
	rec := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), userIDKey, wantUserID)
	req = req.WithContext(ctx)

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusBadRequest)
	}
}

func Test_IntrenalError(t *testing.T) {
	wantUserID := 2
	errDB := errors.New("db down")

	repo := &fakeRepo{
		GetAllFunc: func(ctx context.Context, userID int) ([]models.Task, error) {
			return nil, errDB
		},
		GetByIDFunc: func(ctx context.Context, id, userID int) (models.Task, error) {
			return models.Task{}, errDB
		},
		CreateFunc: func(ctx context.Context, mt models.Task) (models.Task, error) {
			return models.Task{}, nil
		},
		UpdateFunc: func(ctx context.Context, id, userID int, mt models.Task) (models.Task, error) {
			return models.Task{}, errDB
		},
		DeleteFunc: func(ctx context.Context, id, userID int) error {
			return errDB
		},
	}

	svc := service.NewTaskService(repo)
	sloger := logger.New("dev")

	h := NewTaskHandler(svc, sloger)

	tests := []struct {
		name string
		handler http.HandlerFunc
		method string
		id string
		body string
	}{
		{
			name: "GetAll",
			handler: h.GetAll,
			method: "GET",
			id: "",
			body: "",
		},
		{
			name: "GetByID",
			handler: h.GetByID,
			method: "GET",
			id: "2",
			body: "",
		},
		{
			name: "Create",
			handler: h.Create,
			method: "POST",
			id: "",
			body: `{"title" : "Dinner"}`,
		},
		{
			name: "Update",
			handler: h.Update,
			method: "PUT",
			id: "2",
			body: `{"title" : "Dinner"}`,
		},
		{
			name: "Delete",
			handler: h.Delete,
			method: "DELETE",
			id: "2",
			body: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/tasks/", strings.NewReader(tt.body))

			req.SetPathValue("id", tt.id)
			ctx := context.WithValue(req.Context(), userIDKey, wantUserID)
			req = req.WithContext(ctx)

			rec := httptest.NewRecorder()
			tt.handler(rec, req)

			if rec.Code != http.StatusInternalServerError {
				t.Errorf("got: %v, want: %v", rec.Code, http.StatusInternalServerError)
			}
		})
	}

}

func Test_Update(t *testing.T) {
	wantUserID := 2
	wantTaskID := 2
	after := models.Task{
		ID:     2,
		UserID: wantUserID,
		Title:  "Dinner",
		Done:   true,
	}

	repo := &fakeRepo{
		UpdateFunc: func(ctx context.Context, id, userID int, mt models.Task) (models.Task, error) {
			if userID != wantUserID {
				t.Errorf("got: %v, want: %v", userID, wantUserID)
			}

			if id != wantTaskID {
				t.Errorf("got: %v, want: %v", id, wantTaskID)
			}

			mt.ID = id
			mt.UserID = userID
			return mt, nil
		},
	}

	body, err := json.Marshal(after)
	if err != nil {
		t.Fatalf("failed to marshal input: %v", err)
	}

	svc := service.NewTaskService(repo)

	sloger := logger.New("dev")
	h := NewTaskHandler(svc, sloger)

	req := httptest.NewRequest("PUT", "/tasks/2", bytes.NewReader(body))
	req.SetPathValue("id", "2")

	rec := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), userIDKey, wantUserID)
	req = req.WithContext(ctx)

	h.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusOK)
	}

	var data models.Task
	errUnmarshal := json.Unmarshal(rec.Body.Bytes(), &data)
	if errUnmarshal != nil {
		t.Errorf("error for parsing body: %v", errUnmarshal)
	}

	if !reflect.DeepEqual(after, data) {
		t.Errorf("got: %+v, want: %+v", data, after)
	}

}

func Test_Update_InvalidTaskID(t *testing.T) {
	repo := &fakeRepo{
		UpdateFunc: func(ctx context.Context, id, userID int, mt models.Task) (models.Task, error) {
			t.Errorf("repository must not be called on invalid task id")
			return mt, nil
		},
	}
	wantUserID := 1

	svc := service.NewTaskService(repo)
	sloger := logger.New("dev")

	h := NewTaskHandler(svc, sloger)

	req := httptest.NewRequest("PUT", "/tasks/abc", strings.NewReader(`{"title" : "x"}`))
	req.SetPathValue("id", "abc")

	rec := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), userIDKey, wantUserID)
	req = req.WithContext(ctx)

	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusBadRequest)
	}
}

func Test_Update_InvalidJSON(t *testing.T) {
	wantUserID := 1
	repo := &fakeRepo{
		UpdateFunc: func(ctx context.Context, id int, userID int, mt models.Task) (models.Task, error) {
			t.Error("repository must not be called on invalid JSON")
			return mt, nil
		},
	}

	req := httptest.NewRequest("PUT", "/tasks/2", strings.NewReader("{invalid json}"))
	req.SetPathValue("id", "2")

	rec := httptest.NewRecorder()

	svc := service.NewTaskService(repo)
	sloger := logger.New("dev")

	h := NewTaskHandler(svc, sloger)

	ctx := context.WithValue(req.Context(), userIDKey, wantUserID)
	req = req.WithContext(ctx)

	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusBadRequest)
	}
}

func Test_Update_NotFound(t *testing.T) {
	wantUserID := 2
	after := models.Task{
		ID:     2,
		UserID: wantUserID,
		Title:  "Dinner",
		Done:   true,
	}

	body, err := json.Marshal(after)
	if err != nil {
		t.Fatalf("failed to marshal input: %v", err)
	}

	repo := &fakeRepo{
		UpdateFunc: func(ctx context.Context, id, userID int, mt models.Task) (models.Task, error) {
			if userID != wantUserID {
				t.Errorf("got: %v, want: %v", userID, wantUserID)
			}
			return models.Task{}, repository.ErrTaskNotFound
		},
	}

	svc := service.NewTaskService(repo)

	sloger := logger.New("dev")
	h := NewTaskHandler(svc, sloger)

	req := httptest.NewRequest("PUT", "/tasks/2", bytes.NewReader(body))
	req.SetPathValue("id", "2")

	rec := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), userIDKey, wantUserID)
	req = req.WithContext(ctx)

	h.Update(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusNotFound)
	}
}

func Test_Update_EmptyTitle(t *testing.T) {
	wantUserID := 15
	wantTaskID := 5
	repo := &fakeRepo{
		UpdateFunc: func(ctx context.Context, id, userID int, mt models.Task) (models.Task, error) {
			t.Errorf("repository must not be called om empty title")
			return models.Task{}, nil
		},
	}

	svc := service.NewTaskService(repo)
	sloger := logger.New("dev")
	h := NewTaskHandler(svc, sloger)

	tests := []struct {
		name string
		body string
	}{
		{
			name: "empty_string",
			body: `{"title" : ""}`,
		},
		{
			name: "only_spaces",
			body: `{"title" : "  "}`,
		},
		{
			name: "empty_title",
			body: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("PUT", "/tasks/5", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			req.SetPathValue("id", strconv.Itoa(wantTaskID))

			ctx := context.WithValue(req.Context(), userIDKey, wantUserID)
			req = req.WithContext(ctx)
			h.Update(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("got: %v, want: %v", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func Test_Update_BodyTooLarge(t *testing.T) {
	wantUserID := 2

	repo := &fakeRepo{
		UpdateFunc: func(ctx context.Context, id, userID int, mt models.Task) (models.Task, error) {
			t.Errorf("repository must not be called on too large body")
			return models.Task{}, nil
		},
	}

	svc := service.NewTaskService(repo)
	sloger := logger.New("dev")

	h := NewTaskHandler(svc, sloger)

	body := `{"title":"` + strings.Repeat("a", 600) + `"}`
	req := httptest.NewRequest("PUT", "/tasks/2", strings.NewReader(body))
	req.SetPathValue("id", "2")
	rec := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), userIDKey, wantUserID)
	req = req.WithContext(ctx)

	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusBadRequest)
	}
}


func Test_Delete(t *testing.T) {
	wantTaskID := 2
	wantUserID := 2
	repo := &fakeRepo{
		DeleteFunc: func(ctx context.Context, id, userID int) error {
			if userID != wantUserID {
				t.Errorf("got: %v, want: %v", userID, wantUserID)
			}

			if id != wantTaskID {
				t.Errorf("got: %v, want: %v", id, wantTaskID)
			}
			return nil
		},
	}

	svc := service.NewTaskService(repo)

	sloger := logger.New("dev")
	h := NewTaskHandler(svc, sloger)

	req := httptest.NewRequest("DELETE", "/tasks/2", nil)
	req.SetPathValue("id", "2")

	rec := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), userIDKey, wantUserID)
	req = req.WithContext(ctx)

	h.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusNoContent)
	}
}

func Test_Delete_InvalidTaskID(t *testing.T) {
	wantUserId := 2
	repo := &fakeRepo{
		DeleteFunc: func(ctx context.Context, id, userId int) error {
			t.Errorf("repository must not be called on invalid task id")
			return nil
		},
	}

	svc := service.NewTaskService(repo)

	sloger := logger.New("dev")
	h := NewTaskHandler(svc, sloger)

	req := httptest.NewRequest("DELETE", "/tasks/abc", nil)
	req.SetPathValue("id", "abc")

	rec := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), userIDKey, wantUserId)
	req = req.WithContext(ctx)

	h.Delete(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusBadRequest)
	}
}

func Test_Delete_NotFound(t *testing.T) {
	wantUserID := 1
	repo := &fakeRepo{
		DeleteFunc: func(ctx context.Context, id, userID int) error {
			if userID != wantUserID {
				t.Errorf("got: %v, want: %v", userID, wantUserID)
			}
			return repository.ErrTaskNotFound
		},
	}

	svc := service.NewTaskService(repo)
	sloger := logger.New("dev")

	h := NewTaskHandler(svc, sloger)

	req := httptest.NewRequest("DELETE", "/tasks/2", nil)
	req.SetPathValue("id", "2")

	rec := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), userIDKey, wantUserID)
	req = req.WithContext(ctx)

	h.Delete(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusNotFound)
	}
}

