package playcamp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// testServer creates an httptest.Server and a Client/Server pointing at it.
func setupTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	ts := httptest.NewServer(handler)
	client, err := NewClient("test_key:test_secret",
		WithBaseURL(ts.URL),
		WithMaxRetries(0),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client, ts
}

func setupTestServer(t *testing.T, handler http.HandlerFunc) (*Server, *httptest.Server) {
	t.Helper()
	ts := httptest.NewServer(handler)
	server, err := NewServer("test_key:test_secret",
		WithBaseURL(ts.URL),
		WithMaxRetries(0),
	)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	return server, ts
}

// --- Campaign Service Tests ---

func TestCampaignService_List(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/client/campaigns" {
			t.Errorf("path = %s, want /v1/client/campaigns", r.URL.Path)
		}
		if r.URL.Query().Get("limit") != "5" {
			t.Errorf("limit = %s, want 5", r.URL.Query().Get("limit"))
		}
		assertAuthHeader(t, r)
		writeJSON(w, map[string]any{
			"data": []map[string]any{
				{"campaignId": "c1", "projectId": "p1", "status": "EXPOSED"},
				{"campaignId": "c2", "projectId": "p1", "status": "COMPLETED"},
			},
			"pagination": map[string]any{
				"page": 1, "limit": 5, "total": 2, "totalPages": 1,
			},
		})
	})
	defer ts.Close()

	result, err := client.Campaigns.List(context.Background(), &PaginationOptions{
		Limit: Int(5),
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("len(Data) = %d, want 2", len(result.Data))
	}
	if result.Data[0].CampaignID != "c1" {
		t.Errorf("CampaignID = %q, want %q", result.Data[0].CampaignID, "c1")
	}
	if result.Pagination.Total != 2 {
		t.Errorf("Total = %d, want 2", result.Pagination.Total)
	}
	if result.HasNextPage {
		t.Error("HasNextPage should be false")
	}
}

func TestCampaignService_ListAll(t *testing.T) {
	page := 0
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		page++
		switch page {
		case 1:
			writeJSON(w, map[string]any{
				"data":       []map[string]any{{"campaignId": "c1", "projectId": "p1", "status": "EXPOSED"}},
				"pagination": map[string]any{"page": 1, "limit": 1, "total": 2, "totalPages": 2},
			})
		case 2:
			writeJSON(w, map[string]any{
				"data":       []map[string]any{{"campaignId": "c2", "projectId": "p1", "status": "COMPLETED"}},
				"pagination": map[string]any{"page": 2, "limit": 1, "total": 2, "totalPages": 2},
			})
		}
	})
	defer ts.Close()

	iter := client.Campaigns.ListAll(&PaginationOptions{Limit: Int(1)})
	ctx := context.Background()
	var ids []string
	for iter.Next(ctx) {
		ids = append(ids, iter.Item().CampaignID)
		iter.Advance()
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("len(ids) = %d, want 2", len(ids))
	}
	if ids[0] != "c1" || ids[1] != "c2" {
		t.Errorf("ids = %v, want [c1, c2]", ids)
	}
}

func TestCampaignService_ListAll_NilOptions(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"data":       []map[string]any{{"campaignId": "c1", "projectId": "p1", "status": "EXPOSED"}},
			"pagination": map[string]any{"page": 1, "limit": 20, "total": 1, "totalPages": 1},
		})
	})
	defer ts.Close()

	iter := client.Campaigns.ListAll(nil)
	ctx := context.Background()
	if !iter.Next(ctx) {
		t.Fatal("expected at least one item")
	}
	if iter.Item().CampaignID != "c1" {
		t.Errorf("CampaignID = %q, want c1", iter.Item().CampaignID)
	}
	iter.Advance()
	if iter.Next(ctx) {
		t.Error("expected no more items")
	}
}

func TestCampaignService_Get(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/client/campaigns/camp-123" {
			t.Errorf("path = %s, want /v1/client/campaigns/camp-123", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data": map[string]any{
				"campaignId": "camp-123",
				"projectId":  "proj-1",
				"status":     "IN_PROGRESS",
			},
		})
	})
	defer ts.Close()

	campaign, err := client.Campaigns.Get(context.Background(), "camp-123")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if campaign.CampaignID != "camp-123" {
		t.Errorf("CampaignID = %q, want %q", campaign.CampaignID, "camp-123")
	}
	if campaign.Status != CampaignStatusInProgress {
		t.Errorf("Status = %q, want %q", campaign.Status, CampaignStatusInProgress)
	}
}

func TestCampaignService_Get_EmptyID(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := client.Campaigns.Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error")
	}
	var valErr *InputValidationError
	if !errors.As(err, &valErr) {
		t.Errorf("expected InputValidationError, got %T", err)
	}
}

func TestCampaignService_GetCreators(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/client/campaigns/c1/creators" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data": []map[string]any{
				{"creatorId": 1, "creatorName": "Creator1", "creatorKey": "ABCDE", "status": "ACTIVE"},
			},
		})
	})
	defer ts.Close()

	creators, err := client.Campaigns.GetCreators(context.Background(), "c1")
	if err != nil {
		t.Fatalf("GetCreators: %v", err)
	}
	if len(creators) != 1 {
		t.Fatalf("len = %d, want 1", len(creators))
	}
	if creators[0].CreatorKey != "ABCDE" {
		t.Errorf("CreatorKey = %q, want %q", creators[0].CreatorKey, "ABCDE")
	}
}

