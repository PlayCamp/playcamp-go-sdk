package playcamp

import (
	"context"
	"net/url"

	"github.com/playcamp/playcamp-go-sdk/internal/httpclient"
)

// CreatorServerService provides access to creator endpoints for the Server API.
// Shared methods (Get, Search) are inherited from creatorBase.
type CreatorServerService struct{ creatorBase }

func newCreatorServerService(client *httpclient.Client) *CreatorServerService {
	return &CreatorServerService{creatorBase{client: client, basePath: "/v1/server/creators"}}
}

// GetCoupons returns the coupon codes for a creator. (Server API only)
func (s *CreatorServerService) GetCoupons(ctx context.Context, creatorKey string) ([]CreatorCoupon, error) {
	if err := requireNonEmpty("creatorKey", creatorKey); err != nil {
		return nil, err
	}
	var result []CreatorCoupon
	if err := s.client.Get(ctx, s.basePath+"/"+url.PathEscape(creatorKey)+"/coupons", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}
