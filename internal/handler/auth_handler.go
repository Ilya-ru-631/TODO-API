package handler

import (
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"todo_api/internal/apierror"
	"todo_api/internal/service"
)

type AuthHandler struct {
	service *service.UserService
	log     *slog.Logger
}

type TokenRequest struct {
	Token string `json:"token"`
}

func NewAuthHandler(svc *service.UserService, log *slog.Logger) *AuthHandler {
	return &AuthHandler{service: svc, log: log}
}

func (a *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, http.StatusBadRequest, apierror.CodeInvalidJSON, "Incorrect json")
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	if strings.TrimSpace(req.Username) == "" {
		apierror.WriteValidation(w, apierror.CodeValidation, "Username must not be empty", "username")
		return
	}
	if strings.TrimSpace(req.Password) == "" {
		apierror.WriteValidation(w, apierror.CodeValidation, "Password must not be empty", "password")
		return
	}

	reg, err := a.service.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUsernameIsTaken) {
			apierror.Write(w, http.StatusConflict, apierror.CodeUsernameTaken, "Username is already taken")
			return
		}
		a.log.ErrorContext(r.Context(), "registration failed",
			slog.String("username", req.Username),
			slog.Any("err", err),
		)
		apierror.Write(w, http.StatusInternalServerError, apierror.CodeInternal, "Internal server error")
		return
	}

	result := TokenRequest{
		Token: reg,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, http.StatusBadRequest, apierror.CodeInvalidJSON, "Incorrect json")
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	if strings.TrimSpace(req.Username) == "" {
		apierror.WriteValidation(w, apierror.CodeValidation, "Username must not be empty", "username")
		return
	}
	if strings.TrimSpace(req.Password) == "" {
		apierror.WriteValidation(w, apierror.CodeValidation, "Password must not be empty", "password")
		return
	}

	reg, err := a.service.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrPasswordOrLogin) {
			apierror.Write(w, http.StatusUnauthorized, apierror.CodeInvalidCredentials, "Invalid username or password")
			return
		}
		a.log.ErrorContext(r.Context(), "login failed",
			slog.String("username", req.Username),
			slog.Any("err", err),
		)
		apierror.Write(w, http.StatusInternalServerError, apierror.CodeInternal, "Internal server error")
		return
	}

	result := TokenRequest{
		Token: reg,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}
