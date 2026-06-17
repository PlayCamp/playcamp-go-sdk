package playcamp_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/playcamp/playcamp-go-sdk"
	"github.com/playcamp/playcamp-go-sdk/webhookutil"
)

func ExampleNewClient() {
	client, err := playcamp.NewClient("keyId:secret",
		playcamp.WithEnvironment(playcamp.EnvironmentSandbox),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// List campaigns
	campaigns, err := client.Campaigns.List(ctx, &playcamp.PaginationOptions{
		Limit: playcamp.Int(10),
	})
	if err != nil {
		log.Fatal(err)
	}
	_ = campaigns
}

func ExampleNewServer() {
	server, err := playcamp.NewServer("serverKeyId:secret",
		playcamp.WithEnvironment(playcamp.EnvironmentLive),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Redeem a coupon
	result, err := server.Coupons.Redeem(ctx, playcamp.RedeemCouponParams{
		CouponCode: "CODE123",
		UserID:     "user_abc",
	})
	if err != nil {
		log.Fatal(err)
	}
	_ = result
}

func ExampleNewServer_payment() {
	server, err := playcamp.NewServer("serverKeyId:secret",
		playcamp.WithEnvironment(playcamp.EnvironmentLive),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	payment, err := server.Payments.Create(ctx, playcamp.CreatePaymentParams{
		UserID:        "user_abc",
		TransactionID: "txn_123",
		ProductID:     "prod_xyz",
		Amount:        9.99,
		Currency:      "USD",
		Platform:      playcamp.PaymentPlatformIOS,
		PurchasedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	})
	if err != nil {
		log.Fatal(err)
	}
	_ = payment
}

func ExampleNewServer_playtime() {
	server, err := playcamp.NewServer("serverKeyId:secret",
		playcamp.WithEnvironment(playcamp.EnvironmentLive),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Report a single game session's playtime.
	// campaignId/creatorKey/platform are optional attribution overrides;
	// if omitted the server matches the sponsorship and defaults platform to Other.
	session, err := server.PlaytimeSessions.Create(ctx, playcamp.CreatePlaytimeSessionParams{
		SessionID:       "sess_1",
		UserID:          "user_42",
		DurationSeconds: 1830,
		StartedAt:       time.Date(2026, 6, 15, 7, 33, 37, 0, time.UTC),
		EndedAt:         time.Date(2026, 6, 15, 8, 4, 7, 0, time.UTC),
		Platform:        playcamp.PaymentPlatformIOS,
	})
	if err != nil {
		log.Fatal(err)
	}
	_ = session

	// Report multiple sessions in bulk (up to 1000)
	bulk, err := server.PlaytimeSessions.CreateBulk(ctx, playcamp.CreateBulkPlaytimeSessionParams{
		Sessions: []playcamp.CreatePlaytimeSessionParams{
			{
				SessionID:       "sess_2",
				UserID:          "user_42",
				DurationSeconds: 600,
				StartedAt:       time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC),
				EndedAt:         time.Date(2026, 6, 15, 9, 10, 0, 0, time.UTC),
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	_ = bulk
}

func ExampleNewServer_sponsor() {
	server, err := playcamp.NewServer("serverKeyId:secret",
		playcamp.WithEnvironment(playcamp.EnvironmentLive),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Create a sponsor relationship
	sponsor, err := server.Sponsors.Create(ctx, playcamp.CreateSponsorParams{
		UserID:     "user_abc",
		CreatorKey: "ABCDE",
	})
	if err != nil {
		log.Fatal(err)
	}
	_ = sponsor

	// Update sponsor (change creator)
	updated, err := server.Sponsors.Update(ctx, "user_abc", playcamp.UpdateSponsorParams{
		NewCreatorKey: "FGHIJ",
	})
	if err != nil {
		log.Fatal(err)
	}
	_ = updated

	// Delete sponsor
	err = server.Sponsors.Delete(ctx, "user_abc", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func Example_errorHandling() {
	client, err := playcamp.NewClient("keyId:secret")
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.Campaigns.Get(context.Background(), "invalid-id")
	if err != nil {
		var notFound *playcamp.NotFoundError
		var authErr *playcamp.AuthError
		var rateLimited *playcamp.RateLimitError

		switch {
		case errors.As(err, &notFound):
			fmt.Println("Campaign not found")
		case errors.As(err, &authErr):
			fmt.Println("Authentication failed")
		case errors.As(err, &rateLimited):
			fmt.Println("Rate limit exceeded")
		default:
			fmt.Printf("Error: %v\n", err)
		}
	}
}

func Example_pagination() {
	client, err := playcamp.NewClient("keyId:secret")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Fetch a single page
	page, err := client.Campaigns.List(ctx, &playcamp.PaginationOptions{
		Page:  playcamp.Int(1),
		Limit: playcamp.Int(20),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Page %d/%d, total: %d\n",
		page.Pagination.Page, page.Pagination.TotalPages, page.Pagination.Total)
}

func Example_pageIterator() {
	client, err := playcamp.NewClient("keyId:secret")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Iterate over all campaigns across all pages
	iter := client.Campaigns.ListAll(&playcamp.PaginationOptions{
		Limit: playcamp.Int(10),
	})
	for iter.Next(ctx) {
		campaign := iter.Item()
		fmt.Printf("Campaign: %s (%s)\n", campaign.CampaignID, campaign.Status)
		iter.Advance()
	}
	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}
}

func Example_webhookVerify() {
	payload := []byte(`{"events":[]}`)
	secret := "whsec_test"

	// Create a signature for testing
	sig := webhookutil.ConstructSignature(payload, secret, nil)

	// Verify the signature
	result := webhookutil.Verify(webhookutil.VerifyOptions{
		Payload:   payload,
		Signature: sig,
		Secret:    secret,
	})
	fmt.Printf("Valid: %v\n", result.Valid)
	// Output: Valid: true
}

func Example_pointerHelpers() {
	// Use pointer helpers for optional fields
	opts := &playcamp.PaginationOptions{
		Page:  playcamp.Int(1),
		Limit: playcamp.Int(20),
	}
	_ = opts

	params := playcamp.CreateSponsorParams{
		UserID:     "user_abc",
		CreatorKey: "ABCDE",
		CampaignID: playcamp.String("campaign_123"),
		IsTest:     playcamp.Bool(true),
	}
	_ = params
}
