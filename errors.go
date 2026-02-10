package playcamp

import (
	"encoding/json"
	"fmt"
)

// NetworkError represents a connection or timeout error.
type NetworkError struct {
	Message string
	Err     error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("playcamp: %s", e.Message)
}

func (e *NetworkError) Unwrap() error { return e.Err }

// APIError represents an HTTP error response from the API.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("playcamp: API error %d (%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("playcamp: API error %d: %s", e.StatusCode, e.Message)
}

// AuthError represents a 401 Unauthorized error.
type AuthError struct{ APIError }

// ForbiddenError represents a 403 Forbidden error.
type ForbiddenError struct{ APIError }

// NotFoundError represents a 404 Not Found error.
type NotFoundError struct{ APIError }

// ConflictError represents a 409 Conflict error.
type ConflictError struct{ APIError }

// ValidationError represents a 422 Unprocessable Entity error.
type ValidationError struct{ APIError }

// RateLimitError represents a 429 Too Many Requests error.
type RateLimitError struct{ APIError }

// InputValidationError represents a local parameter validation failure.
type InputValidationError struct {
	Field   string
	Message string
}

func (e *InputValidationError) Error() string {
	return fmt.Sprintf("playcamp: validation error on field %q: %s", e.Field, e.Message)
}

// newAPIError creates a typed API error from a status code and response body.
func newAPIError(statusCode int, body []byte) error {
	base := APIError{StatusCode: statusCode}

	var parsed struct {
		Message string `json:"message"`
		Code    string `json:"code"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil {
		base.Message = parsed.Message
		base.Code = parsed.Code
		if base.Message == "" {
			base.Message = parsed.Error
		}
	}
	if base.Message == "" {
		base.Message = fmt.Sprintf("HTTP %d", statusCode)
	}

	switch statusCode {
	case 401:
		return &AuthError{APIError: base}
	case 403:
		return &ForbiddenError{APIError: base}
	case 404:
		return &NotFoundError{APIError: base}
	case 409:
		return &ConflictError{APIError: base}
	case 422:
		return &ValidationError{APIError: base}
	case 429:
		return &RateLimitError{APIError: base}
	default:
		return &base
	}
}

// newNetworkError creates a NetworkError.
func newNetworkError(err error) error {
	msg := "network request failed"
	if err != nil {
		msg = err.Error()
	}
	return &NetworkError{Message: msg, Err: err}
}
