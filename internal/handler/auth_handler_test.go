package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"todo_api/internal/auth"
	"todo_api/internal/logger"
	"todo_api/internal/models"
	"todo_api/internal/repository"
	"todo_api/internal/service"
)

type fakeUserRepo struct {
	CreateFunc        func(ctx context.Context, u models.User) (models.User, error)
	GetByUsernameFunc func(ctx context.Context, username string) (models.User, error)
}

func (f *fakeUserRepo) Create(ctx context.Context, u models.User) (models.User, error) {
	return f.CreateFunc(ctx, u)
}

func (f *fakeUserRepo) GetByUsername(ctx context.Context, username string) (models.User, error) {
	return f.GetByUsernameFunc(ctx, username)
}

var testSecret = []byte("test-secret")

func Test_Register(t *testing.T) {
	wantUserID := 5
	wantUsername := "Shev"
	wantPassword := "123"

	repo := &fakeUserRepo{
		CreateFunc: func(ctx context.Context, u models.User) (models.User, error) {
			if u.Username != wantUsername {
				t.Errorf("got username: %v, want: %v", u.Username, wantUsername)
			}

			if u.PasswordHash == "" || u.PasswordHash == wantPassword {
				t.Errorf("password must be stored as a hash, got: %q", u.PasswordHash)
			}

			u.ID = wantUserID
			return u, nil
		},
		GetByUsernameFunc: func(ctx context.Context, username string) (models.User, error) {
			return models.User{}, repository.ErrUserNotFound
		},
	}

	svc := service.NewUserService(repo, testSecret)
	sloger := logger.New("dev")
	h := NewAuthHandler(svc, sloger)

	body := fmt.Sprintf(`{"username" : %q, "password" : %q}`, wantUsername, wantPassword)

	req := httptest.NewRequest("POST", "/register", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusCreated)
	}

	var resp TokenRequest
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("error for unmarshal: %v", err)
	}

	user_id, err := auth.ParsingToken(resp.Token, testSecret)
	if err != nil {
		t.Fatalf("error for parsing token: %v", err)
	}
	
	if user_id != wantUserID {
		t.Errorf("got userID: %v, want userID: %v", user_id, wantUserID)
	}
}

func Test_Register_InvalidInput(t *testing.T) {
	repo := &fakeUserRepo{
		CreateFunc: func(ctx context.Context, u models.User) (models.User, error) {
			t.Errorf("repository must not be called on empty password or username and invalid json")
			return models.User{}, nil
		},
		GetByUsernameFunc: func(ctx context.Context, username string) (models.User, error) {
			t.Errorf("repository must not be called on empty password or username and invalid json")
			return models.User{}, nil
		},
	}

	tests := []struct {
		name                string
		usernameAndPassword string
	}{
		{
			name:                "empty_username",
			usernameAndPassword: `{"username" : "", "password" : "x"}`,
		},
		{
			name:                "empty_password",
			usernameAndPassword: `{"username" : "x", "password" : ""}`,
		},
		{
			name:                "no_fields",
			usernameAndPassword: `{}`,
		},
		{
			name:                "invalid_json",
			usernameAndPassword: `{invalid json}`,
		},
	}

	svc := service.NewUserService(repo, testSecret)
	sloger := logger.New("dev")
	h := NewAuthHandler(svc, sloger)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/register", strings.NewReader(tt.usernameAndPassword))
			rec := httptest.NewRecorder()

			h.Register(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("got: %v, want: %v", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func Test_Register_RepoErrors(t *testing.T) {
	errDB := errors.New("db down")

	tests := []struct {
		name        string
		getUser     models.User
		getErr      error
		createErr   error
		wantStatus  int
		wantCreated bool
	}{
		{
			name:        "username_taken",
			getUser:     models.User{ID: 1, Username: "Shev"},
			getErr:      nil,
			wantStatus:  http.StatusConflict,
			wantCreated: false,
		},
		{
			name:        "get_by_username_fails",
			getErr:      errDB,
			wantStatus:  http.StatusInternalServerError,
			wantCreated: false,
		},
		{
			name:        "create_fails",
			getErr:      repository.ErrUserNotFound,
			createErr:   errDB,
			wantStatus:  http.StatusInternalServerError,
			wantCreated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created := false
			repo := &fakeUserRepo{
				GetByUsernameFunc: func(ctx context.Context, username string) (models.User, error) {
					return tt.getUser, tt.getErr
				},
				CreateFunc: func(ctx context.Context, u models.User) (models.User, error) {
					created = true
					return models.User{}, tt.createErr
				},
			}

			svc := service.NewUserService(repo, testSecret)
			h := NewAuthHandler(svc, logger.New("dev"))

			req := httptest.NewRequest("POST", "/register",
				strings.NewReader(`{"username":"Shev","password":"123"}`))
			rec := httptest.NewRecorder()

			h.Register(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("got: %v, want: %v", rec.Code, tt.wantStatus)
			}
			if created != tt.wantCreated {
				t.Errorf("repo.Create called: %v, want: %v", created, tt.wantCreated)
			}
		})
	}
}

func Test_Login(t *testing.T) {
	wantUserID := 5
	wantUsername := "Shev"
	wantPassword := "123"

	hash, err := auth.HashPassword(wantPassword)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	repo := &fakeUserRepo{
		GetByUsernameFunc: func(ctx context.Context, username string) (models.User, error) {
			if username != wantUsername {
				t.Errorf("got username: %v, want: %v", username, wantUsername)
			}
			return models.User{ID: wantUserID, Username: wantUsername, PasswordHash: hash}, nil
		},
		CreateFunc: func(ctx context.Context, u models.User) (models.User, error) {
			t.Errorf("repository.Create must not be called on login")
			return models.User{}, nil
		},
	}

	svc := service.NewUserService(repo, testSecret)
	h := NewAuthHandler(svc, logger.New("dev"))

	body := fmt.Sprintf(`{"username":%q,"password":%q}`, wantUsername, wantPassword)
	req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusOK)
	}

	var resp TokenRequest
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error for unmarshal: %v", err)
	}

	gotUserID, err := auth.ParsingToken(resp.Token, testSecret)
	if err != nil {
		t.Fatalf("error for parsing token: %v", err)
	}
	if gotUserID != wantUserID {
		t.Errorf("got userID: %v, want userID: %v", gotUserID, wantUserID)
	}
}

