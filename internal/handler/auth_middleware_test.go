package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"todo_api/internal/apierror"
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

	if rec.Code != http.StatusOK {
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

	const wantStatus = http.StatusUnauthorized
	const nextCalled = false

	tests := []struct {
		name          string
		authorization string
		wantCode      string
	}{
		{
			name:     "no_header",
			wantCode: apierror.CodeInvalidToken,
		},
		{
			name:          "trash_token",
			authorization: "Bearer abc",
			wantCode:      apierror.CodeInvalidToken,
		},
		{
			name:          "alien_secret",
			authorization: "Bearer " + token,
			wantCode:      apierror.CodeInvalidToken,
		},
		{
			name:          "no_bearer",
			authorization: validToken,
			wantCode:      apierror.CodeInvalidToken,
		},
		{
			name:          "invalid_case_bearer",
			authorization: "bearer " + token,
			wantCode:      apierror.CodeInvalidToken,
		},
		{
			name:          "empty_token",
			authorization: "Bearer ",
			wantCode:      apierror.CodeInvalidToken,
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

			if rec.Code != wantStatus {
				t.Errorf("got: %v, want: %v", rec.Code, wantStatus)
			}

			if called != nextCalled {
				t.Errorf("got: %v, want: %v", called, nextCalled)
			}

			testErr := decodeAPIError(t, rec)

			if testErr.Code != tt.wantCode {
				t.Fatalf("got code: %v, want code: %v", testErr.Code, tt.wantCode)
			}

		})
	}
}
