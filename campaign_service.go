package playcamp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/playcamp/playcamp-go-sdk/internal/httpclient"
)

// campaignBase implements shared campaign operations for both client and server APIs.
type campaignBase struct {
	client   *httpclient.Client
	basePath string
}

// CampaignService provides access to campaign endpoints for the Client API.
type CampaignService struct{ campaignBase }

func newCampaignService(client *httpclient.Client) *CampaignService {
	return &CampaignService{campaignBase{client: client, basePath: "/v1/client/campaigns"}}
}

// List returns a paginated list of campaigns.
func (s *campaignBase) List(ctx context.Context, opts *PaginationOptions) (*PageResult[Campaign], error) {
	return getPaginated[Campaign](ctx, s.client, s.basePath, paginationQuery(opts))
}

// Get returns a single campaign by ID.
func (s *campaignBase) Get(ctx context.Context, id string) (*Campaign, error) {
	if err := requireNonEmpty("id", id); err != nil {
		return nil, err
	}
	var result Campaign
	if err := s.client.Get(ctx, s.basePath+"/"+url.PathEscape(id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetCreators returns the creators for a campaign.
func (s *campaignBase) GetCreators(ctx context.Context, campaignID string) ([]Creator, error) {
	if err := requireNonEmpty("campaignId", campaignID); err != nil {
		return nil, err
	}
	var result []Creator
	if err := s.client.Get(ctx, s.basePath+"/"+url.PathEscape(campaignID)+"/creators", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetPackages returns the coupon packages for a campaign. (Client API only)
func (s *CampaignService) GetPackages(ctx context.Context, campaignID string) ([]CouponPackage, error) {
	if err := requireNonEmpty("campaignId", campaignID); err != nil {
		return nil, err
	}
	var result []CouponPackage
	if err := s.client.Get(ctx, s.basePath+"/"+url.PathEscape(campaignID)+"/packages", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// getPaginated is a helper that fetches a paginated endpoint and decodes the result.
func getPaginated[T any](ctx context.Context, client *httpclient.Client, path string, query url.Values) (*PageResult[T], error) {
	raw, err := client.GetPaginated(ctx, path, query)
	if err != nil {
		return nil, err
	}
	var data []T
	if err := json.Unmarshal(raw.Data, &data); err != nil {
		return nil, fmt.Errorf("playcamp: failed to decode paginated data: %w", err)
	}
	var pagination Pagination
	if err := json.Unmarshal(raw.Pagination, &pagination); err != nil {
		return nil, fmt.Errorf("playcamp: failed to decode pagination: %w", err)
	}
	return &PageResult[T]{
		Data:        data,
		Pagination:  pagination,
		HasNextPage: pagination.Page < pagination.TotalPages,
	}, nil
}
