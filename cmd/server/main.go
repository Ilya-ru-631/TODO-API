package main

import (
	"log"
	"net/http"
	"todo_api/internal/handler"
	"todo_api/internal/repository"
)

func main() {
	repo := repository.NewMemoryRepo()
	h := handler.NewTaskHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /tasks", h.GetAll)
	mux.HandleFunc("GET /tasks/{id}", h.GetByID)
	mux.HandleFunc("POST /tasks", h.Create)
	mux.HandleFunc("PUT /tasks/{id}", h.Update)
	mux.HandleFunc("DELETE /tasks/{id}", h.Delete)
	if err := http.ListenAndServe("localhost:8080", mux); err != nil {
		log.Fatal(err)
	}
}
