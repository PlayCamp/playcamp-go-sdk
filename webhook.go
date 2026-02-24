package playcamp

import "encoding/json"

// WebhookEventType represents a webhook event type.
type WebhookEventType string

const (
	// WebhookEventCouponRedeemed is fired when a coupon is redeemed.
	WebhookEventCouponRedeemed WebhookEventType = "coupon.redeemed"
	// WebhookEventPaymentCreated is fired when a payment is created.
	WebhookEventPaymentCreated WebhookEventType = "payment.created"
	// WebhookEventPaymentRefunded is fired when a payment is refunded.
	WebhookEventPaymentRefunded WebhookEventType = "payment.refunded"
	// WebhookEventSponsorCreated is fired when a sponsor relationship is created.
	WebhookEventSponsorCreated WebhookEventType = "sponsor.created"
	// WebhookEventSponsorChanged is fired when a sponsor relationship is changed.
	WebhookEventSponsorChanged WebhookEventType = "sponsor.changed"
	// WebhookEventSponsorEnded is fired when a sponsor relationship is ended.
	WebhookEventSponsorEnded WebhookEventType = "sponsor.ended"
)

// WebhookStatus represents the delivery status of a webhook.
type WebhookStatus string

const (
	// WebhookStatusPending indicates the webhook delivery is queued.
	WebhookStatusPending WebhookStatus = "PENDING"
	// WebhookStatusProcessing indicates the webhook is being delivered.
	WebhookStatusProcessing WebhookStatus = "PROCESSING"
	// WebhookStatusSuccess indicates the webhook was delivered successfully.
	WebhookStatusSuccess WebhookStatus = "SUCCESS"
	// WebhookStatusFailed indicates the webhook delivery failed.
	WebhookStatusFailed WebhookStatus = "FAILED"
	// WebhookStatusRetrying indicates the webhook delivery is being retried.
	WebhookStatusRetrying WebhookStatus = "RETRYING"
)

// Webhook represents a webhook endpoint.
type Webhook struct {
	ID         int              `json:"id"`
	ProjectID  string           `json:"projectId"`
	EventType  WebhookEventType `json:"eventType"`
	URL        string           `json:"url"`
	IsActive   bool             `json:"isActive"`
	RetryCount int              `json:"retryCount"`
	TimeoutMs  int              `json:"timeoutMs"`
	CreatedAt  string           `json:"createdAt"`
	UpdatedAt  string           `json:"updatedAt"`
}

// WebhookWithSecret is a webhook that includes the secret (returned on creation).
type WebhookWithSecret struct {
	Webhook
	Secret string `json:"secret"`
}

// WebhookLog represents a webhook delivery log entry.
type WebhookLog struct {
	ID             string        `json:"id"`
	WebhookID      int           `json:"webhookId"`
	EventType      string        `json:"eventType"`
	Payload        any           `json:"payload"`
	ResponseStatus *int          `json:"responseStatus,omitempty"`
	ResponseBody   *string       `json:"responseBody,omitempty"`
	Attempt        int           `json:"attempt"`
	Status         WebhookStatus `json:"status"`
	CreatedAt      string        `json:"createdAt"`
	CompletedAt    *string       `json:"completedAt,omitempty"`
	NextRetryAt    *string       `json:"nextRetryAt,omitempty"`
	MaxAttempts    int           `json:"maxAttempts"`
}

// CreateWebhookParams specifies parameters for creating a webhook.
type CreateWebhookParams struct {
	EventType  WebhookEventType `json:"eventType"`
	URL        string           `json:"url"`
	RetryCount *int             `json:"retryCount,omitempty"`
	TimeoutMs  *int             `json:"timeoutMs,omitempty"`
}

// UpdateWebhookParams specifies parameters for updating a webhook.
type UpdateWebhookParams struct {
	URL        *string `json:"url,omitempty"`
	IsActive   *bool   `json:"isActive,omitempty"`
	RetryCount *int    `json:"retryCount,omitempty"`
	TimeoutMs  *int    `json:"timeoutMs,omitempty"`
}

// WebhookTestResult represents the result of a webhook test.
type WebhookTestResult struct {
	Success        bool    `json:"success"`
	ResponseStatus *int    `json:"responseStatus,omitempty"`
	ResponseBody   *string `json:"responseBody,omitempty"`
	Error          *string `json:"error,omitempty"`
}

// WebhookPayload is the top-level structure of a webhook delivery.
type WebhookPayload struct {
	Events []WebhookEvent `json:"events"`
}

// WebhookEvent represents a single event in a webhook payload.
type WebhookEvent struct {
	Event      WebhookEventType `json:"event"`
	Timestamp  string           `json:"timestamp"`
	CallbackID string           `json:"callbackId,omitempty"`
	IsTest     *bool            `json:"isTest,omitempty"`
	Data       json.RawMessage  `json:"data"`
}

// CouponRedeemedData is the data payload for a coupon.redeemed event.
type CouponRedeemedData struct {
	CouponCode string          `json:"couponCode"`
	UserID     string          `json:"userId"`
	UsageID    int             `json:"usageId"`
	Reward     json.RawMessage `json:"reward"`
}

// PaymentCreatedData is the data payload for a payment.created event.
type PaymentCreatedData struct {
	TransactionID string  `json:"transactionId"`
	UserID        string  `json:"userId"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	CreatorKey    *string `json:"creatorKey,omitempty"`
	CampaignID    *string `json:"campaignId,omitempty"`
}

// PaymentRefundedData is the data payload for a payment.refunded event.
type PaymentRefundedData struct {
	TransactionID string `json:"transactionId"`
	UserID        string `json:"userId"`
}

// SponsorCreatedData is the data payload for a sponsor.created event.
type SponsorCreatedData struct {
	UserID     string `json:"userId"`
	CampaignID string `json:"campaignId"`
	CreatorKey string `json:"creatorKey"`
}

// SponsorChangedData is the data payload for a sponsor.changed event.
type SponsorChangedData struct {
	UserID        string `json:"userId"`
	CampaignID    string `json:"campaignId"`
	OldCreatorKey string `json:"oldCreatorKey"`
	NewCreatorKey string `json:"newCreatorKey"`
}

// SponsorEndedData is the data payload for a sponsor.ended event.
type SponsorEndedData struct {
	UserID     string `json:"userId"`
	CampaignID string `json:"campaignId"`
	CreatorKey string `json:"creatorKey"`
}
