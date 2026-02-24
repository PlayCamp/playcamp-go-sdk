package playcamp

// WebviewOttParams contains parameters for creating a one-time token.
type WebviewOttParams struct {
	UserID        string                 `json:"userId"`
	CampaignID    string                 `json:"campaignId,omitempty"`
	CodeChallenge string                 `json:"codeChallenge,omitempty"`
	CallbackID    string                 `json:"callbackId,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

// WebviewOttResult contains the response from OTT creation.
type WebviewOttResult struct {
	OTT       string `json:"ott"`
	ExpiresAt string `json:"expiresAt"`
}
