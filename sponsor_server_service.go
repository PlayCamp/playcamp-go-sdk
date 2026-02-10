package playcamp

import (
	"context"
	"fmt"
	"net/url"

	"github.com/playcamp/playcamp-go-sdk/internal/httpclient"
)

// SponsorServerService provides access to sponsor endpoints for the Server API.
type SponsorServerService struct {
	client   *httpclient.Client
	basePath string
}

func newSponsorServerService(client *httpclient.Client) *SponsorServerService {
	return &SponsorServerService{client: client, basePath: "/v1/server/sponsors"}
}

// Create creates a sponsor relationship.
func (s *SponsorServerService) Create(ctx context.Context, params CreateSponsorParams) (*Sponsor, error) {
	if params.UserID == "" {
		return nil, &InputValidationError{Field: "userId", Message: "must be a non-empty string"}
	}
	if params.CreatorKey == "" {
		return nil, &InputValidationError{Field: "creatorKey", Message: "must be a non-empty string"}
	}
	var result Sponsor
	if err := s.client.Post(ctx, s.basePath, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetByUser returns a user's sponsor status.
func (s *SponsorServerService) GetByUser(ctx context.Context, userID string) ([]Sponsor, error) {
	if userID == "" {
		return nil, &InputValidationError{Field: "userId", Message: "must be a non-empty string"}
	}
	var result []Sponsor
	if err := s.client.Get(ctx, s.basePath+"/user/"+userID, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Update updates a sponsor (changes the creator).
func (s *SponsorServerService) Update(ctx context.Context, userID string, params UpdateSponsorParams) (*Sponsor, error) {
	if userID == "" {
		return nil, &InputValidationError{Field: "userId", Message: "must be a non-empty string"}
	}
	if params.NewCreatorKey == "" {
		return nil, &InputValidationError{Field: "newCreatorKey", Message: "must be a non-empty string"}
	}
	var result Sponsor
	if err := s.client.Put(ctx, s.basePath+"/user/"+userID, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete ends a sponsor relationship.
func (s *SponsorServerService) Delete(ctx context.Context, userID string, opts *DeleteSponsorOptions) error {
	if userID == "" {
		return &InputValidationError{Field: "userId", Message: "must be a non-empty string"}
	}
	var query url.Values
	if opts != nil && opts.CampaignID != nil {
		query = url.Values{}
		query.Set("campaignId", *opts.CampaignID)
	}
	return s.client.Delete(ctx, s.basePath+"/user/"+userID, query)
}

// GetHistory returns the sponsor change history for a user.
func (s *SponsorServerService) GetHistory(ctx context.Context, userID string, opts *GetSponsorHistoryOptions) (*PageResult[SponsorHistory], error) {
	if userID == "" {
		return nil, &InputValidationError{Field: "userId", Message: "must be a non-empty string"}
	}
	query := url.Values{}
	if opts != nil {
		if opts.CampaignID != nil {
			query.Set("campaignId", *opts.CampaignID)
		}
		if opts.Page != nil {
			query.Set("page", fmt.Sprintf("%d", *opts.Page))
		}
		if opts.Limit != nil {
			query.Set("limit", fmt.Sprintf("%d", *opts.Limit))
		}
	}
	return getPaginated[SponsorHistory](ctx, s.client, s.basePath+"/user/"+userID+"/history", query)
}
