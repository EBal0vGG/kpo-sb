package repo

import (
	"context"

	"github.com/hse/file-store/internal/domain"
)

// WorkRepository defines persistence operations for works.
type WorkRepository interface {
	Create(ctx context.Context, w *domain.Work) error
	GetByID(ctx context.Context, id int64) (*domain.Work, error)
	FindByHash(ctx context.Context, assignmentID, hash string) ([]domain.Work, error)
	Update(ctx context.Context, w *domain.Work) error
}


