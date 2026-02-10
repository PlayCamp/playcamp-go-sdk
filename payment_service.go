package playcamp

import (
	"context"
	"fmt"
	"net/url"

	"github.com/playcamp/playcamp-go-sdk/internal/httpclient"
)

// PaymentService provides access to payment endpoints for the Server API.
type PaymentService struct {
	client   *httpclient.Client
	basePath string
}

func newPaymentService(client *httpclient.Client) *PaymentService {
	return &PaymentService{client: client, basePath: "/v1/server/payments"}
}

// Create registers a new payment.
func (s *PaymentService) Create(ctx context.Context, params CreatePaymentParams) (*Payment, error) {
	if params.UserID == "" {
		return nil, &InputValidationError{Field: "userId", Message: "must be a non-empty string"}
	}
	if params.TransactionID == "" {
		return nil, &InputValidationError{Field: "transactionId", Message: "must be a non-empty string"}
	}
	if params.ProductID == "" {
		return nil, &InputValidationError{Field: "productId", Message: "must be a non-empty string"}
	}
	if params.Amount <= 0 {
		return nil, &InputValidationError{Field: "amount", Message: "must be a positive number"}
	}
	if params.Currency == "" {
		return nil, &InputValidationError{Field: "currency", Message: "must be a non-empty string"}
	}
	if params.Platform == "" {
		return nil, &InputValidationError{Field: "platform", Message: "must be a non-empty string"}
	}
	if params.PurchasedAt == "" {
		return nil, &InputValidationError{Field: "purchasedAt", Message: "must be a non-empty string"}
	}
	var result Payment
	if err := s.client.Post(ctx, s.basePath, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a payment by transaction ID.
func (s *PaymentService) Get(ctx context.Context, transactionID string) (*Payment, error) {
	if transactionID == "" {
		return nil, &InputValidationError{Field: "transactionId", Message: "must be a non-empty string"}
	}
	var result Payment
	if err := s.client.Get(ctx, s.basePath+"/"+transactionID, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListByUser returns a user's payment history.
func (s *PaymentService) ListByUser(ctx context.Context, userID string, opts *PaginationOptions) (*PageResult[Payment], error) {
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
	return getPaginated[Payment](ctx, s.client, s.basePath+"/user/"+userID, query)
}

// Refund refunds a payment.
func (s *PaymentService) Refund(ctx context.Context, transactionID string, opts *RefundPaymentOptions) (*Payment, error) {
	if transactionID == "" {
		return nil, &InputValidationError{Field: "transactionId", Message: "must be a non-empty string"}
	}
	body := opts
	if body == nil {
		body = &RefundPaymentOptions{}
	}
	var result Payment
	if err := s.client.Post(ctx, s.basePath+"/"+transactionID+"/refund", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
