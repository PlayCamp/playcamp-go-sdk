package playcamp

import (
	"context"
	"fmt"
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
	if params.CouponCode == "" {
		return nil, &InputValidationError{Field: "couponCode", Message: "must be a non-empty string"}
	}
	if params.UserID == "" {
		return nil, &InputValidationError{Field: "userId", Message: "must be a non-empty string"}
	}
	var result CouponValidation
	if err := s.client.Post(ctx, s.basePath+"/validate", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Redeem redeems a coupon code for a user.
func (s *CouponServerService) Redeem(ctx context.Context, params RedeemCouponParams) (*RedeemResult, error) {
	if params.CouponCode == "" {
		return nil, &InputValidationError{Field: "couponCode", Message: "must be a non-empty string"}
	}
	if params.UserID == "" {
		return nil, &InputValidationError{Field: "userId", Message: "must be a non-empty string"}
	}
	var result RedeemResult
	if err := s.client.Post(ctx, s.basePath+"/redeem", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetUserHistory returns a user's coupon usage history.
func (s *CouponServerService) GetUserHistory(ctx context.Context, userID string, opts *PaginationOptions) (*PageResult[CouponUsage], error) {
	if userID == "" {
		return nil, &InputValidationError{Field: "userId", Message: "must be a non-empty string"}
	}
	query := url.Values{}
	if opts != nil {
		if opts.Page != nil {
			query.Set("page", fmt.Sprintf("%d", *opts.Page))
		}
		if opts.Limit != nil {
			query.Set("limit", fmt.Sprintf("%d", *opts.Limit))
		}
	}
	return getPaginated[CouponUsage](ctx, s.client, s.basePath+"/user/"+userID, query)
}
