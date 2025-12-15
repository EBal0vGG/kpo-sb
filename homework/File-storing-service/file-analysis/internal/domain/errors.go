package domain

import "fmt"

// ErrorCode represents a domain error code.
type ErrorCode string

const (
	// ErrCodeNotFound indicates the requested resource was not found.
	ErrCodeNotFound ErrorCode = "NOT_FOUND"
	// ErrCodeQueueFull indicates the job queue is full.
	ErrCodeQueueFull ErrorCode = "QUEUE_FULL"
	// ErrCodeInvalidInput indicates invalid input parameters.
	ErrCodeInvalidInput ErrorCode = "INVALID_INPUT"
	// ErrCodeAnalysisFailed indicates analysis processing failed.
	ErrCodeAnalysisFailed ErrorCode = "ANALYSIS_FAILED"
)

// DomainError represents a domain-level error.
type DomainError struct {
	Code    ErrorCode
	Message string
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewDomainError creates a new domain error.
func NewDomainError(code ErrorCode, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

// IsDomainError checks if an error is a domain error.
func IsDomainError(err error) (*DomainError, bool) {
	domainErr, ok := err.(*DomainError)
	return domainErr, ok
}

