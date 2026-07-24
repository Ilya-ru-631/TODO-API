package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"todo_api/internal/handler"
	"todo_api/internal/repository"
)

func main() {
	repo := repository.NewMemoryRepo()
	h := handler.NewTaskHandler(repo)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/tasks", func(r chi.Router) {
		r.Get("/", h.GetAll)
		r.Get("/{id}", h.GetByID)
		r.Post("/", h.Create)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})

	if err := http.ListenAndServe("localhost:8080", r); err != nil {
		log.Fatal(err)
	}
}
