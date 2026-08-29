package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"todo_api/internal/models"
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

	h := NewTaskHandler(repo)
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

	h := NewTaskHandler(repo)
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

func Test_Create(t *testing.T) {
	input := models.Task{
		Title: "Dinner",
		Done:  false,
	}

	body, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshl input: %v", err)
	}

	want := input
	want.ID = 2

	repo := &fakeRepo{
		CreateFunc: func(ctx context.Context, t models.Task) (models.Task, error) {
			return want, nil
		},
	}

	h := NewTaskHandler(repo)
	req := httptest.NewRequest("POST", "/tasks/", bytes.NewReader(body))
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
