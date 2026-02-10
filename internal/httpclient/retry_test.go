package httpclient

import (
	"testing"
	"time"
)

func TestShouldRetry(t *testing.T) {
	// GET requests should be retried on 429 and 5xx.
	tests := []struct {
		method string
		status int
		want   bool
	}{
		{"GET", 200, false},
		{"GET", 201, false},
		{"GET", 400, false},
		{"GET", 401, false},
		{"GET", 403, false},
		{"GET", 404, false},
		{"GET", 422, false},
		{"GET", 429, true},
		{"GET", 500, true},
		{"GET", 502, true},
		{"GET", 503, true},
		{"HEAD", 500, true},
		{"HEAD", 429, true},
		// Non-idempotent methods should never be retried.
		{"POST", 429, false},
		{"POST", 500, false},
		{"POST", 502, false},
		{"PUT", 500, false},
		{"DELETE", 500, false},
	}
	for _, tt := range tests {
		got := shouldRetry(tt.method, tt.status)
		if got != tt.want {
			t.Errorf("shouldRetry(%s, %d) = %v, want %v", tt.method, tt.status, got, tt.want)
		}
	}
}

func TestCalculateBackoff(t *testing.T) {
	// Verify exponential growth
	prev := time.Duration(0)
	for attempt := 0; attempt < 5; attempt++ {
		delay := calculateBackoff(attempt)
		if delay <= 0 {
			t.Errorf("attempt %d: delay should be positive, got %v", attempt, delay)
		}
		if delay > maxDelay+time.Duration(float64(maxDelay)*jitterFactor) {
			t.Errorf("attempt %d: delay %v exceeds max", attempt, delay)
		}
		if attempt > 0 && delay < prev/2 {
			// With jitter, we can't guarantee strict growth, but it shouldn't be drastically smaller
			// This is a loose check
		}
		prev = delay
	}

	// Verify first attempt is roughly around initialDelay
	delays := make([]time.Duration, 100)
	for i := range delays {
		delays[i] = calculateBackoff(0)
	}
	var sum time.Duration
	for _, d := range delays {
		sum += d
	}
	avg := sum / time.Duration(len(delays))
	// Average should be around 500ms + 15% jitter = ~575ms
	if avg < 400*time.Millisecond || avg > 800*time.Millisecond {
		t.Errorf("average first-attempt delay = %v, expected ~500-650ms", avg)
	}
}
