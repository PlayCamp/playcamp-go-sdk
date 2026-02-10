package playcamp

import (
	"context"
	"fmt"

	"github.com/playcamp/playcamp-go-sdk/internal/httpclient"
)

// WebhookService provides access to webhook endpoints for the Server API.
type WebhookService struct {
	client   *httpclient.Client
	basePath string
}

func newWebhookService(client *httpclient.Client) *WebhookService {
	return &WebhookService{client: client, basePath: "/v1/server/webhooks"}
}

// List returns all registered webhooks.
func (s *WebhookService) List(ctx context.Context) ([]Webhook, error) {
	var result []Webhook
	if err := s.client.Get(ctx, s.basePath, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Create registers a new webhook. The secret is only returned on creation.
func (s *WebhookService) Create(ctx context.Context, params CreateWebhookParams) (*WebhookWithSecret, error) {
	if err := requireNonEmpty("eventType", string(params.EventType)); err != nil {
		return nil, err
	}
	if err := requireNonEmpty("url", params.URL); err != nil {
		return nil, err
	}
	var result WebhookWithSecret
	if err := s.client.Post(ctx, s.basePath, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates a webhook.
func (s *WebhookService) Update(ctx context.Context, id int, params UpdateWebhookParams) (*Webhook, error) {
	if id <= 0 {
		return nil, &InputValidationError{Field: "id", Message: "must be a positive integer"}
	}
	var result Webhook
	if err := s.client.Put(ctx, fmt.Sprintf("%s/%d", s.basePath, id), params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a webhook.
func (s *WebhookService) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return &InputValidationError{Field: "id", Message: "must be a positive integer"}
	}
	return s.client.Delete(ctx, fmt.Sprintf("%s/%d", s.basePath, id), nil)
}

// GetLogs returns the delivery logs for a webhook.
func (s *WebhookService) GetLogs(ctx context.Context, id int) ([]WebhookLog, error) {
	if id <= 0 {
		return nil, &InputValidationError{Field: "id", Message: "must be a positive integer"}
	}
	var result []WebhookLog
	if err := s.client.Get(ctx, fmt.Sprintf("%s/%d/logs", s.basePath, id), nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Test sends a test webhook.
func (s *WebhookService) Test(ctx context.Context, id int) (*WebhookTestResult, error) {
	if id <= 0 {
		return nil, &InputValidationError{Field: "id", Message: "must be a positive integer"}
	}
	var result WebhookTestResult
	if err := s.client.Post(ctx, fmt.Sprintf("%s/%d/test", s.basePath, id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
