package playcamp

import (
	"context"

	"github.com/playcamp/playcamp-go-sdk/internal/httpclient"
)

// WebviewServerService provides access to WebView server endpoints.
type WebviewServerService struct {
	client   *httpclient.Client
	basePath string
}

func newWebviewServerService(client *httpclient.Client) *WebviewServerService {
	return &WebviewServerService{client: client, basePath: "/v1/server/webview"}
}

// CreateOTT creates a one-time token for WebView authentication.
func (s *WebviewServerService) CreateOTT(ctx context.Context, params WebviewOttParams) (*WebviewOttResult, error) {
	if err := requireNonEmpty("userId", params.UserID); err != nil {
		return nil, err
	}
	var result WebviewOttResult
	if err := s.client.Post(ctx, s.basePath+"/ott", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
