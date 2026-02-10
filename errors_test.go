package playcamp

import (
	"errors"
	"testing"
)

func TestNewAPIError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantType   string
	}{
		{"400 bad request", 400, `{"message":"Bad Request"}`, "*playcamp.BadRequestError"},
		{"401 auth error", 401, `{"message":"Unauthorized"}`, "*playcamp.AuthError"},
		{"403 forbidden", 403, `{"message":"Forbidden"}`, "*playcamp.ForbiddenError"},
		{"404 not found", 404, `{"message":"Not found"}`, "*playcamp.NotFoundError"},
		{"409 conflict", 409, `{"message":"Conflict"}`, "*playcamp.ConflictError"},
		{"422 validation", 422, `{"message":"Invalid input"}`, "*playcamp.ValidationError"},
		{"429 rate limit", 429, `{"message":"Too many requests"}`, "*playcamp.RateLimitError"},
		{"500 generic", 500, `{"message":"Internal server error"}`, "*playcamp.APIError"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := newAPIError(tt.statusCode, []byte(tt.body))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestErrorsAs(t *testing.T) {
	t.Run("AuthError", func(t *testing.T) {
		err := newAPIError(401, []byte(`{"message":"Unauthorized"}`))
		var authErr *AuthError
		if !errors.As(err, &authErr) {
			t.Error("errors.As should match *AuthError")
		}
		if authErr.StatusCode != 401 {
			t.Errorf("StatusCode = %d, want 401", authErr.StatusCode)
		}
		if authErr.Message != "Unauthorized" {
			t.Errorf("Message = %q, want %q", authErr.Message, "Unauthorized")
		}
	})

	t.Run("NotFoundError", func(t *testing.T) {
		err := newAPIError(404, []byte(`{"message":"Campaign not found","code":"NOT_FOUND"}`))
		var notFound *NotFoundError
		if !errors.As(err, &notFound) {
			t.Error("errors.As should match *NotFoundError")
		}
		if notFound.Code != "NOT_FOUND" {
			t.Errorf("Code = %q, want %q", notFound.Code, "NOT_FOUND")
		}
	})

	t.Run("RateLimitError", func(t *testing.T) {
		err := newAPIError(429, []byte(`{"message":"Rate limit exceeded"}`))
		var rateLimited *RateLimitError
		if !errors.As(err, &rateLimited) {
			t.Error("errors.As should match *RateLimitError")
		}
	})

	t.Run("generic APIError", func(t *testing.T) {
		err := newAPIError(500, []byte(`{"message":"Internal error"}`))
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Error("errors.As should match *APIError")
		}
	})

	t.Run("does not match wrong type", func(t *testing.T) {
		err := newAPIError(404, []byte(`{"message":"Not found"}`))
		var authErr *AuthError
		if errors.As(err, &authErr) {
			t.Error("404 should not match *AuthError")
		}
	})

	t.Run("malformed JSON body", func(t *testing.T) {
		err := newAPIError(500, []byte(`not json`))
		if err == nil {
			t.Fatal("expected error")
		}
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Error("should still produce APIError")
		}
	})

	t.Run("error field fallback", func(t *testing.T) {
		err := newAPIError(400, []byte(`{"error":"Bad request"}`))
		var badReq *BadRequestError
		if !errors.As(err, &badReq) {
			t.Error("errors.As should match *BadRequestError")
		}
		if badReq.Message != "Bad request" {
			t.Errorf("Message = %q, want %q", badReq.Message, "Bad request")
		}
	})

	t.Run("BadRequestError with details", func(t *testing.T) {
		body := `{"code":"VALIDATION_ERROR","error":"Validation Error","details":[{"message":"Invalid datetime","path":"purchasedAt","target":"body"}]}`
		err := newAPIError(400, []byte(body))
		var badReq *BadRequestError
		if !errors.As(err, &badReq) {
			t.Fatal("errors.As should match *BadRequestError")
		}
		if badReq.Code != "VALIDATION_ERROR" {
			t.Errorf("Code = %q, want VALIDATION_ERROR", badReq.Code)
		}
		if len(badReq.Details) != 1 {
			t.Fatalf("len(Details) = %d, want 1", len(badReq.Details))
		}
		if badReq.Details[0].Path != "purchasedAt" {
			t.Errorf("Details[0].Path = %q, want purchasedAt", badReq.Details[0].Path)
		}
		if badReq.Details[0].Message != "Invalid datetime" {
			t.Errorf("Details[0].Message = %q, want Invalid datetime", badReq.Details[0].Message)
		}
		// Error string should include details
		want := `playcamp: API error 400 (VALIDATION_ERROR): Validation Error; purchasedAt: Invalid datetime`
		if badReq.Error() != want {
			t.Errorf("Error() = %q, want %q", badReq.Error(), want)
		}
	})
}

func TestNetworkError(t *testing.T) {
	cause := errors.New("connection refused")
	err := newNetworkError(cause)

	var netErr *NetworkError
	if !errors.As(err, &netErr) {
		t.Error("errors.As should match *NetworkError")
	}
	if netErr.Err != cause {
		t.Error("Unwrap should return the original cause")
	}
}

func TestInputValidationError(t *testing.T) {
	err := &InputValidationError{Field: "userId", Message: "must be a non-empty string"}
	expected := `playcamp: validation error on field "userId": must be a non-empty string`
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}
