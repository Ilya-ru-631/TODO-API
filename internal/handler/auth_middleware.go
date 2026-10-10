package handler

import (
	"context"
	"net/http"
	"strings"
	"todo_api/internal/auth"
)

type contextKey string

const userIDKey contextKey = "userID"

func AuthMiddleware(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			token, found := strings.CutPrefix(token, "Bearer ")
			if !found {
				writeError(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			userId, err := auth.ParsingToken(token, secret)
			if err != nil {
				writeError(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userId)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
