.PHONY: build tidy test up down logs clean

# Build all services
build:
	cd gateway && go build ./cmd/server
	cd file-store && go build ./cmd/server
	cd file-analysis && go build ./cmd/server

# Tidy dependencies
tidy:
	cd gateway && go mod tidy
	cd file-store && go mod tidy
	cd file-analysis && go mod tidy
	cd pkg/config && go mod tidy
	cd pkg/httpx && go mod tidy
	cd pkg/logger && go mod tidy

# Run tests
test:
	cd gateway && go test ./...
	cd file-store && go test ./...
	cd file-analysis && go test ./...

# Start services with docker-compose
up:
	cd deploy && docker compose up -d

# Stop services
down:
	cd deploy && docker compose down

# View logs
logs:
	cd deploy && docker compose logs -f

# Clean build artifacts
clean:
	rm -f gateway/server
	rm -f file-store/server
	rm -f file-analysis/server


