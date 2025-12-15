package repo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hse/file-analysis/internal/domain"
)

// MemoryRepo is an in-memory repository for local dev/tests.
type MemoryRepo struct {
	mu      sync.RWMutex
	seq     int64
	reports map[int64]domain.Report
	works   map[int64]WorkInfo // workID -> WorkInfo for plagiarism checks
}

func NewMemory() *MemoryRepo {
	return &MemoryRepo{
		reports: make(map[int64]domain.Report),
		works:   make(map[int64]WorkInfo),
	}
}

func (m *MemoryRepo) Create(ctx context.Context, r *domain.Report) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	r.ID = m.seq
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}
	if r.UpdatedAt.IsZero() {
		r.UpdatedAt = r.CreatedAt
	}
	m.reports[r.ID] = *r
	return nil
}

func (m *MemoryRepo) GetByID(ctx context.Context, id int64) (*domain.Report, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.reports[id]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("report with id %d not found", id))
	}
	return &r, nil
}

func (m *MemoryRepo) GetByWorkID(ctx context.Context, workID int64) ([]domain.Report, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var res []domain.Report
	for _, r := range m.reports {
		if r.WorkID == workID {
			res = append(res, r)
		}
	}
	return res, nil
}

func (m *MemoryRepo) Update(ctx context.Context, r *domain.Report) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.reports[r.ID]; !ok {
		return domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("report with id %d not found", r.ID))
	}
	r.UpdatedAt = time.Now().UTC()
	m.reports[r.ID] = *r
	return nil
}



