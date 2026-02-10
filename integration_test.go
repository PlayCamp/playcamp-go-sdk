package playcamp

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// Integration tests run against the real PlayCamp API.
// Skipped unless environment variables are set.
//
// Usage:
//
//	# Client API (read-only)
//	PLAYCAMP_CLIENT_KEY="keyId:secret" go test -v -run TestIntegration_Client ./...
//
//	# Server API (read/write)
//	PLAYCAMP_SERVER_KEY="keyId:secret" go test -v -run TestIntegration_Server ./...
//
//	# Sandbox environment (recommended for testing)
//	PLAYCAMP_CLIENT_KEY="keyId:secret" PLAYCAMP_ENV=sandbox go test -v -run TestIntegration_Client ./...
//
//	# Run all integration tests
//	PLAYCAMP_CLIENT_KEY="keyId:secret" PLAYCAMP_SERVER_KEY="keyId:secret" go test -v -run TestIntegration ./...

func integrationOpts() []Option {
	var opts []Option
	if baseURL := os.Getenv("PLAYCAMP_BASE_URL"); baseURL != "" {
		opts = append(opts, WithBaseURL(baseURL))
	} else if os.Getenv("PLAYCAMP_ENV") == "sandbox" {
		opts = append(opts, WithEnvironment(EnvironmentSandbox))
	}
	opts = append(opts, WithTimeout(15*time.Second))
	return opts
}

