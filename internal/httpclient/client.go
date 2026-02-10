package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// maxResponseSize is the maximum response body size (10 MB).
const maxResponseSize = 10 << 20

// DebugOptions configures debug logging.
type DebugOptions struct {
	Enabled         bool
	Logger          func(format string, args ...any)
	LogRequestBody  bool
	LogResponseBody bool
}

// ErrorFactory creates typed errors from HTTP status codes and response bodies.
type ErrorFactory struct {
	// NewAPIError creates an API error from a status code and response body.
	NewAPIError func(statusCode int, body []byte) error
	// NewNetworkError creates a network error wrapping the underlying cause.
	NewNetworkError func(err error) error
}

// Config holds the HTTP client configuration.
type Config struct {
	BaseURL      string
	APIKey       string
	Timeout      time.Duration
	IsTest       bool
	MaxRetries   int
	HTTPClient   *http.Client
	Debug        *DebugOptions
	ErrorFactory ErrorFactory
}

// Client is the internal HTTP client for the PlayCamp API.
type Client struct {
	baseURL      string
	apiKey       string
	timeout      time.Duration
	isTest       bool
	maxRetries   int
	httpClient   *http.Client
	debug        *DebugOptions
	errorFactory ErrorFactory
}

// New creates a new HTTP client.
func New(cfg Config) *Client {
	hc := cfg.HTTPClient
	if hc == nil {
		hc = &http.Client{}
	}
	c := &Client{
		baseURL:      strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:       cfg.APIKey,
		timeout:      cfg.Timeout,
		isTest:       cfg.IsTest,
		maxRetries:   cfg.MaxRetries,
		httpClient:   hc,
		debug:        cfg.Debug,
		errorFactory: cfg.ErrorFactory,
	}
	if c.debug != nil && c.debug.Enabled {
		c.debugLog("Debug mode enabled")
	}
	return c
}

// dataEnvelope is the standard API response wrapper.
type dataEnvelope struct {
	Data json.RawMessage `json:"data"`
}

// paginatedEnvelope is the paginated API response wrapper.
type paginatedEnvelope struct {
	Data       json.RawMessage `json:"data"`
	Pagination json.RawMessage `json:"pagination"`
}

// Get performs a GET request and unwraps the data envelope.
func (c *Client) Get(ctx context.Context, path string, query url.Values, result any) error {
	return c.doRequest(ctx, http.MethodGet, path, query, nil, result, true)
}

// Post performs a POST request and unwraps the data envelope.
func (c *Client) Post(ctx context.Context, path string, body any, result any) error {
	return c.doRequest(ctx, http.MethodPost, path, nil, body, result, true)
}

// Put performs a PUT request and unwraps the data envelope.
func (c *Client) Put(ctx context.Context, path string, body any, result any) error {
	return c.doRequest(ctx, http.MethodPut, path, nil, body, result, true)
}

// Delete performs a DELETE request with optional query parameters.
func (c *Client) Delete(ctx context.Context, path string, query url.Values) error {
	return c.doRequest(ctx, http.MethodDelete, path, query, nil, nil, false)
}

// PaginatedResult holds a raw paginated response.
type PaginatedResult struct {
	Data       json.RawMessage
	Pagination json.RawMessage
}

// GetPaginated performs a GET request and returns both data and pagination.
func (c *Client) GetPaginated(ctx context.Context, path string, query url.Values) (*PaginatedResult, error) {
	respBody, err := c.doRequestRaw(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return nil, err
	}
	var env paginatedEnvelope
	if err := json.Unmarshal(respBody, &env); err != nil {
		return nil, fmt.Errorf("playcamp: failed to decode response: %w", err)
	}
	return &PaginatedResult{
		Data:       env.Data,
		Pagination: env.Pagination,
	}, nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, query url.Values, body any, result any, unwrap bool) error {
	respBody, err := c.doRequestRaw(ctx, method, path, query, body)
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}
	if unwrap {
		var env dataEnvelope
		if err := json.Unmarshal(respBody, &env); err != nil {
			return fmt.Errorf("playcamp: failed to decode response envelope: %w", err)
		}
		if err := json.Unmarshal(env.Data, result); err != nil {
			return fmt.Errorf("playcamp: failed to decode response data: %w", err)
		}
		return nil
	}
	if err := json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("playcamp: failed to decode response: %w", err)
	}
	return nil
}

