#!/bin/bash
# CRM Platform — Phase 1 API Test Script
# Usage: bash scripts/test_api.sh

BASE_URL="${CRM_URL:-http://localhost:8080}"

echo "========================================"
echo " CRM Platform - Phase 1 API Tests"
echo " Server: $BASE_URL"
echo "========================================"
echo ""

# --- 1. Health Check ---
echo "🔍 [1/9] Health Check"
curl -sf "$BASE_URL/health" | python3 -m json.tool
echo ""

# --- 2. Create Tenant ---
echo "🏢 [2/9] Create Tenant (BUTL)"
TENANT_RESPONSE=$(curl -sf -X POST "$BASE_URL/api/v1/tenants" \
  -H "Content-Type: application/json" \
  -d '{"name": "Ban Uong Toi Lai", "slug": "butl-test"}')
echo "$TENANT_RESPONSE" | python3 -m json.tool
TENANT_ID=$(echo "$TENANT_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin)['data']['id'])")
echo "  → Tenant ID: $TENANT_ID"
echo ""

# --- 3. Create Server API Key ---
echo "🔑 [3/9] Create Server API Key"
KEY_RESPONSE=$(curl -sf -X POST "$BASE_URL/api/v1/tenants/$TENANT_ID/api-keys" \
  -H "Content-Type: application/json" \
  -d '{"key_type": "server", "name": "Test Server Key"}')
echo "$KEY_RESPONSE" | python3 -m json.tool
API_KEY=$(echo "$KEY_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin)['data']['raw_key'])")
echo "  → API Key: $API_KEY"
echo ""

# --- 4. Ingest: User Registered ---
echo "📥 [4/9] Ingest: user_registered"
curl -sf -X POST "$BASE_URL/api/v1/ingest/events" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d "{
    \"event_type\": \"user_registered\",
    \"external_user_id\": \"test-user-001\",
    \"data\": {
      \"phone\": \"0901234567\",
      \"full_name\": \"Nguyen Van Minh\",
      \"source\": \"facebook_ads\",
      \"device_type\": \"ios\"
    }
  }" | python3 -m json.tool
echo ""

# --- 5. Ingest: Trip 1 Completed ---
echo "🚗 [5/9] Ingest: trip_completed #1 (350K VND)"
curl -sf -X POST "$BASE_URL/api/v1/ingest/events" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d "{
    \"event_type\": \"trip_completed\",
    \"external_user_id\": \"test-user-001\",
    \"data\": {
      \"external_trip_id\": \"trip-001\",
      \"amount\": 350000,
      \"pickup_location\": \"Nguyen Hue, Q1\"
    }
  }" | python3 -m json.tool
echo ""

# --- 6. Ingest: Trip 2 Completed ---
echo "🚗 [6/9] Ingest: trip_completed #2 (400K VND)"
curl -sf -X POST "$BASE_URL/api/v1/ingest/events" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d "{
    \"event_type\": \"trip_completed\",
    \"external_user_id\": \"test-user-001\",
    \"data\": {
      \"external_trip_id\": \"trip-002\",
      \"amount\": 400000,
      \"pickup_location\": \"Le Loi, Q1\"
    }
  }" | python3 -m json.tool
echo ""

# --- 7. Ingest: Trip 3 Completed (should trigger Luxury!) ---
echo "🚗 [7/9] Ingest: trip_completed #3 (350K VND — total > 1M → Luxury!)"
curl -sf -X POST "$BASE_URL/api/v1/ingest/events" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d "{
    \"event_type\": \"trip_completed\",
    \"external_user_id\": \"test-user-001\",
    \"data\": {
      \"external_trip_id\": \"trip-003\",
      \"amount\": 350000,
      \"pickup_location\": \"Hai Ba Trung, Q3\"
    }
  }" | python3 -m json.tool
echo ""

# --- 8. Get Customer (should show lifecycle=loyal, tier=luxury) ---
echo "👤 [8/9] List Customers (verify lifecycle + tier)"
curl -sf "$BASE_URL/api/v1/customers" \
  -H "X-API-Key: $API_KEY" | python3 -m json.tool
echo ""

# --- 9. Lifecycle Counts ---
echo "📊 [9/9] Lifecycle Counts"
curl -sf "$BASE_URL/api/v1/customers/lifecycle-counts" \
  -H "X-API-Key: $API_KEY" | python3 -m json.tool
echo ""

echo "========================================"
echo " ✅ All tests completed!"
echo "========================================"
