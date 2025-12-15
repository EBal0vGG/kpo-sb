package repo

import (
	"context"

	"github.com/hse/file-analysis/internal/domain"
)

// ReportRepository defines persistence for reports.
type ReportRepository interface {
	Create(ctx context.Context, r *domain.Report) error
	GetByID(ctx context.Context, id int64) (*domain.Report, error)
	GetByWorkID(ctx context.Context, workID int64) ([]domain.Report, error)
	Update(ctx context.Context, r *domain.Report) error
}


