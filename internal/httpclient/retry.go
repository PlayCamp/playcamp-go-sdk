package httpclient

import (
	"crypto/rand"
	"encoding/binary"
	"math"
	"net/http"
	"time"
)

const (
	initialDelay = 500 * time.Millisecond
	maxDelay     = 10 * time.Second
	jitterFactor = 0.3
)

// isIdempotent returns true if the HTTP method is safe to retry.
func isIdempotent(method string) bool {
	return method == http.MethodGet || method == http.MethodHead
}

// shouldRetry returns true if the request should be retried for the given method and status code.
// Only idempotent methods (GET, HEAD) are retried to avoid duplicate side effects.
func shouldRetry(method string, statusCode int) bool {
	if !isIdempotent(method) {
		return false
	}
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
	jitter := delay * jitterFactor * cryptoRandFloat64()
	delay += jitter

	return time.Duration(delay)
}

// cryptoRandFloat64 returns a cryptographically random float64 in [0, 1).
func cryptoRandFloat64() float64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return float64(binary.LittleEndian.Uint64(b[:])>>1) / float64(1<<63)
}
