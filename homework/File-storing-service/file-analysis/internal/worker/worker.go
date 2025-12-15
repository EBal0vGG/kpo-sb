package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/hse/file-analysis/internal/analyzer"
	"github.com/hse/file-analysis/internal/domain"
	"github.com/hse/file-analysis/internal/repo"
	"github.com/hse/file-analysis/internal/storage"
)

// Worker processes analysis jobs.
type Worker struct {
	reportRepo repo.ReportRepository
	workRepo   repo.WorkRepository
	storage    storage.Storage
	analyzer   *analyzer.Analyzer
	log        *slog.Logger
	jobChan    chan domain.AnalysisJob
	storeURL   string
	httpClient *http.Client
}

func New(reportRepo repo.ReportRepository, workRepo repo.WorkRepository, storage storage.Storage, analyzer *analyzer.Analyzer, log *slog.Logger, storeURL string) *Worker {
	return &Worker{
		reportRepo: reportRepo,
		workRepo:   workRepo,
		storage:    storage,
		analyzer:   analyzer,
		log:        log,
		jobChan:    make(chan domain.AnalysisJob, 100),
		storeURL:   storeURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Enqueue adds a job to the processing queue.
func (w *Worker) Enqueue(job domain.AnalysisJob) error {
	select {
	case w.jobChan <- job:
		return nil
	default:
		w.log.Warn("job queue full, dropping job", slog.Int64("work_id", job.WorkID))
		return fmt.Errorf("job queue full")
	}
}

// Start begins processing jobs from the queue.
func (w *Worker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-w.jobChan:
			w.processJob(ctx, job)
		}
	}
}

