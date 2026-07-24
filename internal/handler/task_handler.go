package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"todo_api/internal/models"
	"todo_api/internal/repository"
)

type TaskHandler struct {
	repo repository.TaskRepository
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewTaskHandler(repo repository.TaskRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}

func writeError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}

func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.repo.GetAll(r.Context())
	if err != nil {
		writeError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, "ID must be a number", http.StatusBadRequest)
		return
	}

	task, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			writeError(w, "Not found task with this id", http.StatusNotFound)
			return
		}
		writeError(w, "Error on the server side", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var n models.Task
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		writeError(w, "Uncorrect Json: ", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	task, err := h.repo.Create(r.Context(), n)
	if err != nil {
		writeError(w, "Error when submitting a task", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	var n models.Task
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		writeError(w, "Uncorrect data: ", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, "ID must be a number", http.StatusBadRequest)
		return
	}

	task, err := h.repo.Update(r.Context(), id, n)
	if err != nil {
		writeError(w, "Not found this task", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, "ID must be a number", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeError(w, "Not found this task", http.StatusNotFound)
		return
	}
}
