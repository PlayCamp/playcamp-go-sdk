package playcamp

import (
	"context"
	"fmt"
	"net/url"

	"github.com/playcamp/playcamp-go-sdk/internal/httpclient"
)

// CreatorServerService provides access to creator endpoints for the Server API.
type CreatorServerService struct {
	client   *httpclient.Client
	basePath string
}

func newCreatorServerService(client *httpclient.Client) *CreatorServerService {
	return &CreatorServerService{client: client, basePath: "/v1/server/creators"}
}

// Get returns a creator by their unique key.
func (s *CreatorServerService) Get(ctx context.Context, creatorKey string) (*Creator, error) {
	if creatorKey == "" {
		return nil, &InputValidationError{Field: "creatorKey", Message: "must be a non-empty string"}
	}
	var result Creator
	if err := s.client.Get(ctx, s.basePath+"/"+creatorKey, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Search searches for creators by keyword.
func (s *CreatorServerService) Search(ctx context.Context, params SearchCreatorsParams) ([]Creator, error) {
	if params.Keyword == "" {
		return nil, &InputValidationError{Field: "keyword", Message: "must be a non-empty string"}
	}
	query := url.Values{}
	query.Set("keyword", params.Keyword)
	if params.CampaignID != nil {
		query.Set("campaignId", *params.CampaignID)
	}
	if params.Limit != nil {
		query.Set("limit", fmt.Sprintf("%d", *params.Limit))
	}
	var result []Creator
	if err := s.client.Get(ctx, s.basePath+"/search", query, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetCoupons returns the coupon codes for a creator.
func (s *CreatorServerService) GetCoupons(ctx context.Context, creatorKey string) ([]CreatorCoupon, error) {
	if creatorKey == "" {
		return nil, &InputValidationError{Field: "creatorKey", Message: "must be a non-empty string"}
	}
	var result []CreatorCoupon
	if err := s.client.Get(ctx, s.basePath+"/"+creatorKey+"/coupons", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}
