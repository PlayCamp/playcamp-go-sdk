package playcamp

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	t.Run("valid API key", func(t *testing.T) {
		client, err := NewClient("key_id:secret")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client.Campaigns == nil {
			t.Error("Campaigns service is nil")
		}
		if client.Creators == nil {
			t.Error("Creators service is nil")
		}
		if client.Coupons == nil {
			t.Error("Coupons service is nil")
		}
		if client.Sponsors == nil {
			t.Error("Sponsors service is nil")
		}
	})

	t.Run("empty API key", func(t *testing.T) {
		_, err := NewClient("")
		if err == nil {
			t.Fatal("expected error for empty API key")
		}
	})

	t.Run("invalid format - no colon", func(t *testing.T) {
		_, err := NewClient("invalid")
		if err == nil {
			t.Fatal("expected error for invalid API key")
		}
	})

	t.Run("invalid format - empty key ID", func(t *testing.T) {
		_, err := NewClient(":secret")
		if err == nil {
			t.Fatal("expected error for empty key ID")
		}
	})

	t.Run("invalid format - empty secret", func(t *testing.T) {
		_, err := NewClient("key:")
		if err == nil {
			t.Fatal("expected error for empty secret")
		}
	})

	t.Run("with options", func(t *testing.T) {
		client, err := NewClient("key:secret",
			WithEnvironment(EnvironmentSandbox),
			WithTestMode(true),
			WithMaxRetries(5),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client == nil {
			t.Fatal("client is nil")
		}
	})
}

func TestNewServer(t *testing.T) {
	t.Run("valid API key", func(t *testing.T) {
		server, err := NewServer("key_id:secret")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if server.Campaigns == nil {
			t.Error("Campaigns service is nil")
		}
		if server.Creators == nil {
			t.Error("Creators service is nil")
		}
		if server.Coupons == nil {
			t.Error("Coupons service is nil")
		}
		if server.Sponsors == nil {
			t.Error("Sponsors service is nil")
		}
		if server.Payments == nil {
			t.Error("Payments service is nil")
		}
		if server.Webhooks == nil {
			t.Error("Webhooks service is nil")
		}
	})

	t.Run("empty API key", func(t *testing.T) {
		_, err := NewServer("")
		if err == nil {
			t.Fatal("expected error for empty API key")
		}
	})
}

func TestEnvironmentURL(t *testing.T) {
	tests := []struct {
		env  Environment
		want string
	}{
		{EnvironmentSandbox, "https://sandbox-sdk-api.playcamp.io"},
		{EnvironmentLive, "https://sdk-api.playcamp.io"},
		{Environment("unknown"), "https://sdk-api.playcamp.io"},
	}
	for _, tt := range tests {
		t.Run(string(tt.env), func(t *testing.T) {
			got := EnvironmentURL(tt.env)
			if got != tt.want {
				t.Errorf("EnvironmentURL(%q) = %q, want %q", tt.env, got, tt.want)
			}
		})
	}
}

func TestPointerHelpers(t *testing.T) {
	intVal := Int(42)
	if *intVal != 42 {
		t.Errorf("Int(42) = %d, want 42", *intVal)
	}

	strVal := String("hello")
	if *strVal != "hello" {
		t.Errorf("String(\"hello\") = %q, want \"hello\"", *strVal)
	}

	boolVal := Bool(true)
	if *boolVal != true {
		t.Errorf("Bool(true) = %v, want true", *boolVal)
	}

	floatVal := Float64(3.14)
	if *floatVal != 3.14 {
		t.Errorf("Float64(3.14) = %f, want 3.14", *floatVal)
	}
}
