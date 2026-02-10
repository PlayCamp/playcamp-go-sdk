package playcamp

import (
	"context"
	"fmt"
	"net/url"

	"github.com/playcamp/playcamp-go-sdk/internal/httpclient"
)

// CreatorService provides access to creator endpoints for the Client API.
type CreatorService struct {
	client   *httpclient.Client
	basePath string
}

func newCreatorService(client *httpclient.Client) *CreatorService {
	return &CreatorService{client: client, basePath: "/v1/client/creators"}
}

// Get returns a creator by their unique key.
func (s *CreatorService) Get(ctx context.Context, creatorKey string) (*Creator, error) {
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
func (s *CreatorService) Search(ctx context.Context, params SearchCreatorsParams) ([]Creator, error) {
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