func (c *Client) doRequestRaw(ctx context.Context, method, path string, query url.Values, body any) ([]byte, error) {
	reqURL := c.buildURL(path, query)
	var bodyReader io.Reader
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("playcamp: failed to encode request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	c.debugLogRequest(method, reqURL, bodyBytes)

	startTime := time.Now()
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}

		reqCtx, cancel := context.WithTimeout(ctx, c.timeout)
		req, err := http.NewRequestWithContext(reqCtx, method, reqURL, bodyReader)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("playcamp: failed to create request: %w", err)
		}
		c.setHeaders(req)

		resp, err := c.httpClient.Do(req)
		cancel()

		if err != nil {
			lastErr = err
			responseTime := time.Since(startTime)
			c.debugLogError(method, reqURL, err, responseTime)

			// Only retry network errors for idempotent methods.
			if attempt < c.maxRetries && isIdempotent(method) {
				delay := calculateBackoff(attempt)
				c.debugLogRetry(method, reqURL, attempt+1, c.maxRetries, delay)
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(delay):
				}
				continue
			}
			return nil, c.errorFactory.NewNetworkError(lastErr)
		}

		respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("playcamp: failed to read response body: %w", err)
		}

		responseTime := time.Since(startTime)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			c.debugLogResponse(method, reqURL, resp.StatusCode, responseTime, respBody)
			return respBody, nil
		}

		c.debugLogResponse(method, reqURL, resp.StatusCode, responseTime, respBody)

		// Only retry on retryable status codes for idempotent methods.
		if attempt < c.maxRetries && shouldRetry(method, resp.StatusCode) {
			delay := calculateBackoff(attempt)
			c.debugLogRetry(method, reqURL, attempt+1, c.maxRetries, delay)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
			continue
		}

		return nil, c.errorFactory.NewAPIError(resp.StatusCode, respBody)
	}

	return nil, c.errorFactory.NewNetworkError(lastErr)
}

func (c *Client) buildURL(path string, query url.Values) string {
	u := c.baseURL + path
	if c.isTest || len(query) > 0 {
		params := url.Values{}
		if c.isTest {
			params.Set("isTest", "true")
		}
		for k, vs := range query {
			for _, v := range vs {
				if v != "" {
					params.Add(k, v)
				}
			}
		}
		if encoded := params.Encode(); encoded != "" {
			u += "?" + encoded
		}
	}
	return u
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
}

func (c *Client) debugLog(format string, args ...any) {
	if c.debug == nil || !c.debug.Enabled {
		return
	}
	logger := c.debug.Logger
	if logger == nil {
		return
	}
	logger("[PlayCamp SDK] "+format, args...)
}

// sensitiveFields are JSON field names that should be redacted in debug logs.
var sensitiveFields = map[string]bool{
	"couponCode":   true,
	"userId":       true,
	"gameUserUuid": true,
	"secret":       true,
	"apiKey":       true,
}

// redactBody returns a JSON string with sensitive fields masked.
func redactBody(data []byte) string {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return "[non-JSON body]"
	}
	for k := range m {
		if sensitiveFields[k] {
			m[k] = "[REDACTED]"
		}
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func (c *Client) debugLogRequest(method, u string, body []byte) {
	if c.debug == nil || !c.debug.Enabled {
		return
	}
	msg := fmt.Sprintf("→ %s %s", method, u)
	if c.debug.LogRequestBody && len(body) > 0 {
		msg += fmt.Sprintf("\n  body: %s", redactBody(body))
	}
	c.debugLog("%s", msg)
}

func (c *Client) debugLogResponse(method, u string, status int, responseTime time.Duration, body []byte) {
	if c.debug == nil || !c.debug.Enabled {
		return
	}
	msg := fmt.Sprintf("← %d %s %s (%dms)", status, method, u, responseTime.Milliseconds())
	if c.debug.LogResponseBody && len(body) > 0 {
		msg += fmt.Sprintf("\n  response: %s", redactBody(body))
	}
	c.debugLog("%s", msg)
}

func (c *Client) debugLogError(method, u string, err error, responseTime time.Duration) {
	if c.debug == nil || !c.debug.Enabled {
		return
	}
	c.debugLog("✗ %s %s - %v (%dms)", method, u, err, responseTime.Milliseconds())
}

func (c *Client) debugLogRetry(method, u string, attempt, maxRetries int, delay time.Duration) {
	if c.debug == nil || !c.debug.Enabled {
		return
	}
	c.debugLog("↻ %s %s - Retrying in %dms (attempt %d/%d)", method, u, delay.Milliseconds(), attempt, maxRetries)
}
