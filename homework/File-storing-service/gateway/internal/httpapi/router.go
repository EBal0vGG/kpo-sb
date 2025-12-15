package httpapi

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hse/file-gateway/internal/client"
	"github.com/hse/pkg-config"
	"github.com/hse/pkg-httpx"
)

// New constructs the HTTP router for the gateway service.
func New(log *slog.Logger, cfg config.Config) http.Handler {
	storeURL := os.Getenv("STORE_SERVICE_URL")
	if storeURL == "" {
		storeURL = "http://file-store:8081"
	}

	analysisURL := os.Getenv("ANALYSIS_SERVICE_URL")
	if analysisURL == "" {
		analysisURL = "http://file-analysis:8082"
	}

	storeClient := client.NewStoreClient(storeURL)
	analysisClient := client.NewAnalysisClient(analysisURL)
	handlers := NewHandlers(storeClient, analysisClient, log)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
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
		r.Get("/{work_id}/reports", handlers.GetReports)
	})

	return r
}