func TestCampaignService_GetPackages(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/client/campaigns/c1/packages" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data": []map[string]any{
				{"packageId": 1, "packageNo": 1, "itemName": map[string]string{"en": "Test Item"}, "itemId": "ITEM001", "itemQuantity": 1, "perCodeUsageLimit": 1, "maxTotalUsage": 10000},
			},
		})
	})
	defer ts.Close()

	packages, err := client.Campaigns.GetPackages(context.Background(), "c1")
	if err != nil {
		t.Fatalf("GetPackages: %v", err)
	}
	if len(packages) != 1 {
		t.Fatalf("len = %d, want 1", len(packages))
	}
	if packages[0].ItemID != "ITEM001" {
		t.Errorf("ItemID = %q, want %q", packages[0].ItemID, "ITEM001")
	}
}

// --- Creator Service Tests ---

func TestCreatorService_Get(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/client/creators/ABCDE" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data": map[string]any{
				"creatorId": 42, "creatorName": "TestCreator", "creatorKey": "ABCDE", "status": "ACTIVE",
			},
		})
	})
	defer ts.Close()

	creator, err := client.Creators.Get(context.Background(), "ABCDE")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if creator.CreatorName != "TestCreator" {
		t.Errorf("CreatorName = %q, want %q", creator.CreatorName, "TestCreator")
	}
}

func TestCreatorService_Search(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/client/creators/search" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("keyword") != "game" {
			t.Errorf("keyword = %s, want game", r.URL.Query().Get("keyword"))
		}
		writeJSON(w, map[string]any{
			"data": []map[string]any{
				{"creatorId": 1, "creatorName": "Gamer", "creatorKey": "GAMER", "status": "ACTIVE"},
			},
		})
	})
	defer ts.Close()

	creators, err := client.Creators.Search(context.Background(), SearchCreatorsParams{
		Keyword: "game",
		Limit:   Int(10),
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(creators) != 1 {
		t.Fatalf("len = %d, want 1", len(creators))
	}
}

func TestCreatorService_Search_EmptyKeyword(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := client.Creators.Search(context.Background(), SearchCreatorsParams{Keyword: ""})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

// --- Coupon Service Tests ---

func TestCouponService_Validate(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/client/coupons/validate" {
			t.Errorf("path = %s", r.URL.Path)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["couponCode"] != "TESTCODE" {
			t.Errorf("couponCode = %q, want TESTCODE", body["couponCode"])
		}
		writeJSON(w, map[string]any{
			"data": map[string]any{
				"valid": true, "couponCode": "TESTCODE",
				"itemName": map[string]string{"en": "Item1"},
				"creatorKey": "ABCDE", "campaignId": "c1",
			},
		})
	})
	defer ts.Close()

	result, err := client.Coupons.Validate(context.Background(), ValidateCouponParams{
		CouponCode: "TESTCODE",
	})
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !result.Valid {
		t.Error("expected valid=true")
	}
	if result.CouponCode != "TESTCODE" {
		t.Errorf("CouponCode = %q", result.CouponCode)
	}
}

// --- Sponsor Service Tests ---

func TestSponsorService_Get(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/client/sponsors" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("userId") != "user1" {
			t.Errorf("userId = %s", r.URL.Query().Get("userId"))
		}
		writeJSON(w, map[string]any{
			"data": []map[string]any{
				{"userId": "user1", "campaignId": "c1", "creatorKey": "ABCDE", "isActive": true, "sponsoredAt": "2024-01-01T00:00:00Z"},
			},
		})
	})
	defer ts.Close()

	sponsors, err := client.Sponsors.Get(context.Background(), GetSponsorParams{UserID: "user1"})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(sponsors) != 1 {
		t.Fatalf("len = %d, want 1", len(sponsors))
	}
	if !sponsors[0].IsActive {
		t.Error("expected IsActive=true")
	}
}

// --- Server: Coupon Redeem ---

func TestCouponServerService_Redeem(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/server/coupons/redeem" {
			t.Errorf("path = %s", r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["couponCode"] != "CODE123" {
			t.Errorf("couponCode = %v", body["couponCode"])
		}
		if body["userId"] != "user_abc" {
			t.Errorf("userId = %v", body["userId"])
		}
		writeJSON(w, map[string]any{
			"data": map[string]any{
				"success": true, "usageId": 1, "couponCode": "CODE123",
				"reward": []map[string]any{{"itemName": map[string]string{"en": "item1"}, "itemId": "item1", "itemQuantity": 5}},
				"itemName": map[string]string{"en": "Test Item"},
				"creatorKey": "ABCDE", "campaignId": "c1",
				"redeemedAt": "2024-01-15T10:00:00Z",
			},
		})
	})
	defer ts.Close()

	result, err := server.Coupons.Redeem(context.Background(), RedeemCouponParams{
		CouponCode: "CODE123",
		UserID:     "user_abc",
	})
	if err != nil {
		t.Fatalf("Redeem: %v", err)
	}
	if !result.Success {
		t.Error("expected success=true")
	}
	if len(result.Reward) != 1 {
		t.Fatalf("len(Reward) = %d, want 1", len(result.Reward))
	}
	if result.Reward[0].ItemID != "item1" {
		t.Errorf("ItemID = %q", result.Reward[0].ItemID)
	}
}

func TestCouponServerService_Redeem_Validation(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Coupons.Redeem(context.Background(), RedeemCouponParams{
		CouponCode: "",
		UserID:     "user1",
	})
	if err == nil {
		t.Fatal("expected validation error for empty couponCode")
	}

	_, err = server.Coupons.Redeem(context.Background(), RedeemCouponParams{
		CouponCode: "CODE",
		UserID:     "",
	})
	if err == nil {
		t.Fatal("expected validation error for empty userId")
	}
}

func TestCouponServerService_GetUserHistory(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/server/coupons/user/user1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data": []map[string]any{
				{"id": 1, "userId": "user1", "couponCode": "CODE1", "packageId": 1, "usedAt": "2024-01-01T00:00:00Z", "rewardDelivered": true},
			},
			"pagination": map[string]any{"page": 1, "limit": 20, "total": 1, "totalPages": 1},
		})
	})
	defer ts.Close()

	result, err := server.Coupons.GetUserHistory(context.Background(), "user1", nil)
	if err != nil {
		t.Fatalf("GetUserHistory: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("len(Data) = %d, want 1", len(result.Data))
	}
}

