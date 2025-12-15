package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// StoreClient communicates with file-store service.
type StoreClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewStoreClient(baseURL string) *StoreClient {
	return &StoreClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// UploadWorkResponse represents the response from upload endpoint.
type UploadWorkResponse struct {
	ID           int64     `json:"id"`
	StudentID    string    `json:"student_id"`
	AssignmentID string    `json:"assignment_id"`
	Filename     string    `json:"filename"`
	Hash         string    `json:"hash"`
	Size         int64     `json:"size"`
	CreatedAt    time.Time `json:"created_at"`
	StoragePath  string    `json:"storage_path"`
}

// UploadWork uploads a file to the store service.
func (c *StoreClient) UploadWork(ctx context.Context, fileContent []byte, filename, studentID, assignmentID string) (*UploadWorkResponse, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}

	if _, err := io.Copy(fw, bytes.NewReader(fileContent)); err != nil {
		return nil, fmt.Errorf("write file content: %w", err)
	}

	if err := w.WriteField("student_id", studentID); err != nil {
		return nil, fmt.Errorf("write student_id: %w", err)
	}
	if err := w.WriteField("assignment_id", assignmentID); err != nil {
		return nil, fmt.Errorf("write assignment_id: %w", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/works", &buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusRequestEntityTooLarge {
			return nil, fmt.Errorf("request entity too large: %s", string(body))
		}
		if resp.StatusCode == http.StatusUnsupportedMediaType {
			return nil, fmt.Errorf("unsupported media type: %s", string(body))
		}
		return nil, fmt.Errorf("upload failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result UploadWorkResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// GetWork retrieves work metadata.
func (c *StoreClient) GetWork(ctx context.Context, workID int64) (*UploadWorkResponse, error) {
	url := fmt.Sprintf("%s/works/%d", c.baseURL, workID)
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
			return nil, fmt.Errorf("work not found: %s", string(body))
		}
		return nil, fmt.Errorf("get failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result UploadWorkResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}


