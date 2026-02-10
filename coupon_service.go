package playcamp

import (
	"context"

	"github.com/playcamp/playcamp-go-sdk/internal/httpclient"
)

// CouponService provides access to coupon endpoints for the Client API.
type CouponService struct {
	client   *httpclient.Client
	basePath string
}

func newCouponService(client *httpclient.Client) *CouponService {
	return &CouponService{client: client, basePath: "/v1/client/coupons"}
}

// Validate validates a coupon code.
func (s *CouponService) Validate(ctx context.Context, params ValidateCouponParams) (*CouponValidation, error) {
	if err := requireNonEmpty("couponCode", params.CouponCode); err != nil {
		return nil, err
	}
	var result CouponValidation
	if err := s.client.Post(ctx, s.basePath+"/validate", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
