package logger

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{
		ResponseWriter: w,
		statusCode: 200,
	}
}

func(r *responseWriterWrapper) WriteHeader(code int)  {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func LoggerMiddleware(sloger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := newResponseWriter(w)
			next.ServeHTTP(ww, r)
			sloger.Info("request handled",
				slog.String("method", r.Method),
				slog.String("url-path", r.URL.Path),
				slog.Int("status", ww.statusCode),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

func New(env string) *slog.Logger {
	level := slog.LevelInfo
	if env == "dev" {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}
