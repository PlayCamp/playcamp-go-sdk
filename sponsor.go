package playcamp

// SponsorAction represents the type of sponsor action.
type SponsorAction string

const (
	// SponsorActionCreated indicates a new sponsor relationship was created.
	SponsorActionCreated SponsorAction = "CREATED"
	// SponsorActionChanged indicates the sponsor's creator was changed.
	SponsorActionChanged SponsorAction = "CHANGED"
	// SponsorActionEnded indicates the sponsor relationship was ended.
	SponsorActionEnded SponsorAction = "ENDED"
)

// Sponsor represents a sponsor relationship.
type Sponsor struct {
	UserID      string  `json:"userId"`
	CampaignID  string  `json:"campaignId"`
	CreatorKey  string  `json:"creatorKey"`
	IsActive    bool    `json:"isActive"`
	SponsoredAt string  `json:"sponsoredAt"`
	EndedAt     *string `json:"endedAt,omitempty"`
}

// SponsorHistory represents a sponsor history entry.
type SponsorHistory struct {
	ID                 int           `json:"id"`
	UserID             string        `json:"userId"`
	CampaignID         string        `json:"campaignId"`
	CreatorKey         string        `json:"creatorKey"`
	Action             SponsorAction `json:"action"`
	PreviousCreatorKey *string       `json:"previousCreatorKey,omitempty"`
	CreatedAt          string        `json:"createdAt"`
}

// GetSponsorParams specifies parameters for getting sponsor status (Client API).
type GetSponsorParams struct {
	UserID     string  `json:"userId"`
	CampaignID *string `json:"campaignId,omitempty"`
}

// CreateSponsorParams specifies parameters for creating a sponsor (Server API).
type CreateSponsorParams struct {
	UserID     string  `json:"userId"`
	CreatorKey string  `json:"creatorKey"`
	CampaignID *string `json:"campaignId,omitempty"`
	IsTest     *bool   `json:"isTest,omitempty"`
}

// UpdateSponsorParams specifies parameters for updating a sponsor (Server API).
type UpdateSponsorParams struct {
	CampaignID    *string `json:"campaignId,omitempty"`
	NewCreatorKey string  `json:"newCreatorKey"`
	IsTest        *bool   `json:"isTest,omitempty"`
}

// DeleteSponsorOptions specifies options for deleting a sponsor (Server API).
type DeleteSponsorOptions struct {
	CampaignID *string
}

// GetSponsorHistoryOptions specifies options for getting sponsor history (Server API).
type GetSponsorHistoryOptions struct {
	CampaignID *string
	Page       *int
	Limit      *int
}
