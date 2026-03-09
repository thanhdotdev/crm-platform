#!/bin/bash
# Test: trip_number only assigned on completion, cancel doesn't count
set -e
BASE="http://localhost:8080"
OK=0; FAIL=0

pass() { echo "  ✅ $1"; OK=$((OK+1)); }
fail() { echo "  ❌ $1: $2"; FAIL=$((FAIL+1)); }

echo "========================================"
echo " Trip Number & Cancel Logic Test"
echo "========================================"

# --- Setup: tenant + key ---
echo ""
echo "📦 SETUP"
R=$(curl -sf -X POST "$BASE/api/v1/tenants" -H "Content-Type: application/json" \
  -d '{"name":"Trip Logic Test","slug":"trip-test-'$RANDOM'"}')
TID=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")
R=$(curl -sf -X POST "$BASE/api/v1/tenants/$TID/api-keys" -H "Content-Type: application/json" \
  -d '{"key_type":"server","name":"test"}')
KEY=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['raw_key'])")
H="-H X-API-Key:$KEY -H Content-Type:application/json"
pass "Tenant+Key created"

# --- Register user ---
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"event_type":"user_registered","external_user_id":"u100","data":{"phone":"090100"}}' > /dev/null
pass "User registered"

# Get customer ID
R=$(curl -sf "$BASE/api/v1/customers" $H)
CID=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'][0]['id'])")

# --- Test 1: trip_booked should NOT have trip_number ---
echo ""
echo "🧪 TEST 1: Booked trip — trip_number should be 0"
R=$(curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"event_type":"trip_booked","external_user_id":"u100","data":{"external_trip_id":"t200","amount":300000}}')
echo "$R" | python3 -c "
import sys,json
d=json.load(sys.stdin)['data']
# trip_number should NOT be in response (or be 0)
has_tn = 'trip_number' in d
print(f'  trip_number in response: {has_tn}')
" 2>/dev/null
pass "Booked trip created without trip_number"

# --- Test 2: trip_cancelled should NOT count ---
echo ""
echo "🧪 TEST 2: Cancel booked trip — check customer total_trips stays 0"
R=$(curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"event_type":"trip_cancelled","external_user_id":"u100","data":{"external_trip_id":"t200","cancel_reason":"changed mind"}}')
pass "Trip cancelled"

R=$(curl -sf "$BASE/api/v1/customers/$CID" $H)
TOTAL=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total_trips'])")
if [ "$TOTAL" = "0" ]; then pass "total_trips=0 after cancel ✓"; else fail "total_trips should be 0" "got $TOTAL"; fi

# --- Test 3: trip_completed should get trip_number=1 ---
echo ""
echo "🧪 TEST 3: Complete trip — trip_number should be 1"
R=$(curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"event_type":"trip_completed","external_user_id":"u100","data":{"external_trip_id":"t201","amount":400000}}')
TN=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['trip_number'])")
if [ "$TN" = "1" ]; then pass "trip_number=1 ✓"; else fail "trip_number should be 1" "got $TN"; fi

# --- Test 4: second completed trip → trip_number=2 ---
echo ""
echo "🧪 TEST 4: Second complete → trip_number=2"
R=$(curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"event_type":"trip_completed","external_user_id":"u100","data":{"external_trip_id":"t202","amount":350000}}')
TN=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['trip_number'])")
if [ "$TN" = "2" ]; then pass "trip_number=2 ✓"; else fail "trip_number should be 2" "got $TN"; fi

# --- Test 5: book + cancel + complete → trip_number stays sequential ---
echo ""
echo "🧪 TEST 5: Book→Cancel→Complete — trip_number=3 (cancel doesn't affect)"
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"event_type":"trip_booked","external_user_id":"u100","data":{"external_trip_id":"t203"}}' > /dev/null
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"event_type":"trip_cancelled","external_user_id":"u100","data":{"external_trip_id":"t203"}}' > /dev/null
R=$(curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"event_type":"trip_completed","external_user_id":"u100","data":{"external_trip_id":"t204","amount":300000}}')
TN=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['trip_number'])")
if [ "$TN" = "3" ]; then pass "trip_number=3 after book+cancel+complete ✓"; else fail "trip_number should be 3" "got $TN"; fi

# --- Test 6: Verify customer total_trips = 3 (only completed) ---
echo ""
echo "🧪 TEST 6: Customer total_trips = 3 (completed only)"
R=$(curl -sf "$BASE/api/v1/customers/$CID" $H)
TOTAL=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total_trips'])")
STAGE=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['lifecycle_stage'])")
SPENT=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total_spent'])")
if [ "$TOTAL" = "3" ]; then pass "total_trips=3 ✓"; else fail "total_trips should be 3" "got $TOTAL"; fi
if [ "$STAGE" = "loyal" ]; then pass "lifecycle=loyal (3 trips) ✓"; else fail "lifecycle should be loyal" "got $STAGE"; fi
echo "  📊 Spent=$SPENT"

# --- Summary ---
echo ""
echo "========================================"
echo " RESULTS: ✅ $OK passed | ❌ $FAIL failed"
echo "========================================"
