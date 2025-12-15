package storage

import "context"

// Uploader defines storing binary content.
type Uploader interface {
	Save(ctx context.Context, objectKey string, content []byte, contentType string) error
}

// Downloader defines retrieving stored binaries.
type Downloader interface {
	Load(ctx context.Context, objectKey string) ([]byte, string, error) // returns data, contentType
}

// Storage combines upload/download operations.
type Storage interface {
	Uploader
	Downloader
}



