package domain

import "time"

// ReportStatus represents analysis status.
type ReportStatus string

const (
	StatusPending ReportStatus = "pending"
	StatusRunning ReportStatus = "running"
	StatusDone    ReportStatus = "done"
	StatusFailed  ReportStatus = "failed"
)

// Report describes plagiarism analysis result.
type Report struct {
	ID             int64
	WorkID         int64
	Status         ReportStatus
	PlagiarismFlag bool
	Score          float64 // 0.0 to 1.0
	DetailsPath    string  // path to JSON report file
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// AnalysisJob represents a pending analysis task.
type AnalysisJob struct {
	WorkID       int64
	AssignmentID string
	StudentID    string
	FileHash     string
	StoragePath  string
}



