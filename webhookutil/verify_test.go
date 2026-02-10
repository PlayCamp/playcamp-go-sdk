package webhookutil

import (
	"encoding/json"
	"testing"
	"time"
)

func TestVerify_SimpleSignature(t *testing.T) {
	payload := []byte(`{"events":[{"event":"coupon.redeemed","timestamp":"2024-01-01T00:00:00Z","data":{"couponCode":"TEST","userId":"user1","usageId":1,"reward":[]}}]}`)
	secret := "test_secret"

	sig := ConstructSignature(payload, secret, nil)

	result := Verify(VerifyOptions{
		Payload:   payload,
		Signature: sig,
		Secret:    secret,
	})

	if !result.Valid {
		t.Fatalf("expected valid, got error: %s", result.Error)
	}
	if result.Payload == nil {
		t.Fatal("payload should not be nil")
	}
	if len(result.Payload.Events) != 1 {
		t.Errorf("len(Events) = %d, want 1", len(result.Payload.Events))
	}
}

func TestVerify_TimestampedSignature(t *testing.T) {
	payload := []byte(`{"events":[]}`)
	secret := "test_secret"
	now := time.Now().Unix()

	sig := ConstructSignature(payload, secret, &SignatureOptions{
		Timestamped: true,
		Timestamp:   now,
	})

	result := Verify(VerifyOptions{
		Payload:   payload,
		Signature: sig,
		Secret:    secret,
	})

	if !result.Valid {
		t.Fatalf("expected valid, got error: %s", result.Error)
	}
}

func TestVerify_InvalidSignature(t *testing.T) {
	payload := []byte(`{"events":[]}`)

	result := Verify(VerifyOptions{
		Payload:   payload,
		Signature: "0000000000000000000000000000000000000000000000000000000000000000",
		Secret:    "test_secret",
	})

	if result.Valid {
		t.Fatal("expected invalid")
	}
	if result.Error != "Invalid signature" {
		t.Errorf("Error = %q, want %q", result.Error, "Invalid signature")
	}
}

func TestVerify_InvalidSignatureFormat(t *testing.T) {
	payload := []byte(`{"events":[]}`)

	result := Verify(VerifyOptions{
		Payload:   payload,
		Signature: "not-a-hex-string",
		Secret:    "test_secret",
	})

	if result.Valid {
		t.Fatal("expected invalid")
	}
}

func TestVerify_TimestampedSignature_Expired(t *testing.T) {
	payload := []byte(`{"events":[]}`)
	secret := "test_secret"
	oldTimestamp := time.Now().Unix() - 600 // 10 minutes ago

	sig := ConstructSignature(payload, secret, &SignatureOptions{
		Timestamped: true,
		Timestamp:   oldTimestamp,
	})

	result := Verify(VerifyOptions{
		Payload:   payload,
		Signature: sig,
		Secret:    secret,
		Tolerance: 300,
	})

	if result.Valid {
		t.Fatal("expected invalid for expired timestamp")
	}
	if result.Error != "Webhook timestamp outside tolerance window" {
		t.Errorf("Error = %q", result.Error)
	}
}

func TestVerify_TimestampedSignature_CustomTolerance(t *testing.T) {
	payload := []byte(`{"events":[]}`)
	secret := "test_secret"
	oldTimestamp := time.Now().Unix() - 600

	sig := ConstructSignature(payload, secret, &SignatureOptions{
		Timestamped: true,
		Timestamp:   oldTimestamp,
	})

	// With a large tolerance, it should pass
	result := Verify(VerifyOptions{
		Payload:   payload,
		Signature: sig,
		Secret:    secret,
		Tolerance: 3600, // 1 hour
	})

	if !result.Valid {
		t.Fatalf("expected valid with large tolerance, got error: %s", result.Error)
	}
}

func TestVerify_InvalidTimestampedFormat(t *testing.T) {
	payload := []byte(`{"events":[]}`)

	result := Verify(VerifyOptions{
		Payload:   payload,
		Signature: "t=notanumber,v1=abc123",
		Secret:    "test_secret",
	})

	if result.Valid {
		t.Fatal("expected invalid")
	}
	if result.Error != "Invalid timestamp in signature" {
		t.Errorf("Error = %q", result.Error)
	}
}

