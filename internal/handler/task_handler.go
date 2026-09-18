package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	//"todo_api/internal/models"

	"todo_api/internal/repository"
	"todo_api/internal/service"
)

type TaskHandler struct {
	service *service.TaskService
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var LimitForBodyLength = 500

func NewTaskHandler(svc *service.TaskService) *TaskHandler {
	return &TaskHandler{service: svc}
}

func writeError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}

func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.service.GetAll(r.Context())
	if err != nil {
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

	task, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			writeError(w, "not found task with this id", http.StatusNotFound)
			return
		}
		writeError(w, "error on the server side", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	task, err := h.service.Create(r.Context(), req.ToTask())
	if err != nil {
		if errors.Is(err, service.ErrDuplicateTitle) {
			writeError(w, "a task with such a title already exists", http.StatusConflict)
			return
		}
		writeError(w, "error when submitting a task", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	task, err := h.service.Update(r.Context(), id, req.ToTask())
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			writeError(w, "not found this task", http.StatusNotFound)
			return
		}
		if errors.Is(err, repository.ErrValidation) {
			writeError(w, err.Error(), http.StatusBadRequest)
			return
		}
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

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			writeError(w, "not found this task", http.StatusNotFound)
			return
		}
		if errors.Is(err, repository.ErrValidation) {
			writeError(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeError(w, "error on the server side", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
