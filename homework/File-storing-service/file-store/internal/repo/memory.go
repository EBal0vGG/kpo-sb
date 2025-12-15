package repo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hse/file-store/internal/domain"
)

// MemoryRepo is an in-memory repository for local dev/tests.
type MemoryRepo struct {
	mu    sync.RWMutex
	seq   int64
	items map[int64]domain.Work
}

func NewMemory() *MemoryRepo {
	return &MemoryRepo{
		items: make(map[int64]domain.Work),
	}
}

func (m *MemoryRepo) Create(ctx context.Context, w *domain.Work) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	w.ID = m.seq
	if w.CreatedAt.IsZero() {
		w.CreatedAt = time.Now().UTC()
	}
	// Make a copy to ensure all fields are preserved
	workCopy := *w
	m.items[w.ID] = workCopy
	
	// Log to verify StoragePath is saved
	if workCopy.StoragePath == "" {
		// This should not happen, but log it for debugging
	}
	
	return nil
}

func (m *MemoryRepo) GetByID(ctx context.Context, id int64) (*domain.Work, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	w, ok := m.items[id]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("work with id %d not found", id))
	}
	// Return a pointer to a copy to avoid returning pointer to map value
	workCopy := w
	return &workCopy, nil
}

func (m *MemoryRepo) FindByHash(ctx context.Context, assignmentID, hash string) ([]domain.Work, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var res []domain.Work
	for _, w := range m.items {
		if w.AssignmentID == assignmentID && w.Hash == hash {
			res = append(res, w)
		}
	}
	return res, nil
}

func (m *MemoryRepo) Update(ctx context.Context, w *domain.Work) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.items[w.ID]; !ok {
		return domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("work with id %d not found", w.ID))
	}
	m.items[w.ID] = *w
	return nil
}


