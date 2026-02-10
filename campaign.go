package playcamp

// CampaignStatus represents the status of a campaign.
type CampaignStatus string

const (
	// CampaignStatusPendingExposure indicates the campaign is awaiting exposure.
	CampaignStatusPendingExposure CampaignStatus = "PENDING_EXPOSURE"
	// CampaignStatusExposed indicates the campaign is exposed and visible.
	CampaignStatusExposed CampaignStatus = "EXPOSED"
	// CampaignStatusInProgress indicates the campaign is actively running.
	CampaignStatusInProgress CampaignStatus = "IN_PROGRESS"
	// CampaignStatusCompleted indicates the campaign has finished.
	CampaignStatusCompleted CampaignStatus = "COMPLETED"
	// CampaignStatusCancelled indicates the campaign was cancelled.
	CampaignStatusCancelled CampaignStatus = "CANCELLED"
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
