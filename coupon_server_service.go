package playcamp

import (
	"context"
	"net/url"

	"github.com/playcamp/playcamp-go-sdk/internal/httpclient"
)

// CouponServerService provides access to coupon endpoints for the Server API.
type CouponServerService struct {
	client   *httpclient.Client
	basePath string
}

func newCouponServerService(client *httpclient.Client) *CouponServerService {
	return &CouponServerService{client: client, basePath: "/v1/server/coupons"}
}

// Validate validates a coupon code with user context.
func (s *CouponServerService) Validate(ctx context.Context, params ValidateCouponServerParams) (*CouponValidation, error) {
	if err := requireNonEmpty("couponCode", params.CouponCode); err != nil {
		return nil, err
	}
	if err := requireNonEmpty("userId", params.UserID); err != nil {
		return nil, err
	}
	var result CouponValidation
	if err := s.client.Post(ctx, s.basePath+"/validate", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Redeem redeems a coupon code for a user.
func (s *CouponServerService) Redeem(ctx context.Context, params RedeemCouponParams) (*RedeemResult, error) {
	if err := requireNonEmpty("couponCode", params.CouponCode); err != nil {
		return nil, err
	}
	if err := requireNonEmpty("userId", params.UserID); err != nil {
		return nil, err
	}
	var result RedeemResult
	if err := s.client.Post(ctx, s.basePath+"/redeem", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetUserHistory returns a user's coupon usage history.
func (s *CouponServerService) GetUserHistory(ctx context.Context, userID string, opts *PaginationOptions) (*PageResult[CouponUsage], error) {
	if err := requireNonEmpty("userId", userID); err != nil {
		return nil, err
	}
	return getPaginated[CouponUsage](ctx, s.client, s.basePath+"/user/"+url.PathEscape(userID), paginationQuery(opts))
}

// ListAllUserHistory returns an iterator that yields all coupon usage history for a user across all pages.
func (s *CouponServerService) ListAllUserHistory(userID string, opts *PaginationOptions) *PageIterator[CouponUsage] {
	var limit *int
	if opts != nil {
		limit = opts.Limit
	}
	return NewPageIterator(func(ctx context.Context, page int) (*PageResult[CouponUsage], error) {
		return s.GetUserHistory(ctx, userID, &PaginationOptions{Page: Int(page), Limit: limit})
	})
}
