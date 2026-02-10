package playcamp

// CouponErrorCode represents coupon validation/redemption error codes.
type CouponErrorCode string

const (
	// CouponErrorNotFound indicates the coupon code does not exist.
	CouponErrorNotFound CouponErrorCode = "COUPON_NOT_FOUND"
	// CouponErrorInactive indicates the coupon is no longer active.
	CouponErrorInactive CouponErrorCode = "COUPON_INACTIVE"
	// CouponErrorNotYetValid indicates the coupon is not yet within its validity period.
	CouponErrorNotYetValid CouponErrorCode = "COUPON_NOT_YET_VALID"
	// CouponErrorExpired indicates the coupon has expired.
	CouponErrorExpired CouponErrorCode = "COUPON_EXPIRED"
	// CouponErrorUserCodeLimit indicates the user has reached the per-code usage limit.
	CouponErrorUserCodeLimit CouponErrorCode = "USER_CODE_LIMIT"
	// CouponErrorUserPkgLimit indicates the user has reached the per-package usage limit.
	CouponErrorUserPkgLimit CouponErrorCode = "USER_PACKAGE_LIMIT"
	// CouponErrorTotalUsageLmt indicates the coupon has reached its total usage limit.
	CouponErrorTotalUsageLmt CouponErrorCode = "TOTAL_USAGE_LIMIT"
)

// CouponValidation represents the result of coupon validation.
type CouponValidation struct {
	Valid        bool             `json:"valid"`
	CouponCode   string          `json:"couponCode"`
	ItemName     LocalizedString  `json:"itemName"`
	CreatorKey   string          `json:"creatorKey"`
	CampaignID   string          `json:"campaignId"`
	ErrorCode    *CouponErrorCode `json:"errorCode,omitempty"`
	ErrorMessage *string          `json:"errorMessage,omitempty"`
}

// ValidateCouponParams specifies parameters for validating a coupon (Client API).
type ValidateCouponParams struct {
	CouponCode string `json:"couponCode"`
}

// ValidateCouponServerParams specifies parameters for validating a coupon (Server API).
type ValidateCouponServerParams struct {
	CouponCode string `json:"couponCode"`
	UserID     string `json:"userId"`
	IsTest     *bool  `json:"isTest,omitempty"`
}

// RewardItem represents a reward item in a redemption result.
type RewardItem struct {
	ItemName     LocalizedString `json:"itemName"`
	ItemID       string          `json:"itemId"`
	ItemQuantity int             `json:"itemQuantity"`
}

// RedeemResult represents the result of coupon redemption.
type RedeemResult struct {
	Success      bool             `json:"success"`
	UsageID      int              `json:"usageId"`
	CouponCode   string          `json:"couponCode"`
	Reward       []RewardItem    `json:"reward"`
	ItemName     LocalizedString  `json:"itemName"`
	CreatorKey   string          `json:"creatorKey"`
	CampaignID   string          `json:"campaignId"`
	RedeemedAt   string          `json:"redeemedAt"`
	ErrorCode    *CouponErrorCode `json:"errorCode,omitempty"`
	ErrorMessage *string          `json:"errorMessage,omitempty"`
}

// RedeemCouponParams specifies parameters for redeeming a coupon.
type RedeemCouponParams struct {
	CouponCode   string  `json:"couponCode"`
	UserID       string  `json:"userId"`
	GameUserUUID *string `json:"gameUserUuid,omitempty"`
	IsTest       *bool   `json:"isTest,omitempty"`
}

// CouponUsage represents a coupon usage record.
type CouponUsage struct {
	ID                int     `json:"id"`
	UserID            string  `json:"userId"`
	CouponCode        string  `json:"couponCode"`
	PackageID         int     `json:"packageId"`
	CampaignID        *string `json:"campaignId,omitempty"`
	CreatorKey        *string `json:"creatorKey,omitempty"`
	UsedAt            string  `json:"usedAt"`
	RewardDelivered   bool    `json:"rewardDelivered"`
	RewardDeliveredAt *string `json:"rewardDeliveredAt,omitempty"`
}
