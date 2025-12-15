package storage

import "context"

// Storage defines file storage operations.
type Storage interface {
	Save(ctx context.Context, objectKey string, content []byte, contentType string) error
	Load(ctx context.Context, objectKey string) ([]byte, string, error) // returns data, contentType
}



