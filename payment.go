package playcamp

// PaymentPlatform represents the payment platform.
type PaymentPlatform string

const (
	// PaymentPlatformIOS represents the iOS App Store platform.
	PaymentPlatformIOS PaymentPlatform = "iOS"
	// PaymentPlatformAndroid represents the Android Google Play platform.
	PaymentPlatformAndroid PaymentPlatform = "Android"
	// PaymentPlatformWeb represents a web-based payment platform.
	PaymentPlatformWeb PaymentPlatform = "Web"
	// PaymentPlatformRoblox represents the Roblox platform.
	PaymentPlatformRoblox PaymentPlatform = "Roblox"
	// PaymentPlatformOther represents other payment platforms.
	PaymentPlatformOther PaymentPlatform = "Other"
)

// DistributionType represents the distribution type.
type DistributionType string

const (
	// DistributionMobileStore represents distribution via a mobile app store.
	DistributionMobileStore DistributionType = "MOBILE_STORE"
	// DistributionMobileSelfStore represents self-distributed mobile apps.
	DistributionMobileSelfStore DistributionType = "MOBILE_SELF_STORE"
	// DistributionPCStore represents distribution via a PC store.
	DistributionPCStore DistributionType = "PC_STORE"
	// DistributionPCSelfStore represents self-distributed PC apps.
	DistributionPCSelfStore DistributionType = "PC_SELF_STORE"
)

// PaymentStatus represents the status of a payment.
type PaymentStatus string

const (
	// PaymentStatusCompleted indicates the payment was successfully completed.
	PaymentStatusCompleted PaymentStatus = "COMPLETED"
	// PaymentStatusRefunded indicates the payment was refunded.
	PaymentStatusRefunded PaymentStatus = "REFUNDED"
	// PaymentStatusPending indicates the payment is pending processing.
	PaymentStatusPending PaymentStatus = "PENDING"
	// PaymentStatusCancelled indicates the payment was cancelled.
	PaymentStatusCancelled PaymentStatus = "CANCELLED"
)

// Payment represents a payment record.
type Payment struct {
	ID                int               `json:"id"`
	TransactionID     string            `json:"transactionId"`
	UserID            string            `json:"userId"`
	ProductID         string            `json:"productId"`
	ProductName       *string           `json:"productName,omitempty"`
	Amount            float64           `json:"amount"`
	Currency          string            `json:"currency"`
	AmountUSD         *float64          `json:"amountUsd,omitempty"`
	ExchangeRateToUSD *string           `json:"exchangeRateToUsd,omitempty"`
	ExchangeRateDate  *string           `json:"exchangeRateDate,omitempty"`
	Platform          PaymentPlatform   `json:"platform"`
	DistributionType  *DistributionType `json:"distributionType,omitempty"`
	Receipt           *string           `json:"receipt,omitempty"`
	Status            PaymentStatus     `json:"status"`
	CampaignID        *string           `json:"campaignId,omitempty"`
	CreatorKey        *string           `json:"creatorKey,omitempty"`
	PurchasedAt       string            `json:"purchasedAt"`
	CreatedAt         string            `json:"createdAt"`
}

// CreatePaymentParams specifies parameters for creating a payment.
type CreatePaymentParams struct {
	UserID           string            `json:"userId"`
	TransactionID    string            `json:"transactionId"`
	ProductID        string            `json:"productId"`
	ProductName      *string           `json:"productName,omitempty"`
	Amount           float64           `json:"amount"`
	Currency         string            `json:"currency"`
	Platform         PaymentPlatform   `json:"platform"`
	DistributionType *DistributionType `json:"distributionType,omitempty"`
	PurchasedAt      string            `json:"purchasedAt"`
	Receipt          *string           `json:"receipt,omitempty"`
	CampaignID       *string           `json:"campaignId,omitempty"`
	CreatorKey       *string           `json:"creatorKey,omitempty"`
	IsTest           *bool             `json:"isTest,omitempty"`
}

// RefundPaymentOptions specifies options for refunding a payment.
type RefundPaymentOptions struct {
	IsTest *bool `json:"isTest,omitempty"`
}
