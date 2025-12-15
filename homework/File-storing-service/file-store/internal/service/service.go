package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hse/file-store/internal/domain"
	"github.com/hse/file-store/internal/repo"
	"github.com/hse/file-store/internal/storage"
)

// Service orchestrates upload flow.
type Service struct {
	repo    repo.WorkRepository
	storage storage.Storage
	maxSize int64
	log     *slog.Logger
}

func New(repo repo.WorkRepository, storage storage.Storage, maxSize int64) *Service {
	return &Service{repo: repo, storage: storage, maxSize: maxSize, log: slog.Default()}
}

func NewWithLogger(repo repo.WorkRepository, storage storage.Storage, maxSize int64, log *slog.Logger) *Service {
	return &Service{repo: repo, storage: storage, maxSize: maxSize, log: log}
}

// Upload handles validation, storage and metadata persistence.
func (s *Service) Upload(ctx context.Context, input UploadRequest) (*domain.Work, error) {
	if input.Size > s.maxSize {
		return nil, domain.NewDomainError(domain.ErrCodeTooLarge, fmt.Sprintf("file size %d exceeds maximum %d", input.Size, s.maxSize))
	}
	if !isAllowedMime(input.Mime) {
		return nil, domain.NewDomainError(domain.ErrCodeUnsupportedType, fmt.Sprintf("unsupported MIME type: %s", input.Mime))
	}

	objectKey := fmt.Sprintf("%s/%s", input.AssignmentID, input.Filename)

	if err := s.storage.Save(ctx, objectKey, input.Content, input.Mime); err != nil {
		return nil, fmt.Errorf("store file: %w", err)
	}

	work := &domain.Work{
		StudentID:    input.StudentID,
		AssignmentID: input.AssignmentID,
		Filename:     input.Filename,
		StoragePath:  objectKey,
		Mime:         input.Mime,
		Size:         input.Size,
		Hash:         input.Hash,
	}

	if err := s.repo.Create(ctx, work); err != nil {
		return nil, fmt.Errorf("repo create: %w", err)
	}
	
	// Verify StoragePath was saved correctly
	if work.StoragePath == "" {
		s.log.Warn("StoragePath is empty after Create, this should not happen", 
			slog.Int64("id", work.ID), 
			slog.String("object_key", objectKey))
	}
	
	return work, nil
}

// GetByID retrieves work metadata by ID.
func (s *Service) GetByID(ctx context.Context, id int64) (*domain.Work, error) {
	work, err := s.repo.GetByID(ctx, id)
	if err != nil {
		// Preserve domain errors
		return nil, err
	}
	// Ensure StoragePath is set (for backward compatibility with old records)
	if work.StoragePath == "" {
		if work.AssignmentID != "" && work.Filename != "" {
			work.StoragePath = fmt.Sprintf("%s/%s", work.AssignmentID, work.Filename)
			s.log.Info("restoring StoragePath", slog.Int64("id", id), slog.String("storage_path", work.StoragePath))
			// Update repository with restored StoragePath
			if err := s.repo.Update(ctx, work); err != nil {
				s.log.Warn("failed to update StoragePath in repository", slog.Int64("id", id), slog.Any("err", err))
				// Continue - storage_path is restored for this request
			}
		} else {
			s.log.Warn("StoragePath is empty and cannot be restored", 
				slog.Int64("id", id), 
				slog.String("assignment_id", work.AssignmentID), 
				slog.String("filename", work.Filename))
		}
	}
	return work, nil
}

// Download retrieves file content by work ID.
func (s *Service) Download(ctx context.Context, id int64) ([]byte, string, error) {
	work, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, "", err
	}
	
	// Ensure StoragePath is set (for backward compatibility with old records)
	if work.StoragePath == "" {
		if work.AssignmentID != "" && work.Filename != "" {
			work.StoragePath = fmt.Sprintf("%s/%s", work.AssignmentID, work.Filename)
			s.log.Info("restoring StoragePath in Download", slog.Int64("id", id), slog.String("storage_path", work.StoragePath))
			// Update repository with restored StoragePath
			if err := s.repo.Update(ctx, work); err != nil {
				s.log.Warn("failed to update StoragePath in repository", slog.Int64("id", id), slog.Any("err", err))
				// Continue - storage_path is restored for this request
			}
		} else {
			return nil, "", domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("work %d has no storage path and cannot be restored", id))
		}
	}
	
	data, contentType, err := s.storage.Load(ctx, work.StoragePath)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("file not found: %s", work.StoragePath))
	}
	return data, contentType, nil
}

// UploadRequest encapsulates upload parameters.
type UploadRequest struct {
	StudentID    string
	AssignmentID string
	Filename     string
	Mime         string
	Size         int64
	Hash         string
	Content      []byte
}

func isAllowedMime(m string) bool {
	switch m {
	case "text/plain", "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return true
	default:
		return false
	}
}