func TestCouponServerService_ListAllUserHistory(t *testing.T) {
	page := 0
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/server/coupons/user/user1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		page++
		switch page {
		case 1:
			writeJSON(w, map[string]any{
				"data": []map[string]any{
					{"id": 1, "userId": "user1", "couponCode": "CODE1", "packageId": 1, "usedAt": "2024-01-01T00:00:00Z", "rewardDelivered": true},
				},
				"pagination": map[string]any{"page": 1, "limit": 1, "total": 2, "totalPages": 2},
			})
		case 2:
			writeJSON(w, map[string]any{
				"data": []map[string]any{
					{"id": 2, "userId": "user1", "couponCode": "CODE2", "packageId": 2, "usedAt": "2024-01-02T00:00:00Z", "rewardDelivered": true},
				},
				"pagination": map[string]any{"page": 2, "limit": 1, "total": 2, "totalPages": 2},
			})
		}
	})
	defer ts.Close()

	iter := server.Coupons.ListAllUserHistory("user1", &PaginationOptions{Limit: Int(1)})
	ctx := context.Background()
	var codes []string
	for iter.Next(ctx) {
		codes = append(codes, iter.Item().CouponCode)
		iter.Advance()
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	if len(codes) != 2 {
		t.Fatalf("len(codes) = %d, want 2", len(codes))
	}
	if codes[0] != "CODE1" || codes[1] != "CODE2" {
		t.Errorf("codes = %v, want [CODE1, CODE2]", codes)
	}
}

// --- Server: Sponsor Tests ---

func TestSponsorServerService_Create(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/server/sponsors" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data": map[string]any{
				"userId": "user1", "campaignId": "c1", "creatorKey": "ABCDE",
				"isActive": true, "sponsoredAt": "2024-01-01T00:00:00Z",
			},
		})
	})
	defer ts.Close()

	sponsor, err := server.Sponsors.Create(context.Background(), CreateSponsorParams{
		UserID:     "user1",
		CreatorKey: "ABCDE",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if sponsor.CreatorKey != "ABCDE" {
		t.Errorf("CreatorKey = %q", sponsor.CreatorKey)
	}
}

func TestSponsorServerService_Update(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/v1/server/sponsors/user/user1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["newCreatorKey"] != "FGHIJ" {
			t.Errorf("newCreatorKey = %v", body["newCreatorKey"])
		}
		writeJSON(w, map[string]any{
			"data": map[string]any{
				"userId": "user1", "campaignId": "c1", "creatorKey": "FGHIJ",
				"isActive": true, "sponsoredAt": "2024-01-01T00:00:00Z",
			},
		})
	})
	defer ts.Close()

	sponsor, err := server.Sponsors.Update(context.Background(), "user1", UpdateSponsorParams{
		NewCreatorKey: "FGHIJ",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if sponsor.CreatorKey != "FGHIJ" {
		t.Errorf("CreatorKey = %q", sponsor.CreatorKey)
	}
}

func TestSponsorServerService_Delete(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/v1/server/sponsors/user/user1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer ts.Close()

	err := server.Sponsors.Delete(context.Background(), "user1", nil)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestSponsorServerService_GetHistory(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/server/sponsors/user/user1/history" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data": []map[string]any{
				{"id": 1, "userId": "user1", "campaignId": "c1", "creatorKey": "ABCDE", "action": "CREATED", "createdAt": "2024-01-01T00:00:00Z"},
			},
			"pagination": map[string]any{"page": 1, "limit": 20, "total": 1, "totalPages": 1},
		})
	})
	defer ts.Close()

	result, err := server.Sponsors.GetHistory(context.Background(), "user1", nil)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("len(Data) = %d, want 1", len(result.Data))
	}
	if result.Data[0].Action != SponsorActionCreated {
		t.Errorf("Action = %q, want CREATED", result.Data[0].Action)
	}
}

func TestSponsorServerService_ListAllHistory(t *testing.T) {
	page := 0
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/server/sponsors/user/user1/history" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("campaignId") != "c1" {
			t.Errorf("campaignId = %s, want c1", r.URL.Query().Get("campaignId"))
		}
		page++
		switch page {
		case 1:
			writeJSON(w, map[string]any{
				"data": []map[string]any{
					{"id": 1, "userId": "user1", "campaignId": "c1", "creatorKey": "ABCDE", "action": "CREATED", "createdAt": "2024-01-01T00:00:00Z"},
				},
				"pagination": map[string]any{"page": 1, "limit": 1, "total": 2, "totalPages": 2},
			})
		case 2:
			writeJSON(w, map[string]any{
				"data": []map[string]any{
					{"id": 2, "userId": "user1", "campaignId": "c1", "creatorKey": "FGHIJ", "action": "CHANGED", "createdAt": "2024-02-01T00:00:00Z"},
				},
				"pagination": map[string]any{"page": 2, "limit": 1, "total": 2, "totalPages": 2},
			})
		}
	})
	defer ts.Close()

	iter := server.Sponsors.ListAllHistory("user1", &GetSponsorHistoryOptions{
		CampaignID: String("c1"),
		Limit:      Int(1),
	})
	ctx := context.Background()
	var actions []SponsorAction
	for iter.Next(ctx) {
		actions = append(actions, iter.Item().Action)
		iter.Advance()
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	if len(actions) != 2 {
		t.Fatalf("len(actions) = %d, want 2", len(actions))
	}
	if actions[0] != SponsorActionCreated || actions[1] != SponsorActionChanged {
		t.Errorf("actions = %v, want [CREATED, CHANGED]", actions)
	}
}

// --- Server: Payment Tests ---

