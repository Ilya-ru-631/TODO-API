package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"todo_api/internal/auth"
)

func Test_AuthMiddleware_ValidToken(t *testing.T) {
	wantUserID := 5
	token, err := auth.GenerateToken(wantUserID, testSecret)
	if err != nil {
		t.Fatalf("error for generate token: %v", err)
	}

	called := false
	var gotUserID int
	var ok bool

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		gotUserID, ok = r.Context().Value(userIDKey).(int)
	})

	h := AuthMiddleware(testSecret)(next)

	req := httptest.NewRequest("POST", "/tasks/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK{
		t.Errorf("got: %v, want: %v", rec.Code, http.StatusOK)
	}

	if !called {
		t.Fatalf("next must be called with a valid tokend")
	}

	if !ok {
		t.Errorf("userID is empty or userID not int")
	}

	if gotUserID != wantUserID {
		t.Errorf("got: %v, want: %v", gotUserID, wantUserID)
	}
}

func Test_AuthMiddleware_Crash(t *testing.T) {
	wantUserID := 5
	otherSecret := []byte("other-secret")
	token, err := auth.GenerateToken(wantUserID, otherSecret)
	if err != nil {
		t.Fatalf("error for generate token: %v", err)
	}

	validToken, err := auth.GenerateToken(wantUserID, testSecret)
	if err != nil {
		t.Fatalf("error for generate token: %v", err)
	}

	tests := []struct {
		name          string
		authorization string
		wantStatus    int
		nextCalled    bool
	}{
		{
			name:       "no_header",
			wantStatus: http.StatusUnauthorized,
			nextCalled: false,
		},
		{
			name:          "trash_token",
			authorization: "Bearer abc",
			wantStatus:    http.StatusUnauthorized,
			nextCalled:    false,
		},
		{
			name:          "alien_secret",
			authorization: "Bearer " + token,
			wantStatus:    http.StatusUnauthorized,
			nextCalled:    false,
		},
		{
			name:          "no_bearer",
			authorization: validToken,
			wantStatus:    http.StatusUnauthorized,
			nextCalled:    false,
		},
		{
			name:          "invalid_case_bearer",
			authorization: "bearer " + token,
			wantStatus:    http.StatusUnauthorized,
			nextCalled:    false,
		},
		{
			name:          "empty_token",
			authorization: "Bearer ",
			wantStatus:    http.StatusUnauthorized,
			nextCalled:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
			})

			h := AuthMiddleware(testSecret)(next)
			req := httptest.NewRequest("POST", "/tasks/", nil)
			req.Header.Set("Authorization", tt.authorization)

			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("got: %v, want: %v", rec.Code, tt.wantStatus)
			}

			if called != tt.nextCalled {
				t.Errorf("got: %v, want: %v", called, tt.nextCalled)
			}

			var testErr ErrorResponse
			if err := json.NewDecoder(rec.Body).Decode(&testErr); err != nil {
				t.Fatalf("error for decode body: %v", err)
			}

			if testErr.Error == "" {
				t.Errorf("empty error field")
			}

			rr := rec.Result()
			contentType := rr.Header.Get("Content-type")
			if !strings.HasPrefix(contentType, "application/json") {
				t.Errorf("header do not have application/json")
			}

		})
	}
}
