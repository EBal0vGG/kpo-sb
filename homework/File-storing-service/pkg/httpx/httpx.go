package httpx

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// JSON writes JSON response with code.
func JSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// Error writes a JSON error response.
func Error(w http.ResponseWriter, code int, msg string) {
	JSON(w, code, ErrorResponse{Error: msg})
}

// MaxBytes wraps a handler to enforce request size limits.
func MaxBytes(next http.Handler, n int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, n)
		next.ServeHTTP(w, r)
		// Note: MaxBytesReader will close the connection if limit is exceeded,
		// so we can't catch the error here. The error will be visible when reading the body.
	})
}


