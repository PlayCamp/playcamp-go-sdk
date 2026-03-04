package playcamp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestWebviewServerService_CreateOTT(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/server/webview/ott" {
			t.Errorf("path = %s, want /v1/server/webview/ott", r.URL.Path)
		}
		assertAuthHeader(t, r)

		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["userId"] != "user1" {
			t.Errorf("userId = %v, want user1", body["userId"])
		}

		writeJSON(w, map[string]any{
			"data": map[string]any{
				"ott":       "ott_abc123",
				"expiresAt": "2024-12-31T23:59:59Z",
			},
		})
	})
	defer ts.Close()

	result, err := server.Webview.CreateOTT(context.Background(), WebviewOttParams{
		UserID:     "user1",
		CampaignID: "campaign1",
	})
	if err != nil {
		t.Fatalf("CreateOTT: %v", err)
	}
	if result.OTT != "ott_abc123" {
		t.Errorf("OTT = %q, want %q", result.OTT, "ott_abc123")
	}
	if result.ExpiresAt != "2024-12-31T23:59:59Z" {
		t.Errorf("ExpiresAt = %q, want %q", result.ExpiresAt, "2024-12-31T23:59:59Z")
	}
}

func TestWebviewServerService_CreateOTT_EmptyUserID(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Webview.CreateOTT(context.Background(), WebviewOttParams{
		UserID: "",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	var valErr *InputValidationError
	if !errors.As(err, &valErr) {
		t.Errorf("expected InputValidationError, got %T", err)
	}
}
