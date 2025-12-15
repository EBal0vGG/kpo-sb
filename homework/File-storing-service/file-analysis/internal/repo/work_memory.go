package repo

import (
	"context"
	"sync"
)

// MemoryWorkRepo is an in-memory work repository for plagiarism checks.
type MemoryWorkRepo struct {
	mu    sync.RWMutex
	works map[int64]WorkInfo // workID -> WorkInfo
}

// NewMemoryWorkRepo creates a new in-memory work repository.
func NewMemoryWorkRepo() *MemoryWorkRepo {
	return &MemoryWorkRepo{
		works: make(map[int64]WorkInfo),
	}
}

// FindByHash finds works by assignment ID and hash.
// For plagiarism check, we need to find all works in the same assignment (not just same hash).
// This allows comparing works that are similar but not identical.
func (m *MemoryWorkRepo) FindByHash(ctx context.Context, assignmentID, hash string) ([]WorkInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var res []WorkInfo
	// Find all works in the same assignment (for text similarity comparison)
	// This allows detecting plagiarism even if hashes differ slightly
	for _, w := range m.works {
		if w.AssignmentID == assignmentID {
			res = append(res, w)
		}
	}
	return res, nil
}

// RegisterWork stores work info for plagiarism checks.
func (m *MemoryWorkRepo) RegisterWork(workID int64, info WorkInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.works[workID] = info
}

