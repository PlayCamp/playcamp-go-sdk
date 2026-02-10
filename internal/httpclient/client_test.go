package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newTestErrorFactory() ErrorFactory {
	return ErrorFactory{
		NewAPIError: func(statusCode int, body []byte) error {
			return fmt.Errorf("api error %d: %s", statusCode, string(body))
		},
		NewNetworkError: func(err error) error {
			return fmt.Errorf("network error: %w", err)
		},
	}
}

func newTestClient(t *testing.T, ts *httptest.Server, opts ...func(*Config)) *Client {
	t.Helper()
	cfg := Config{
		BaseURL:      ts.URL,
		APIKey:       "test_key:test_secret",
		Timeout:      5 * time.Second,
		MaxRetries:   0,
		ErrorFactory: newTestErrorFactory(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return New(cfg)
}

// --- New ---

func TestNew_DefaultHTTPClient(t *testing.T) {
	c := New(Config{
		BaseURL:      "https://example.com",
		APIKey:       "key:secret",
		Timeout:      10 * time.Second,
		ErrorFactory: newTestErrorFactory(),
	})
	if c.httpClient == nil {
		t.Fatal("httpClient should not be nil")
	}
	if c.baseURL != "https://example.com" {
		t.Errorf("baseURL = %q", c.baseURL)
	}
}

func TestNew_CustomHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 1 * time.Second}
	c := New(Config{
		BaseURL:      "https://example.com/",
		APIKey:       "key:secret",
		HTTPClient:   custom,
		ErrorFactory: newTestErrorFactory(),
	})
	if c.httpClient != custom {
		t.Fatal("expected custom httpClient")
	}
	// Trailing slash should be trimmed
	if c.baseURL != "https://example.com" {
		t.Errorf("baseURL = %q, trailing slash not trimmed", c.baseURL)
	}
}

func TestNew_DebugEnabled(t *testing.T) {
	var logged bool
	New(Config{
		BaseURL:      "https://example.com",
		APIKey:       "key:secret",
		ErrorFactory: newTestErrorFactory(),
		Debug: &DebugOptions{
			Enabled: true,
			Logger: func(format string, args ...any) {
				logged = true
			},
		},
	})
	if !logged {
		t.Error("expected debug log on init")
	}
}

// --- Get ---

func TestGet_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/resource" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"id": "123", "name": "test"},
		})
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	var result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	err := c.Get(context.Background(), "/api/resource", nil, &result)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if result.ID != "123" {
		t.Errorf("ID = %q, want %q", result.ID, "123")
	}
	if result.Name != "test" {
		t.Errorf("Name = %q, want %q", result.Name, "test")
	}
}

func TestGet_WithQueryParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("page = %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("limit = %s", r.URL.Query().Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": "1"}})
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	query := url.Values{}
	query.Set("page", "2")
	query.Set("limit", "10")
	var result struct{ ID string `json:"id"` }
	err := c.Get(context.Background(), "/api/items", query, &result)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
}

func TestGet_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{"message": "Not found"})
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	var result any
	err := c.Get(context.Background(), "/api/missing", nil, &result)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "api error 404") {
		t.Errorf("error = %q, want api error 404", err.Error())
	}
}

// --- Post ---

func TestPost_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "test" {
			t.Errorf("name = %q", body["name"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"id": "new-1", "name": "test"},
		})
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	var result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	err := c.Post(context.Background(), "/api/create", map[string]string{"name": "test"}, &result)
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	if result.ID != "new-1" {
		t.Errorf("ID = %q", result.ID)
	}
}

// --- Put ---

func TestPut_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"id": "1", "name": "updated"},
		})
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	var result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	err := c.Put(context.Background(), "/api/update/1", map[string]string{"name": "updated"}, &result)
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if result.Name != "updated" {
		t.Errorf("Name = %q", result.Name)
	}
}

// --- Delete ---

func TestDelete_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	err := c.Delete(context.Background(), "/api/resource/1", nil)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestDelete_WithQuery(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("campaignId") != "c1" {
			t.Errorf("campaignId = %s", r.URL.Query().Get("campaignId"))
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	query := url.Values{}
	query.Set("campaignId", "c1")
	err := c.Delete(context.Background(), "/api/resource/1", query)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

// --- GetPaginated ---

func TestGetPaginated_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data":       []map[string]any{{"id": "1"}, {"id": "2"}},
			"pagination": map[string]any{"page": 1, "limit": 10, "total": 2, "totalPages": 1},
		})
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	result, err := c.GetPaginated(context.Background(), "/api/items", nil)
	if err != nil {
		t.Fatalf("GetPaginated: %v", err)
	}
	if result.Data == nil {
		t.Fatal("Data should not be nil")
	}
	if result.Pagination == nil {
		t.Fatal("Pagination should not be nil")
	}
}

