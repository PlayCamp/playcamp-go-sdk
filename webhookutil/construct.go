package webhookutil

import (
	"fmt"
	"time"
)

// SignatureOptions specifies options for constructing a webhook signature.
type SignatureOptions struct {
	// Timestamped controls whether to use the "t=...,v1=..." format.
	Timestamped bool
	// Timestamp overrides the timestamp used in timestamped signatures.
	// If zero, the current time is used.
	Timestamp int64
}

// ConstructSignature creates a webhook signature for testing purposes.
//
// By default it creates a simple HMAC-SHA256 hex signature.
// Set opts.Timestamped to true for the "t=timestamp,v1=signature" format.
func ConstructSignature(payload []byte, secret string, opts *SignatureOptions) string {
	if opts != nil && opts.Timestamped {
		ts := opts.Timestamp
		if ts == 0 {
			ts = time.Now().Unix()
		}
		signedPayload := fmt.Sprintf("%d.%s", ts, string(payload))
		sig := computeHMAC([]byte(signedPayload), secret)
		return fmt.Sprintf("t=%d,v1=%s", ts, sig)
	}

	return computeHMAC(payload, secret)
}
