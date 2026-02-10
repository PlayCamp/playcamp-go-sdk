package playcamp

import (
	"context"
	"fmt"
	"net/url"

	"github.com/playcamp/playcamp-go-sdk/internal/httpclient"
)

// creatorBase implements shared creator operations for both client and server APIs.
type creatorBase struct {
	client   *httpclient.Client
	basePath string
}

// CreatorService provides access to creator endpoints for the Client API.
type CreatorService struct{ creatorBase }

func newCreatorService(client *httpclient.Client) *CreatorService {
	return &CreatorService{creatorBase{client: client, basePath: "/v1/client/creators"}}
}

// Get returns a creator by their unique key.
func (s *creatorBase) Get(ctx context.Context, creatorKey string) (*Creator, error) {
	if err := requireNonEmpty("creatorKey", creatorKey); err != nil {
		return nil, err
	}
	var result Creator
	if err := s.client.Get(ctx, s.basePath+"/"+url.PathEscape(creatorKey), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Search searches for creators by keyword.
func (s *creatorBase) Search(ctx context.Context, params SearchCreatorsParams) ([]Creator, error) {
	if err := requireNonEmpty("keyword", params.Keyword); err != nil {
		return nil, err
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
