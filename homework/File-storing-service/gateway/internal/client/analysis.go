package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AnalysisClient communicates with file-analysis service.
type AnalysisClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAnalysisClient(baseURL string) *AnalysisClient {
	return &AnalysisClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SubmitJobRequest represents analysis job submission.
type SubmitJobRequest struct {
	WorkID       int64  `json:"work_id"`
	AssignmentID string `json:"assignment_id"`
	StudentID    string `json:"student_id"`
	FileHash     string `json:"file_hash"`
	StoragePath  string `json:"storage_path"`
}

// Report represents an analysis report.
type Report struct {
	ID             int64     `json:"id"`
	Status         string    `json:"status"`
	PlagiarismFlag bool      `json:"plagiarism_flag"`
	Score          float64   `json:"score"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// SubmitJob submits an analysis job.
func (c *AnalysisClient) SubmitJob(ctx context.Context, req SubmitJobRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/analysis/jobs", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		bodyBytes, _ := json.Marshal(map[string]string{"error": "failed to submit job"})
		return fmt.Errorf("submit failed: status %d, %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// GetReports retrieves reports for a work.
func (c *AnalysisClient) GetReports(ctx context.Context, workID int64) ([]Report, error) {
	url := fmt.Sprintf("%s/works/%d/reports", c.baseURL, workID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("reports not found: %s", string(body))
		}
		return nil, fmt.Errorf("get reports failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var reports []Report
	if err := json.NewDecoder(resp.Body).Decode(&reports); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return reports, nil
}


