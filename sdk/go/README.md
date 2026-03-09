# CRM Go SDK

A lightweight Go SDK for sending events to the CRM Platform.

## Installation

```bash
go get github.com/vothanh/crm-platform/sdk/go
```

## Usage

### 1. Initialization

By default, the SDK reads the API key and endpoint from environment variables (`CRM_API_KEY` and `CRM_ENDPOINT`).

```go
package main

import (
	"context"
	"log"
	
	"github.com/vothanh/crm-platform/sdk/go" // import path depends on your project setup
)

func main() {
	// Initialize using environment variables
	// export CRM_API_KEY="sk_live_..."
	// export CRM_ENDPOINT="http://localhost:8080"
	client := crmsdk.NewClient()
    
    // OR initialize manually
    // client := crmsdk.NewClient("sk_live_...", "http://localhost:8080")
    
    // ...
}
```

### 2. Tracking Events

Use the `Track` method to send events to the CRM. The parameters are:
- `ctx`: context for the request
- `eventType`: string identifying the event (e.g. "user_registered", "trip_completed")
- `externalUserID`: unique identifier for the user in your system
- `userType`: type of user (e.g. "customer", "driver")
- `data`: a map containing custom event properties

```go
err := client.Track(
    context.Background(),
    "trip_completed",     // event type
    "user-123",           // user internal ID
    "customer",           // user type
    map[string]interface{}{ // custom data
        "amount": 350000,
        "external_trip_id": "trip-999",
        "dropoff_location": "District 1",
    },
)

if err != nil {
    log.Printf("Failed to track event: %v", err)
}
```
