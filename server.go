package playcamp

import (
	"github.com/playcamp/playcamp-go-sdk/internal/httpclient"
)

// Server provides full read/write access to the PlayCamp API using a SERVER API key.
type Server struct {
	Campaigns *CampaignServerService
	Creators  *CreatorServerService
	Coupons   *CouponServerService
	Sponsors  *SponsorServerService
	Payments  *PaymentService
	Webhooks  *WebhookService
	Webview   *WebviewServerService
}

// NewServer creates a new PlayCamp Server for read/write operations.
//
// The apiKey must be in "keyId:secret" format and must be a SERVER key.
func NewServer(apiKey string, opts ...Option) (*Server, error) {
	if err := validateAPIKey(apiKey); err != nil {
		return nil, err
	}

	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	httpCfg, err := buildHTTPConfig(apiKey, cfg)
	if err != nil {
		return nil, err
	}
	hc := httpclient.New(httpCfg)

	return &Server{
		Campaigns: newCampaignServerService(hc),
		Creators:  newCreatorServerService(hc),
		Coupons:   newCouponServerService(hc),
		Sponsors:  newSponsorServerService(hc),
		Payments:  newPaymentService(hc),
		Webhooks:  newWebhookService(hc),
		Webview:   newWebviewServerService(hc),
	}, nil
}
