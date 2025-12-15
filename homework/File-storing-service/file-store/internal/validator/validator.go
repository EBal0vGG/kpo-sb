package validator

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateUploadRequest validates upload request parameters.
func ValidateUploadRequest(filename, studentID, assignmentID string, size int64, maxSize int64) error {
	if filename == "" {
		return fmt.Errorf("filename is required")
	}
	if studentID == "" {
		return fmt.Errorf("student_id is required")
	}
	if assignmentID == "" {
		return fmt.Errorf("assignment_id is required")
	}
	if size <= 0 {
		return fmt.Errorf("file size must be greater than 0")
	}
	if size > maxSize {
		return fmt.Errorf("file size %d exceeds maximum %d", size, maxSize)
	}
	return nil
}

// ValidateMimeType checks if MIME type is allowed.
func ValidateMimeType(mime string) error {
	allowed := []string{
		"text/plain",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}
	for _, allowedMime := range allowed {
		if mime == allowedMime {
			return nil
		}
	}
	return fmt.Errorf("unsupported MIME type: %s", mime)
}

// DetectMimeType detects MIME type from filename extension.
func DetectMimeType(filename, providedMime string) string {
	if providedMime != "" && providedMime != "application/octet-stream" {
		return providedMime
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".txt":
		return "text/plain"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	default:
		return "application/octet-stream"
	}
}