// --- Headers ---

func TestSetHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer mykey:mysecret" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{}})
	}))
	defer ts.Close()

	c := newTestClient(t, ts, func(cfg *Config) {
		cfg.APIKey = "mykey:mysecret"
	})
	err := c.Get(context.Background(), "/test", nil, &json.RawMessage{})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
}

// --- buildURL ---

func TestBuildURL_NoQuery(t *testing.T) {
	c := &Client{baseURL: "https://api.example.com"}
	got := c.buildURL("/v1/test", nil)
	if got != "https://api.example.com/v1/test" {
		t.Errorf("buildURL = %q", got)
	}
}

func TestBuildURL_WithQuery(t *testing.T) {
	c := &Client{baseURL: "https://api.example.com"}
	q := url.Values{}
	q.Set("page", "1")
	q.Set("limit", "10")
	got := c.buildURL("/v1/items", q)
	if !strings.Contains(got, "page=1") || !strings.Contains(got, "limit=10") {
		t.Errorf("buildURL = %q, missing query params", got)
	}
}

func TestBuildURL_IsTest(t *testing.T) {
	c := &Client{baseURL: "https://api.example.com", isTest: true}
	got := c.buildURL("/v1/test", nil)
	if !strings.Contains(got, "isTest=true") {
		t.Errorf("buildURL = %q, missing isTest", got)
	}
}

func TestBuildURL_IsTestWithQuery(t *testing.T) {
	c := &Client{baseURL: "https://api.example.com", isTest: true}
	q := url.Values{}
	q.Set("page", "1")
	got := c.buildURL("/v1/items", q)
	if !strings.Contains(got, "isTest=true") || !strings.Contains(got, "page=1") {
		t.Errorf("buildURL = %q", got)
	}
}

func TestBuildURL_EmptyValueSkipped(t *testing.T) {
	c := &Client{baseURL: "https://api.example.com"}
	q := url.Values{}
	q.Set("key", "")
	q.Set("other", "val")
	got := c.buildURL("/v1/test", q)
	if strings.Contains(got, "key=") {
		t.Errorf("buildURL = %q, should skip empty values", got)
	}
	if !strings.Contains(got, "other=val") {
		t.Errorf("buildURL = %q, missing non-empty param", got)
	}
}

// --- Retry ---

func TestRetry_OnServerError(t *testing.T) {
	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"server error"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"ok": true}})
	}))
	defer ts.Close()

	c := newTestClient(t, ts, func(cfg *Config) {
		cfg.MaxRetries = 3
	})
	var result struct{ OK bool `json:"ok"` }
	err := c.Get(context.Background(), "/test", nil, &result)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !result.OK {
		t.Error("expected ok=true")
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("attempts = %d, want 3", atomic.LoadInt32(&attempts))
	}
}

func TestRetry_On429(t *testing.T) {
	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"message":"rate limited"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"ok": true}})
	}))
	defer ts.Close()

	c := newTestClient(t, ts, func(cfg *Config) {
		cfg.MaxRetries = 2
	})
	var result struct{ OK bool `json:"ok"` }
	err := c.Get(context.Background(), "/test", nil, &result)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("attempts = %d, want 2", atomic.LoadInt32(&attempts))
	}
}

func TestRetry_NoRetryOn4xx(t *testing.T) {
	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"bad request"}`))
	}))
	defer ts.Close()

	c := newTestClient(t, ts, func(cfg *Config) {
		cfg.MaxRetries = 3
	})
	var result any
	err := c.Get(context.Background(), "/test", nil, &result)
	if err == nil {
		t.Fatal("expected error")
	}
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("attempts = %d, want 1 (no retry on 400)", atomic.LoadInt32(&attempts))
	}
}

func TestRetry_ExhaustedRetries(t *testing.T) {
	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"server error"}`))
	}))
	defer ts.Close()

	c := newTestClient(t, ts, func(cfg *Config) {
		cfg.MaxRetries = 2
	})
	var result any
	err := c.Get(context.Background(), "/test", nil, &result)
	if err == nil {
		t.Fatal("expected error after exhausted retries")
	}
	// 1 initial + 2 retries = 3 attempts
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("attempts = %d, want 3", atomic.LoadInt32(&attempts))
	}
}