func TestPaymentService_Create(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/server/payments" {
			t.Errorf("path = %s", r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["transactionId"] != "txn_1" {
			t.Errorf("transactionId = %v", body["transactionId"])
		}
		writeJSON(w, map[string]any{
			"data": map[string]any{
				"id": 1, "transactionId": "txn_1", "userId": "user1",
				"productId": "prod1", "amount": 9.99, "currency": "USD",
				"platform": "iOS", "status": "COMPLETED",
				"purchasedAt": "2024-01-15T10:00:00Z", "createdAt": "2024-01-15T10:00:00Z",
			},
		})
	})
	defer ts.Close()

	payment, err := server.Payments.Create(context.Background(), CreatePaymentParams{
		UserID:        "user1",
		TransactionID: "txn_1",
		ProductID:     "prod1",
		Amount:        9.99,
		Currency:      "USD",
		Platform:      PaymentPlatformIOS,
		PurchasedAt:   "2024-01-15T10:00:00Z",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if payment.TransactionID != "txn_1" {
		t.Errorf("TransactionID = %q", payment.TransactionID)
	}
	if payment.Status != PaymentStatusCompleted {
		t.Errorf("Status = %q", payment.Status)
	}
}

func TestPaymentService_Create_Validation(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Payments.Create(context.Background(), CreatePaymentParams{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestPaymentService_Get(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/server/payments/txn_1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data": map[string]any{
				"id": 1, "transactionId": "txn_1", "userId": "user1",
				"productId": "prod1", "amount": 9.99, "currency": "USD",
				"platform": "iOS", "status": "COMPLETED",
				"purchasedAt": "2024-01-15T10:00:00Z", "createdAt": "2024-01-15T10:00:00Z",
			},
		})
	})
	defer ts.Close()

	payment, err := server.Payments.Get(context.Background(), "txn_1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if payment.TransactionID != "txn_1" {
		t.Errorf("TransactionID = %q", payment.TransactionID)
	}
}

func TestPaymentService_Refund(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/server/payments/txn_1/refund" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data": map[string]any{
				"id": 1, "transactionId": "txn_1", "userId": "user1",
				"productId": "prod1", "amount": 9.99, "currency": "USD",
				"platform": "iOS", "status": "REFUNDED",
				"purchasedAt": "2024-01-15T10:00:00Z", "createdAt": "2024-01-15T10:00:00Z",
			},
		})
	})
	defer ts.Close()

	payment, err := server.Payments.Refund(context.Background(), "txn_1", nil)
	if err != nil {
		t.Fatalf("Refund: %v", err)
	}
	if payment.Status != PaymentStatusRefunded {
		t.Errorf("Status = %q, want REFUNDED", payment.Status)
	}
}

func TestPaymentService_ListByUser(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/server/payments/user/user1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data": []map[string]any{
				{"id": 1, "transactionId": "txn_1", "userId": "user1", "productId": "p1", "amount": 9.99, "currency": "USD", "platform": "iOS", "status": "COMPLETED", "purchasedAt": "2024-01-15T10:00:00Z", "createdAt": "2024-01-15T10:00:00Z"},
			},
			"pagination": map[string]any{"page": 1, "limit": 20, "total": 1, "totalPages": 1},
		})
	})
	defer ts.Close()

	result, err := server.Payments.ListByUser(context.Background(), "user1", nil)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("len(Data) = %d, want 1", len(result.Data))
	}
}

func TestPaymentService_ListAllByUser(t *testing.T) {
	page := 0
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/server/payments/user/user1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		page++
		switch page {
		case 1:
			writeJSON(w, map[string]any{
				"data": []map[string]any{
					{"id": 1, "transactionId": "txn_1", "userId": "user1", "productId": "p1", "amount": 9.99, "currency": "USD", "platform": "iOS", "status": "COMPLETED", "purchasedAt": "2024-01-15T10:00:00Z", "createdAt": "2024-01-15T10:00:00Z"},
				},
				"pagination": map[string]any{"page": 1, "limit": 1, "total": 2, "totalPages": 2},
			})
		case 2:
			writeJSON(w, map[string]any{
				"data": []map[string]any{
					{"id": 2, "transactionId": "txn_2", "userId": "user1", "productId": "p2", "amount": 19.99, "currency": "USD", "platform": "iOS", "status": "COMPLETED", "purchasedAt": "2024-02-15T10:00:00Z", "createdAt": "2024-02-15T10:00:00Z"},
				},
				"pagination": map[string]any{"page": 2, "limit": 1, "total": 2, "totalPages": 2},
			})
		}
	})
	defer ts.Close()

	iter := server.Payments.ListAllByUser("user1", &PaginationOptions{Limit: Int(1)})
	ctx := context.Background()
	var txnIDs []string
	for iter.Next(ctx) {
		txnIDs = append(txnIDs, iter.Item().TransactionID)
		iter.Advance()
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	if len(txnIDs) != 2 {
		t.Fatalf("len(txnIDs) = %d, want 2", len(txnIDs))
	}
	if txnIDs[0] != "txn_1" || txnIDs[1] != "txn_2" {
		t.Errorf("txnIDs = %v, want [txn_1, txn_2]", txnIDs)
	}
}

// --- Server: Webhook Tests ---

func TestWebhookService_List(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/server/webhooks" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data": []map[string]any{
				{"id": 1, "projectId": "p1", "eventType": "coupon.redeemed", "url": "https://example.com/hook", "isActive": true, "retryCount": 3, "timeoutMs": 5000, "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"},
			},
		})
	})
	defer ts.Close()

	webhooks, err := server.Webhooks.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(webhooks) != 1 {
		t.Fatalf("len = %d, want 1", len(webhooks))
	}
	if webhooks[0].EventType != WebhookEventCouponRedeemed {
		t.Errorf("EventType = %q", webhooks[0].EventType)
	}
}

