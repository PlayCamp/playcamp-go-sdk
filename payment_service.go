package playcamp

import (
	"context"
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
	if err := requireNonEmpty("userId", params.UserID); err != nil {
		return nil, err
	}
	if err := requireNonEmpty("transactionId", params.TransactionID); err != nil {
		return nil, err
	}
	if err := requireNonEmpty("productId", params.ProductID); err != nil {
		return nil, err
	}
	if params.Amount <= 0 {
		return nil, &InputValidationError{Field: "amount", Message: "must be a positive number"}
	}
	if err := requireNonEmpty("currency", params.Currency); err != nil {
		return nil, err
	}
	if err := requireNonEmpty("platform", string(params.Platform)); err != nil {
		return nil, err
	}
	if params.PurchasedAt.IsZero() {
		return nil, &InputValidationError{Field: "purchasedAt", Message: "must be a non-zero time.Time"}
	}
	var result Payment
	if err := s.client.Post(ctx, s.basePath, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a payment by transaction ID.
func (s *PaymentService) Get(ctx context.Context, transactionID string) (*Payment, error) {
	if err := requireNonEmpty("transactionId", transactionID); err != nil {
		return nil, err
	}
	var result Payment
	if err := s.client.Get(ctx, s.basePath+"/"+url.PathEscape(transactionID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListByUser returns a user's payment history.
func (s *PaymentService) ListByUser(ctx context.Context, userID string, opts *PaginationOptions) (*PageResult[Payment], error) {
	if err := requireNonEmpty("userId", userID); err != nil {
		return nil, err
	}
	return getPaginated[Payment](ctx, s.client, s.basePath+"/user/"+url.PathEscape(userID), paginationQuery(opts))
}

// ListAllByUser returns an iterator that yields all payments for a user across all pages.
func (s *PaymentService) ListAllByUser(userID string, opts *PaginationOptions) *PageIterator[Payment] {
	var limit *int
	if opts != nil {
		limit = opts.Limit
	}
	return NewPageIterator(func(ctx context.Context, page int) (*PageResult[Payment], error) {
		return s.ListByUser(ctx, userID, &PaginationOptions{Page: Int(page), Limit: limit})
	})
}

// CreateBulk registers payments in bulk (up to 1000).
func (s *PaymentService) CreateBulk(ctx context.Context, params CreateBulkPaymentParams) (*BulkPaymentResult, error) {
	if len(params.Payments) == 0 {
		return nil, &InputValidationError{Field: "payments", Message: "must contain at least one payment"}
	}
	var result BulkPaymentResult
	if err := s.client.Post(ctx, s.basePath+"/bulk", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Refund refunds a payment.
func (s *PaymentService) Refund(ctx context.Context, transactionID string, opts *RefundPaymentOptions) (*Payment, error) {
	if err := requireNonEmpty("transactionId", transactionID); err != nil {
		return nil, err
	}
	body := opts
	if body == nil {
		body = &RefundPaymentOptions{}
	}
	var result Payment
	if err := s.client.Post(ctx, s.basePath+"/"+url.PathEscape(transactionID)+"/refund", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
