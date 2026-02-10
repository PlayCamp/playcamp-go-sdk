package playcamp

// Creator represents a PlayCamp creator.
type Creator struct {
	CreatorID   int     `json:"creatorId"`
	CreatorName string  `json:"creatorName"`
	Genre       *string `json:"genre"`
	CreatorKey  string  `json:"creatorKey"`
	Status      string  `json:"status"`
}

// SearchCreatorsParams specifies parameters for searching creators.
type SearchCreatorsParams struct {
	Keyword    string  `json:"keyword"`
	CampaignID *string `json:"campaignId,omitempty"`
	Limit      *int    `json:"limit,omitempty"`
}

// CreatorCoupon represents a creator's coupon code info (Server API only).
type CreatorCoupon struct {
	CouponCode string `json:"code"`
	PackageNo  int    `json:"packageNo"`
	Status     string `json:"status"`
}
