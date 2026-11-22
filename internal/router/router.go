package router

import (
	"log/slog"
	"net/http"
	"todo_api/internal/handler"
	"todo_api/internal/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(sloger *slog.Logger, h *handler.TaskHandler, authH *handler.AuthHandler, secret string) *chi.Mux {
	r := chi.NewRouter()

	r.Use(logger.LoggerMiddleware(sloger))
	r.Use(middleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Post("/register", authH.Register)
			r.Post("/login", authH.Login)

			r.Route("/tasks", func(r chi.Router) {
				r.Use(handler.AuthMiddleware([]byte(secret)))
				r.Get("/", h.GetAll)
				r.Get("/{id}", h.GetByID)
				r.Post("/", h.Create)
				r.Put("/{id}", h.Update)
				r.Delete("/{id}", h.Delete)
			})
		})
	})

	r.Handle("/*", http.FileServer(http.Dir("web")))
	return r
}
