# PlayCamp Go SDK

PlayCamp SDK API의 공식 Go 클라이언트 라이브러리입니다.

## 설치

```bash
go get github.com/playcamp/playcamp-go-sdk
```

Go 1.21 이상이 필요합니다.

## 빠른 시작

### Client (읽기 전용)

CLIENT API 키를 사용하여 캠페인, 크리에이터, 쿠폰 조회 및 스폰서 상태를 확인합니다.

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

    // 캠페인 목록 조회
    campaigns, err := client.Campaigns.List(ctx, &playcamp.PaginationOptions{
        Limit: playcamp.Int(10),
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("캠페인 %d개 조회\n", len(campaigns.Data))

    // 크리에이터 조회
    creator, err := client.Creators.Get(ctx, "ABCDE")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(creator.CreatorName)
}
```

### Server (읽기/쓰기)

SERVER API 키를 사용하여 쿠폰 사용, 스폰서 관리, 결제 기록, 웹훅 설정 등 모든 작업을 수행합니다.

```go
server, err := playcamp.NewServer("serverKeyId:secret",
    playcamp.WithEnvironment(playcamp.EnvironmentLive),
)
if err != nil {
    log.Fatal(err)
}

ctx := context.Background()

// 쿠폰 사용
result, err := server.Coupons.Redeem(ctx, playcamp.RedeemCouponParams{
    CouponCode: "CODE123",
    UserID:     "user_abc",
})

// 결제 등록
payment, err := server.Payments.Create(ctx, playcamp.CreatePaymentParams{
    UserID:        "user_abc",
    TransactionID: "txn_123",
    ProductID:     "prod_xyz",
    Amount:        9.99,
    Currency:      "USD",
    Platform:      playcamp.PaymentPlatformIOS,
    PurchasedAt:   "2024-01-15T10:30:00Z",
})

// 스폰서 생성
sponsor, err := server.Sponsors.Create(ctx, playcamp.CreateSponsorParams{
    UserID:     "user_abc",
    CreatorKey: "ABCDE",
})
```

## 설정 옵션

```go
client, err := playcamp.NewClient("keyId:secret",
    playcamp.WithEnvironment(playcamp.EnvironmentSandbox), // sandbox 또는 live (기본값: live)
    playcamp.WithBaseURL("https://custom-api.example.com"), // 커스텀 URL (environment 오버라이드)
    playcamp.WithTimeout(10 * time.Second),                 // 요청 타임아웃 (기본값: 30초)
    playcamp.WithTestMode(true),                            // 테스트 모드 (기본값: false)
    playcamp.WithMaxRetries(5),                             // 최대 재시도 횟수 (기본값: 3)
    playcamp.WithHTTPClient(customHTTPClient),               // 커스텀 http.Client
    playcamp.WithDebug(playcamp.DebugOptions{               // 디버그 로깅
        Enabled:         true,
        LogRequestBody:  true,
        LogResponseBody: true,
    }),
)
```

## API 레퍼런스

### Client 서비스

| 서비스 | 메서드 | 설명 |
|--------|--------|------|
| `Campaigns` | `List`, `Get`, `GetCreators`, `GetPackages` | 캠페인 조회 |
| `Creators` | `Get`, `Search` | 크리에이터 조회/검색 |
| `Coupons` | `Validate` | 쿠폰 유효성 검증 |
| `Sponsors` | `Get` | 스폰서 상태 조회 |

### Server 서비스

| 서비스 | 메서드 | 설명 |
|--------|--------|------|
| `Campaigns` | `List`, `Get`, `GetCreators` | 캠페인 조회 |
| `Creators` | `Get`, `Search`, `GetCoupons` | 크리에이터 조회/검색, 쿠폰 코드 조회 |
| `Coupons` | `Validate`, `Redeem`, `GetUserHistory` | 쿠폰 검증/사용/이력 조회 |
| `Sponsors` | `Create`, `GetByUser`, `Update`, `Delete`, `GetHistory` | 스폰서 관리 |
| `Payments` | `Create`, `Get`, `ListByUser`, `Refund` | 결제 관리 |
| `Webhooks` | `List`, `Create`, `Update`, `Delete`, `GetLogs`, `Test` | 웹훅 관리 |

## 에러 처리

SDK는 구조화된 에러 타입을 제공하며, `errors.As`를 사용하여 특정 에러를 처리할 수 있습니다.

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
        fmt.Println("캠페인을 찾을 수 없습니다")
    case errors.As(err, &authErr):
        fmt.Println("인증에 실패했습니다")
    case errors.As(err, &rateLimited):
        fmt.Println("요청 제한에 도달했습니다")
    case errors.As(err, &validationErr):
        fmt.Println("유효하지 않은 요청입니다")
    default:
        fmt.Printf("에러: %v\n", err)
    }
}
```

**에러 타입:**

| 타입 | HTTP 상태 코드 | 설명 |
|------|---------------|------|
| `AuthError` | 401 | 인증 실패 |
| `ForbiddenError` | 403 | 권한 없음 |
| `NotFoundError` | 404 | 리소스를 찾을 수 없음 |
| `ConflictError` | 409 | 충돌 |
| `ValidationError` | 422 | 서버측 유효성 검증 실패 |
| `RateLimitError` | 429 | 요청 제한 초과 |
| `NetworkError` | - | 네트워크/타임아웃 에러 |
| `InputValidationError` | - | 클라이언트측 파라미터 검증 실패 |

## 웹훅 검증

`webhookutil` 패키지를 사용하여 웹훅 서명을 검증합니다.

```go
import "github.com/playcamp/playcamp-go-sdk/webhookutil"

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)

    result := webhookutil.Verify(webhookutil.VerifyOptions{
        Payload:   body,
        Signature: r.Header.Get("X-Webhook-Signature"),
        Secret:    "whsec_your_webhook_secret",
        Tolerance: 300, // 초 단위 (기본값: 300 = 5분)
    })

    if !result.Valid {
        http.Error(w, result.Error, http.StatusUnauthorized)
        return
    }

    // 이벤트 처리
    for _, event := range result.Payload.Events {
        switch event.Event {
        case playcamp.WebhookEventCouponRedeemed:
            var data playcamp.CouponRedeemedData
            json.Unmarshal(event.Data, &data)
            fmt.Printf("쿠폰 사용: %s\n", data.CouponCode)
        case playcamp.WebhookEventPaymentCreated:
            var data playcamp.PaymentCreatedData
            json.Unmarshal(event.Data, &data)
            fmt.Printf("결제 생성: %s\n", data.TransactionID)
        }
    }

    w.WriteHeader(http.StatusOK)
}
```

### 테스트용 서명 생성

```go
payload := []byte(`{"events":[...]}`)
secret := "whsec_test"

// 단순 서명
sig := webhookutil.ConstructSignature(payload, secret, nil)

// 타임스탬프 포함 서명
sig := webhookutil.ConstructSignature(payload, secret, &webhookutil.SignatureOptions{
    Timestamped: true,
})
```

## 페이지네이션

페이지네이션이 있는 엔드포인트는 `PageResult[T]`를 반환합니다.

```go
// 첫 페이지 조회
page, err := client.Campaigns.List(ctx, &playcamp.PaginationOptions{
    Page:  playcamp.Int(1),
    Limit: playcamp.Int(20),
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("총 %d개 중 %d개 조회\n", page.Pagination.Total, len(page.Data))
fmt.Printf("다음 페이지 있음: %v\n", page.HasNextPage)
```

## 포인터 헬퍼

optional 필드를 설정할 때 포인터 헬퍼 함수를 사용합니다.

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

## 라이선스

Copyright PlayCamp. All rights reserved.
