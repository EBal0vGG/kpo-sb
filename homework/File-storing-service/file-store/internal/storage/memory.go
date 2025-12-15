package storage

import (
	"context"
	"fmt"
	"sync"
)

// MemoryStorage is a simple in-memory storage for local dev and tests.
type MemoryStorage struct {
	mu    sync.RWMutex
	files map[string]blob
}

type blob struct {
	data        []byte
	contentType string
}

func NewMemory() *MemoryStorage {
	return &MemoryStorage{files: make(map[string]blob)}
}

func (m *MemoryStorage) Save(ctx context.Context, objectKey string, content []byte, contentType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[objectKey] = blob{data: append([]byte{}, content...), contentType: contentType}
	return nil
}

func (m *MemoryStorage) Load(ctx context.Context, objectKey string) ([]byte, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	bl, ok := m.files[objectKey]
	if !ok {
		return nil, "", fmt.Errorf("not found")
	}
	return append([]byte{}, bl.data...), bl.contentType, nil
}



