package playcamp

// CampaignStatus represents the status of a campaign.
type CampaignStatus string

const (
	CampaignStatusPendingExposure CampaignStatus = "PENDING_EXPOSURE"
	CampaignStatusExposed         CampaignStatus = "EXPOSED"
	CampaignStatusInProgress      CampaignStatus = "IN_PROGRESS"
	CampaignStatusCompleted       CampaignStatus = "COMPLETED"
	CampaignStatusCancelled       CampaignStatus = "CANCELLED"
)

// LocalizedString is a map of locale codes to translated strings.
type LocalizedString map[string]string

// Campaign represents a PlayCamp campaign.
type Campaign struct {
	CampaignID   string          `json:"campaignId"`
	ProjectID    string          `json:"projectId"`
	CampaignName LocalizedString `json:"campaignName"`
	Description  LocalizedString `json:"description"`
	StartDate    *string         `json:"startDate"`
	EndDate      *string         `json:"endDate"`
	Status       CampaignStatus  `json:"status"`
}

// ListCampaignsOptions specifies options for listing campaigns.
type ListCampaignsOptions struct {
	Page  *int
	Limit *int
}

// CouponPackage represents a coupon package within a campaign.
type CouponPackage struct {
	PackageID         int              `json:"packageId"`
	PackageNo         int              `json:"packageNo"`
	ItemName          LocalizedString  `json:"itemName"`
	ItemDescription   LocalizedString  `json:"itemDescription"`
	ItemID            string           `json:"itemId"`
	ItemQuantity      int              `json:"itemQuantity"`
	PerCodeUsageLimit int              `json:"perCodeUsageLimit"`
	CrossCreatorLimit int              `json:"crossCreatorLimit"`
	MaxTotalUsage     int              `json:"maxTotalUsage"`
}