func TestWebhookService_Create(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		writeJSON(w, map[string]any{
			"data": map[string]any{
				"id": 1, "projectId": "p1", "eventType": "payment.created",
				"url": "https://example.com/hook", "isActive": true,
				"retryCount": 3, "timeoutMs": 5000,
				"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
				"secret": "whsec_test123",
			},
		})
	})
	defer ts.Close()

	webhook, err := server.Webhooks.Create(context.Background(), CreateWebhookParams{
		EventType: WebhookEventPaymentCreated,
		URL:       "https://example.com/hook",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if webhook.Secret != "whsec_test123" {
		t.Errorf("Secret = %q", webhook.Secret)
	}
}

func TestWebhookService_Delete(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/v1/server/webhooks/1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer ts.Close()

	err := server.Webhooks.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestWebhookService_Delete_InvalidID(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	err := server.Webhooks.Delete(context.Background(), 0)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

// --- HTTP Error Tests ---

func TestHTTPError_404(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"message": "Campaign not found",
			"code":    "NOT_FOUND",
		})
	})
	defer ts.Close()

	_, err := client.Campaigns.Get(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestHTTPError_401(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{"message": "Invalid API key"})
	})
	defer ts.Close()

	_, err := client.Campaigns.Get(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error")
	}
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Errorf("expected AuthError, got %T: %v", err, err)
	}
}

func TestHTTPError_429(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]any{"message": "Rate limit exceeded"})
	})
	defer ts.Close()

	_, err := client.Campaigns.Get(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error")
	}
	var rateLimited *RateLimitError
	if !errors.As(err, &rateLimited) {
		t.Errorf("expected RateLimitError, got %T: %v", err, err)
	}
}

// --- Test Mode ---

func TestTestMode(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not be called on default client")
	})
	ts.Close()

	// Create a new client with test mode
	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("isTest") != "true" {
			t.Error("isTest query param should be set")
		}
		writeJSON(w, map[string]any{
			"data": map[string]any{
				"campaignId": "c1", "projectId": "p1", "status": "EXPOSED",
			},
		})
	}))
	defer ts2.Close()

	testClient, _ := NewClient("key:secret",
		WithBaseURL(ts2.URL),
		WithTestMode(true),
		WithMaxRetries(0),
	)
	_ = client // suppress unused

	_, err := testClient.Campaigns.Get(context.Background(), "c1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
}

// --- Auth Header ---

func TestAuthHeader(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test_key:test_secret" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test_key:test_secret")
		}
		writeJSON(w, map[string]any{
			"data": map[string]any{"campaignId": "c1", "projectId": "p1", "status": "EXPOSED"},
		})
	})
	defer ts.Close()

	_, err := client.Campaigns.Get(context.Background(), "c1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
}

// --- Server: Creator Server Service ---

func TestCreatorServerService_GetCoupons(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/server/creators/ABCDE/coupons" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data": []map[string]any{
				{"code": "CODE1", "packageNo": 1, "status": "ACTIVE"},
				{"code": "CODE2", "packageNo": 2, "status": "USED"},
			},
		})
	})
	defer ts.Close()

	coupons, err := server.Creators.GetCoupons(context.Background(), "ABCDE")
	if err != nil {
		t.Fatalf("GetCoupons: %v", err)
	}
	if len(coupons) != 2 {
		t.Fatalf("len = %d, want 2", len(coupons))
	}
	if coupons[0].CouponCode != "CODE1" {
		t.Errorf("CouponCode = %q", coupons[0].CouponCode)
	}
}

// --- Server: Campaign Server Service ---

func TestCampaignServerService_List(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/server/campaigns" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data":       []map[string]any{{"campaignId": "c1", "projectId": "p1", "status": "EXPOSED"}},
			"pagination": map[string]any{"page": 1, "limit": 20, "total": 1, "totalPages": 1},
		})
	})
	defer ts.Close()

	result, err := server.Campaigns.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("len(Data) = %d, want 1", len(result.Data))
	}
}

func TestCampaignServerService_Get(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/server/campaigns/c1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data": map[string]any{"campaignId": "c1", "projectId": "p1", "status": "COMPLETED"},
		})
	})
	defer ts.Close()

	campaign, err := server.Campaigns.Get(context.Background(), "c1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if campaign.Status != CampaignStatusCompleted {
		t.Errorf("Status = %q", campaign.Status)
	}
}

func TestCampaignServerService_Get_EmptyID(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Campaigns.Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCampaignServerService_GetCreators(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/server/campaigns/c1/creators" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data": []map[string]any{{"creatorId": 1, "creatorName": "C1", "creatorKey": "KEY1", "status": "ACTIVE"}},
		})
	})
	defer ts.Close()

	creators, err := server.Campaigns.GetCreators(context.Background(), "c1")
	if err != nil {
		t.Fatalf("GetCreators: %v", err)
	}
	if len(creators) != 1 {
		t.Errorf("len = %d, want 1", len(creators))
	}
}

func TestCampaignServerService_GetCreators_EmptyID(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Campaigns.GetCreators(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

// --- Server: Coupon Validate ---

func TestCouponServerService_Validate(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/server/coupons/validate" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data": map[string]any{"valid": true, "couponCode": "CODE1", "itemName": map[string]string{"en": "Item"}},
		})
	})
	defer ts.Close()

	result, err := server.Coupons.Validate(context.Background(), ValidateCouponServerParams{
		CouponCode: "CODE1",
		UserID:     "user1",
	})
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !result.Valid {
		t.Error("expected valid=true")
	}
}