func TestRetry_NetworkError(t *testing.T) {
	// Use a server that immediately closes
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.Close() // Close immediately to force network errors

	c := newTestClient(t, ts, func(cfg *Config) {
		cfg.MaxRetries = 1
		cfg.Timeout = 1 * time.Second
	})
	var result any
	err := c.Get(context.Background(), "/test", nil, &result)
	if err == nil {
		t.Fatal("expected network error")
	}
	if !strings.Contains(err.Error(), "network error") {
		t.Errorf("error = %q, expected network error", err.Error())
	}
}

// --- Context Cancellation ---

func TestContextCancellation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := newTestClient(t, ts, func(cfg *Config) {
		cfg.Timeout = 5 * time.Second
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var result any
	err := c.Get(ctx, "/test", nil, &result)
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
}

// --- Debug Logging ---

func TestDebugLogging_Request(t *testing.T) {
	var logs []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{}})
	}))
	defer ts.Close()

	c := newTestClient(t, ts, func(cfg *Config) {
		cfg.Debug = &DebugOptions{
			Enabled:         true,
			Logger:          func(format string, args ...any) { logs = append(logs, fmt.Sprintf(format, args...)) },
			LogRequestBody:  true,
			LogResponseBody: true,
		}
	})

	err := c.Post(context.Background(), "/test", map[string]string{"key": "val"}, &json.RawMessage{})
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	if len(logs) < 2 {
		t.Fatalf("expected at least 2 log entries (init + request + response), got %d", len(logs))
	}

	// Check request log
	foundRequest := false
	foundResponse := false
	for _, log := range logs {
		if strings.Contains(log, "POST") && strings.Contains(log, "/test") {
			foundRequest = true
		}
		if strings.Contains(log, "200") {
			foundResponse = true
		}
	}
	if !foundRequest {
		t.Error("missing request log")
	}
	if !foundResponse {
		t.Error("missing response log")
	}
}

func TestDebugLogging_Disabled(t *testing.T) {
	var logCount int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{}})
	}))
	defer ts.Close()

	c := newTestClient(t, ts, func(cfg *Config) {
		cfg.Debug = &DebugOptions{
			Enabled: false,
			Logger:  func(format string, args ...any) { logCount++ },
		}
	})

	c.Get(context.Background(), "/test", nil, &json.RawMessage{})
	if logCount != 0 {
		t.Errorf("expected 0 logs when disabled, got %d", logCount)
	}
}

func TestDebugLogging_NilLogger(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{}})
	}))
	defer ts.Close()

	// Should not panic with nil logger
	c := newTestClient(t, ts, func(cfg *Config) {
		cfg.Debug = &DebugOptions{
			Enabled: true,
			Logger:  nil,
		}
	})

	err := c.Get(context.Background(), "/test", nil, &json.RawMessage{})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
}

func TestDebugLogging_Retry(t *testing.T) {
	var logs []string
	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"error"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{}})
	}))
	defer ts.Close()

	c := newTestClient(t, ts, func(cfg *Config) {
		cfg.MaxRetries = 1
		cfg.Debug = &DebugOptions{
			Enabled:         true,
			Logger:          func(format string, args ...any) { logs = append(logs, fmt.Sprintf(format, args...)) },
			LogResponseBody: true,
		}
	})

	err := c.Get(context.Background(), "/test", nil, &json.RawMessage{})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	foundRetry := false
	for _, log := range logs {
		if strings.Contains(log, "Retrying") {
			foundRetry = true
		}
	}
	if !foundRetry {
		t.Error("expected retry log entry")
	}
}

// --- Invalid response envelope ---

func TestGet_InvalidEnvelope(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`not json`))
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	var result any
	err := c.Get(context.Background(), "/test", nil, &result)
	if err == nil {
		t.Fatal("expected error for invalid envelope")
	}
	if !strings.Contains(err.Error(), "decode response envelope") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestGet_InvalidDataInEnvelope(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Valid envelope but data cannot be decoded into a struct
		json.NewEncoder(w).Encode(map[string]any{
			"data": "not-an-object",
		})
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	var result struct{ ID string `json:"id"` }
	err := c.Get(context.Background(), "/test", nil, &result)
	if err == nil {
		t.Fatal("expected error for invalid data in envelope")
	}
	if !strings.Contains(err.Error(), "decode response data") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestGetPaginated_InvalidEnvelope(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`not json`))
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	_, err := c.GetPaginated(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("expected error for invalid paginated envelope")
	}
	if !strings.Contains(err.Error(), "decode response") {
		t.Errorf("error = %q", err.Error())
	}
}

// --- Delete with nil result (no body parsing) ---

func TestDelete_NilResult(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return 200 with body - Delete should ignore since result is nil
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"id":"1"}}`))
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	err := c.Delete(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
}
