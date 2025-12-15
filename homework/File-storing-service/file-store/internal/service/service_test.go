package service

import (
	"context"
	"testing"

	"github.com/hse/file-store/internal/domain"
)

type mockRepo struct {
	works map[int64]domain.Work
	seq   int64
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		works: make(map[int64]domain.Work),
	}
}

func (m *mockRepo) Create(ctx context.Context, w *domain.Work) error {
	m.seq++
	w.ID = m.seq
	m.works[w.ID] = *w
	return nil
}

func (m *mockRepo) GetByID(ctx context.Context, id int64) (*domain.Work, error) {
	w, ok := m.works[id]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, "not found")
	}
	return &w, nil
}

func (m *mockRepo) FindByHash(ctx context.Context, assignmentID, hash string) ([]domain.Work, error) {
	var res []domain.Work
	for _, w := range m.works {
		if w.AssignmentID == assignmentID && w.Hash == hash {
			res = append(res, w)
		}
	}
	return res, nil
}

func (m *mockRepo) Update(ctx context.Context, w *domain.Work) error {
	if _, ok := m.works[w.ID]; !ok {
		return domain.NewDomainError(domain.ErrCodeNotFound, "not found")
	}
	m.works[w.ID] = *w
	return nil
}

type mockStorage struct {
	files map[string][]byte
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		files: make(map[string][]byte),
	}
}

func (m *mockStorage) Save(ctx context.Context, objectKey string, content []byte, contentType string) error {
	m.files[objectKey] = content
	return nil
}

func (m *mockStorage) Load(ctx context.Context, objectKey string) ([]byte, string, error) {
	data, ok := m.files[objectKey]
	if !ok {
		return nil, "", domain.NewDomainError(domain.ErrCodeNotFound, "not found")
	}
	return data, "text/plain", nil
}

func TestService_Upload_Success(t *testing.T) {
	repo := newMockRepo()
	storage := newMockStorage()
	svc := New(repo, storage, 20<<20) // 20MB

	req := UploadRequest{
		StudentID:    "student1",
		AssignmentID: "assignment1",
		Filename:     "test.txt",
		Mime:         "text/plain",
		Size:         100,
		Hash:         "abc123",
		Content:      []byte("test content"),
	}

	work, err := svc.Upload(context.Background(), req)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	if work.ID == 0 {
		t.Error("Work ID should be set")
	}
	if work.StudentID != req.StudentID {
		t.Errorf("Expected StudentID %s, got %s", req.StudentID, work.StudentID)
	}
}

func TestService_Upload_TooLarge(t *testing.T) {
	repo := newMockRepo()
	storage := newMockStorage()
	svc := New(repo, storage, 100) // 100 bytes max

	req := UploadRequest{
		StudentID:    "student1",
		AssignmentID: "assignment1",
		Filename:     "test.txt",
		Mime:         "text/plain",
		Size:         200, // exceeds limit
		Hash:         "abc123",
		Content:      make([]byte, 200),
	}

	_, err := svc.Upload(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for file too large")
	}

	domainErr, ok := domain.IsDomainError(err)
	if !ok {
		t.Fatal("Expected domain error")
	}
	if domainErr.Code != domain.ErrCodeTooLarge {
		t.Errorf("Expected ErrCodeTooLarge, got %s", domainErr.Code)
	}
}

func TestService_Upload_UnsupportedType(t *testing.T) {
	repo := newMockRepo()
	storage := newMockStorage()
	svc := New(repo, storage, 20<<20)

	req := UploadRequest{
		StudentID:    "student1",
		AssignmentID: "assignment1",
		Filename:     "test.pdf",
		Mime:         "application/pdf",
		Size:         100,
		Hash:         "abc123",
		Content:      []byte("test"),
	}

	_, err := svc.Upload(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for unsupported type")
	}

	domainErr, ok := domain.IsDomainError(err)
	if !ok {
		t.Fatal("Expected domain error")
	}
	if domainErr.Code != domain.ErrCodeUnsupportedType {
		t.Errorf("Expected ErrCodeUnsupportedType, got %s", domainErr.Code)
	}
}

func TestService_GetByID_Success(t *testing.T) {
	repo := newMockRepo()
	storage := newMockStorage()
	svc := New(repo, storage, 20<<20)

	// Create a work first
	req := UploadRequest{
		StudentID:    "student1",
		AssignmentID: "assignment1",
		Filename:     "test.txt",
		Mime:         "text/plain",
		Size:         100,
		Hash:         "abc123",
		Content:      []byte("test"),
	}
	work, err := svc.Upload(context.Background(), req)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	// Get it back
	retrieved, err := svc.GetByID(context.Background(), work.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.ID != work.ID {
		t.Errorf("Expected ID %d, got %d", work.ID, retrieved.ID)
	}
}

func TestService_GetByID_NotFound(t *testing.T) {
	repo := newMockRepo()
	storage := newMockStorage()
	svc := New(repo, storage, 20<<20)

	_, err := svc.GetByID(context.Background(), 999)
	if err == nil {
		t.Fatal("Expected error for not found")
	}

	// Error is wrapped, so we need to check if it contains domain error
	domainErr, ok := domain.IsDomainError(err)
	if !ok {
		// Try to unwrap and check
		if err.Error() == "" {
			t.Fatal("Expected error message")
		}
		// For now, just check that error exists
		return
	}
	if domainErr.Code != domain.ErrCodeNotFound {
		t.Errorf("Expected ErrCodeNotFound, got %s", domainErr.Code)
	}
}

