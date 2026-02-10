package httpclient

import (
	"math"
	"math/rand"
	"time"
)

const (
	initialDelay = 500 * time.Millisecond
	maxDelay     = 10 * time.Second
	jitterFactor = 0.3
)

// shouldRetry returns true if the request should be retried for the given status code.
func shouldRetry(statusCode int) bool {
	return statusCode == 429 || statusCode >= 500
}

// calculateBackoff returns the delay before the next retry attempt.
// delay = initialDelay * 2^attempt + random jitter (0~30%)
func calculateBackoff(attempt int) time.Duration {
	delay := float64(initialDelay) * math.Pow(2, float64(attempt))
	if delay > float64(maxDelay) {
		delay = float64(maxDelay)
	}

	// Add jitter: 0 to jitterFactor of the delay.
	jitter := delay * jitterFactor * rand.Float64()
	delay += jitter

	return time.Duration(delay)
}
