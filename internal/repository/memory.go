package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"proga/todo_api/internal/models"
	"strconv"
	"sync"
)

type MemoryRepo struct {
	notes  map[int]models.Task
	nextID int
	mu     sync.RWMutex
}

func NewNoteStore() *MemoryRepo {
	return &MemoryRepo{
		notes:  make(map[int]models.Task),
		nextID: 1,
	}
}

func (m *MemoryRepo) GetAll(ctx context.Context) ([]models.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]models.Task, 0, len(m.notes))
	for _, v := range m.notes {
		tasks = append(tasks, v)
	}

	return tasks, nil
}

func (m *MemoryRepo) GetByID(ctx context.Context, id int) (models.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.notes[id]
	if !ok {
		return models.Task{}, fmt.Errorf("ID must be a number: %w", ErrTaskNotFound)
	}

	return t, nil
}

func (m *MemoryRepo) Create(ctx context.Context, t models.Task) (models.Task, error) {
	var n Task

	if n.Text == nil || *n.Text == "" {
		return models.Task{}, fmt.Errorf("Text should  not be empty: %w", ErrTaskNotFound)


	// if n.Description == nil || *n.Description == "" {
	// 	return models.Task{}, fmt.Errorf("The header should  not be empty: %w", ErrTaskNotFound)
	// }

	m.mu.Lock()
	n.ID = m.nextID
	m.notes[n.ID] = n
	m.nextID++
	m.mu.Unlock()

	return m, nil

}

func (ns *NoteStore) PutNotes(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	num, err := strconv.Atoi(id)
	if err != nil {
		writeError(w, "id должно быть числом", http.StatusBadRequest)
		return
	}

	ns.mu.RLock()
	value, ok := ns.notes[num]
	ns.mu.RUnlock()
	if !ok {
		writeError(w, "Запись не найдена", http.StatusNotFound)
		return
	}

	var s Task
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeError(w, "Некорректный Json: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	ns.mu.Lock()
	s.ID = value.ID
	ns.notes[value.ID] = s
	ns.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(s)
}

func (ns *NoteStore) DeleteNotes(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	num, err := strconv.Atoi(id)
	if err != nil {
		writeError(w, "id должно быть числом", http.StatusBadRequest)
		return
	}

	ns.mu.RLock()
	value, ok := ns.notes[num]
	ns.mu.RUnlock()
	if !ok {
		writeError(w, "Запись не найдена", http.StatusNotFound)
		return
	}

	ns.mu.Lock()
	delete(ns.notes, value.ID)
	ns.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func (ns *NoteStore) PatchNotes(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	num, err := strconv.Atoi(id)
	if err != nil {
		writeError(w, "id должно быть числом", http.StatusBadRequest)
		return
	}

	var s Task
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeError(w, "Некорретный Json: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	ns.mu.Lock()
	defer ns.mu.Unlock()
	notes, ok := ns.notes[num]
	if !ok {
		writeError(w, "Запись не найдена", http.StatusNotFound)
		return
	}

	if s.Text != nil {
		notes.Text = s.Text
	}
	if s.Title != nil {
		notes.Title = s.Title
	}
	ns.notes[num] = notes

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notes)
}

func main() {
	st := NewNoteStore()
	mux := http.NewServeMux()

	mux.HandleFunc("GET /notes", st.GetNotes)
	mux.HandleFunc("GET /notes/{id}", st.GetNotesID)
	mux.HandleFunc("POST /notes", st.PostNotes)
	mux.HandleFunc("PUT /notes/{id}", st.PutNotes)
	mux.HandleFunc("DELETE /notes/{id}", st.DeleteNotes)
	mux.HandleFunc("PATCH /notes/{id}", st.PatchNotes)
	if err := http.ListenAndServe("localhost:8080", mux); err != nil {
		log.Fatal(err)
	}
}
