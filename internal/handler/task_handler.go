package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	//"todo_api/internal/models"

	"todo_api/internal/repository"
	"todo_api/internal/service"
)

type TaskHandler struct {
	service *service.TaskService
	log     *slog.Logger
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var LimitForBodyLength = 500

func NewTaskHandler(svc *service.TaskService, log *slog.Logger) *TaskHandler {
	return &TaskHandler{service: svc, log: log}
}

func writeError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}

func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tasks, err := h.service.GetAll(r.Context(), userID)
	if err != nil {
		h.log.ErrorContext(r.Context(), "failed to get tasks",
			slog.Int("user_id", userID),
			slog.Any("err", err),
		)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, "id must be a number", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	task, err := h.service.GetByID(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			writeError(w, "not found task with this id", http.StatusNotFound)
			return
		}

		h.log.ErrorContext(r.Context(), "failed to get task",
			slog.Int("task_id", id),
			slog.Int("user_id", userID),
			slog.Any("err", err),
		)
		writeError(w, "error on the server side", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateTaskRequest

	r.Body = http.MaxBytesReader(w, r.Body, int64(LimitForBodyLength))

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "incorrect Json: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if strings.TrimSpace(req.Title) == "" {
		writeError(w, "empty title", http.StatusBadRequest)
		return
	}

	task, err := h.service.Create(r.Context(), userID, req.ToTask())
	if err != nil {
		if errors.Is(err, service.ErrDuplicateTitle) {
			writeError(w, "a task with such a title already exists", http.StatusConflict)
			return
		}
		h.log.ErrorContext(r.Context(), "failed to create task",
			slog.Int("user_id", userID),
			slog.Any("err", err),
		)
		writeError(w, "error on a server side", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdateTaskRequest

	r.Body = http.MaxBytesReader(w, r.Body, int64(LimitForBodyLength))

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "incorrect data: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, "id must be a number", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		writeError(w, "empty title", http.StatusBadRequest)
		return
	}

	task, err := h.service.Update(r.Context(), id, userID, req.ToTask())
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			writeError(w, "not found this task", http.StatusNotFound)
			return
		}
		if errors.Is(err, repository.ErrValidation) {
			writeError(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.log.ErrorContext(r.Context(), "failed to update task",
			slog.Int("task_id", id),
			slog.Int("user_id", userID),
			slog.Any("err", err),
		)
		writeError(w, "error on the server side", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, "id must be a number", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			writeError(w, "not found this task", http.StatusNotFound)
			return
		}
		if errors.Is(err, repository.ErrValidation) {
			writeError(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.log.ErrorContext(r.Context(), "failed to delete task",
			slog.Int("task_id", id),
			slog.Int("user_id", userID),
			slog.Any("err", err),
		)
		writeError(w, "error on the server side", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
