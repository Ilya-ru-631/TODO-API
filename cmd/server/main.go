package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"todo_api/internal/handler"
	"todo_api/internal/repository"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	db, err := sql.Open("pgx", "postgres://postgres:mysecretpassword@localhost:5432/todo_api?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		log.Fatal("failed to connect to database: ", err)
	}

	repo := repository.NewPostgresRepo(db)
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
