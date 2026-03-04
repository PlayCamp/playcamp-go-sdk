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
	if err := requireNonEmpty("userId", params.UserID); err != nil {
		return nil, err
	}
	if err := requireNonEmpty("creatorKey", params.CreatorKey); err != nil {
		return nil, err
	}
	var result Sponsor
	if err := s.client.Post(ctx, s.basePath, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetByUser returns a user's sponsor status.
func (s *SponsorServerService) GetByUser(ctx context.Context, userID string) ([]Sponsor, error) {
	if err := requireNonEmpty("userId", userID); err != nil {
		return nil, err
	}
	var result []Sponsor
	if err := s.client.Get(ctx, s.basePath+"/user/"+url.PathEscape(userID), nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Update updates a sponsor (changes the creator).
func (s *SponsorServerService) Update(ctx context.Context, userID string, params UpdateSponsorParams) (*Sponsor, error) {
	if err := requireNonEmpty("userId", userID); err != nil {
		return nil, err
	}
	if err := requireNonEmpty("newCreatorKey", params.NewCreatorKey); err != nil {
		return nil, err
	}
	var result Sponsor
	if err := s.client.Put(ctx, s.basePath+"/user/"+url.PathEscape(userID), params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete ends a sponsor relationship.
func (s *SponsorServerService) Delete(ctx context.Context, userID string, opts *DeleteSponsorOptions) error {
	if err := requireNonEmpty("userId", userID); err != nil {
		return err
	}
	var query url.Values
	if opts != nil {
		if opts.CampaignID != nil || opts.CallbackID != "" {
			query = url.Values{}
		}
		if opts.CampaignID != nil {
			query.Set("campaignId", *opts.CampaignID)
		}
		if opts.CallbackID != "" {
			query.Set("callbackId", opts.CallbackID)
		}
	}
	return s.client.Delete(ctx, s.basePath+"/user/"+url.PathEscape(userID), query)
}

// GetHistory returns the sponsor change history for a user.
func (s *SponsorServerService) GetHistory(ctx context.Context, userID string, opts *GetSponsorHistoryOptions) (*PageResult[SponsorHistory], error) {
	if err := requireNonEmpty("userId", userID); err != nil {
		return nil, err
	}
	query := paginationQuery(nil)
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
	return getPaginated[SponsorHistory](ctx, s.client, s.basePath+"/user/"+url.PathEscape(userID)+"/history", query)
}

// ListAllHistory returns an iterator that yields all sponsor history for a user across all pages.
func (s *SponsorServerService) ListAllHistory(userID string, opts *GetSponsorHistoryOptions) *PageIterator[SponsorHistory] {
	var limit *int
	var campaignID *string
	if opts != nil {
		limit = opts.Limit
		campaignID = opts.CampaignID
	}
	return NewPageIterator(func(ctx context.Context, page int) (*PageResult[SponsorHistory], error) {
		return s.GetHistory(ctx, userID, &GetSponsorHistoryOptions{
			Page:       Int(page),
			Limit:      limit,
			CampaignID: campaignID,
		})
	})
}
