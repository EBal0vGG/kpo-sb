module github.com/hse/file-store

go 1.22.2

require (
	github.com/go-chi/chi/v5 v5.2.3
	github.com/hse/pkg-config v0.0.0
	github.com/hse/pkg-httpx v0.0.0
	github.com/hse/pkg-logger v0.0.0
)

replace github.com/hse/pkg-config => ../pkg/config

replace github.com/hse/pkg-httpx => ../pkg/httpx

replace github.com/hse/pkg-logger => ../pkg/logger