func integrationClient(t *testing.T) *Client {
	t.Helper()
	key := os.Getenv("PLAYCAMP_CLIENT_KEY")
	if key == "" {
		t.Skip("PLAYCAMP_CLIENT_KEY not set, skipping integration test")
	}
	client, err := NewClient(key, integrationOpts()...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

func integrationServer(t *testing.T) *Server {
	t.Helper()
	key := os.Getenv("PLAYCAMP_SERVER_KEY")
	if key == "" {
		t.Skip("PLAYCAMP_SERVER_KEY not set, skipping integration test")
	}
	server, err := NewServer(key, integrationOpts()...)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	return server
}

// --- Client Integration Tests ---

func TestIntegration_Client_Campaigns_List(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	result, err := client.Campaigns.List(ctx, &PaginationOptions{Limit: Int(5)})
	if err != nil {
		t.Fatalf("Campaigns.List: %v", err)
	}
	t.Logf("Campaigns: %d total, page %d/%d", result.Pagination.Total, result.Pagination.Page, result.Pagination.TotalPages)
	for _, c := range result.Data {
		t.Logf("  - %s (%s) status=%s", c.CampaignID, c.CampaignName, c.Status)
	}
}

func TestIntegration_Client_Campaigns_Get(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	// First list to get an ID
	result, err := client.Campaigns.List(ctx, &PaginationOptions{Limit: Int(1)})
	if err != nil {
		t.Fatalf("Campaigns.List: %v", err)
	}
	if len(result.Data) == 0 {
		t.Skip("no campaigns available")
	}

	id := result.Data[0].CampaignID
	campaign, err := client.Campaigns.Get(ctx, id)
	if err != nil {
		t.Fatalf("Campaigns.Get(%q): %v", id, err)
	}
	t.Logf("Campaign: %s status=%s", campaign.CampaignID, campaign.Status)
}

func TestIntegration_Client_Campaigns_GetCreators(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	result, err := client.Campaigns.List(ctx, &PaginationOptions{Limit: Int(1)})
	if err != nil {
		t.Fatalf("Campaigns.List: %v", err)
	}
	if len(result.Data) == 0 {
		t.Skip("no campaigns available")
	}

	id := result.Data[0].CampaignID
	creators, err := client.Campaigns.GetCreators(ctx, id)
	if err != nil {
		t.Fatalf("Campaigns.GetCreators(%q): %v", id, err)
	}
	t.Logf("Creators for campaign %s: %d", id, len(creators))
	for _, c := range creators {
		t.Logf("  - %s (key=%s)", c.CreatorName, c.CreatorKey)
	}
}

func TestIntegration_Client_Campaigns_GetPackages(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	result, err := client.Campaigns.List(ctx, &PaginationOptions{Limit: Int(1)})
	if err != nil {
		t.Fatalf("Campaigns.List: %v", err)
	}
	if len(result.Data) == 0 {
		t.Skip("no campaigns available")
	}

	id := result.Data[0].CampaignID
	packages, err := client.Campaigns.GetPackages(ctx, id)
	if err != nil {
		t.Fatalf("Campaigns.GetPackages(%q): %v", id, err)
	}
	t.Logf("Packages for campaign %s: %d", id, len(packages))
	for _, p := range packages {
		t.Logf("  - id=%d itemId=%s qty=%d", p.PackageID, p.ItemID, p.ItemQuantity)
	}
}

func TestIntegration_Client_Creators_Search(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	keyword := os.Getenv("PLAYCAMP_SEARCH_KEYWORD")
	if keyword == "" {
		keyword = "test"
	}

	creators, err := client.Creators.Search(ctx, SearchCreatorsParams{
		Keyword: keyword,
		Limit:   Int(5),
	})
	if err != nil {
		t.Fatalf("Creators.Search(%q): %v", keyword, err)
	}
	t.Logf("Search %q: %d results", keyword, len(creators))
	for _, c := range creators {
		t.Logf("  - %s (key=%s)", c.CreatorName, c.CreatorKey)
	}
}

func TestIntegration_Client_Creators_Get(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	creatorKey := os.Getenv("PLAYCAMP_CREATOR_KEY")
	if creatorKey == "" {
		t.Skip("PLAYCAMP_CREATOR_KEY not set")
	}

	creator, err := client.Creators.Get(ctx, creatorKey)
	if err != nil {
		t.Fatalf("Creators.Get(%q): %v", creatorKey, err)
	}
	t.Logf("Creator: %s (key=%s, status=%s)", creator.CreatorName, creator.CreatorKey, creator.Status)
}

func TestIntegration_Client_Coupons_Validate(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	couponCode := os.Getenv("PLAYCAMP_COUPON_CODE")
	if couponCode == "" {
		t.Skip("PLAYCAMP_COUPON_CODE not set")
	}

	result, err := client.Coupons.Validate(ctx, ValidateCouponParams{
		CouponCode: couponCode,
	})
	if err != nil {
		t.Fatalf("Coupons.Validate(%q): %v", couponCode, err)
	}
	t.Logf("Coupon %q: valid=%v item=%v creator=%s", couponCode, result.Valid, result.ItemName, result.CreatorKey)
}

// --- Server Integration Tests ---

func TestIntegration_Server_Campaigns_List(t *testing.T) {
	server := integrationServer(t)
	ctx := context.Background()

	result, err := server.Campaigns.List(ctx, &PaginationOptions{Limit: Int(5)})
	if err != nil {
		t.Fatalf("Campaigns.List: %v", err)
	}
	t.Logf("Server Campaigns: %d total", result.Pagination.Total)
}

func TestIntegration_Server_Webhooks_List(t *testing.T) {
	server := integrationServer(t)
	ctx := context.Background()

	webhooks, err := server.Webhooks.List(ctx)
	if err != nil {
		t.Fatalf("Webhooks.List: %v", err)
	}
	t.Logf("Webhooks: %d", len(webhooks))
	for _, wh := range webhooks {
		t.Logf("  - id=%d event=%s url=%s active=%v", wh.ID, wh.EventType, wh.URL, wh.IsActive)
	}
}

// --- Debug Mode Integration Test ---

func TestIntegration_Client_Debug(t *testing.T) {
	key := os.Getenv("PLAYCAMP_CLIENT_KEY")
	if key == "" {
		t.Skip("PLAYCAMP_CLIENT_KEY not set")
	}

	opts := append(integrationOpts(), WithDebug(DebugOptions{
		Enabled:         true,
		LogRequestBody:  true,
		LogResponseBody: true,
	}))

	client, err := NewClient(key, opts...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.Campaigns.List(context.Background(), &PaginationOptions{Limit: Int(1)})
	if err != nil {
		t.Fatalf("Campaigns.List: %v", err)
	}
	t.Log("Debug mode request completed (check log output above)")
}

// --- Error Handling Integration Test ---

func TestIntegration_Client_NotFound(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	_, err := client.Campaigns.Get(ctx, "nonexistent-campaign-id-12345")
	if err == nil {
		t.Fatal("expected error for nonexistent campaign")
	}
	t.Logf("Expected error: %v", err)
}

func TestIntegration_Client_InvalidKey(t *testing.T) {
	opts := append(integrationOpts(), WithMaxRetries(0))
	client, err := NewClient("fake_key:fake_secret", opts...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.Campaigns.List(context.Background(), nil)
	if err == nil {
		t.Fatal("expected auth error with fake key")
	}
	t.Logf("Expected error: %v", err)
}

// --- Full Flow: Client API (data-driven chained test) ---

func TestIntegration_Client_FullFlow(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	// 1. List campaigns
	campaigns, err := client.Campaigns.List(ctx, &PaginationOptions{Limit: Int(10)})
	if err != nil {
		t.Fatalf("Campaigns.List: %v", err)
	}
	if len(campaigns.Data) == 0 {
		t.Fatal("expected at least 1 campaign, got 0")
	}
	campaign := campaigns.Data[0]
	t.Logf("[1] Campaign: id=%s status=%s", campaign.CampaignID, campaign.Status)

	// 2. Get campaign detail
	detail, err := client.Campaigns.Get(ctx, campaign.CampaignID)
	if err != nil {
		t.Fatalf("Campaigns.Get: %v", err)
	}
	if detail.CampaignID != campaign.CampaignID {
		t.Fatalf("CampaignID mismatch: %s != %s", detail.CampaignID, campaign.CampaignID)
	}
	t.Logf("[2] Campaign detail: id=%s status=%s", detail.CampaignID, detail.Status)

	// 3. Get campaign creators
	creators, err := client.Campaigns.GetCreators(ctx, campaign.CampaignID)
	if err != nil {
		t.Fatalf("Campaigns.GetCreators: %v", err)
	}
	if len(creators) == 0 {
		t.Fatal("expected at least 1 creator, got 0")
	}
	t.Logf("[3] Creators: %d found", len(creators))
	for _, c := range creators {
		t.Logf("     - %s (key=%s, status=%s)", c.CreatorName, c.CreatorKey, c.Status)
	}

	// 4. Get creator by key
	creatorKey := creators[0].CreatorKey
	creator, err := client.Creators.Get(ctx, creatorKey)
	if err != nil {
		t.Fatalf("Creators.Get(%q): %v", creatorKey, err)
	}
	if creator.CreatorKey != creatorKey {
		t.Fatalf("CreatorKey mismatch: %s != %s", creator.CreatorKey, creatorKey)
	}
	t.Logf("[4] Creator detail: %s (key=%s)", creator.CreatorName, creator.CreatorKey)

	// 5. Search creators (keyword max 10 chars)
	keyword := "test"
	searchResults, err := client.Creators.Search(ctx, SearchCreatorsParams{
		Keyword: keyword,
		Limit:   Int(5),
	})
	if err != nil {
		t.Fatalf("Creators.Search: %v", err)
	}
	if len(searchResults) == 0 {
		t.Fatal("expected at least 1 search result, got 0")
	}
	t.Logf("[5] Search %q: %d results", keyword, len(searchResults))

	// 6. Get campaign packages
	packages, err := client.Campaigns.GetPackages(ctx, campaign.CampaignID)
	if err != nil {
		t.Fatalf("Campaigns.GetPackages: %v", err)
	}
	t.Logf("[6] Packages: %d found", len(packages))
	for _, p := range packages {
		t.Logf("     - id=%d itemId=%s itemName=%v qty=%d", p.PackageID, p.ItemID, p.ItemName, p.ItemQuantity)
	}

	// 7. Get sponsors (verify empty array returned when no data)
	sponsors, err := client.Sponsors.Get(ctx, GetSponsorParams{
		UserID:     "integration_test_user",
		CampaignID: &campaign.CampaignID,
	})
	if err != nil {
		t.Fatalf("Sponsors.Get: %v", err)
	}
	t.Logf("[7] Sponsors for test user: %d", len(sponsors))
}

// --- Full Flow: Server API (read/write full flow) ---

func TestIntegration_Server_FullFlow(t *testing.T) {
	server := integrationServer(t)
	ctx := context.Background()

	// 1. List server campaigns
	campaigns, err := server.Campaigns.List(ctx, &PaginationOptions{Limit: Int(10)})
	if err != nil {
		t.Fatalf("Campaigns.List: %v", err)
	}
	if len(campaigns.Data) == 0 {
		t.Fatal("expected at least 1 campaign, got 0")
	}
	campaign := campaigns.Data[0]
	t.Logf("[1] Server Campaign: id=%s", campaign.CampaignID)

	// 2. Get server campaign detail
	detail, err := server.Campaigns.Get(ctx, campaign.CampaignID)
	if err != nil {
		t.Fatalf("Campaigns.Get: %v", err)
	}
	t.Logf("[2] Server Campaign detail: id=%s status=%s", detail.CampaignID, detail.Status)

	// 3. Get server campaign creators
	creators, err := server.Campaigns.GetCreators(ctx, campaign.CampaignID)
	if err != nil {
		t.Fatalf("Campaigns.GetCreators: %v", err)
	}
	if len(creators) == 0 {
		t.Fatal("expected at least 1 creator")
	}
	creatorKey := creators[0].CreatorKey
	t.Logf("[3] Server Creators: %d (first=%s)", len(creators), creatorKey)

	// 4. Get server creator by key
	creator, err := server.Creators.Get(ctx, creatorKey)
	if err != nil {
		t.Fatalf("Creators.Get: %v", err)
	}
	t.Logf("[4] Server Creator: %s (key=%s)", creator.CreatorName, creator.CreatorKey)

	// 5. Search server creators
	searchResults, err := server.Creators.Search(ctx, SearchCreatorsParams{
		Keyword: "test",
		Limit:   Int(5),
	})
	if err != nil {
		t.Fatalf("Creators.Search: %v", err)
	}
	t.Logf("[5] Server Creator search: %d results", len(searchResults))

	// 6. Get server creator coupons
	coupons, err := server.Creators.GetCoupons(ctx, creatorKey)
	if err != nil {
		t.Fatalf("Creators.GetCoupons(%s): %v", creatorKey, err)
	}
	t.Logf("[6] Server Creator coupons for %s: %d", creatorKey, len(coupons))
	for _, c := range coupons {
		t.Logf("     - code=%s status=%s", c.CouponCode, c.Status)
	}

	// 7. List webhooks
	webhooks, err := server.Webhooks.List(ctx)
	if err != nil {
		t.Fatalf("Webhooks.List: %v", err)
	}
	t.Logf("[7] Webhooks: %d", len(webhooks))

	// 8. Get webhook logs (if any webhooks exist)
	if len(webhooks) > 0 {
		logs, err := server.Webhooks.GetLogs(ctx, webhooks[0].ID)
		if err != nil {
			t.Fatalf("Webhooks.GetLogs(%d): %v", webhooks[0].ID, err)
		}
		t.Logf("[8] Webhook logs for id=%d: %d entries", webhooks[0].ID, len(logs))
	}
}

// --- Server: Sponsor CRUD Flow ---

func TestIntegration_Server_Sponsor_CRUD(t *testing.T) {
	server := integrationServer(t)
	ctx := context.Background()

	// Verify campaigns and creators exist
	campaigns, err := server.Campaigns.List(ctx, nil)
	if err != nil {
		t.Fatalf("Campaigns.List: %v", err)
	}
	if len(campaigns.Data) == 0 {
		t.Skip("no campaigns")
	}
	campaignID := campaigns.Data[0].CampaignID

	creators, err := server.Campaigns.GetCreators(ctx, campaignID)
	if err != nil {
		t.Fatalf("GetCreators: %v", err)
	}
	if len(creators) < 2 {
		t.Skip("need at least 2 creators for sponsor CRUD test")
	}

	testUserID := "go_sdk_integration_test_user"
	creatorKey1 := creators[0].CreatorKey
	creatorKey2 := creators[1].CreatorKey

	// 1. Create sponsor
	sponsor, err := server.Sponsors.Create(ctx, CreateSponsorParams{
		UserID:     testUserID,
		CampaignID: &campaignID,
		CreatorKey: creatorKey1,
	})
	if err != nil {
		t.Fatalf("[1] Sponsors.Create: %v", err)
	}
	if sponsor.CreatorKey != creatorKey1 {
		t.Errorf("CreatorKey = %q, want %q", sponsor.CreatorKey, creatorKey1)
	}
	t.Logf("[1] Created sponsor: user=%s creator=%s", sponsor.UserID, sponsor.CreatorKey)

	// 2. GetByUser
	sponsors, err := server.Sponsors.GetByUser(ctx, testUserID)
	if err != nil {
		t.Fatalf("[2] Sponsors.GetByUser: %v", err)
	}
	if len(sponsors) == 0 {
		t.Fatal("[2] expected at least 1 sponsor")
	}
	t.Logf("[2] GetByUser: %d sponsors", len(sponsors))

	// 3. Update sponsor (may fail due to 30-day change restriction)
	updated, err := server.Sponsors.Update(ctx, testUserID, UpdateSponsorParams{
		NewCreatorKey: creatorKey2,
		CampaignID:    &campaignID,
	})
	if err != nil {
		t.Logf("[3] Sponsors.Update (expected if 30-day rule): %v", err)
	} else {
		t.Logf("[3] Updated sponsor: creator=%s -> %s", creatorKey1, updated.CreatorKey)
	}

	// 4. GetHistory
	history, err := server.Sponsors.GetHistory(ctx, testUserID, &GetSponsorHistoryOptions{
		CampaignID: &campaignID,
	})
	if err != nil {
		t.Fatalf("[4] Sponsors.GetHistory: %v", err)
	}
	t.Logf("[4] Sponsor history: %d entries", len(history.Data))
	for _, h := range history.Data {
		prev := ""
		if h.PreviousCreatorKey != nil {
			prev = *h.PreviousCreatorKey
		}
		t.Logf("     - action=%s creator=%s prev=%s at=%s", h.Action, h.CreatorKey, prev, h.CreatedAt)
	}

	// 5. Delete sponsor
	err = server.Sponsors.Delete(ctx, testUserID, &DeleteSponsorOptions{
		CampaignID: &campaignID,
	})
	if err != nil {
		t.Fatalf("[5] Sponsors.Delete: %v", err)
	}
	t.Logf("[5] Deleted sponsor for user=%s", testUserID)

	// 6. Verify deletion
	sponsorsAfter, err := server.Sponsors.GetByUser(ctx, testUserID)
	if err != nil {
		t.Fatalf("[6] Sponsors.GetByUser after delete: %v", err)
	}
	activeCount := 0
	for _, s := range sponsorsAfter {
		if s.IsActive && s.CampaignID == campaignID {
			activeCount++
		}
	}
	if activeCount > 0 {
		t.Errorf("[6] expected 0 active sponsors for campaign, got %d", activeCount)
	}
	t.Logf("[6] Verified: no active sponsor for campaign %s", campaignID)
}

// --- Server: Coupon Validate & Redeem Flow ---

func TestIntegration_Server_Coupon_Flow(t *testing.T) {
	server := integrationServer(t)
	ctx := context.Background()

	// Get creator coupon codes
	campaigns, err := server.Campaigns.List(ctx, nil)
	if err != nil {
		t.Fatalf("Campaigns.List: %v", err)
	}
	if len(campaigns.Data) == 0 {
		t.Skip("no campaigns")
	}

	creators, err := server.Campaigns.GetCreators(ctx, campaigns.Data[0].CampaignID)
	if err != nil {
		t.Fatalf("GetCreators: %v", err)
	}
	if len(creators) == 0 {
		t.Skip("no creators")
	}

	coupons, err := server.Creators.GetCoupons(ctx, creators[0].CreatorKey)
	if err != nil {
		t.Fatalf("GetCoupons: %v", err)
	}
	if len(coupons) == 0 {
		t.Skip("no coupons available")
	}

	couponCode := coupons[0].CouponCode
	testUserID := "go_sdk_coupon_test_user"
	t.Logf("Testing with coupon: %s", couponCode)

	// 1. Server Validate
	validation, err := server.Coupons.Validate(ctx, ValidateCouponServerParams{
		CouponCode: couponCode,
		UserID:     testUserID,
	})
	if err != nil {
		t.Fatalf("[1] Coupons.Validate: %v", err)
	}
	errCode, errMsg := "", ""
	if validation.ErrorCode != nil {
		errCode = string(*validation.ErrorCode)
	}
	if validation.ErrorMessage != nil {
		errMsg = *validation.ErrorMessage
	}
	t.Logf("[1] Validate: valid=%v code=%s item=%s errorCode=%s errorMsg=%s",
		validation.Valid, validation.CouponCode, validation.ItemName,
		errCode, errMsg)

	// 2. Redeem (only if valid)
	if validation.Valid {
		result, err := server.Coupons.Redeem(ctx, RedeemCouponParams{
			CouponCode: couponCode,
			UserID:     testUserID,
		})
		if err != nil {
			t.Fatalf("[2] Coupons.Redeem: %v", err)
		}
		t.Logf("[2] Redeem: success=%v usageId=%d code=%s item=%s rewards=%d",
			result.Success, result.UsageID, result.CouponCode, result.ItemName["en"], len(result.Reward))
	} else {
		t.Logf("[2] Skipping Redeem (coupon not valid: %s)", errMsg)
	}

	// 3. User history
	history, err := server.Coupons.GetUserHistory(ctx, testUserID, &PaginationOptions{
		Limit: Int(10),
	})
	if err != nil {
		t.Fatalf("[3] Coupons.GetUserHistory: %v", err)
	}
	t.Logf("[3] Coupon history for %s: %d entries (total=%d)", testUserID, len(history.Data), history.Pagination.Total)
	for _, u := range history.Data {
		t.Logf("     - code=%s usedAt=%s delivered=%v", u.CouponCode, u.UsedAt, u.RewardDelivered)
	}
}

// --- Server: Payment Flow ---

func TestIntegration_Server_Payment_Flow(t *testing.T) {
	server := integrationServer(t)
	ctx := context.Background()

	testUserID := "go_sdk_payment_test_user"
	txnID := fmt.Sprintf("go_sdk_test_txn_%d", time.Now().UnixNano())

	// 1. Create payment
	payment, err := server.Payments.Create(ctx, CreatePaymentParams{
		UserID:        testUserID,
		TransactionID: txnID,
		ProductID:     "test_product_001",
		ProductName:   String("Test Product"),
		Amount:        9.99,
		Currency:      "USD",
		Platform:      PaymentPlatformWeb,
		PurchasedAt:   time.Now().UTC().Format("2006-01-02T15:04:05Z"),
	})
	if err != nil {
		t.Fatalf("[1] Payments.Create: %v", err)
	}
	t.Logf("[1] Created payment: txn=%s status=%s amount=%.2f %s", payment.TransactionID, payment.Status, payment.Amount, payment.Currency)

	// 2. Get payment
	fetched, err := server.Payments.Get(ctx, txnID)
	if err != nil {
		t.Fatalf("[2] Payments.Get: %v", err)
	}
	if fetched.TransactionID != txnID {
		t.Fatalf("[2] TransactionID mismatch: %s != %s", fetched.TransactionID, txnID)
	}
	t.Logf("[2] Fetched payment: txn=%s status=%s", fetched.TransactionID, fetched.Status)

	// 3. List by user
	userPayments, err := server.Payments.ListByUser(ctx, testUserID, &PaginationOptions{Limit: Int(10)})
	if err != nil {
		t.Fatalf("[3] Payments.ListByUser: %v", err)
	}
	if len(userPayments.Data) == 0 {
		t.Fatal("[3] expected at least 1 payment")
	}
	t.Logf("[3] User payments: %d (total=%d)", len(userPayments.Data), userPayments.Pagination.Total)

	// 4. Refund
	refunded, err := server.Payments.Refund(ctx, txnID, nil)
	if err != nil {
		t.Fatalf("[4] Payments.Refund: %v", err)
	}
	if refunded.Status != PaymentStatusRefunded {
		t.Errorf("[4] Status = %q, want REFUNDED", refunded.Status)
	}
	t.Logf("[4] Refunded: txn=%s status=%s", refunded.TransactionID, refunded.Status)
}
