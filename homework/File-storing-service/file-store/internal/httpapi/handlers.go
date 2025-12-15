package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hse/file-store/internal/domain"
	"github.com/hse/file-store/internal/service"
	"github.com/hse/file-store/internal/validator"
	"github.com/hse/pkg-httpx"
)

type Handlers struct {
	svc *service.Service
	log *slog.Logger
}

func NewHandlers(svc *service.Service, log *slog.Logger) *Handlers {
	return &Handlers{svc: svc, log: log}
}

// UploadWork handles multipart form upload: file, student_id, assignment_id.
func (h *Handlers) UploadWork(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with large limit (actual limit enforced by MaxBytes middleware)
	// Use 50MB as ParseMultipartForm limit to allow multipart overhead
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		// Check if error is due to request being too large
		if err == io.EOF || err == http.ErrNotSupported || 
		   err.Error() == "http: request body too large" ||
		   err.Error() == "multipart: NextPart: http: request body too large" {
			httpx.Error(w, http.StatusRequestEntityTooLarge, "request entity too large")
		} else {
			httpx.Error(w, http.StatusBadRequest, fmt.Sprintf("parse form: %v", err))
		}
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		// Check if error is due to request being too large
		if err == io.EOF || err.Error() == "http: request body too large" {
			httpx.Error(w, http.StatusRequestEntityTooLarge, "request entity too large")
		} else {
			httpx.Error(w, http.StatusBadRequest, fmt.Sprintf("get file: %v", err))
		}
		return
	}
	defer file.Close()

	studentID := r.FormValue("student_id")
	assignmentID := r.FormValue("assignment_id")

	content, err := io.ReadAll(file)
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

	// Validate request
	if err := validator.ValidateUploadRequest(header.Filename, studentID, assignmentID, int64(len(content)), 20<<20); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Compute SHA256 hash
	hash := sha256.Sum256(content)
	hashStr := hex.EncodeToString(hash[:])

	// Detect MIME type
	mime := validator.DetectMimeType(header.Filename, header.Header.Get("Content-Type"))
	if err := validator.ValidateMimeType(mime); err != nil {
		httpx.Error(w, http.StatusUnsupportedMediaType, err.Error())
		return
	}

	req := service.UploadRequest{
		StudentID:    studentID,
		AssignmentID: assignmentID,
		Filename:     header.Filename,
		Mime:         mime,
		Size:         int64(len(content)),
		Hash:         hashStr,
		Content:      content,
	}

	work, err := h.svc.Upload(r.Context(), req)
	if err != nil {
		if domainErr, ok := domain.IsDomainError(err); ok {
			switch domainErr.Code {
			case domain.ErrCodeTooLarge:
				httpx.Error(w, http.StatusRequestEntityTooLarge, domainErr.Message)
			case domain.ErrCodeUnsupportedType:
				httpx.Error(w, http.StatusUnsupportedMediaType, domainErr.Message)
			case domain.ErrCodeNotFound:
				httpx.Error(w, http.StatusNotFound, domainErr.Message)
			case domain.ErrCodeInvalidInput:
				httpx.Error(w, http.StatusBadRequest, domainErr.Message)
			default:
				h.log.Error("upload failed", slog.Any("err", err))
				httpx.Error(w, http.StatusInternalServerError, "upload failed")
			}
		} else {
			h.log.Error("upload failed", slog.Any("err", err))
			httpx.Error(w, http.StatusInternalServerError, "upload failed")
		}
		return
	}

	httpx.JSON(w, http.StatusCreated, map[string]interface{}{
		"id":            work.ID,
		"student_id":    work.StudentID,
		"assignment_id": work.AssignmentID,
		"filename":      work.Filename,
		"hash":          work.Hash,
		"size":          work.Size,
		"created_at":    work.CreatedAt,
		"storage_path":  work.StoragePath,
	})
}

// GetWork returns work metadata by ID.
func (h *Handlers) GetWork(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid work id")
		return
	}

	work, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if domainErr, ok := domain.IsDomainError(err); ok && domainErr.Code == domain.ErrCodeNotFound {
			httpx.Error(w, http.StatusNotFound, domainErr.Message)
		} else {
			h.log.Error("get work failed", slog.Int64("id", id), slog.Any("err", err))
			httpx.Error(w, http.StatusInternalServerError, "failed to get work")
		}
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]interface{}{
		"id":            work.ID,
		"student_id":    work.StudentID,
		"assignment_id": work.AssignmentID,
		"filename":      work.Filename,
		"hash":          work.Hash,
		"size":          work.Size,
		"mime":          work.Mime,
		"created_at":    work.CreatedAt,
		"storage_path":  work.StoragePath,
	})
}

// DownloadFile streams the file content.
func (h *Handlers) DownloadFile(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid work id")
		return
	}

	data, contentType, err := h.svc.Download(r.Context(), id)
	if err != nil {
		if domainErr, ok := domain.IsDomainError(err); ok && domainErr.Code == domain.ErrCodeNotFound {
			httpx.Error(w, http.StatusNotFound, domainErr.Message)
		} else {
			h.log.Error("download failed", slog.Int64("id", id), slog.Any("err", err))
			httpx.Error(w, http.StatusInternalServerError, "failed to download file")
		}
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}


