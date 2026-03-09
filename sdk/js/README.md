# CRM JS/TS SDK

A lightweight TypeScript SDK for tracking events from client applications (Web, React Native, Node.js).

## Installation

```bash
npm install @vothanh/crm-sdk
```

## Usage

### 1. Initialization

The SDK can automatically read configuration from the environment, making it easy to instantiate without passing arguments manually.

#### In Node.js / React Native
Set environment variables `CRM_API_KEY` and `CRM_ENDPOINT`.

```typescript
import { CRMClient } from '@vothanh/crm-sdk';

// Automatically reads process.env.CRM_API_KEY and CRM_ENDPOINT
const crm = new CRMClient();
```

#### In the Browser
Set global variables before initializing.

```html
<script>
  window.__CRM_CONFIG = {
    apiKey: 'ck_live_...',
    endpoint: 'http://api.yourdomain.com'
  };
</script>
```
```typescript
import { CRMClient } from '@vothanh/crm-sdk';

// Automatically reads window.__CRM_CONFIG
const crm = new CRMClient();
```

#### Manual Initialization
You can always fall back to passing configuration directly:

```typescript
const crm = new CRMClient('ck_live_...', 'http://localhost:8080');
```

### 2. Tracking Events

Use the `track` method to send an event. 

Parameters:
1. `eventType`: Event name (standard types or custom string)
2. `externalUserId`: Your system's unique user ID
3. `userType`: Type of user (e.g., 'customer', 'driver')
4. `data` (optional): Additional event properties

```typescript
// Example: Tracking a completed trip
await crm.track(
  'trip_completed', // Event type
  'user-123',       // User ID
  'customer',       // User type
  {                 // Custom data
    amount: 350000,
    external_trip_id: 'trip-999',
    dropoff_location: 'District 1'
  }
);
```

### Supported Standard Event Types
You can pass any string as an event type, but the SDK exports standard types for convenience:
- `user_registered`
- `app_installed`
- `app_opened`
- `search_trip`
- `trip_booked`
- `trip_completed`
- `trip_cancelled`
- `button_clicked`
- `page_viewed`
- `enter_location`
