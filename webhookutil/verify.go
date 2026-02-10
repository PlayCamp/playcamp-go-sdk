// Package webhookutil provides utilities for verifying and constructing
// PlayCamp webhook signatures.
package webhookutil

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/playcamp/playcamp-go-sdk"
)

// VerifyOptions specifies options for webhook signature verification.
type VerifyOptions struct {
	// Payload is the raw request body.
	Payload []byte
	// Signature is the value from the X-Webhook-Signature header.
	Signature string
	// Secret is the webhook secret received when creating the webhook.
	Secret string
	// Tolerance is the maximum allowed age in seconds (default: 300).
	// Only used for timestamped signatures (t=...,v1=... format).
	Tolerance int
}

// VerifyResult is the result of webhook signature verification.
type VerifyResult struct {
	// Valid is true if the signature is valid.
	Valid bool
	// Payload is the parsed webhook payload (only set if Valid is true).
	Payload *playcamp.WebhookPayload
	// Error is a description of the verification failure.
	Error string
}

// Verify verifies a webhook signature.
//
// It supports two signature formats:
//   - Simple hex signature (PlayCamp default): HMAC-SHA256 hex digest
//   - Timestamped signature: "t=timestamp,v1=signature" format
func Verify(opts VerifyOptions) VerifyResult {
	tolerance := opts.Tolerance
	if tolerance <= 0 {
		tolerance = 300
	}

	sig := opts.Signature
	if strings.Contains(sig, "t=") && strings.Contains(sig, "v1=") {
		return verifyTimestamped(opts.Payload, sig, opts.Secret, tolerance)
	}
	return verifySimple(opts.Payload, sig, opts.Secret)
}

func verifySimple(payload []byte, signature, secret string) VerifyResult {
	expected := computeHMAC(payload, secret)

	providedBytes, err := hex.DecodeString(signature)
	if err != nil {
		return VerifyResult{Error: "Invalid signature format"}
	}
	expectedBytes, err := hex.DecodeString(expected)
	if err != nil {
		return VerifyResult{Error: "Invalid signature format"}
	}

	if len(providedBytes) != len(expectedBytes) {
		return VerifyResult{Error: "Invalid signature"}
	}
	if subtle.ConstantTimeCompare(providedBytes, expectedBytes) != 1 {
		return VerifyResult{Error: "Invalid signature"}
	}

	var parsed playcamp.WebhookPayload
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return VerifyResult{Error: "Invalid JSON payload"}
	}
	return VerifyResult{Valid: true, Payload: &parsed}
}

func verifyTimestamped(payload []byte, signature, secret string, tolerance int) VerifyResult {
	parts := strings.Split(signature, ",")
	var timestampStr, sigHex string
	for _, p := range parts {
		if strings.HasPrefix(p, "t=") {
			timestampStr = p[2:]
		}
		if strings.HasPrefix(p, "v1=") {
			sigHex = p[3:]
		}
	}
	if timestampStr == "" || sigHex == "" {
		return VerifyResult{Error: "Invalid signature format"}
	}

	ts, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return VerifyResult{Error: "Invalid timestamp in signature"}
	}

	now := time.Now().Unix()
	if abs(now-ts) > int64(tolerance) {
		return VerifyResult{Error: "Webhook timestamp outside tolerance window"}
	}

	signedPayload := fmt.Sprintf("%d.%s", ts, string(payload))
	expected := computeHMAC([]byte(signedPayload), secret)

	providedBytes, err := hex.DecodeString(sigHex)
	if err != nil {
		return VerifyResult{Error: "Invalid signature format"}
	}
	expectedBytes, err := hex.DecodeString(expected)
	if err != nil {
		return VerifyResult{Error: "Invalid signature format"}
	}

	if len(providedBytes) != len(expectedBytes) {
		return VerifyResult{Error: "Invalid signature"}
	}
	if subtle.ConstantTimeCompare(providedBytes, expectedBytes) != 1 {
		return VerifyResult{Error: "Invalid signature"}
	}

	var parsed playcamp.WebhookPayload
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return VerifyResult{Error: "Invalid JSON payload"}
	}
	return VerifyResult{Valid: true, Payload: &parsed}
}

func computeHMAC(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
