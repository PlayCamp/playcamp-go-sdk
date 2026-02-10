package playcamp

import (
	"context"
	"fmt"
	"net/url"

	"github.com/playcamp/playcamp-go-sdk/internal/httpclient"
)

// CampaignServerService provides access to campaign endpoints for the Server API.
type CampaignServerService struct {
	client   *httpclient.Client
	basePath string
}

func newCampaignServerService(client *httpclient.Client) *CampaignServerService {
	return &CampaignServerService{client: client, basePath: "/v1/server/campaigns"}
}

// List returns a paginated list of campaigns.
func (s *CampaignServerService) List(ctx context.Context, opts *ListCampaignsOptions) (*PageResult[Campaign], error) {
	query := url.Values{}
	if opts != nil {
		if opts.Page != nil {
			query.Set("page", fmt.Sprintf("%d", *opts.Page))
		}
		if opts.Limit != nil {
			query.Set("limit", fmt.Sprintf("%d", *opts.Limit))
		}
	}
	return getPaginated[Campaign](ctx, s.client, s.basePath, query)
}

// Get returns a single campaign by ID.
func (s *CampaignServerService) Get(ctx context.Context, id string) (*Campaign, error) {
	if id == "" {
		return nil, &InputValidationError{Field: "id", Message: "must be a non-empty string"}
	}
	var result Campaign
	if err := s.client.Get(ctx, s.basePath+"/"+id, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetCreators returns the creators for a campaign.
func (s *CampaignServerService) GetCreators(ctx context.Context, campaignID string) ([]Creator, error) {
	if campaignID == "" {
		return nil, &InputValidationError{Field: "campaignId", Message: "must be a non-empty string"}
	}
	var result []Creator
	if err := s.client.Get(ctx, s.basePath+"/"+campaignID+"/creators", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}
