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

// Error returns the error message.
func (e *NetworkError) Error() string {
	return fmt.Sprintf("playcamp: %s", e.Message)
}

// Unwrap returns the underlying error for use with errors.Is and errors.As.
func (e *NetworkError) Unwrap() error { return e.Err }

// APIError represents an HTTP error response from the API.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Details    []ValidationDetail
}

// ValidationDetail contains field-level validation error information.
type ValidationDetail struct {
	Message string `json:"message"`
	Path    string `json:"path"`
	Target  string `json:"target"`
}

// Error returns a formatted error message including the status code and message.
// If field-level validation details are present, they are appended.
func (e *APIError) Error() string {
	var msg string
	if e.Code != "" {
		msg = fmt.Sprintf("playcamp: API error %d (%s): %s", e.StatusCode, e.Code, e.Message)
	} else {
		msg = fmt.Sprintf("playcamp: API error %d: %s", e.StatusCode, e.Message)
	}
	for _, d := range e.Details {
		msg += fmt.Sprintf("; %s: %s", d.Path, d.Message)
	}
	return msg
}

// AuthError represents a 401 Unauthorized error.
type AuthError struct{ APIError }

// ForbiddenError represents a 403 Forbidden error.
type ForbiddenError struct{ APIError }

// NotFoundError represents a 404 Not Found error.
type NotFoundError struct{ APIError }

// ConflictError represents a 409 Conflict error.
type ConflictError struct{ APIError }

// BadRequestError represents a 400 Bad Request error.
type BadRequestError struct{ APIError }

// ValidationError represents a 422 Unprocessable Entity error.
type ValidationError struct{ APIError }

// RateLimitError represents a 429 Too Many Requests error.
type RateLimitError struct{ APIError }

// InputValidationError represents a local parameter validation failure.
type InputValidationError struct {
	Field   string
	Message string
}

// Error returns a formatted validation error message.
func (e *InputValidationError) Error() string {
	return fmt.Sprintf("playcamp: validation error on field %q: %s", e.Field, e.Message)
}

// newAPIError creates a typed API error from a status code and response body.
func newAPIError(statusCode int, body []byte) error {
	base := APIError{StatusCode: statusCode}

	var parsed struct {
		Message string             `json:"message"`
		Code    string             `json:"code"`
		Error   string             `json:"error"`
		Details []ValidationDetail `json:"details"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil {
		base.Message = parsed.Message
		base.Code = parsed.Code
		base.Details = parsed.Details
		if base.Message == "" {
			base.Message = parsed.Error
		}
	}
	if base.Message == "" {
		base.Message = fmt.Sprintf("HTTP %d", statusCode)
	}

	switch statusCode {
	case 400:
		return &BadRequestError{APIError: base}
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

// requireNonEmpty returns an InputValidationError if value is empty.
func requireNonEmpty(field, value string) error {
	if value == "" {
		return &InputValidationError{Field: field, Message: "must be a non-empty string"}
	}
	return nil
}

// newNetworkError creates a NetworkError with a generic message.
// The underlying error is available via Unwrap for debugging.
func newNetworkError(err error) error {
	return &NetworkError{Message: "network request failed", Err: err}
}
