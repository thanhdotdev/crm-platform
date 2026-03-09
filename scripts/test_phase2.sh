#!/bin/bash
# CRM Platform — Phase 2 API Test Script
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
echo " CRM Phase 2 — Full API Test"
echo "========================================"

# --- 1. Create tenant ---
echo ""
echo "📦 TENANT + API KEY"
R=$(curl -sf -X POST "$BASE/api/v1/tenants" -H "Content-Type: application/json" \
  -d '{"name":"Phase2 Test","slug":"p2test-'$RANDOM'"}')
check "Create Tenant" "$R"
TID=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")

R=$(curl -sf -X POST "$BASE/api/v1/tenants/$TID/api-keys" -H "Content-Type: application/json" \
  -d '{"key_type":"server","name":"test key"}')
check "Create API Key" "$R"
KEY=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['raw_key'])")
echo "  🔑 Key: ${KEY:0:20}..."

H="-H X-API-Key:$KEY -H Content-Type:application/json"

# --- 2. Ingest events ---
echo ""
echo "📥 INGESTION"
R=$(curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"event_type":"user_registered","external_user_id":"u001","data":{"phone":"090111","full_name":"Test User","source":"facebook_ads"}}')
check "Ingest user_registered" "$R"

R=$(curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"event_type":"trip_completed","external_user_id":"u001","data":{"external_trip_id":"t001","amount":350000}}')
check "Ingest trip_completed #1" "$R"

R=$(curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"event_type":"trip_completed","external_user_id":"u001","data":{"external_trip_id":"t002","amount":400000}}')
check "Ingest trip_completed #2" "$R"

R=$(curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"event_type":"trip_completed","external_user_id":"u001","data":{"external_trip_id":"t003","amount":350000}}')
check "Ingest trip_completed #3 (Luxury?)" "$R"

# --- 3. Customer verification ---
echo ""
echo "👤 CUSTOMER VERIFY"
R=$(curl -sf "$BASE/api/v1/customers" $H)
check "List Customers" "$R"
CID=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'][0]['id'])")
STAGE=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'][0]['lifecycle_stage'])")
TIER=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'][0]['tier'])")
TRIPS=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'][0]['total_trips'])")
SPENT=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'][0]['total_spent'])")
echo "  📊 Stage=$STAGE | Tier=$TIER | Trips=$TRIPS | Spent=$SPENT"

R=$(curl -sf "$BASE/api/v1/customers/$CID/timeline" $H)
check "Customer Timeline" "$R"
TIMELINE_COUNT=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['meta']['total'])")
echo "  📋 Timeline events: $TIMELINE_COUNT"

R=$(curl -sf "$BASE/api/v1/customers/lifecycle-counts" $H)
check "Lifecycle Counts" "$R"

# --- 4. Segmentation ---
echo ""
echo "🎯 SEGMENTATION"
R=$(curl -sf -X POST "$BASE/api/v1/segments" $H \
  -d '{"name":"High Spenders","type":"trip_based","rules":{"min_trips":2,"min_spent":500000},"description":"Customers with 2+ trips and 500K+ spent"}')
check "Create Segment" "$R"
SID=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")

R=$(curl -sf "$BASE/api/v1/segments" $H)
check "List Segments" "$R"

R=$(curl -sf "$BASE/api/v1/segments/$SID" $H)
check "Get Segment Detail" "$R"

# --- 5. Campaign ---
echo ""
echo "🎪 CAMPAIGN"
R=$(curl -sf -X POST "$BASE/api/v1/campaigns" $H \
  -d '{"name":"Campaign 150 ngày","type":"new_user","start_date":"2026-03-01","end_date":"2026-08-01","budget":50000000,"description":"Chiến dịch 150 ngày"}')
check "Create Campaign" "$R"
CAMP_ID=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")

R=$(curl -sf "$BASE/api/v1/campaigns" $H)
check "List Campaigns" "$R"

R=$(curl -sf "$BASE/api/v1/campaigns/$CAMP_ID" $H)
check "Get Campaign Detail" "$R"

# Generate vouchers (50-30-20 policy)
R=$(curl -sf -X POST "$BASE/api/v1/campaigns/$CAMP_ID/vouchers/generate" $H \
  -d "{\"customer_id\":\"$CID\",\"trip_number\":1}")
check "Generate Voucher (Trip 1 = 50K)" "$R"
V_CODE=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['code'])")
V_DISCOUNT=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['discount_value'])")
echo "  🎟️  Code=$V_CODE | Discount=${V_DISCOUNT}đ"

R=$(curl -sf -X POST "$BASE/api/v1/campaigns/$CAMP_ID/vouchers/generate" $H \
  -d "{\"customer_id\":\"$CID\",\"trip_number\":2}")
check "Generate Voucher (Trip 2 = 30K)" "$R"
V2_DISCOUNT=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['discount_value'])")
echo "  🎟️  Discount=${V2_DISCOUNT}đ"

R=$(curl -sf -X POST "$BASE/api/v1/campaigns/$CAMP_ID/vouchers/generate" $H \
  -d "{\"customer_id\":\"$CID\",\"trip_number\":3}")
check "Generate Voucher (Trip 3 = 20K)" "$R"
V3_DISCOUNT=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['discount_value'])")
echo "  🎟️  Discount=${V3_DISCOUNT}đ"

