# CRM Platform - SDK Integration Guide

This guide explains how to install, configure, and use the CRM SDKs (Go and JS/TS) to track user events from your applications directly into the CRM Data Hub.

---

## 🔑 1. Prerequisites

Before integration, we will provide you with:
1. **API Key**: e.g. `sk_live_...` or `ck_live_...`
2. **CRM API Endpoint**: e.g., `https://api.crm-platform.com`

You will need to configure these in your environment variables.

---

## 📦 2. Go SDK Integration

A lightweight Go SDK for sending events strictly typed or dynamically.

### Installation
```bash
go get github.com/vothanh/crm-platform/sdk/go
```

### Initialization & Usage
By default, the SDK reads from environment variables (`CRM_API_KEY` and `CRM_ENDPOINT`).

```go
package main

import (
	"context"
	"log"
	
	crmsdk "github.com/vothanh/crm-platform/sdk/go"
)

func main() {
    // The SDK initialized automatically using ENV vars
    // export CRM_API_KEY="sk_live_..."
    // export CRM_ENDPOINT="http://localhost:8080"
    client := crmsdk.NewClient()
    
    // Track an event and push data to CRM
    err := client.Track(
        context.Background(),
        "custom_event_name",  // Event Type
        "user-123",           // User ID in your database
        "customer",           // User Role/Type
        map[string]interface{}{ // Any custom JSON payload
            "amount": 350000,
            "status": "success",
            "source": "web",
        },
    )

    if err != nil {
        log.Printf("Failed to track event: %v", err)
    }
}
```

---

## 🌐 3. JS/TS SDK Integration

A lightweight TypeScript SDK for tracking events from Web browsers, React Native, or Node.js backends.

### Installation
```bash
npm install @vothanh/crm-sdk
```

### Initialization & Usage

#### For Node.js / React Native (Server-side & App)
Set environment variables `CRM_API_KEY` and `CRM_ENDPOINT` in your `.env` file.

```typescript
import { CRMClient } from '@vothanh/crm-sdk';

// Automatically reads from process.env
const crm = new CRMClient();

// Send an event asynchronously
await crm.track(
  'custom_event_name', // Event Type
  'user-123',          // User ID in your database
  'customer',          // User Role/Type
  {                    // Any custom JSON payload
    amount: 150000,
    status: 'success',
    source: 'mobile_app'
  }
);
```

#### For Web Browser (Client-side HTML)
If you are integrating directly in the browser via HTML/Scripts, set the global variables before initializing the script.

```html
<script>
  window.__CRM_CONFIG = {
    apiKey: 'ck_live_...',   // Use the Client Key provided
    endpoint: 'https://api.crm-platform.com'
  };
</script>
```
```typescript
import { CRMClient } from '@vothanh/crm-sdk';

// Automatically reads window.__CRM_CONFIG
const crm = new CRMClient();

// Track an event
crm.track('page_viewed', 'user-123', 'customer', { page_name: 'home' });
```
