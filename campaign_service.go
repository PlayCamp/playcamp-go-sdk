package playcamp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/playcamp/playcamp-go-sdk/internal/httpclient"
)

// CampaignService provides access to campaign endpoints for the Client API.
type CampaignService struct {
	client   *httpclient.Client
	basePath string
}

func newCampaignService(client *httpclient.Client) *CampaignService {
	return &CampaignService{client: client, basePath: "/v1/client/campaigns"}
}

// List returns a paginated list of campaigns.
func (s *CampaignService) List(ctx context.Context, opts *ListCampaignsOptions) (*PageResult[Campaign], error) {
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
func (s *CampaignService) Get(ctx context.Context, id string) (*Campaign, error) {
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
func (s *CampaignService) GetCreators(ctx context.Context, campaignID string) ([]Creator, error) {
	if campaignID == "" {
		return nil, &InputValidationError{Field: "campaignId", Message: "must be a non-empty string"}
	}
	var result []Creator
	if err := s.client.Get(ctx, s.basePath+"/"+campaignID+"/creators", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetPackages returns the coupon packages for a campaign.
func (s *CampaignService) GetPackages(ctx context.Context, campaignID string) ([]CouponPackage, error) {
	if campaignID == "" {
		return nil, &InputValidationError{Field: "campaignId", Message: "must be a non-empty string"}
	}
	var result []CouponPackage
	if err := s.client.Get(ctx, s.basePath+"/"+campaignID+"/packages", nil, &result); err != nil {
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
