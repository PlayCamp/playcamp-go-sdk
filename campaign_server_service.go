package playcamp

import "github.com/playcamp/playcamp-go-sdk/internal/httpclient"

// CampaignServerService provides access to campaign endpoints for the Server API.
// Shared methods (List, Get, GetCreators) are inherited from campaignBase.
type CampaignServerService struct{ campaignBase }

func newCampaignServerService(client *httpclient.Client) *CampaignServerService {
	return &CampaignServerService{campaignBase{client: client, basePath: "/v1/server/campaigns"}}
}
