package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"todo_api/internal/models"
	"todo_api/internal/repository"
	"todo_api/internal/service"
)

type fakeRepo struct {
	GetAllFunc  func(ctx context.Context) ([]models.Task, error)
	GetByIDFunc func(ctx context.Context, id int) (models.Task, error)
	CreateFunc  func(ctx context.Context, t models.Task) (models.Task, error)
	UpdateFunc  func(ctx context.Context, id int, t models.Task) (models.Task, error)
	DeleteFunc  func(ctx context.Context, id int) error
}

func (f *fakeRepo) GetAll(ctx context.Context) ([]models.Task, error) {
	return f.GetAllFunc(ctx)
}

func (f *fakeRepo) GetByID(ctx context.Context, id int) (models.Task, error) {
	return f.GetByIDFunc(ctx, id)
}

func (f *fakeRepo) Create(ctx context.Context, t models.Task) (models.Task, error) {
	return f.CreateFunc(ctx, t)
}

func (f *fakeRepo) Update(ctx context.Context, id int, t models.Task) (models.Task, error) {
	return f.UpdateFunc(ctx, id, t)
}

func (f *fakeRepo) Delete(ctx context.Context, id int) error {
	return f.DeleteFunc(ctx, id)
}

func Test_GetAll(t *testing.T) {
	want := []models.Task{
		{ID: 1, Title: "Homework", Done: false},
		{ID: 2, Title: "Dinner", Done: true},
	}

	repo := &fakeRepo{
		GetAllFunc: func(ctx context.Context) ([]models.Task, error) {
			return want, nil
		},
	}

	svc := service.NewTaskService(repo)
	h := NewTaskHandler(svc)
	req := httptest.NewRequest("GET", "/tasks/", nil)
	rec := httptest.NewRecorder()
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

func Test_GetByID(t *testing.T) {
	want := models.Task{
		ID:    2,
		Title: "Homework",
		Done:  false,
	}

	repo := &fakeRepo{
		GetByIDFunc: func(ctx context.Context, id int) (models.Task, error) {
			return want, nil
		},
	}

	svc := service.NewTaskService(repo)
	h := NewTaskHandler(svc)
	req := httptest.NewRequest("GET", "/tasks/2", nil)
	req.SetPathValue("id", "2")
	rec := httptest.NewRecorder()
	h.GetByID(rec, req)

	var data models.Task
	err := json.Unmarshal(rec.Body.Bytes(), &data)
	if err != nil {
		t.Errorf("error for parsing body: %v", err)
	}

	if !reflect.DeepEqual(data, want) {
		t.Errorf("got: %+v, want: %+v", data, want)
	}
}

func Test_GetByID_InvalidID(t *testing.T) {
	repo := &fakeRepo{}

	svc := service.NewTaskService(repo)
	h := NewTaskHandler(svc)
	req := httptest.NewRequest("GET", "/tasks/abc", nil)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()
	h.GetByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusBadRequest)
	}
}

func Test_GetByID_NotFound(t *testing.T) {
	repo := &fakeRepo{
		GetByIDFunc: func(ctx context.Context, id int) (models.Task, error) {
			return models.Task{}, repository.ErrTaskNotFound
		},
	}

	svc := service.NewTaskService(repo)
	h := NewTaskHandler(svc)
	req := httptest.NewRequest("GET", "/tasks/2", nil)
	req.SetPathValue("id", "2")
	rec := httptest.NewRecorder()
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

	repo := &fakeRepo{
		CreateFunc: func(ctx context.Context, t models.Task) (models.Task, error) {
			return want, nil
		},
		GetAllFunc: func(ctx context.Context) ([]models.Task, error) {
			return []models.Task{}, nil
		},
	}

	svc := service.NewTaskService(repo)
	h := NewTaskHandler(svc)
	req := httptest.NewRequest("POST", "/tasks/2", bytes.NewReader(body))
	req.SetPathValue("id", "2")
	rec := httptest.NewRecorder()

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
	repo := &fakeRepo{}
	req := httptest.NewRequest("POST", "/tasks/", strings.NewReader("{invalid json}"))
	rec := httptest.NewRecorder()

	svc := service.NewTaskService(repo)
	h := NewTaskHandler(svc)
	h.Create(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusBadRequest)
	}
}

func Test_Update(t *testing.T) {
	after := models.Task{
		ID:    2,
		Title: "Dinner",
		Done:  true,
	}

	repo := &fakeRepo{
		UpdateFunc: func(ctx context.Context, id int, t models.Task) (models.Task, error) {
			t.ID = id
			return t, nil
		},
	}

	body, err := json.Marshal(after)
	if err != nil {
		t.Fatalf("failed to marshal input: %v", err)
	}

	svc := service.NewTaskService(repo)
	h := NewTaskHandler(svc)
	req := httptest.NewRequest("PUT", "/tasks/2", bytes.NewReader(body))
	req.SetPathValue("id", "2")
	rec := httptest.NewRecorder()

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

func Test_Update_InvalidID(t *testing.T) {
	repo := &fakeRepo{}

	svc := service.NewTaskService(repo)
	h := NewTaskHandler(svc)
	req := httptest.NewRequest("PUT", "/tasks/abc", nil)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusBadRequest)
	}
}

func Test_Update_InvalidJSON(t *testing.T) {
	repo := &fakeRepo{}
	req := httptest.NewRequest("PUT", "/tasks/2", strings.NewReader("{invalid json}"))
	req.SetPathValue("id", "2")
	rec := httptest.NewRecorder()

	svc := service.NewTaskService(repo)
	h := NewTaskHandler(svc)
	h.Update(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusBadRequest)
	}
}

func Test_Update_NotFound(t *testing.T) {
	after := models.Task{
		ID:    2,
		Title: "Dinner",
		Done:  true,
	}

	body, err := json.Marshal(after)
	if err != nil {
		t.Fatalf("failed to marshal input: %v", err)
	}

	repo := &fakeRepo{
		UpdateFunc: func(ctx context.Context, id int, t models.Task) (models.Task, error) {
			return models.Task{}, repository.ErrTaskNotFound
		},
	}

	svc := service.NewTaskService(repo)
	h := NewTaskHandler(svc)
	req := httptest.NewRequest("PUT", "/tasks/2", bytes.NewReader(body))
	req.SetPathValue("id", "2")
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusNotFound)
	}
}

func Test_Delete(t *testing.T) {
	repo := &fakeRepo{
		DeleteFunc: func(ctx context.Context, id int) error {
			return nil
		},
	}

	svc := service.NewTaskService(repo)
	h := NewTaskHandler(svc)
	req := httptest.NewRequest("DELETE", "/tasks/2", nil)
	req.SetPathValue("id", "2")
	rec := httptest.NewRecorder()
	h.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusNoContent)
	}
}

func Test_Delete_InvalidID(t *testing.T) {
	repo := &fakeRepo{}

	svc := service.NewTaskService(repo)
	h := NewTaskHandler(svc)
	req := httptest.NewRequest("DELETE", "/tasks/abc", nil)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()
	h.Delete(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusBadRequest)
	}
}

func Test_Delete_NotFound(t *testing.T) {
	repo := &fakeRepo{
		DeleteFunc: func(ctx context.Context, id int) error {
			return repository.ErrTaskNotFound
		},
	}

	svc := service.NewTaskService(repo)
	h := NewTaskHandler(svc)
	req := httptest.NewRequest("DELETE", "/tasks/2", nil)
	req.SetPathValue("id", "2")
	rec := httptest.NewRecorder()
	h.Delete(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusNotFound)
	}
}
