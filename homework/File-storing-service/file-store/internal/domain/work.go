package domain

import "time"

// Work describes uploaded assignment metadata.
type Work struct {
	ID           int64
	StudentID    string
	AssignmentID string
	Filename     string
	StoragePath  string
	Mime         string
	Size         int64
	Hash         string
	CreatedAt    time.Time
}



