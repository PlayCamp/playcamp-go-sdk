package playcamp_test

import (
	"context"
	"fmt"
	"log"

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
	campaigns, err := client.Campaigns.List(ctx, &playcamp.ListCampaignsOptions{
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
		PurchasedAt:   "2024-01-15T10:30:00Z",
	})
	if err != nil {
		log.Fatal(err)
	}
	_ = payment
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
