package playcamp

import (
	"fmt"
	"strings"

	"github.com/playcamp/playcamp-go-sdk/internal/httpclient"
)

// Client provides read-only access to the PlayCamp API using a CLIENT API key.
type Client struct {
	Campaigns *CampaignService
	Creators  *CreatorService
	Coupons   *CouponService
	Sponsors  *SponsorService
}

// NewClient creates a new PlayCamp Client for read-only operations.
//
// The apiKey must be in "keyId:secret" format.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
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

	return &Client{
		Campaigns: newCampaignService(hc),
		Creators:  newCreatorService(hc),
		Coupons:   newCouponService(hc),
		Sponsors:  newSponsorService(hc),
	}, nil
}

func validateAPIKey(apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("playcamp: API key is required")
	}
	parts := strings.SplitN(apiKey, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("playcamp: API key must be in format \"keyId:secret\"")
	}
	return nil
}

func buildHTTPConfig(apiKey string, cfg *config) (httpclient.Config, error) {
	baseURL := cfg.baseURL
	if baseURL == "" {
		baseURL = EnvironmentURL(cfg.environment)
		if baseURL == "" {
			return httpclient.Config{}, fmt.Errorf("playcamp: unsupported environment %q", cfg.environment)
		}
	} else if !strings.HasPrefix(baseURL, "https://") && !isLocalhostURL(baseURL) {
		return httpclient.Config{}, fmt.Errorf("playcamp: base URL must use HTTPS scheme, got %q", baseURL)
	}

	var debug *httpclient.DebugOptions
	if cfg.debug != nil {
		debug = &httpclient.DebugOptions{
			Enabled:         cfg.debug.Enabled,
			Logger:          cfg.debug.Logger,
			LogRequestBody:  cfg.debug.LogRequestBody,
			LogResponseBody: cfg.debug.LogResponseBody,
		}
	}

	return httpclient.Config{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		Timeout:    cfg.timeout,
		IsTest:     cfg.isTest,
		MaxRetries: cfg.maxRetries,
		HTTPClient: cfg.httpClient,
		Debug:      debug,
		ErrorFactory: httpclient.ErrorFactory{
			NewAPIError:     newAPIError,
			NewNetworkError: newNetworkError,
		},
	}, nil
}

// isLocalhostURL returns true if the URL points to localhost (for testing).
func isLocalhostURL(u string) bool {
	return strings.HasPrefix(u, "http://localhost") || strings.HasPrefix(u, "http://127.0.0.1")
}
