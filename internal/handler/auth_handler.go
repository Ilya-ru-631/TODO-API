package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"todo_api/internal/service"
)

type AuthHandler struct {
	service *service.UserService
}

type TokenRequest struct {
	Token string `json:"token"`
}

func NewAuthHandler(svc *service.UserService) *AuthHandler {
	return &AuthHandler{service: svc}
}

func (a *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "incorrect json: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Username == "" || req.Password == "" {
		writeError(w, "username and password not be empty", http.StatusBadRequest)
		return
	}

	reg, err := a.service.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUsernameTaken) {
			writeError(w, "username is already taken", http.StatusConflict)
			return
		}
		writeError(w, "error for a server side", http.StatusInternalServerError)
		return
	}

	result := TokenRequest{
		Token: reg,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)

}

func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "incorrect json"+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Username == "" || req.Password == "" {
		writeError(w, "username and password must not be empty", http.StatusBadRequest)
		return
	}

	reg, err := a.service.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		writeError(w, "wrong username or password", http.StatusUnauthorized)
		return
	}

	result := TokenRequest{
		Token: reg,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
