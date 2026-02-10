package playcamp

import (
	"context"
	"net/url"

	"github.com/playcamp/playcamp-go-sdk/internal/httpclient"
)

// SponsorService provides access to sponsor endpoints for the Client API.
type SponsorService struct {
	client   *httpclient.Client
	basePath string
}

func newSponsorService(client *httpclient.Client) *SponsorService {
	return &SponsorService{client: client, basePath: "/v1/client/sponsors"}
}

// Get returns the sponsor status for a user. The result is normalized to a slice.
func (s *SponsorService) Get(ctx context.Context, params GetSponsorParams) ([]Sponsor, error) {
	if params.UserID == "" {
		return nil, &InputValidationError{Field: "userId", Message: "must be a non-empty string"}
	}
	query := url.Values{}
	query.Set("userId", params.UserID)
	if params.CampaignID != nil {
		query.Set("campaignId", *params.CampaignID)
	}
	var result []Sponsor
	if err := s.client.Get(ctx, s.basePath, query, &result); err != nil {
		return nil, err
	}
	return result, nil
}
