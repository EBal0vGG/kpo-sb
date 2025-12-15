package httpapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hse/file-analysis/internal/domain"
	"github.com/hse/file-analysis/internal/service"
	"github.com/hse/pkg-httpx"
)

type Handlers struct {
	svc *service.Service
	log *slog.Logger
}

func NewHandlers(svc *service.Service, log *slog.Logger) *Handlers {
	return &Handlers{svc: svc, log: log}
}

// SubmitAnalysisJob handles POST /analysis/jobs.
func (h *Handlers) SubmitAnalysisJob(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WorkID       int64  `json:"work_id"`
		AssignmentID string `json:"assignment_id"`
		StudentID    string `json:"student_id"`
		FileHash     string `json:"file_hash"`
		StoragePath  string `json:"storage_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	if req.WorkID == 0 || req.AssignmentID == "" || req.StudentID == "" || req.FileHash == "" {
		httpx.Error(w, http.StatusBadRequest, "missing required fields")
		return
	}

	job := domain.AnalysisJob{
		WorkID:       req.WorkID,
		AssignmentID: req.AssignmentID,
		StudentID:    req.StudentID,
		FileHash:     req.FileHash,
		StoragePath:  req.StoragePath,
	}

	if err := h.svc.SubmitJob(r.Context(), job); err != nil {
		if domainErr, ok := domain.IsDomainError(err); ok {
			switch domainErr.Code {
			case domain.ErrCodeQueueFull:
				httpx.Error(w, http.StatusServiceUnavailable, domainErr.Message)
			case domain.ErrCodeInvalidInput:
				httpx.Error(w, http.StatusBadRequest, domainErr.Message)
			default:
				h.log.Error("submit job failed", slog.Any("err", err))
				httpx.Error(w, http.StatusInternalServerError, "failed to submit job")
			}
		} else {
			h.log.Error("submit job failed", slog.Any("err", err))
			httpx.Error(w, http.StatusInternalServerError, "failed to submit job")
		}
		return
	}

	httpx.JSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}

// GetReports handles GET /works/{work_id}/reports.
func (h *Handlers) GetReports(w http.ResponseWriter, r *http.Request) {
	workIDStr := chi.URLParam(r, "work_id")
	workID, err := strconv.ParseInt(workIDStr, 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid work_id")
		return
	}

	reports, err := h.svc.GetReportsByWorkID(r.Context(), workID)
	if err != nil {
		if domainErr, ok := domain.IsDomainError(err); ok && domainErr.Code == domain.ErrCodeNotFound {
			httpx.Error(w, http.StatusNotFound, domainErr.Message)
		} else {
			h.log.Error("get reports failed", slog.Int64("work_id", workID), slog.Any("err", err))
			httpx.Error(w, http.StatusInternalServerError, "failed to get reports")
		}
		return
	}

	result := make([]map[string]interface{}, 0, len(reports))
	for _, r := range reports {
		result = append(result, map[string]interface{}{
			"id":              r.ID,
			"status":          string(r.Status),
			"plagiarism_flag": r.PlagiarismFlag,
			"score":           r.Score,
			"created_at":      r.CreatedAt,
			"updated_at":      r.UpdatedAt,
		})
	}

	httpx.JSON(w, http.StatusOK, result)
}


