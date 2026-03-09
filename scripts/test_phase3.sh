#!/bin/bash
# Phase 3: Advanced Analytics Test
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
echo " Phase 3: Advanced Analytics Test"
echo "========================================"

# --- Setup ---
echo ""
echo "📦 SETUP (tenant + key + data)"
R=$(curl -sf -X POST "$BASE/api/v1/tenants" -H "Content-Type: application/json" \
  -d '{"name":"P3 Test","slug":"p3-test-'$RANDOM'"}')
TID=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")
R=$(curl -sf -X POST "$BASE/api/v1/tenants/$TID/api-keys" -H "Content-Type: application/json" \
  -d '{"key_type":"server","name":"test"}')
KEY=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['raw_key'])")
H="-H X-API-Key:$KEY -H Content-Type:application/json"
pass "Tenant+Key"

# Create 3 users with different trip counts
for u in 1 2 3; do
  curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
    -d "{\"event_type\":\"user_registered\",\"external_user_id\":\"p3-u$u\",\"data\":{\"phone\":\"090$u\",\"full_name\":\"User $u\",\"source\":\"facebook_ads\"}}" > /dev/null
done
pass "3 users registered"

# User 1: 3 completed trips (loyal) + 1 cancel
for t in 1 2 3; do
  curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
    -d "{\"event_type\":\"trip_completed\",\"external_user_id\":\"p3-u1\",\"data\":{\"external_trip_id\":\"p3-t1-$t\",\"amount\":400000}}" > /dev/null
done
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"event_type":"trip_booked","external_user_id":"p3-u1","data":{"external_trip_id":"p3-t1-c1"}}' > /dev/null
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"event_type":"trip_cancelled","external_user_id":"p3-u1","data":{"external_trip_id":"p3-t1-c1"}}' > /dev/null
pass "User1: 3 completed + 1 cancel"

# User 2: 1 completed trip + 3 cancels (suspicious!)
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"event_type":"trip_completed","external_user_id":"p3-u2","data":{"external_trip_id":"p3-t2-1","amount":300000}}' > /dev/null
for canc in 1 2 3; do
  curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
    -d "{\"event_type\":\"trip_booked\",\"external_user_id\":\"p3-u2\",\"data\":{\"external_trip_id\":\"p3-t2-c$canc\"}}" > /dev/null
  curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
    -d "{\"event_type\":\"trip_cancelled\",\"external_user_id\":\"p3-u2\",\"data\":{\"external_trip_id\":\"p3-t2-c$canc\"}}" > /dev/null
done
pass "User2: 1 completed + 3 cancels (abuse pattern)"

# User 3: 0 trips (new user, no activity)
pass "User3: 0 trips"

# Create campaign + voucher for ROI test
R=$(curl -sf -X POST "$BASE/api/v1/campaigns" $H \
  -d '{"name":"ROI Test Campaign","type":"retention","start_date":"2026-03-01","end_date":"2026-08-01","budget":10000000}')
CAMP_ID=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")
CID1=$(curl -sf "$BASE/api/v1/customers" $H | python3 -c "import sys,json; cs=json.load(sys.stdin)['data']; print([c['id'] for c in cs if c['external_id']=='p3-u1'][0])")

R=$(curl -sf -X POST "$BASE/api/v1/campaigns/$CAMP_ID/vouchers/generate" $H \
  -d "{\"customer_id\":\"$CID1\",\"trip_number\":1}")
V_CODE=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['code'])")

TRIP_ID=$(curl -sf "$BASE/api/v1/trips/customer/$CID1" $H | python3 -c "import sys,json; print(json.load(sys.stdin)['data'][0]['id'])")
curl -sf -X POST "$BASE/api/v1/campaigns/vouchers/redeem" $H \
  -d "{\"code\":\"$V_CODE\",\"customer_id\":\"$CID1\",\"trip_id\":\"$TRIP_ID\"}" > /dev/null
pass "Campaign + voucher created and redeemed"

# --- Test Phase 3 endpoints ---
echo ""
echo "📊 PHASE 3 ENDPOINTS"

# 1. Cohort Analysis
echo ""
echo "🔬 Cohort Analysis"
R=$(curl -sf "$BASE/api/v1/analytics/cohorts" $H)
check "GET /cohorts" "$R"
echo "$R" | python3 -c "
import sys,json
d=json.load(sys.stdin)['data']
if d:
  for c in d:
    print(f\"  Week {c['cohort_week']}: {c['registered']} reg, {c['activated']} activated ({c['activation_rate']:.0f}%), {c['loyal']} loyal\")
else:
  print('  No cohort data')
" 2>/dev/null

# 2. Campaign ROI
echo ""
echo "💰 Campaign ROI"
R=$(curl -sf "$BASE/api/v1/analytics/campaign-roi" $H)
check "GET /campaign-roi" "$R"
echo "$R" | python3 -c "
import sys,json
d=json.load(sys.stdin)['data']
if d:
  for r in d:
    print(f\"  {r['campaign_name']}: {r['total_vouchers']} vouchers, {r['used_vouchers']} used ({r['redemption_rate']:.0f}%), {r['total_discount_given']:,.0f}đ discount\")
else:
  print('  No campaign data')
" 2>/dev/null

# 3. Source Quality
echo ""
echo "🎯 Source Quality"
R=$(curl -sf "$BASE/api/v1/analytics/source-quality" $H)
check "GET /source-quality" "$R"
echo "$R" | python3 -c "
import sys,json
d=json.load(sys.stdin)['data']
if d:
  for s in d:
    print(f\"  {s['source']}: {s['customers']} users, conv {s['conversion_rate']:.0f}%, LTV {s['ltv']:,.0f}đ, cancel {s['cancel_rate']:.0f}%\")
else:
  print('  No source data')
" 2>/dev/null

# 4. Abuse Detection
echo ""
echo "🚨 Abuse Detection"
R=$(curl -sf "$BASE/api/v1/analytics/abuse-detection" $H)
check "GET /abuse-detection" "$R"
echo "$R" | python3 -c "
import sys,json
d=json.load(sys.stdin)['data']
print(f\"  Total: {d['total_customers']} customers, {d['flagged_count']} flagged ({d['flagged_rate']:.1f}%)\")
if d.get('suspicious_users'):
  for u in d['suspicious_users']:
    print(f\"  ⚠️  {u['external_id']} ({u['full_name']}): {u['reason']}, risk={u['risk_score']}, cancels={u['cancel_count']}\")
" 2>/dev/null

# Did abuse detection find User 2?
FLAGGED=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['flagged_count'])")
if [ "$FLAGGED" -ge "1" ]; then pass "Abuse detection flagged suspicious user ✓"; else fail "Should flag User2" "flagged=$FLAGGED"; fi

# --- Summary ---
echo ""
echo "========================================"
echo " RESULTS: ✅ $OK passed | ❌ $FAIL failed"
echo "========================================"
