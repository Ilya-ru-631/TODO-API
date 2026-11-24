package handler

import (
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"todo_api/internal/apierror"
	"todo_api/internal/service"
)

type TaskHandler struct {
	service *service.TaskService
	log     *slog.Logger
}

var LimitForBodyLength = 500

func NewTaskHandler(svc *service.TaskService, log *slog.Logger) *TaskHandler {
	return &TaskHandler{service: svc, log: log}
}

func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		apierror.Write(w, http.StatusUnauthorized, apierror.CodeUnauthorized, "Unauthorized user")
		return
	}

	tasks, err := h.service.GetAll(r.Context(), userID)
	if err != nil {
		h.log.ErrorContext(r.Context(), "failed to get tasks",
			slog.Int("user_id", userID),
			slog.Any("err", err),
		)
		apierror.Write(w, http.StatusInternalServerError, apierror.CodeInternal, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		apierror.Write(w, http.StatusBadRequest, apierror.CodeInvalidID, "Id must be a number")
		return
	}

	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		apierror.Write(w, http.StatusUnauthorized, apierror.CodeUnauthorized, "Unauthorized user")
		return
	}

	task, err := h.service.GetByID(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			apierror.Write(w, http.StatusNotFound, apierror.CodeTaskNotFound, "Not found task with this id")
			return
		}

		h.log.ErrorContext(r.Context(), "failed to get task",
			slog.Int("task_id", id),
			slog.Int("user_id", userID),
			slog.Any("err", err),
		)
		apierror.Write(w, http.StatusInternalServerError, apierror.CodeInternal, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		apierror.Write(w, http.StatusUnauthorized, apierror.CodeUnauthorized, "Unauthorized user")
		return
	}

	var req CreateTaskRequest

	r.Body = http.MaxBytesReader(w, r.Body, int64(LimitForBodyLength))

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, http.StatusBadRequest, apierror.CodeInvalidJSON, "Incorrect json")
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	titleClear := strings.TrimSpace(req.Title)
	if titleClear == "" {
		apierror.WriteValidation(w, apierror.CodeValidation, "Empty title", "title")
		return
	}

	req.Title = titleClear

	task, err := h.service.Create(r.Context(), userID, req.ToTask())
	if err != nil {
		if errors.Is(err, service.ErrDuplicateTitle) {
			apierror.Write(w, http.StatusConflict, apierror.CodeDuplicateTitle, "A task with such a title already exists")
			return
		}
		h.log.ErrorContext(r.Context(), "failed to create task",
			slog.Int("user_id", userID),
			slog.Any("err", err),
		)
		apierror.Write(w, http.StatusInternalServerError, apierror.CodeInternal, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		apierror.Write(w, http.StatusUnauthorized, apierror.CodeUnauthorized, "Unauthorized user")
		return
	}

	var req UpdateTaskRequest

	r.Body = http.MaxBytesReader(w, r.Body, int64(LimitForBodyLength))

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, http.StatusBadRequest, apierror.CodeInvalidJSON, "Incorrect json")
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {	
		apierror.Write(w, http.StatusBadRequest, apierror.CodeInvalidID, "Id must be a number")
		return
	}

	titleClear := strings.TrimSpace(req.Title)
	if titleClear == "" {
		apierror.WriteValidation(w, apierror.CodeValidation, "Empty title", "title")
		return
	}

	req.Title = titleClear

	task, err := h.service.Update(r.Context(), id, userID, req.ToTask())
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			apierror.Write(w, http.StatusNotFound, apierror.CodeTaskNotFound, "Not found task with this id")
			return
		}
		if errors.Is(err, service.ErrValidation) {
			apierror.WriteValidation(w, apierror.CodeValidation, err.Error(), "title")
			return
		}
		h.log.ErrorContext(r.Context(), "failed to update task",
			slog.Int("task_id", id),
			slog.Int("user_id", userID),
			slog.Any("err", err),
		)
		apierror.Write(w, http.StatusInternalServerError, apierror.CodeInternal, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		apierror.Write(w, http.StatusBadRequest, apierror.CodeInvalidID, "Id must be a number")
		return
	}

	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		apierror.Write(w, http.StatusUnauthorized, apierror.CodeUnauthorized, "Unauthorized user")
		return
	}

	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			apierror.Write(w, http.StatusNotFound, apierror.CodeTaskNotFound, "Not found task with this id")
			return
		}
		h.log.ErrorContext(r.Context(), "failed to delete task",
			slog.Int("task_id", id),
			slog.Int("user_id", userID),
			slog.Any("err", err),
		)
		apierror.Write(w, http.StatusInternalServerError, apierror.CodeInternal, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