func Test_Login_WrongCredentials(t *testing.T) {
	const realPassword = "123"

	hash, err := auth.HashPassword(realPassword)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	tests := []struct {
		name     string
		getUser  models.User
		getErr   error
		password string
	}{
		{
			name:     "wrong_password",
			getUser:  models.User{ID: 5, Username: "Shev", PasswordHash: hash},
			getErr:   nil,
			password: "wrong",
		},
		{
			name:     "user_not_found",
			getUser:  models.User{},
			getErr:   repository.ErrUserNotFound,
			password: realPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepo{
				GetByUsernameFunc: func(ctx context.Context, username string) (models.User, error) {
					return tt.getUser, tt.getErr
				},
				CreateFunc: func(ctx context.Context, u models.User) (models.User, error) {
					t.Errorf("repository.Create must not be called on login")
					return models.User{}, nil
				},
			}

			svc := service.NewUserService(repo, testSecret)
			h := NewAuthHandler(svc, logger.New("dev"))

			body := fmt.Sprintf(`{"username":"Shev","password":%q}`, tt.password)
			req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
			rec := httptest.NewRecorder()

			h.Login(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("got: %v, want: %v", rec.Code, http.StatusUnauthorized)
			}
		})
	}
}

func Test_Login_InvalidInput(t *testing.T) {
	repo := &fakeUserRepo{
		GetByUsernameFunc: func(ctx context.Context, username string) (models.User, error) {
			t.Errorf("repository must not be called on invalid input")
			return models.User{}, nil
		},
		CreateFunc: func(ctx context.Context, u models.User) (models.User, error) {
			t.Errorf("repository must not be called on invalid input")
			return models.User{}, nil
		},
	}

	svc := service.NewUserService(repo, testSecret)
	h := NewAuthHandler(svc, logger.New("dev"))

	tests := []struct {
		name string
		body string
	}{
		{
			name: "empty_username", 
			body: `{"username":"","password":"x"}`,
		},
		{
			name: "empty_password", 
			body: `{"username":"x","password":""}`,
		},
		{
			name: "no_fields", 
			body: `{}`,
		},
		{
			name: "invalid_json", 
			body: `{invalid json}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/login", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			h.Login(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("got: %v, want: %v", rec.Code, http.StatusBadRequest)
			}
		})
	}
}
