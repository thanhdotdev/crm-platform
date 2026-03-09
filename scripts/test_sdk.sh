#!/bin/bash

# Configuration
BASE_URL="http://localhost:8080"
TENANT_NAME="SDK Test Tenant $(date +%s)"
ADMIN_EMAIL="admin_$(date +%s)@test.com"

echo "==========================================="
echo "CRM SDK E2E Test"
echo "==========================================="

echo "1. Registering Tenant to get API Key..."
# Step 1: Register Tenant
RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/tenants" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "'"${TENANT_NAME}"'",
    "slug": "test-tenant-'"$(date +%s)"'",
    "admin_email": "'"${ADMIN_EMAIL}"'",
    "admin_password": "password123"
  }')

# Extract Tenant ID
TENANT_ID=$(echo $RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)

if [ -z "$TENANT_ID" ]; then
    echo "❌ Failed to create tenant: $RESPONSE"
    exit 1
fi

echo "✅ Created Tenant: $TENANT_ID"

# Step 2: Generate Server API Key
KEY_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/tenants/${TENANT_ID}/api-keys" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Server Key",
    "key_type": "server"
  }')

API_KEY=$(echo $KEY_RESPONSE | grep -o '"raw_key":"[^"]*' | head -1 | cut -d'"' -f4)

if [ -z "$API_KEY" ]; then
    echo "❌ Failed to create API key: $KEY_RESPONSE"
    exit 1
fi

echo "✅ Created API Key: ${API_KEY:0:10}..."

echo ""
echo "==========================================="
echo "2. Running Go SDK Test..."
echo "==========================================="
cd sdk/go
export CRM_API_KEY="$API_KEY"
export CRM_ENDPOINT="$BASE_URL"
go run example/main.go

if [ $? -ne 0 ]; then
    echo "❌ Go SDK Test Failed"
    exit 1
fi

echo ""
echo "==========================================="
echo "3. Running JS SDK Test..."
echo "==========================================="
cd ../js
# npx tsx reads the CRM_API_KEY from environment directly inside test.ts
npx tsx example/test.ts

if [ $? -ne 0 ]; then
    echo "❌ JS SDK Test Failed"
    exit 1
fi

echo ""
echo "🎉 ALL SDK TESTS PASSED SUCCESSFULLYY"
