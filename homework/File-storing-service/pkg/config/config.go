package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type HTTP struct {
	Addr          string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	MaxBodyBytes  int64
	ShutdownGrace time.Duration
}

type Postgres struct {
	URL string
}

type MinIO struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type Config struct {
	ServiceName string
	HTTP        HTTP
	Postgres    Postgres
	MinIO       MinIO
}

// Load builds config from environment variables with sane defaults for local dev.
func Load(service string) (Config, error) {
	get := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return def
	}

	parseDuration := func(key, def string) (time.Duration, error) {
		val := get(key, def)
		d, err := time.ParseDuration(val)
		if err != nil {
			return 0, fmt.Errorf("parse %s: %w", key, err)
		}
		return d, nil
	}

	parseBool := func(key, def string) (bool, error) {
		val := get(key, def)
		b, err := strconv.ParseBool(val)
		if err != nil {
			return false, fmt.Errorf("parse %s: %w", key, err)
		}
		return b, nil
	}

	readTimeout, err := parseDuration("HTTP_READ_TIMEOUT", "5s")
	if err != nil {
		return Config{}, err
	}
	writeTimeout, err := parseDuration("HTTP_WRITE_TIMEOUT", "10s")
	if err != nil {
		return Config{}, err
	}
	shutdownGrace, err := parseDuration("HTTP_SHUTDOWN_GRACE", "10s")
	if err != nil {
		return Config{}, err
	}

	useSSL, err := parseBool("MINIO_USE_SSL", "false")
	if err != nil {
		return Config{}, err
	}

	maxBody := get("HTTP_MAX_BODY_BYTES", "20971520") // 20MB default
	maxBodyInt, err := strconv.ParseInt(maxBody, 10, 64)
	if err != nil {
		return Config{}, fmt.Errorf("parse HTTP_MAX_BODY_BYTES: %w", err)
	}

	// Default port based on service name
	defaultPort := ":8080"
	switch service {
	case "file-store":
		defaultPort = ":8081"
	case "file-analysis":
		defaultPort = ":8082"
	case "gateway":
		defaultPort = ":8080"
	}

	return Config{
		ServiceName: service,
		HTTP: HTTP{
			Addr:          get("HTTP_ADDR", defaultPort),
			ReadTimeout:   readTimeout,
			WriteTimeout:  writeTimeout,
			MaxBodyBytes:  maxBodyInt,
			ShutdownGrace: shutdownGrace,
		},
		Postgres: Postgres{
			URL: get("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/"+service+"?sslmode=disable"),
		},
		MinIO: MinIO{
			Endpoint:  get("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: get("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: get("MINIO_SECRET_KEY", "minioadmin"),
			Bucket:    get("MINIO_BUCKET", service),
			UseSSL:    useSSL,
		},
	}, nil
}