R=$(curl -sf "$BASE/api/v1/campaigns/$CAMP_ID/vouchers" $H)
check "List Campaign Vouchers" "$R"
V_COUNT=$(echo "$R" | python3 -c "import sys,json; print(len(json.load(sys.stdin)['data']))")
echo "  📋 Vouchers in campaign: $V_COUNT"

# Redeem voucher
TRIP_ID=$(curl -sf "$BASE/api/v1/trips/customer/$CID" $H | python3 -c "import sys,json; print(json.load(sys.stdin)['data'][0]['id'])")
R=$(curl -sf -X POST "$BASE/api/v1/campaigns/vouchers/redeem" $H \
  -d "{\"code\":\"$V_CODE\",\"customer_id\":\"$CID\",\"trip_id\":\"$TRIP_ID\"}")
check "Redeem Voucher" "$R"

# Try double redeem — should fail
R2=$(curl -s -X POST "$BASE/api/v1/campaigns/vouchers/redeem" $H \
  -d "{\"code\":\"$V_CODE\",\"customer_id\":\"$CID\",\"trip_id\":\"$TRIP_ID\"}")
IS_ERR=$(echo "$R2" | python3 -c "import sys,json; d=json.load(sys.stdin); print('yes' if not d.get('success',False) else 'no')" 2>/dev/null || echo "yes")
if [ "$IS_ERR" = "yes" ]; then pass "Double-redeem blocked ✓"; else fail "Double-redeem NOT blocked" "$R2"; fi

# --- 6. Automation ---
echo ""
echo "⚙️  AUTOMATION"
R=$(curl -sf -X POST "$BASE/api/v1/automation/rules" $H \
  -d '{"name":"Welcome Push","trigger_type":"event","trigger_config":{"event":"user_registered"},"action_type":"send_notification","action_config":{"channel":"push","title":"Welcome!"},"execution_mode":"automatic"}')
check "Create Auto Rule" "$R"
RULE_ID=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")

R=$(curl -sf -X POST "$BASE/api/v1/automation/rules" $H \
  -d '{"name":"VIP Manual Review","trigger_type":"event","trigger_config":{"event":"trip_completed"},"action_type":"assign_voucher","action_config":{"trip_number":1},"execution_mode":"manual"}')
check "Create Manual Rule" "$R"

R=$(curl -sf "$BASE/api/v1/automation/rules" $H)
check "List Rules" "$R"
RULE_COUNT=$(echo "$R" | python3 -c "import sys,json; print(len(json.load(sys.stdin)['data']))")
echo "  📋 Rules: $RULE_COUNT"

R=$(curl -sf "$BASE/api/v1/automation/rules/$RULE_ID" $H)
check "Get Rule Detail" "$R"

# --- 7. Notification ---
echo ""
echo "🔔 NOTIFICATION"
R=$(curl -sf -X POST "$BASE/api/v1/notifications" $H \
  -d "{\"customer_id\":\"$CID\",\"channel\":\"push\",\"title\":\"Welcome!\",\"content\":\"Cảm ơn bạn đã đăng ký\"}")
check "Send Push Notification" "$R"

R=$(curl -sf -X POST "$BASE/api/v1/notifications" $H \
  -d "{\"customer_id\":\"$CID\",\"channel\":\"sms\",\"title\":\"Promo\",\"content\":\"Ưu đãi 50K cho chuyến xe tiếp theo\"}")
check "Send SMS Notification" "$R"

R=$(curl -sf -X POST "$BASE/api/v1/notifications" $H \
  -d "{\"customer_id\":\"$CID\",\"channel\":\"email\",\"title\":\"Trip Summary\",\"content\":\"Bạn đã hoàn thành 3 chuyến\"}")
check "Send Email Notification" "$R"

R=$(curl -sf "$BASE/api/v1/notifications/customer/$CID" $H)
check "List Customer Notifications" "$R"
N_COUNT=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['meta']['total'])")
echo "  📋 Notifications sent: $N_COUNT"

R=$(curl -sf "$BASE/api/v1/notifications/channels" $H)
check "List Channels" "$R"

# --- 8. Analytics ---
echo ""
echo "📊 ANALYTICS"
R=$(curl -sf "$BASE/api/v1/analytics/overview" $H)
check "Overview KPIs" "$R"
echo "$R" | python3 -c "
import sys,json
d=json.load(sys.stdin)['data']
print(f\"  📈 Customers={d['total_customers']} | Trips={d['total_trips']} | Revenue={d['total_revenue']:,.0f}đ\")
" 2>/dev/null || true

R=$(curl -sf "$BASE/api/v1/analytics/acquisition" $H)
check "Acquisition by Source" "$R"

R=$(curl -sf "$BASE/api/v1/analytics/retention" $H)
check "Retention Funnel" "$R"
echo "$R" | python3 -c "
import sys,json
d=json.load(sys.stdin)['data']
f=d.get('conversion_funnel',{})
print(f\"  🔄 Reg→1st: {f.get('registered_to_1st_trip',0):.0f}% | 1st→2nd: {f.get('1st_to_2nd_trip',0):.0f}% | 2nd→3rd: {f.get('2nd_to_3rd_trip',0):.0f}%\")
" 2>/dev/null || true

# --- Summary ---
echo ""
echo "========================================"
echo " RESULTS: ✅ $OK passed | ❌ $FAIL failed"
echo "========================================"
