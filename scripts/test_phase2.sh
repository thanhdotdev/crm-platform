#!/bin/bash
# CRM Platform — Full API Test (Phase 1 + 2, post trip_number removal)
set -e
BASE="http://localhost:8080"
OK=0; FAIL=0

pass() { echo "  ✅ $1"; OK=$((OK+1)); }
fail() { echo "  ❌ $1: $2"; FAIL=$((FAIL+1)); }
check() {
  local name="$1" result="$2"
  if echo "$result" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d.get('success',False)" 2>/dev/null; then
    pass "$name"
  else
    fail "$name" "$result"
  fi
}

echo "========================================"
echo " CRM Full API Test (Phase 1 + 2)"
echo "========================================"

# --- 1. Tenant + Key ---
echo ""
echo "📦 TENANT + API KEY"
R=$(curl -sf -X POST "$BASE/api/v1/tenants" -H "Content-Type: application/json" \
  -d '{"name":"Full Test","slug":"full-test-'$RANDOM'"}')
check "Create Tenant" "$R"
TID=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")

R=$(curl -sf -X POST "$BASE/api/v1/tenants/$TID/api-keys" -H "Content-Type: application/json" \
  -d '{"key_type":"server","name":"test key"}')
check "Create API Key" "$R"
KEY=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['raw_key'])")
H="-H X-API-Key:$KEY -H Content-Type:application/json"

# --- 2. Ingestion ---
echo ""
echo "📥 INGESTION"
R=$(curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"user_registered","external_user_id":"u001","data":{"phone":"090111","full_name":"Nguyen Van A","source":"facebook_ads"}}')
check "user_registered" "$R"

R=$(curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"trip_booked","external_user_id":"u001","data":{"external_trip_id":"t001","amount":350000}}')
check "trip_booked" "$R"

R=$(curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"trip_completed","external_user_id":"u001","data":{"external_trip_id":"t001","amount":350000}}')
check "trip_completed #1" "$R"

R=$(curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"trip_completed","external_user_id":"u001","data":{"external_trip_id":"t002","amount":400000}}')
check "trip_completed #2" "$R"

R=$(curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"trip_completed","external_user_id":"u001","data":{"external_trip_id":"t003","amount":350000}}')
check "trip_completed #3" "$R"

# --- 3. Customer verify ---
echo ""
echo "👤 CUSTOMER"
R=$(curl -sf "$BASE/api/v1/customers" $H)
check "List Customers" "$R"
CID=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'][0]['id'])")
STAGE=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'][0]['lifecycle_stage'])")
TIER=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'][0]['tier'])")
TRIPS=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'][0]['total_trips'])")
SPENT=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'][0]['total_spent'])")
echo "  📊 Stage=$STAGE | Tier=$TIER | Trips=$TRIPS | Spent=$SPENT"

if [ "$TRIPS" = "3" ]; then pass "total_trips=3 (completed only)"; else fail "total_trips" "got $TRIPS"; fi
if [ "$STAGE" = "loyal" ]; then pass "lifecycle=loyal"; else fail "lifecycle" "got $STAGE"; fi
if [ "$TIER" = "luxury" ]; then pass "tier=luxury (≥3 trips, ≥1M)"; else fail "tier" "got $TIER"; fi

# Cancel test
R=$(curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"trip_booked","external_user_id":"u001","data":{"external_trip_id":"t-cancel"}}')
R=$(curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"trip_cancelled","external_user_id":"u001","data":{"external_trip_id":"t-cancel","cancel_reason":"rain"}}')
R=$(curl -sf "$BASE/api/v1/customers/$CID" $H)
TRIPS2=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total_trips'])")
if [ "$TRIPS2" = "3" ]; then pass "Cancel doesn't increment total_trips ✓"; else fail "cancel" "total_trips=$TRIPS2"; fi

# --- 4. Segmentation ---
echo ""
echo "🎯 SEGMENTATION"
R=$(curl -sf -X POST "$BASE/api/v1/segments" $H \
  -d '{"name":"High Value","type":"trip_based","rules":{"min_trips":2,"min_spent":500000}}')
check "Create Segment" "$R"
R=$(curl -sf "$BASE/api/v1/segments" $H)
check "List Segments" "$R"

# --- 5. Campaign + Voucher ---
echo ""
echo "🎪 CAMPAIGN"
R=$(curl -sf -X POST "$BASE/api/v1/campaigns" $H \
  -d '{"name":"Campaign 150","type":"new_user","start_date":"2026-03-01","end_date":"2026-08-01","budget":50000000}')
check "Create Campaign" "$R"
CAMP_ID=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")

R=$(curl -sf -X POST "$BASE/api/v1/campaigns/$CAMP_ID/vouchers/generate" $H \
  -d "{\"customer_id\":\"$CID\",\"trip_number\":1}")
check "Generate Voucher 50K" "$R"
V_CODE=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['code'])")
V_VAL=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['discount_value'])")
echo "  🎟️  $V_CODE = ${V_VAL}đ"

TRIP_ID=$(curl -sf "$BASE/api/v1/trips/customer/$CID" $H | python3 -c "import sys,json; print(json.load(sys.stdin)['data'][0]['id'])")
R=$(curl -sf -X POST "$BASE/api/v1/campaigns/vouchers/redeem" $H \
  -d "{\"code\":\"$V_CODE\",\"customer_id\":\"$CID\",\"trip_id\":\"$TRIP_ID\"}")
check "Redeem Voucher" "$R"

# --- 6. Automation ---
echo ""
echo "⚙️  AUTOMATION"
R=$(curl -sf -X POST "$BASE/api/v1/automation/rules" $H \
  -d '{"name":"Welcome Push","trigger_type":"event","trigger_config":{"event":"user_registered"},"action_type":"send_notification","execution_mode":"automatic"}')
check "Create Rule" "$R"
R=$(curl -sf "$BASE/api/v1/automation/rules" $H)
check "List Rules" "$R"

# --- 7. Notification ---
echo ""
echo "🔔 NOTIFICATION"
R=$(curl -sf -X POST "$BASE/api/v1/notifications" $H \
  -d "{\"customer_id\":\"$CID\",\"channel\":\"push\",\"title\":\"Hello\",\"content\":\"Test\"}")
check "Send Push" "$R"
R=$(curl -sf "$BASE/api/v1/notifications/channels" $H)
check "List Channels" "$R"

# --- 8. Analytics ---
echo ""
echo "📊 ANALYTICS"
R=$(curl -sf "$BASE/api/v1/analytics/overview" $H)
check "Overview" "$R"
echo "$R" | python3 -c "
import sys,json; d=json.load(sys.stdin)['data']
print(f\"  📈 Customers={d['total_customers']} | Trips={d['total_trips']} | Revenue={d['total_revenue']:,.0f}đ\")
" 2>/dev/null

R=$(curl -sf "$BASE/api/v1/analytics/acquisition" $H)
check "Acquisition" "$R"

R=$(curl -sf "$BASE/api/v1/analytics/retention" $H)
check "Retention" "$R"
echo "$R" | python3 -c "
import sys,json; f=json.load(sys.stdin)['data']['conversion_funnel']
print(f\"  🔄 Reg→1st: {f['registered_to_1st_trip']:.0f}% | 1st→2nd: {f['1st_to_2nd_trip']:.0f}% | 2nd→3rd: {f['2nd_to_3rd_trip']:.0f}%\")
" 2>/dev/null

# --- Summary ---
echo ""
echo "========================================"
echo " RESULTS: ✅ $OK passed | ❌ $FAIL failed"
echo "========================================"