func TestVerify_MissingTimestampParts(t *testing.T) {
	payload := []byte(`{"events":[]}`)

	result := Verify(VerifyOptions{
		Payload:   payload,
		Signature: "t=12345",
		Secret:    "test_secret",
	})

	// This doesn't match timestamped format (no v1=), so it falls to simple
	if result.Valid {
		t.Fatal("expected invalid for partial signature")
	}
}

func TestVerify_InvalidJSON(t *testing.T) {
	payload := []byte(`not json at all`)
	secret := "test_secret"
	sig := ConstructSignature(payload, secret, nil)

	result := Verify(VerifyOptions{
		Payload:   payload,
		Signature: sig,
		Secret:    secret,
	})

	if result.Valid {
		t.Fatal("expected invalid for bad JSON")
	}
	if result.Error != "Invalid JSON payload" {
		t.Errorf("Error = %q, want %q", result.Error, "Invalid JSON payload")
	}
}

func TestVerify_WrongSecret(t *testing.T) {
	payload := []byte(`{"events":[]}`)
	sig := ConstructSignature(payload, "correct_secret", nil)

	result := Verify(VerifyOptions{
		Payload:   payload,
		Signature: sig,
		Secret:    "wrong_secret",
	})

	if result.Valid {
		t.Fatal("expected invalid with wrong secret")
	}
}

func TestConstructSignature_Simple(t *testing.T) {
	payload := []byte(`{"events":[]}`)
	secret := "test_secret"

	sig := ConstructSignature(payload, secret, nil)
	if sig == "" {
		t.Fatal("signature should not be empty")
	}
	// Should be a 64-char hex string (SHA-256)
	if len(sig) != 64 {
		t.Errorf("len(sig) = %d, want 64", len(sig))
	}
}

func TestConstructSignature_Timestamped(t *testing.T) {
	payload := []byte(`{"events":[]}`)
	secret := "test_secret"
	ts := int64(1700000000)

	sig := ConstructSignature(payload, secret, &SignatureOptions{
		Timestamped: true,
		Timestamp:   ts,
	})

	if sig == "" {
		t.Fatal("signature should not be empty")
	}
	// Should start with t= and contain v1=
	if len(sig) < 10 {
		t.Error("timestamped signature too short")
	}
}

func TestConstructSignature_Deterministic(t *testing.T) {
	payload := []byte(`{"events":[]}`)
	secret := "test_secret"

	sig1 := ConstructSignature(payload, secret, nil)
	sig2 := ConstructSignature(payload, secret, nil)

	if sig1 != sig2 {
		t.Error("simple signatures should be deterministic")
	}
}

func TestVerify_EventDataParsing(t *testing.T) {
	payload := []byte(`{
		"events": [
			{
				"event": "coupon.redeemed",
				"timestamp": "2024-01-01T00:00:00Z",
				"data": {"couponCode": "TEST", "userId": "u1", "usageId": 1, "reward": []}
			},
			{
				"event": "payment.created",
				"timestamp": "2024-01-01T00:00:00Z",
				"data": {"transactionId": "txn1", "userId": "u1", "amount": 9.99, "currency": "USD"}
			}
		]
	}`)
	secret := "test_secret"
	sig := ConstructSignature(payload, secret, nil)

	result := Verify(VerifyOptions{
		Payload:   payload,
		Signature: sig,
		Secret:    secret,
	})

	if !result.Valid {
		t.Fatalf("expected valid, got error: %s", result.Error)
	}
	if len(result.Payload.Events) != 2 {
		t.Fatalf("len(Events) = %d, want 2", len(result.Payload.Events))
	}
	if result.Payload.Events[0].Event != "coupon.redeemed" {
		t.Errorf("Event[0] = %q", result.Payload.Events[0].Event)
	}

	// Verify we can unmarshal event data
	var couponData struct {
		CouponCode string `json:"couponCode"`
	}
	if err := json.Unmarshal(result.Payload.Events[0].Data, &couponData); err != nil {
		t.Fatalf("unmarshal coupon data: %v", err)
	}
	if couponData.CouponCode != "TEST" {
		t.Errorf("CouponCode = %q", couponData.CouponCode)
	}
}