func (w *Worker) processJob(ctx context.Context, job domain.AnalysisJob) {
	// Create report with pending status
	report := &domain.Report{
		WorkID:         job.WorkID,
		Status:         domain.StatusRunning,
		PlagiarismFlag: false,
		Score:          0.0,
	}

	if err := w.reportRepo.Create(ctx, report); err != nil {
		w.log.Error("failed to create report", slog.Any("err", err))
		return
	}

	// Register work in WorkRepository for plagiarism checks
	w.workRepo.RegisterWork(job.WorkID, repo.WorkInfo{
		ID:           job.WorkID,
		StudentID:    job.StudentID,
		AssignmentID: job.AssignmentID,
		Hash:         job.FileHash,
		CreatedAt:    time.Now().UTC(),
	})

	// Try to load file from our storage first
	objectKey := job.StoragePath
	var content []byte
	var mime string
	var err error

	// If StoragePath is empty, try to fetch from Store service directly
	if objectKey == "" {
		w.log.Info("storage path is empty, fetching from store service", slog.Int64("work_id", job.WorkID))
		content, mime, err = w.fetchFileFromStore(ctx, job.WorkID)
		if err != nil {
			w.log.Error("failed to load file from store service", slog.Int64("work_id", job.WorkID), slog.Any("err", err))
			report.Status = domain.StatusFailed
			_ = w.reportRepo.Update(ctx, report)
			return
		}
		// Reconstruct storage path for saving
		objectKey = fmt.Sprintf("%s/work_%d", job.AssignmentID, job.WorkID)
		// Save to our storage for future use
		if err := w.storage.Save(ctx, objectKey, content, mime); err != nil {
			w.log.Warn("failed to save file to local storage", slog.Any("err", err))
		}
	} else {
		content, mime, err = w.storage.Load(ctx, objectKey)
		if err != nil {
			// If not found, fetch from Store service
			w.log.Info("file not in local storage, fetching from store service", 
				slog.String("storage_path", objectKey),
				slog.Int64("work_id", job.WorkID))
			content, mime, err = w.fetchFileFromStore(ctx, job.WorkID)
			if err != nil {
				w.log.Error("failed to load file", 
					slog.Int64("work_id", job.WorkID), 
					slog.String("storage_path", objectKey), 
					slog.Any("err", err))
				report.Status = domain.StatusFailed
				_ = w.reportRepo.Update(ctx, report)
				return
			}
			// Save to our storage for future use
			if err := w.storage.Save(ctx, objectKey, content, mime); err != nil {
				w.log.Warn("failed to save file to local storage", slog.String("storage_path", objectKey), slog.Any("err", err))
			} else {
				w.log.Info("file saved to local storage", slog.String("storage_path", objectKey), slog.Int64("work_id", job.WorkID))
			}
		} else {
			w.log.Info("file loaded from local storage", slog.String("storage_path", objectKey), slog.Int64("work_id", job.WorkID))
		}
	}

	// Extract text for similarity check
	text, err := analyzer.ExtractText(content, mime)
	if err != nil {
		w.log.Warn("text extraction failed, using hash-only check", slog.Any("err", err))
		text = ""
	}

	// Find existing works in the same assignment for comparison
	// We compare all works in the assignment, not just those with the same hash
	// This allows detecting plagiarism even if files differ slightly
	existingWorks, err := w.workRepo.FindByHash(ctx, job.AssignmentID, job.FileHash)
	if err != nil {
		w.log.Error("failed to find existing works", slog.Any("err", err))
		existingWorks = []repo.WorkInfo{} // Continue with empty list
	}

	// Extract hashes and texts from existing works
	existingHashes := make([]string, 0, len(existingWorks))
	existingTexts := make([]string, 0)
	for _, work := range existingWorks {
		if work.StudentID != job.StudentID { // exclude same student
			existingHashes = append(existingHashes, work.Hash)
			
			// Load text from existing work for similarity check
			// Try to load from local storage first (works are saved with storage_path)
			existingStoragePath := fmt.Sprintf("%s/work_%d", job.AssignmentID, work.ID)
			existingContent, existingMime, err := w.storage.Load(ctx, existingStoragePath)
			if err != nil {
				// If not in local storage, try to fetch from store service
				w.log.Info("existing work not in local storage, fetching from store", 
					slog.Int64("existing_work_id", work.ID),
					slog.Int64("current_work_id", job.WorkID))
				existingContent, existingMime, err = w.fetchFileFromStore(ctx, work.ID)
				if err != nil {
					w.log.Warn("failed to load existing work for comparison", 
						slog.Int64("existing_work_id", work.ID),
						slog.Any("err", err))
					continue // Skip this work, continue with others
				}
				// Save to local storage for future use
				if err := w.storage.Save(ctx, existingStoragePath, existingContent, existingMime); err != nil {
					w.log.Warn("failed to save existing work to local storage", slog.Int64("existing_work_id", work.ID))
				}
			}
			
			// Extract text from existing work
			existingText, err := analyzer.ExtractText(existingContent, existingMime)
			if err != nil {
				w.log.Warn("failed to extract text from existing work", 
					slog.Int64("existing_work_id", work.ID),
					slog.Any("err", err))
				continue // Skip this work if text extraction fails
			}
			
			if existingText != "" {
				existingTexts = append(existingTexts, existingText)
			}
		}
	}

	// Run plagiarism check
	isPlagiarism, score, err := w.analyzer.CheckPlagiarism(ctx, job.FileHash, existingHashes, text, existingTexts)
	if err != nil {
		w.log.Error("plagiarism check failed", slog.Any("err", err))
		report.Status = domain.StatusFailed
		_ = w.reportRepo.Update(ctx, report)
		return
	}

	report.PlagiarismFlag = isPlagiarism
	report.Score = score
	report.Status = domain.StatusDone

	// Save report details as JSON
	details := map[string]interface{}{
		"work_id":         job.WorkID,
		"plagiarism_flag": isPlagiarism,
		"score":           score,
		"checked_at":      time.Now().UTC(),
		"hash":            job.FileHash,
	}

	detailsJSON, _ := json.Marshal(details)
	detailsPath := fmt.Sprintf("reports/%d.json", report.ID)
	if err := w.storage.Save(ctx, detailsPath, detailsJSON, "application/json"); err != nil {
		w.log.Error("failed to save report details", slog.Any("err", err))
	} else {
		report.DetailsPath = detailsPath
	}

	if err := w.reportRepo.Update(ctx, report); err != nil {
		w.log.Error("failed to update report", slog.Any("err", err))
		return
	}

	w.log.Info("analysis completed",
		slog.Int64("work_id", job.WorkID),
		slog.Bool("plagiarism", isPlagiarism),
		slog.Float64("score", score))
}

// fetchFileFromStore fetches file content from Store service.
func (w *Worker) fetchFileFromStore(ctx context.Context, workID int64) ([]byte, string, error) {
	url := fmt.Sprintf("%s/works/%d/file", w.storeURL, workID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		w.log.Error("failed to create request", slog.String("url", url), slog.Any("err", err))
		return nil, "", fmt.Errorf("create request: %w", err)
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		w.log.Error("HTTP request failed", slog.String("url", url), slog.Int64("work_id", workID), slog.Any("err", err))
		return nil, "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		w.log.Error("store service returned error", 
			slog.Int("status", resp.StatusCode), 
			slog.String("body", string(body)),
			slog.String("url", url),
			slog.Int64("work_id", workID))
		return nil, "", fmt.Errorf("fetch failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read body: %w", err)
	}

	mime := resp.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}

	w.log.Info("successfully fetched file from store service", 
		slog.Int64("work_id", workID), 
		slog.Int("size", len(content)),
		slog.String("mime", mime))
	
	return content, mime, nil
}


