package repo

import (
	"context"
	"time"
)

// WorkRepository defines operations for accessing work information for plagiarism checks.
type WorkRepository interface {
	FindByHash(ctx context.Context, assignmentID, hash string) ([]WorkInfo, error)
	RegisterWork(workID int64, info WorkInfo)
}

// WorkInfo is minimal work info needed for plagiarism check.
type WorkInfo struct {
	ID           int64
	StudentID    string
	AssignmentID string
	Hash         string
	CreatedAt    time.Time
}

