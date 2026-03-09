# CRM SDK — Go

Go SDK cho CRM Platform. Dùng ở **backend đối tác** với **server key** (`sk_live_*`).

## Cài đặt

```bash
go get github.com/vothanh/crm-platform/sdk/go
```

## Sử dụng

```go
import crmsdk "github.com/vothanh/crm-platform/sdk/go"

// Đọc từ env: CRM_API_KEY, CRM_ENDPOINT
client := crmsdk.NewClient()

// Hoặc truyền trực tiếp:
client := crmsdk.NewClient(
    crmsdk.WithAPIKey("sk_live_xxxx"),
    crmsdk.WithEndpoint("https://crm.example.com"),
)

// Gửi event
client.Track(ctx, crmsdk.EventUserRegistered{
    ExternalUserID: "user-123",
    Phone:          "0901234567",
    FullName:       "Nguyễn Văn A",
    Source:         crmsdk.SourceFacebookAds,
})

client.Track(ctx, crmsdk.EventTripCompleted{
    ExternalUserID: "user-123",
    ExternalTripID: "trip-456",
    Amount:         350000,
})

// Batch events
client.TrackBatch(ctx, []crmsdk.Event{
    crmsdk.EventAppOpened{ExternalUserID: "user-123"},
    crmsdk.EventPageViewed{ExternalUserID: "user-123", PageName: "home"},
})
```

## Event Types

| Event | Description |
|---|---|
| `EventUserRegistered` | User đăng ký |
| `EventAppInstalled` | Cài đặt app |
| `EventAppOpened` | Mở app |
| `EventSearchTrip` | Tìm chuyến |
| `EventTripBooked` | Đặt chuyến |
| `EventTripCompleted` | Hoàn thành chuyến |
| `EventTripCancelled` | Huỷ chuyến |
| `EventButtonClicked` | Click button |
| `EventPageViewed` | Xem trang |
| `EventEnterLocation` | Nhập địa điểm |

## Environment Variables

| Variable | Description |
|---|---|
| `CRM_API_KEY` | Server API key (`sk_live_*`) |
| `CRM_ENDPOINT` | CRM API endpoint |
