# PlayCamp Go SDK

Official Go client library for the PlayCamp SDK API.

## Installation

```bash
go get github.com/playcamp/playcamp-go-sdk
```

Requires Go 1.21 or later.

## Quick Start

### Client (Read-only)

Use a CLIENT API key to query campaigns, creators, coupons, and check sponsor status.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/playcamp/playcamp-go-sdk"
)

func main() {
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
    fmt.Printf("Found %d campaigns\n", len(campaigns.Data))

    // Get a creator
    creator, err := client.Creators.Get(ctx, "ABCDE")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(creator.CreatorName)
}
```

### Server (Read/Write)

Use a SERVER API key to perform all operations including coupon redemption, sponsor management, payment recording, and webhook configuration.

```go
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

// Create a payment
payment, err := server.Payments.Create(ctx, playcamp.CreatePaymentParams{
    UserID:        "user_abc",
    TransactionID: "txn_123",
    ProductID:     "prod_xyz",
    Amount:        9.99,
    Currency:      "USD",
    Platform:      playcamp.PaymentPlatformIOS,
    PurchasedAt:   "2024-01-15T10:30:00Z",
})

// Create a sponsor
sponsor, err := server.Sponsors.Create(ctx, playcamp.CreateSponsorParams{
    UserID:     "user_abc",
    CreatorKey: "ABCDE",
})
```

## Configuration Options

```go
client, err := playcamp.NewClient("keyId:secret",
    playcamp.WithEnvironment(playcamp.EnvironmentSandbox), // sandbox or live (default: live)
    playcamp.WithBaseURL("https://custom-api.example.com"), // custom URL (overrides environment)
    playcamp.WithTimeout(10 * time.Second),                 // request timeout (default: 30s)
    playcamp.WithTestMode(true),                            // test mode (default: false)
    playcamp.WithMaxRetries(5),                             // max retry attempts (default: 3)
    playcamp.WithHTTPClient(customHTTPClient),               // custom http.Client
    playcamp.WithDebug(playcamp.DebugOptions{               // debug logging
        Enabled:         true,
        LogRequestBody:  true,
        LogResponseBody: true,
    }),
)
```

## API Reference

### Client Services

| Service | Methods | Description |
|---------|---------|-------------|
| `Campaigns` | `List`, `ListAll`, `Get`, `GetCreators`, `GetPackages` | Campaign queries |
| `Creators` | `Get`, `Search` | Creator lookup/search |
| `Coupons` | `Validate` | Coupon validation |
| `Sponsors` | `Get` | Sponsor status lookup |

### Server Services

| Service | Methods | Description |
|---------|---------|-------------|
| `Campaigns` | `List`, `ListAll`, `Get`, `GetCreators` | Campaign queries |
| `Creators` | `Get`, `Search`, `GetCoupons` | Creator lookup/search, coupon code retrieval |
| `Coupons` | `Validate`, `Redeem`, `GetUserHistory`, `ListAllUserHistory` | Coupon validation/redemption/history |
| `Sponsors` | `Create`, `GetByUser`, `Update`, `Delete`, `GetHistory`, `ListAllHistory` | Sponsor management |
| `Payments` | `Create`, `Get`, `ListByUser`, `ListAllByUser`, `Refund` | Payment management |
| `Webhooks` | `List`, `Create`, `Update`, `Delete`, `GetLogs`, `Test` | Webhook management |

## Error Handling

The SDK provides structured error types. Use `errors.As` to handle specific errors.

```go
import "errors"

result, err := client.Campaigns.Get(ctx, "invalid-id")
if err != nil {
    var notFound *playcamp.NotFoundError
    var authErr *playcamp.AuthError
    var rateLimited *playcamp.RateLimitError
    var validationErr *playcamp.ValidationError

    switch {
    case errors.As(err, &notFound):
        fmt.Println("Campaign not found")
    case errors.As(err, &authErr):
        fmt.Println("Authentication failed")
    case errors.As(err, &rateLimited):
        fmt.Println("Rate limit exceeded")
    case errors.As(err, &validationErr):
        fmt.Println("Invalid request")
    default:
        fmt.Printf("Error: %v\n", err)
    }
}
```

**Error Types:**

| Type | HTTP Status Code | Description |
|------|-----------------|-------------|
| `BadRequestError` | 400 | Bad request (includes field-level `Details`) |
| `AuthError` | 401 | Authentication failure |
| `ForbiddenError` | 403 | Insufficient permissions |
| `NotFoundError` | 404 | Resource not found |
| `ConflictError` | 409 | Conflict |
| `ValidationError` | 422 | Server-side validation failure |
| `RateLimitError` | 429 | Rate limit exceeded |
| `NetworkError` | - | Network/timeout error |
| `InputValidationError` | - | Client-side parameter validation failure |

## Webhook Verification

Use the `webhookutil` package to verify webhook signatures.

```go
import "github.com/playcamp/playcamp-go-sdk/webhookutil"

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)

    result := webhookutil.Verify(webhookutil.VerifyOptions{
        Payload:   body,
        Signature: r.Header.Get("X-Webhook-Signature"),
        Secret:    "whsec_your_webhook_secret",
        Tolerance: 300, // seconds (default: 300 = 5 minutes)
    })

    if !result.Valid {
        http.Error(w, result.Error, http.StatusUnauthorized)
        return
    }

    // Handle events
    for _, event := range result.Payload.Events {
        switch event.Event {
        case playcamp.WebhookEventCouponRedeemed:
            var data playcamp.CouponRedeemedData
            json.Unmarshal(event.Data, &data)
            fmt.Printf("Coupon redeemed: %s\n", data.CouponCode)
        case playcamp.WebhookEventPaymentCreated:
            var data playcamp.PaymentCreatedData
            json.Unmarshal(event.Data, &data)
            fmt.Printf("Payment created: %s\n", data.TransactionID)
        }
    }

    w.WriteHeader(http.StatusOK)
}
```

### Constructing Test Signatures

```go
payload := []byte(`{"events":[...]}`)
secret := "whsec_test"

// Simple signature
sig := webhookutil.ConstructSignature(payload, secret, nil)

// Timestamped signature
sig := webhookutil.ConstructSignature(payload, secret, &webhookutil.SignatureOptions{
    Timestamped: true,
})
```

## Pagination

Paginated endpoints return `PageResult[T]`.

```go
// Fetch the first page
page, err := client.Campaigns.List(ctx, &playcamp.PaginationOptions{
    Page:  playcamp.Int(1),
    Limit: playcamp.Int(20),
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Fetched %d of %d total\n", len(page.Data), page.Pagination.Total)
fmt.Printf("Has next page: %v\n", page.HasNextPage)
```

## Pointer Helpers

Use pointer helper functions to set optional fields.

```go
playcamp.Int(10)          // *int
playcamp.String("value")  // *string
playcamp.Bool(true)       // *bool
playcamp.Float64(9.99)    // *float64
```

## Environment

| Environment | URL |
|-------------|-----|
| `EnvironmentSandbox` | `https://sandbox-sdk-api.playcamp.io` |
| `EnvironmentLive` | `https://sdk-api.playcamp.io` |

## License

Copyright PlayCamp. All rights reserved.
