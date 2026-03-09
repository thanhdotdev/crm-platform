#!/bin/bash
# Stress Test: Ingestion API throughput
# Tests with hey (Go HTTP benchmark tool)
set -e
BASE="http://localhost:8080"

echo "========================================"
echo " CRM Ingestion Stress Test"
echo "========================================"

# Check if hey is installed
if ! command -v hey &>/dev/null; then
  echo "Installing hey..."
  go install github.com/rakyll/hey@latest
fi

# Setup: create tenant + API key
echo ""
echo "📦 Setup..."
R=$(curl -sf -X POST "$BASE/api/v1/tenants" -H "Content-Type: application/json" \
  -d '{"name":"Stress Test","slug":"stress-'$RANDOM'"}')
TID=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")

R=$(curl -sf -X POST "$BASE/api/v1/tenants/$TID/api-keys" -H "Content-Type: application/json" \
  -d '{"key_type":"server","name":"stress key"}')
KEY=$(echo "$R" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['raw_key'])")
echo "  ✅ Tenant + Key ready"

# Warmup: 1 request to prime cache
echo ""
echo "🔥 Warmup (1 request to prime API key cache)..."
curl -sf -X POST "$BASE/api/v1/ingest/events" \
  -H "X-API-Key: $KEY" -H "Content-Type: application/json" \
  -d '{"user_type":"customer","event_type":"warmup","external_user_id":"warmup-user","data":{}}' > /dev/null
echo "  ✅ Cache primed"

BODY='{"user_type":"customer","event_type":"user_registered","external_user_id":"stress-user-'$RANDOM'","data":{"phone":"090999","full_name":"Stress User","source":"stress_test"}}'

# Test 1: Baseline — 10 concurrent, 1000 requests
echo ""
echo "═══════════════════════════════════════"
echo " TEST 1: 10 concurrent × 1000 requests"
echo "═══════════════════════════════════════"
hey -n 1000 -c 10 -m POST \
  -H "X-API-Key: $KEY" \
  -H "Content-Type: application/json" \
  -d "$BODY" \
  "$BASE/api/v1/ingest/events"

# Test 2: Medium — 50 concurrent, 2000 requests
echo ""
echo "═══════════════════════════════════════"
echo " TEST 2: 50 concurrent × 2000 requests"
echo "═══════════════════════════════════════"
hey -n 2000 -c 50 -m POST \
  -H "X-API-Key: $KEY" \
  -H "Content-Type: application/json" \
  -d "$BODY" \
  "$BASE/api/v1/ingest/events"

# Test 3: High — 100 concurrent, 5000 requests
echo ""
echo "═══════════════════════════════════════"
echo " TEST 3: 100 concurrent × 5000 requests"
echo "═══════════════════════════════════════"
hey -n 5000 -c 100 -m POST \
  -H "X-API-Key: $KEY" \
  -H "Content-Type: application/json" \
  -d "$BODY" \
  "$BASE/api/v1/ingest/events"

echo ""
echo "========================================"
echo " DONE"
echo "========================================"
