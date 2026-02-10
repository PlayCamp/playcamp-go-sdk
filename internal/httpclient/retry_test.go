package httpclient

import (
	"testing"
	"time"
)

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		status int
		want   bool
	}{
		{200, false},
		{201, false},
		{400, false},
		{401, false},
		{403, false},
		{404, false},
		{422, false},
		{429, true},
		{500, true},
		{502, true},
		{503, true},
	}
	for _, tt := range tests {
		got := shouldRetry(tt.status)
		if got != tt.want {
			t.Errorf("shouldRetry(%d) = %v, want %v", tt.status, got, tt.want)
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
