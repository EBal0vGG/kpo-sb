package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hse/file-store/internal/repo"
	"github.com/hse/file-store/internal/service"
	"github.com/hse/file-store/internal/storage"
	"github.com/hse/pkg-config"
	"github.com/hse/pkg-httpx"
)

// New constructs the HTTP router for the file-store service.
func New(log *slog.Logger, cfg config.Config) http.Handler {
	// Initialize dependencies (in-memory for MVP)
	workRepo := repo.NewMemory()
	fileStorage := storage.NewMemory()
	svc := service.NewWithLogger(workRepo, fileStorage, cfg.HTTP.MaxBodyBytes, log)

	handlers := NewHandlers(svc, log)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger) // simple logging for now
	r.Use(middleware.Recoverer)
	r.Use(func(next http.Handler) http.Handler {
		return httpx.MaxBytes(next, cfg.HTTP.MaxBodyBytes)
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/works", func(r chi.Router) {
		r.Post("/", handlers.UploadWork)
		r.Get("/{id}", handlers.GetWork)
		r.Get("/{id}/file", handlers.DownloadFile)
	})

	return r
}


