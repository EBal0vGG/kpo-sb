package domain

import "fmt"

// ErrorCode represents a domain error code.
type ErrorCode string

const (
	// ErrCodeNotFound indicates the requested resource was not found.
	ErrCodeNotFound ErrorCode = "NOT_FOUND"
	// ErrCodeTooLarge indicates the file exceeds the maximum allowed size.
	ErrCodeTooLarge ErrorCode = "FILE_TOO_LARGE"
	// ErrCodeUnsupportedType indicates the file type is not supported.
	ErrCodeUnsupportedType ErrorCode = "UNSUPPORTED_FILE_TYPE"
	// ErrCodeInvalidInput indicates invalid input parameters.
	ErrCodeInvalidInput ErrorCode = "INVALID_INPUT"
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

