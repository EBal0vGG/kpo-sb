package service

import (
	"context"

	"github.com/hse/file-analysis/internal/domain"
	"github.com/hse/file-analysis/internal/repo"
	"github.com/hse/file-analysis/internal/worker"
)

// Service orchestrates analysis operations.
type Service struct {
	repo   repo.ReportRepository
	worker *worker.Worker
}

func New(repo repo.ReportRepository, w *worker.Worker) *Service {
	return &Service{repo: repo, worker: w}
}

// SubmitJob enqueues an analysis job.
func (s *Service) SubmitJob(ctx context.Context, job domain.AnalysisJob) error {
	if err := s.worker.Enqueue(job); err != nil {
		return domain.NewDomainError(domain.ErrCodeQueueFull, "job queue is full, please try again later")
	}
	return nil
}

// GetReportsByWorkID returns all reports for a work.
func (s *Service) GetReportsByWorkID(ctx context.Context, workID int64) ([]domain.Report, error) {
	return s.repo.GetByWorkID(ctx, workID)
}

// GetReportByID returns a specific report.
func (s *Service) GetReportByID(ctx context.Context, id int64) (*domain.Report, error) {
	return s.repo.GetByID(ctx, id)
}


