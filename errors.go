package muxi

import "fmt"

// MuxiError is the base error type for all MUXI SDK errors
type MuxiError struct {
	Code       string
	Message    string
	StatusCode int
	Details    map[string]interface{}
}

func (e *MuxiError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return e.Message
}

// AuthenticationError represents a 401 Unauthorized error
type AuthenticationError struct {
	*MuxiError
}

// AuthorizationError represents a 403 Forbidden error
type AuthorizationError struct {
	*MuxiError
}

// NotFoundError represents a 404 Not Found error
type NotFoundError struct {
	*MuxiError
}

// ConflictError represents a 409 Conflict error
type ConflictError struct {
	*MuxiError
}

// ValidationError represents a 422 Validation error
type ValidationError struct {
	*MuxiError
}

// RateLimitError represents a 429 Too Many Requests error
type RateLimitError struct {
	*MuxiError
	RetryAfter int
}

// ServerError represents a 5xx server error
type ServerError struct {
	*MuxiError
}

// ConnectionError represents a network/connection error
type ConnectionError struct {
	*MuxiError
}

// newMuxiError creates a new MuxiError
func newMuxiError(code, message string, statusCode int) *MuxiError {
	return &MuxiError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Error codes
const (
	ErrUnauthorized     = "UNAUTHORIZED"
	ErrForbidden        = "FORBIDDEN"
	ErrNotFound         = "NOT_FOUND"
	ErrConflict         = "CONFLICT"
	ErrValidationError  = "VALIDATION_ERROR"
	ErrRateLimited      = "RATE_LIMITED"
	ErrServerError      = "SERVER_ERROR"
	ErrConnectionError  = "CONNECTION_ERROR"
)
