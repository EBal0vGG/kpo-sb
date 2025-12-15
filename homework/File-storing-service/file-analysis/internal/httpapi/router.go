package httpapi

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hse/file-analysis/internal/analyzer"
	"github.com/hse/file-analysis/internal/repo"
	"github.com/hse/file-analysis/internal/service"
	"github.com/hse/file-analysis/internal/storage"
	"github.com/hse/file-analysis/internal/worker"
	"github.com/hse/pkg-config"
	"github.com/hse/pkg-httpx"
)

// New constructs the HTTP router for the file-analysis service.
func New(log *slog.Logger, cfg config.Config) (http.Handler, *worker.Worker) {
	// Initialize dependencies (in-memory for MVP)
	reportRepo := repo.NewMemory()
	workRepo := repo.NewMemoryWorkRepo()
	fileStorage := storage.NewMemory()

	// Parse plagiarism threshold from env
	threshold := 0.7
	if t := os.Getenv("PLAGIARISM_THRESHOLD"); t != "" {
		if parsed, err := strconv.ParseFloat(t, 64); err == nil {
			threshold = parsed
		}
	}

	analyzer := analyzer.New(threshold)
	
	// Get Store service URL from environment
	storeURL := os.Getenv("STORE_SERVICE_URL")
	if storeURL == "" {
		storeURL = "http://file-store:8081"
	}
	
	w := worker.New(reportRepo, workRepo, fileStorage, analyzer, log, storeURL)

	svc := service.New(reportRepo, w)
	handlers := NewHandlers(svc, log)

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

	r.Post("/analysis/jobs", handlers.SubmitAnalysisJob)
	r.Get("/works/{work_id}/reports", handlers.GetReports)

	return r, w
}