func TestCouponServerService_Validate_EmptyCouponCode(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Coupons.Validate(context.Background(), ValidateCouponServerParams{
		CouponCode: "",
		UserID:     "user1",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCouponServerService_Validate_EmptyUserID(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Coupons.Validate(context.Background(), ValidateCouponServerParams{
		CouponCode: "CODE",
		UserID:     "",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

// --- Additional Validation Tests ---

func TestCouponService_Validate_EmptyCouponCode(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := client.Coupons.Validate(context.Background(), ValidateCouponParams{CouponCode: ""})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSponsorService_Get_EmptyUserID(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := client.Sponsors.Get(context.Background(), GetSponsorParams{UserID: ""})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSponsorService_Get_WithCampaignID(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("campaignId") != "c1" {
			t.Errorf("campaignId = %s", r.URL.Query().Get("campaignId"))
		}
		writeJSON(w, map[string]any{
			"data": []map[string]any{{"userId": "user1", "campaignId": "c1", "creatorKey": "ABCDE", "isActive": true, "sponsoredAt": "2024-01-01T00:00:00Z"}},
		})
	})
	defer ts.Close()

	_, err := client.Sponsors.Get(context.Background(), GetSponsorParams{
		UserID:     "user1",
		CampaignID: String("c1"),
	})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
}

func TestSponsorServerService_Create_EmptyUserID(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Sponsors.Create(context.Background(), CreateSponsorParams{
		UserID:     "",
		CreatorKey: "ABCDE",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSponsorServerService_Create_EmptyCreatorKey(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Sponsors.Create(context.Background(), CreateSponsorParams{
		UserID:     "user1",
		CreatorKey: "",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSponsorServerService_GetByUser(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/server/sponsors/user/user1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data": []map[string]any{{"userId": "user1", "campaignId": "c1", "creatorKey": "KEY", "isActive": true, "sponsoredAt": "2024-01-01T00:00:00Z"}},
		})
	})
	defer ts.Close()

	sponsors, err := server.Sponsors.GetByUser(context.Background(), "user1")
	if err != nil {
		t.Fatalf("GetByUser: %v", err)
	}
	if len(sponsors) != 1 {
		t.Errorf("len = %d, want 1", len(sponsors))
	}
}

func TestSponsorServerService_GetByUser_EmptyUserID(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Sponsors.GetByUser(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSponsorServerService_Update_EmptyUserID(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Sponsors.Update(context.Background(), "", UpdateSponsorParams{NewCreatorKey: "KEY"})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSponsorServerService_Update_EmptyCreatorKey(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Sponsors.Update(context.Background(), "user1", UpdateSponsorParams{NewCreatorKey: ""})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSponsorServerService_Delete_EmptyUserID(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	err := server.Sponsors.Delete(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSponsorServerService_Delete_WithCampaignID(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("campaignId") != "c1" {
			t.Errorf("campaignId = %s", r.URL.Query().Get("campaignId"))
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer ts.Close()

	err := server.Sponsors.Delete(context.Background(), "user1", &DeleteSponsorOptions{
		CampaignID: String("c1"),
	})
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestSponsorServerService_GetHistory_EmptyUserID(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Sponsors.GetHistory(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSponsorServerService_GetHistory_WithOptions(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("campaignId") != "c1" {
			t.Errorf("campaignId = %s", r.URL.Query().Get("campaignId"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("page = %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("limit") != "5" {
			t.Errorf("limit = %s", r.URL.Query().Get("limit"))
		}
		writeJSON(w, map[string]any{
			"data":       []map[string]any{},
			"pagination": map[string]any{"page": 2, "limit": 5, "total": 0, "totalPages": 0},
		})
	})
	defer ts.Close()

	_, err := server.Sponsors.GetHistory(context.Background(), "user1", &GetSponsorHistoryOptions{
		CampaignID: String("c1"),
		Page:       Int(2),
		Limit:      Int(5),
	})
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
}

func TestCouponServerService_GetUserHistory_EmptyUserID(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Coupons.GetUserHistory(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCouponServerService_GetUserHistory_WithPagination(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("page = %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("limit") != "5" {
			t.Errorf("limit = %s", r.URL.Query().Get("limit"))
		}
		writeJSON(w, map[string]any{
			"data":       []map[string]any{},
			"pagination": map[string]any{"page": 2, "limit": 5, "total": 0, "totalPages": 0},
		})
	})
	defer ts.Close()

	_, err := server.Coupons.GetUserHistory(context.Background(), "user1", &PaginationOptions{
		Page:  Int(2),
		Limit: Int(5),
	})
	if err != nil {
		t.Fatalf("GetUserHistory: %v", err)
	}
}

// --- Payment Validation Tests ---

func TestPaymentService_Get_EmptyTransactionID(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Payments.Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestPaymentService_Refund_EmptyTransactionID(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Payments.Refund(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestPaymentService_ListByUser_EmptyUserID(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Payments.ListByUser(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestPaymentService_ListByUser_WithPagination(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "1" {
			t.Errorf("page = %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("limit = %s", r.URL.Query().Get("limit"))
		}
		writeJSON(w, map[string]any{
			"data":       []map[string]any{},
			"pagination": map[string]any{"page": 1, "limit": 10, "total": 0, "totalPages": 0},
		})
	})
	defer ts.Close()

	_, err := server.Payments.ListByUser(context.Background(), "user1", &PaginationOptions{
		Page:  Int(1),
		Limit: Int(10),
	})
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
}

func TestPaymentService_Refund_WithOptions(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["isTest"] != true {
			t.Errorf("isTest = %v, want true", body["isTest"])
		}
		writeJSON(w, map[string]any{
			"data": map[string]any{
				"id": 1, "transactionId": "txn_1", "userId": "user1",
				"productId": "prod1", "amount": 9.99, "currency": "USD",
				"platform": "iOS", "status": "REFUNDED",
				"purchasedAt": "2024-01-15T10:00:00Z", "createdAt": "2024-01-15T10:00:00Z",
			},
		})
	})
	defer ts.Close()

	_, err := server.Payments.Refund(context.Background(), "txn_1", &RefundPaymentOptions{
		IsTest: Bool(true),
	})
	if err != nil {
		t.Fatalf("Refund: %v", err)
	}
}

func TestPaymentService_Create_AllValidations(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	tests := []struct {
		name   string
		params CreatePaymentParams
		field  string
	}{
		{"empty userId", CreatePaymentParams{TransactionID: "t", ProductID: "p", Amount: 1, Currency: "USD", Platform: "iOS", PurchasedAt: "2024-01-01"}, "userId"},
		{"empty transactionId", CreatePaymentParams{UserID: "u", ProductID: "p", Amount: 1, Currency: "USD", Platform: "iOS", PurchasedAt: "2024-01-01"}, "transactionId"},
		{"empty productId", CreatePaymentParams{UserID: "u", TransactionID: "t", Amount: 1, Currency: "USD", Platform: "iOS", PurchasedAt: "2024-01-01"}, "productId"},
		{"zero amount", CreatePaymentParams{UserID: "u", TransactionID: "t", ProductID: "p", Amount: 0, Currency: "USD", Platform: "iOS", PurchasedAt: "2024-01-01"}, "amount"},
		{"negative amount", CreatePaymentParams{UserID: "u", TransactionID: "t", ProductID: "p", Amount: -1, Currency: "USD", Platform: "iOS", PurchasedAt: "2024-01-01"}, "amount"},
		{"empty currency", CreatePaymentParams{UserID: "u", TransactionID: "t", ProductID: "p", Amount: 1, Platform: "iOS", PurchasedAt: "2024-01-01"}, "currency"},
		{"empty platform", CreatePaymentParams{UserID: "u", TransactionID: "t", ProductID: "p", Amount: 1, Currency: "USD", PurchasedAt: "2024-01-01"}, "platform"},
		{"empty purchasedAt", CreatePaymentParams{UserID: "u", TransactionID: "t", ProductID: "p", Amount: 1, Currency: "USD", Platform: "iOS"}, "purchasedAt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := server.Payments.Create(context.Background(), tt.params)
			if err == nil {
				t.Fatal("expected validation error")
			}
			var valErr *InputValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("expected InputValidationError, got %T", err)
			}
			if valErr.Field != tt.field {
				t.Errorf("Field = %q, want %q", valErr.Field, tt.field)
			}
		})
	}
}

// --- Webhook Service Additional Tests ---

func TestWebhookService_Update(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/v1/server/webhooks/1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data": map[string]any{
				"id": 1, "projectId": "p1", "eventType": "coupon.redeemed",
				"url": "https://new-url.com/hook", "isActive": true,
				"retryCount": 5, "timeoutMs": 10000,
				"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-02T00:00:00Z",
			},
		})
	})
	defer ts.Close()

	webhook, err := server.Webhooks.Update(context.Background(), 1, UpdateWebhookParams{
		URL: String("https://new-url.com/hook"),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if webhook.URL != "https://new-url.com/hook" {
		t.Errorf("URL = %q", webhook.URL)
	}
}

func TestWebhookService_Update_InvalidID(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Webhooks.Update(context.Background(), 0, UpdateWebhookParams{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestWebhookService_GetLogs(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/server/webhooks/1/logs" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data": []map[string]any{
				{"id": "log-1", "webhookId": 1, "eventType": "coupon.redeemed", "status": "SUCCESS", "attempt": 1, "maxAttempts": 3, "createdAt": "2024-01-01T00:00:00Z"},
			},
		})
	})
	defer ts.Close()

	logs, err := server.Webhooks.GetLogs(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetLogs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("len = %d, want 1", len(logs))
	}
	if logs[0].Status != WebhookStatusSuccess {
		t.Errorf("Status = %q, want SUCCESS", logs[0].Status)
	}
}

func TestWebhookService_GetLogs_InvalidID(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Webhooks.GetLogs(context.Background(), 0)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestWebhookService_Test(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/server/webhooks/1/test" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, map[string]any{
			"data": map[string]any{"success": true, "statusCode": 200},
		})
	})
	defer ts.Close()

	result, err := server.Webhooks.Test(context.Background(), 1)
	if err != nil {
		t.Fatalf("Test: %v", err)
	}
	if !result.Success {
		t.Error("expected success=true")
	}
}

func TestWebhookService_Test_InvalidID(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Webhooks.Test(context.Background(), 0)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestWebhookService_Create_EmptyEventType(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Webhooks.Create(context.Background(), CreateWebhookParams{
		EventType: "",
		URL:       "https://example.com/hook",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestWebhookService_Create_EmptyURL(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Webhooks.Create(context.Background(), CreateWebhookParams{
		EventType: "coupon.redeemed",
		URL:       "",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

// --- Creator Validation Tests ---

func TestCreatorService_Get_EmptyKey(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := client.Creators.Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCreatorServerService_Get_EmptyKey(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Creators.Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCreatorServerService_Search_EmptyKeyword(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Creators.Search(context.Background(), SearchCreatorsParams{Keyword: ""})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCreatorServerService_GetCoupons_EmptyKey(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := server.Creators.GetCoupons(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCreatorService_Search_WithCampaignID(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("campaignId") != "c1" {
			t.Errorf("campaignId = %s", r.URL.Query().Get("campaignId"))
		}
		writeJSON(w, map[string]any{
			"data": []map[string]any{{"creatorId": 1, "creatorName": "C", "creatorKey": "KEY", "status": "ACTIVE"}},
		})
	})
	defer ts.Close()

	_, err := client.Creators.Search(context.Background(), SearchCreatorsParams{
		Keyword:    "test",
		CampaignID: String("c1"),
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
}

// --- Campaign List with Page option ---

func TestCampaignService_List_WithPageOption(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "3" {
			t.Errorf("page = %s", r.URL.Query().Get("page"))
		}
		writeJSON(w, map[string]any{
			"data":       []map[string]any{},
			"pagination": map[string]any{"page": 3, "limit": 20, "total": 0, "totalPages": 0},
		})
	})
	defer ts.Close()

	_, err := client.Campaigns.List(context.Background(), &PaginationOptions{
		Page: Int(3),
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestCampaignService_List_NilOptions(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"data":       []map[string]any{},
			"pagination": map[string]any{"page": 1, "limit": 20, "total": 0, "totalPages": 0},
		})
	})
	defer ts.Close()

	_, err := client.Campaigns.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestCampaignService_GetCreators_EmptyID(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := client.Campaigns.GetCreators(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCampaignService_GetPackages_EmptyID(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been made")
	})
	defer ts.Close()

	_, err := client.Campaigns.GetPackages(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

// --- HTTP Error: ForbiddenError, ConflictError, ValidationError ---

func TestHTTPError_400(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"code":  "VALIDATION_ERROR",
			"error": "Validation Error",
			"details": []map[string]any{
				{"message": "Invalid datetime", "path": "purchasedAt", "target": "body"},
			},
		})
	})
	defer ts.Close()

	_, err := server.Payments.Create(context.Background(), CreatePaymentParams{
		UserID: "u1", TransactionID: "t1", ProductID: "p1",
		Amount: 9.99, Currency: "USD", Platform: "iOS", PurchasedAt: "now",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var badReq *BadRequestError
	if !errors.As(err, &badReq) {
		t.Errorf("expected BadRequestError, got %T: %v", err, err)
	}
	if len(badReq.Details) != 1 {
		t.Fatalf("len(Details) = %d, want 1", len(badReq.Details))
	}
	if badReq.Details[0].Path != "purchasedAt" {
		t.Errorf("Details[0].Path = %q, want purchasedAt", badReq.Details[0].Path)
	}
}

func TestHTTPError_403(t *testing.T) {
	client, ts := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]any{"message": "Forbidden"})
	})
	defer ts.Close()

	_, err := client.Campaigns.Get(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error")
	}
	var forbiddenErr *ForbiddenError
	if !errors.As(err, &forbiddenErr) {
		t.Errorf("expected ForbiddenError, got %T: %v", err, err)
	}
}

func TestHTTPError_409(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]any{"message": "Conflict"})
	})
	defer ts.Close()

	_, err := server.Sponsors.Create(context.Background(), CreateSponsorParams{
		UserID: "u1", CreatorKey: "K1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var conflictErr *ConflictError
	if !errors.As(err, &conflictErr) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestHTTPError_422(t *testing.T) {
	server, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(422)
		json.NewEncoder(w).Encode(map[string]any{"message": "Invalid input"})
	})
	defer ts.Close()

	_, err := server.Payments.Create(context.Background(), CreatePaymentParams{
		UserID: "u1", TransactionID: "t1", ProductID: "p1",
		Amount: 9.99, Currency: "USD", Platform: "iOS", PurchasedAt: "2024-01-01",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

// --- Options Tests ---

func TestOptions(t *testing.T) {
	t.Run("WithBaseURL", func(t *testing.T) {
		client, err := NewClient("key:secret", WithBaseURL("https://custom.api.com"))
		if err != nil {
			t.Fatal(err)
		}
		if client == nil {
			t.Fatal("client is nil")
		}
	})

	t.Run("WithTimeout", func(t *testing.T) {
		client, err := NewClient("key:secret", WithTimeout(10*time.Second))
		if err != nil {
			t.Fatal(err)
		}
		if client == nil {
			t.Fatal("client is nil")
		}
	})

	t.Run("WithHTTPClient", func(t *testing.T) {
		custom := &http.Client{Timeout: 1 * time.Second}
		client, err := NewClient("key:secret", WithHTTPClient(custom))
		if err != nil {
			t.Fatal(err)
		}
		if client == nil {
			t.Fatal("client is nil")
		}
	})

	t.Run("WithDebug", func(t *testing.T) {
		client, err := NewClient("key:secret", WithDebug(DebugOptions{
			Enabled:         true,
			LogRequestBody:  true,
			LogResponseBody: true,
		}))
		if err != nil {
			t.Fatal(err)
		}
		if client == nil {
			t.Fatal("client is nil")
		}
	})

	t.Run("WithDebug custom logger", func(t *testing.T) {
		var logged bool
		client, err := NewClient("key:secret", WithDebug(DebugOptions{
			Enabled: true,
			Logger:  func(format string, args ...any) { logged = true },
		}))
		if err != nil {
			t.Fatal(err)
		}
		if client == nil {
			t.Fatal("client is nil")
		}
		// Logger is set but won't be called until a request is made
		_ = logged
	})
}

// --- Error message formatting ---

func TestAPIError_ErrorString(t *testing.T) {
	t.Run("with code", func(t *testing.T) {
		err := &APIError{StatusCode: 404, Code: "NOT_FOUND", Message: "Not found"}
		got := err.Error()
		want := "playcamp: API error 404 (NOT_FOUND): Not found"
		if got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("without code", func(t *testing.T) {
		err := &APIError{StatusCode: 500, Message: "Internal error"}
		got := err.Error()
		want := "playcamp: API error 500: Internal error"
		if got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})
}

func TestNetworkError_ErrorString(t *testing.T) {
	err := newNetworkError(nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var netErr *NetworkError
	if !errors.As(err, &netErr) {
		t.Fatal("expected NetworkError")
	}
	if netErr.Message != "network request failed" {
		t.Errorf("Message = %q", netErr.Message)
	}
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func assertAuthHeader(t *testing.T, r *http.Request) {
	t.Helper()
	auth := r.Header.Get("Authorization")
	if auth == "" {
		t.Error("missing Authorization header")
	}
}
