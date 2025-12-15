package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hse/file-gateway/internal/client"
	"github.com/hse/pkg-httpx"
)

type Handlers struct {
	storeClient    *client.StoreClient
	analysisClient *client.AnalysisClient
	log            *slog.Logger
	jobTimeout     time.Duration
}

func NewHandlers(storeClient *client.StoreClient, analysisClient *client.AnalysisClient, log *slog.Logger) *Handlers {
	return &Handlers{
		storeClient:    storeClient,
		analysisClient: analysisClient,
		log:            log,
		jobTimeout:     30 * time.Second, // Timeout for async job submission
	}
}

// UploadWork handles POST /works - proxies to store service and triggers analysis.
func (h *Handlers) UploadWork(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with large limit (actual limit enforced by MaxBytes middleware)
	// Use 50MB as ParseMultipartForm limit to allow multipart overhead
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		// Check if error is due to request being too large
		// MaxBytesReader returns io.EOF or "http: request body too large" when limit exceeded
		if err == io.EOF || err == http.ErrNotSupported || 
		   err.Error() == "http: request body too large" ||
		   err.Error() == "multipart: NextPart: http: request body too large" {
			httpx.Error(w, http.StatusRequestEntityTooLarge, "request entity too large")
		} else {
			httpx.Error(w, http.StatusBadRequest, "failed to parse multipart form")
		}
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	studentID := r.FormValue("student_id")
	if studentID == "" {
		httpx.Error(w, http.StatusBadRequest, "student_id is required")
		return
	}

	assignmentID := r.FormValue("assignment_id")
	if assignmentID == "" {
		httpx.Error(w, http.StatusBadRequest, "assignment_id is required")
		return
	}

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		// Check if error is due to request being too large
		if err == io.EOF || err.Error() == "http: request body too large" {
			httpx.Error(w, http.StatusRequestEntityTooLarge, "request entity too large")
		} else {
			h.log.Error("failed to read file", slog.Any("err", err))
			httpx.Error(w, http.StatusInternalServerError, "failed to read file")
		}
		return
	}

	// Upload to store service
	uploadResp, err := h.storeClient.UploadWork(r.Context(), fileContent, header.Filename, studentID, assignmentID)
	if err != nil {
		// Check if error is due to request being too large or unsupported media type
		errMsg := err.Error()
		if errMsg != "" {
			if len(errMsg) >= 23 && errMsg[:23] == "request entity too large" {
				httpx.Error(w, http.StatusRequestEntityTooLarge, "request entity too large")
			} else if len(errMsg) >= 20 && errMsg[:20] == "unsupported media type" {
				httpx.Error(w, http.StatusUnsupportedMediaType, "unsupported media type")
			} else {
				h.log.Error("store service upload failed", slog.Any("err", err))
				httpx.Error(w, http.StatusBadGateway, "store service unavailable")
			}
		} else {
			h.log.Error("store service upload failed", slog.Any("err", err))
			httpx.Error(w, http.StatusBadGateway, "store service unavailable")
		}
		return
	}

	// Return response
	httpx.JSON(w, http.StatusCreated, uploadResp)

	// Trigger analysis asynchronously with timeout
	go h.submitAnalysisJobAsync(uploadResp)
}

// submitAnalysisJobAsync submits analysis job asynchronously with timeout and error handling.
func (h *Handlers) submitAnalysisJobAsync(uploadResp *client.UploadWorkResponse) {
	ctx, cancel := context.WithTimeout(context.Background(), h.jobTimeout)
	defer cancel()

	jobReq := client.SubmitJobRequest{
		WorkID:       uploadResp.ID,
		AssignmentID: uploadResp.AssignmentID,
		StudentID:    uploadResp.StudentID,
		FileHash:     uploadResp.Hash,
		StoragePath:  uploadResp.StoragePath,
	}

	if err := h.analysisClient.SubmitJob(ctx, jobReq); err != nil {
		h.log.Error("failed to submit analysis job",
			slog.Int64("work_id", uploadResp.ID),
			slog.Any("err", err))
	}
}

// GetReports handles GET /works/{work_id}/reports - proxies to analysis service.
func (h *Handlers) GetReports(w http.ResponseWriter, r *http.Request) {
	workIDStr := chi.URLParam(r, "work_id")
	workID, err := strconv.ParseInt(workIDStr, 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid work_id")
		return
	}

	reports, err := h.analysisClient.GetReports(r.Context(), workID)
	if err != nil {
		// Check if error is "not found"
		errMsg := err.Error()
		if errMsg != "" && len(errMsg) >= 17 && errMsg[:17] == "reports not found" {
			httpx.Error(w, http.StatusNotFound, "reports not found")
		} else {
			h.log.Error("get reports failed", slog.Int64("work_id", workID), slog.Any("err", err))
			httpx.Error(w, http.StatusBadGateway, "analysis service unavailable")
		}
		return
	}

	httpx.JSON(w, http.StatusOK, reports)
}

// GetWork handles GET /works/{id} - proxies to store service.
func (h *Handlers) GetWork(w http.ResponseWriter, r *http.Request) {
	workIDStr := chi.URLParam(r, "id")
	workID, err := strconv.ParseInt(workIDStr, 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid work id")
		return
	}

	work, err := h.storeClient.GetWork(r.Context(), workID)
	if err != nil {
		// Check if error is "not found"
		errMsg := err.Error()
		if errMsg != "" && len(errMsg) >= 13 && errMsg[:13] == "work not found" {
			httpx.Error(w, http.StatusNotFound, "work not found")
		} else {
			h.log.Error("get work failed", slog.Int64("work_id", workID), slog.Any("err", err))
			httpx.Error(w, http.StatusBadGateway, "store service unavailable")
		}
		return
	}

	httpx.JSON(w, http.StatusOK, work)
}
