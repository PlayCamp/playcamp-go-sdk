package playcamp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestPlaytimeSessionService_Create(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/server/playtime/sessions" {
			t.Errorf("path = %s, want /v1/server/playtime/sessions", r.URL.Path)
		}
		assertAuthHeader(t, r)

		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["sessionId"] != "sess_1" {
			t.Errorf("sessionId = %v, want sess_1", body["sessionId"])
		}
		if body["startedAt"] != "2026-06-15T07:33:37Z" {
			t.Errorf("startedAt = %v, want 2026-06-15T07:33:37Z", body["startedAt"])
		}

		writeJSON(w, map[string]any{
			"data": map[string]any{
				"sessionId":       "sess_1",
				"userId":          "user_42",
				"durationSeconds": 1830,
				"recorded":        true,
				"createdAt":       "2026-06-15T08:04:07.000Z",
			},
		})
	})
	defer ts.Close()

	result, err := server.PlaytimeSessions.Create(context.Background(), CreatePlaytimeSessionParams{
		SessionID:       "sess_1",
		UserID:          "user_42",
		DurationSeconds: 1830,
		StartedAt:       time.Date(2026, 6, 15, 7, 33, 37, 0, time.UTC),
		EndedAt:         time.Date(2026, 6, 15, 8, 4, 7, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if result.SessionID != "sess_1" {
		t.Errorf("SessionID = %q, want %q", result.SessionID, "sess_1")
	}
	if !result.Recorded {
		t.Errorf("Recorded = %v, want true", result.Recorded)
	}
}

func TestPlaytimeSessionService_Create_Validation(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	cases := []struct {
		name   string
		params CreatePlaytimeSessionParams
	}{
		{"empty sessionId", CreatePlaytimeSessionParams{UserID: "u", DurationSeconds: 1, StartedAt: time.Now().UTC(), EndedAt: time.Now().UTC()}},
		{"empty userId", CreatePlaytimeSessionParams{SessionID: "s", DurationSeconds: 1, StartedAt: time.Now().UTC(), EndedAt: time.Now().UTC()}},
		{"non-positive duration", CreatePlaytimeSessionParams{SessionID: "s", UserID: "u", DurationSeconds: 0, StartedAt: time.Now().UTC(), EndedAt: time.Now().UTC()}},
		{"zero startedAt", CreatePlaytimeSessionParams{SessionID: "s", UserID: "u", DurationSeconds: 1, EndedAt: time.Now().UTC()}},
		{"endedAt before startedAt", CreatePlaytimeSessionParams{SessionID: "s", UserID: "u", DurationSeconds: 1, StartedAt: time.Date(2026, 6, 15, 8, 0, 0, 0, time.UTC), EndedAt: time.Date(2026, 6, 15, 7, 0, 0, 0, time.UTC)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := server.PlaytimeSessions.Create(context.Background(), tc.params)
			if err == nil {
				t.Fatal("expected validation error")
			}
			var valErr *InputValidationError
			if !errors.As(err, &valErr) {
				t.Errorf("expected InputValidationError, got %T", err)
			}
		})
	}
}

func TestPlaytimeSessionService_CreateBulk(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/server/playtime/sessions/bulk" {
			t.Errorf("path = %s, want /v1/server/playtime/sessions/bulk", r.URL.Path)
		}
		assertAuthHeader(t, r)

		writeJSON(w, map[string]any{
			"data": map[string]any{
				"totalRequested": 2,
				"successful":     1,
				"failed":         0,
				"skipped":        1,
				"results": []map[string]any{
					{"sessionId": "sess_1", "status": "SUCCESS"},
					{"sessionId": "sess_2", "status": "SKIPPED"},
				},
			},
		})
	})
	defer ts.Close()

	result, err := server.PlaytimeSessions.CreateBulk(context.Background(), CreateBulkPlaytimeSessionParams{
		Sessions: []CreatePlaytimeSessionParams{
			{SessionID: "sess_1", UserID: "user_42", DurationSeconds: 1830, StartedAt: time.Now().UTC(), EndedAt: time.Now().UTC()},
			{SessionID: "sess_2", UserID: "user_42", DurationSeconds: 600, StartedAt: time.Now().UTC(), EndedAt: time.Now().UTC()},
		},
	})
	if err != nil {
		t.Fatalf("CreateBulk: %v", err)
	}
	if result.TotalRequested != 2 {
		t.Errorf("TotalRequested = %d, want 2", result.TotalRequested)
	}
	if result.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", result.Skipped)
	}
	if len(result.Results) != 2 {
		t.Fatalf("len(Results) = %d, want 2", len(result.Results))
	}
	if result.Results[1].Status != "SKIPPED" {
		t.Errorf("Results[1].Status = %q, want SKIPPED", result.Results[1].Status)
	}
}

func TestPlaytimeSessionService_CreateBulk_Empty(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.PlaytimeSessions.CreateBulk(context.Background(), CreateBulkPlaytimeSessionParams{})
	if err == nil {
		t.Fatal("expected validation error")
	}
	var valErr *InputValidationError
	if !errors.As(err, &valErr) {
		t.Errorf("expected InputValidationError, got %T", err)
	}
}
