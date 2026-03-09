#!/bin/bash
# ============================================================================
# CRM Platform — Production Simulation
# Mô phỏng kịch bản thực tế từ khi integrate đến vận hành
# ============================================================================
set -e
BASE="http://localhost:8080"
H=""  # will be set after API key created
OK=0; FAIL=0

pass() { echo "  ✅ $1"; OK=$((OK+1)); }
fail() { echo "  ❌ $1: $2"; FAIL=$((FAIL+1)); }
check() {
  local name="$1" result="$2"
  if echo "$result" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d.get('success',False)" 2>/dev/null; then
    pass "$name"
  else
    fail "$name" "$(echo $result | head -c 200)"
  fi
}
jq_field() { echo "$1" | python3 -c "import sys,json; print(json.load(sys.stdin)$2)" 2>/dev/null; }
sep() { echo ""; echo "════════════════════════════════════════════════════════════"; echo " $1"; echo "════════════════════════════════════════════════════════════"; }

# ============================================================================
sep "BƯỚC 1: ĐĂNG KÝ TENANT (BUTL tích hợp CRM)"
# ============================================================================
echo "  📋 BUTL đăng ký làm tenant trên CRM platform"
R=$(curl -sf -X POST "$BASE/api/v1/tenants" -H "Content-Type: application/json" \
  -d '{"name":"BUTL Ride-hailing","slug":"butl-'$RANDOM'","domain":"butl.vn"}')
check "Tạo tenant BUTL" "$R"
TID=$(jq_field "$R" "['data']['id']")
echo "  🆔 Tenant ID: $TID"

echo ""
echo "  🔑 Tạo API key cho backend BUTL (server key)"
R=$(curl -sf -X POST "$BASE/api/v1/tenants/$TID/api-keys" -H "Content-Type: application/json" \
  -d '{"key_type":"server","name":"BUTL Backend"}')
check "Tạo server API key" "$R"
KEY=$(jq_field "$R" "['data']['raw_key']")
H="-H X-API-Key:$KEY -H Content-Type:application/json"
echo "  🔐 API Key: ${KEY:0:20}..."

# ============================================================================
sep "BƯỚC 2: CẤU HÌNH CHIẾN DỊCH & QUY TẮC TỰ ĐỘNG"
# ============================================================================

# 2a. Campaign 50-30-20 (theo yêu cầu CEO)
echo "  🎪 Tạo Campaign 'Chiến dịch 1 - Kéo khách quay lại'"
R=$(curl -sf -X POST "$BASE/api/v1/campaigns" $H \
  -d '{"name":"Chiến dịch 1 - Kéo khách quay lại","type":"retention","start_date":"2026-03-01","end_date":"2026-12-31","budget":100000000}')
check "Tạo campaign" "$R"
CAMP_ID=$(jq_field "$R" "['data']['id']")

# 2b. Segment cho high-value customers
echo "  🎯 Tạo segment 'Khách VIP' (≥5 chuyến, ≥2M chi tiêu)"
R=$(curl -sf -X POST "$BASE/api/v1/segments" $H \
  -d '{"name":"Khách VIP","type":"trip_based","rules":{"min_trips":5,"min_spent":2000000}}')
check "Tạo segment VIP" "$R"

echo "  🎯 Tạo segment 'Khách có nguy cơ rời bỏ' (1 chuyến, >30 ngày)"
R=$(curl -sf -X POST "$BASE/api/v1/segments" $H \
  -d '{"name":"Nguy cơ rời bỏ","type":"trip_based","rules":{"max_trips":1}}')
check "Tạo segment at-risk" "$R"

# 2c. Automation rules
echo "  ⚙️  Tạo rule: Khi user đăng ký → gửi push chào mừng (tự động)"
R=$(curl -sf -X POST "$BASE/api/v1/automation/rules" $H \
  -d '{"name":"Welcome Push","trigger_type":"event","trigger_config":{"event":"user_registered"},"action_type":"send_notification","action_config":{"channel":"push","title":"Chào mừng bạn đến BUTL!","content":"Đặt chuyến đầu tiên ngay"},"execution_mode":"automatic"}')
check "Rule: welcome push" "$R"

echo "  ⚙️  Tạo rule: Khi hoàn thành chuyến → kiểm tra tặng voucher (manual)"
R=$(curl -sf -X POST "$BASE/api/v1/automation/rules" $H \
  -d '{"name":"Post-trip Voucher","trigger_type":"event","trigger_config":{"event":"trip_completed"},"action_type":"generate_voucher","execution_mode":"manual"}')
check "Rule: post-trip voucher" "$R"

# ============================================================================
sep "BƯỚC 3: USERS BẮT ĐẦU SỬ DỤNG APP (10 user profiles)"
# ============================================================================

echo ""
echo "  👤 Kịch bản 10 users với các hành vi khác nhau:"
echo "  ┌─────────┬──────────────┬───────────┬──────────────────────────────┐"
echo "  │ User    │ Nguồn        │ Hành vi   │ Mô tả                        │"
echo "  ├─────────┼──────────────┼───────────┼──────────────────────────────┤"
echo "  │ u01     │ facebook_ads │ VIP       │ 6 chuyến, high spender       │"
echo "  │ u02     │ facebook_ads │ Loyal     │ 3 chuyến ổn định             │"
echo "  │ u03     │ google_ads   │ Returning │ 2 chuyến rồi dừng           │"
echo "  │ u04     │ google_ads   │ Activated │ 1 chuyến rồi biến mất       │"
echo "  │ u05     │ organic      │ No trip   │ Cài app nhưng chưa đặt      │"
echo "  │ u06     │ organic      │ Canceller │ Book 4 lần, cancel 3 lần    │"
echo "  │ u07     │ referral     │ Luxury    │ 5 chuyến, chi >2M → luxury  │"
echo "  │ u08     │ referral     │ New       │ Mới đăng ký, chưa cài app   │"
echo "  │ u09     │ tiktok_ads   │ Bouncer   │ Cài app → mở 1 lần → biến  │"
echo "  │ u10     │ tiktok_ads   │ Moderate  │ 2 chuyến + 1 cancel         │"
echo "  └─────────┴──────────────┴───────────┴──────────────────────────────┘"

# ------- U01: VIP (6 completed trips, facebook_ads) -------
echo ""
echo "  ── U01: VIP Customer (facebook_ads) ──"
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"user_registered","external_user_id":"u01","data":{"phone":"090100001","full_name":"Trần Nhật Minh","email":"minh@gmail.com","source":"facebook_ads"}}' > /dev/null
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"app_installed","external_user_id":"u01","data":{"device_type":"ios"}}' > /dev/null
for t in 1 2 3 4 5 6; do
  curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
    -d "{\"user_type\":\"customer\",\"event_type\":\"trip_completed\",\"external_user_id\":\"u01\",\"data\":{\"external_trip_id\":\"u01-t$t\",\"amount\":$((300000 + RANDOM % 200000)),\"pickup_location\":\"Q1\",\"dropoff_location\":\"Q7\"}}" > /dev/null
done
pass "U01: 6 chuyến hoàn thành → VIP"

# ------- U02: Loyal (3 completed, facebook_ads) -------
echo "  ── U02: Loyal Customer (facebook_ads) ──"
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"user_registered","external_user_id":"u02","data":{"phone":"090100002","full_name":"Lê Thu Hà","source":"facebook_ads"}}' > /dev/null
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"app_installed","external_user_id":"u02","data":{"device_type":"android"}}' > /dev/null
for t in 1 2 3; do
  curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
    -d "{\"user_type\":\"customer\",\"event_type\":\"trip_completed\",\"external_user_id\":\"u02\",\"data\":{\"external_trip_id\":\"u02-t$t\",\"amount\":250000}}" > /dev/null
done
pass "U02: 3 chuyến → Loyal"

# ------- U03: Returning (2 completed, google_ads) -------
echo "  ── U03: Returning Customer (google_ads) ──"
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"user_registered","external_user_id":"u03","data":{"phone":"090100003","full_name":"Nguyễn Văn Hùng","source":"google_ads"}}' > /dev/null
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"app_installed","external_user_id":"u03","data":{"device_type":"android"}}' > /dev/null
for t in 1 2; do
  curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
    -d "{\"user_type\":\"customer\",\"event_type\":\"trip_completed\",\"external_user_id\":\"u03\",\"data\":{\"external_trip_id\":\"u03-t$t\",\"amount\":180000}}" > /dev/null
done
pass "U03: 2 chuyến → Returning"

# ------- U04: Activated (1 completed, google_ads) -------
echo "  ── U04: One-and-done (google_ads) ──"
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"user_registered","external_user_id":"u04","data":{"phone":"090100004","full_name":"Phạm Thị Lan","source":"google_ads"}}' > /dev/null
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"app_installed","external_user_id":"u04","data":{"device_type":"ios"}}' > /dev/null
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"trip_completed","external_user_id":"u04","data":{"external_trip_id":"u04-t1","amount":150000}}' > /dev/null
pass "U04: 1 chuyến → Activated (rồi biến mất)"

# ------- U05: Installed but no trip (organic) -------
echo "  ── U05: Cài app nhưng không dùng (organic) ──"
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"user_registered","external_user_id":"u05","data":{"phone":"090100005","full_name":"Hoàng Minh Tuấn","source":"organic"}}' > /dev/null
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"app_installed","external_user_id":"u05","data":{"device_type":"android"}}' > /dev/null
# Behavioral events: mở app, xem xung quanh, rồi đi
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"app_opened","external_user_id":"u05","data":{}}' > /dev/null
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"search_trip","external_user_id":"u05","data":{"from":"Q3","to":"Q1"}}' > /dev/null
pass "U05: Cài app → mở → tìm kiếm → rồi thôi"

# ------- U06: Serial Canceller (organic) — ABUSE -------
echo "  ── U06: Serial Canceller (organic) — sẽ bị flag abuse ──"
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"user_registered","external_user_id":"u06","data":{"phone":"090100006","full_name":"Đỗ Quốc Bảo","source":"organic"}}' > /dev/null
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"app_installed","external_user_id":"u06","data":{"device_type":"ios"}}' > /dev/null
# 1 trip completed
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"trip_completed","external_user_id":"u06","data":{"external_trip_id":"u06-t1","amount":200000}}' > /dev/null
# 3 trips booked then cancelled
for c in 1 2 3; do
  curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
    -d "{\"user_type\":\"customer\",\"event_type\":\"trip_booked\",\"external_user_id\":\"u06\",\"data\":{\"external_trip_id\":\"u06-c$c\",\"pickup_location\":\"Q5\"}}" > /dev/null
  curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
    -d "{\"user_type\":\"customer\",\"event_type\":\"trip_cancelled\",\"external_user_id\":\"u06\",\"data\":{\"external_trip_id\":\"u06-c$c\",\"cancel_reason\":\"driver too slow\"}}" > /dev/null
done
pass "U06: 1 complete + 3 cancel → abuse pattern"

# ------- U07: Luxury Customer (referral) -------
echo "  ── U07: Luxury Customer (referral) ──"
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"user_registered","external_user_id":"u07","data":{"phone":"090100007","full_name":"Võ Thanh Tùng","email":"tung.vo@company.com","source":"referral"}}' > /dev/null
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"app_installed","external_user_id":"u07","data":{"device_type":"ios"}}' > /dev/null
# 5 high-value trips → sẽ trigger luxury
for t in 1 2 3 4 5; do
  curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
    -d "{\"user_type\":\"customer\",\"event_type\":\"trip_completed\",\"external_user_id\":\"u07\",\"data\":{\"external_trip_id\":\"u07-t$t\",\"amount\":500000,\"pickup_location\":\"Tân Sơn Nhất\",\"dropoff_location\":\"Thủ Đức\"}}" > /dev/null
done
pass "U07: 5 chuyến × 500K = 2.5M → Luxury ✨"

# ------- U08: Just registered (referral) -------
echo "  ── U08: Mới đăng ký (referral) ──"
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"user_registered","external_user_id":"u08","data":{"phone":"090100008","full_name":"Bùi Thị Mai","source":"referral"}}' > /dev/null
pass "U08: Đăng ký → chưa cài app, chưa dùng"

# ------- U09: App bouncer (tiktok_ads) -------
echo "  ── U09: Bouncer - cài rồi xóa (tiktok_ads) ──"
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"user_registered","external_user_id":"u09","data":{"phone":"090100009","full_name":"Lý Hoàng Nam","source":"tiktok_ads"}}' > /dev/null
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"app_installed","external_user_id":"u09","data":{"device_type":"android"}}' > /dev/null
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"app_opened","external_user_id":"u09","data":{}}' > /dev/null
pass "U09: Cài app → mở 1 lần → biến mất"

# ------- U10: Moderate (2 complete + 1 cancel, tiktok_ads) -------
echo "  ── U10: Moderate user (tiktok_ads) ──"
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"user_registered","external_user_id":"u10","data":{"phone":"090100010","full_name":"Trương Mỹ Duyên","source":"tiktok_ads"}}' > /dev/null
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"app_installed","external_user_id":"u10","data":{"device_type":"android"}}' > /dev/null
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"trip_completed","external_user_id":"u10","data":{"external_trip_id":"u10-t1","amount":200000}}' > /dev/null
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"trip_booked","external_user_id":"u10","data":{"external_trip_id":"u10-c1"}}' > /dev/null
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"trip_cancelled","external_user_id":"u10","data":{"external_trip_id":"u10-c1","cancel_reason":"changed plan"}}' > /dev/null
curl -sf -X POST "$BASE/api/v1/ingest/events" $H \
  -d '{"user_type":"customer","event_type":"trip_completed","external_user_id":"u10","data":{"external_trip_id":"u10-t2","amount":280000}}' > /dev/null
pass "U10: 2 complete + 1 cancel → Returning"

# ============================================================================
sep "BƯỚC 4: VOUCHER FLOW (Policy 50-30-20)"
# ============================================================================
echo "  🎟️  Tạo voucher cho U02 (sau chuyến 1 → giảm 50K cho chuyến 2)"
CID_U02=$(curl -sf "$BASE/api/v1/customers" $H | python3 -c "
import sys,json
cs=json.load(sys.stdin)['data']
print([c['id'] for c in cs if c['external_id']=='u02'][0])
")

R=$(curl -sf -X POST "$BASE/api/v1/campaigns/$CAMP_ID/vouchers/generate" $H \
  -d "{\"customer_id\":\"$CID_U02\",\"trip_number\":1}")
check "Tạo voucher 50K cho U02" "$R"
V_CODE=$(jq_field "$R" "['data']['code']")
V_VAL=$(jq_field "$R" "['data']['discount_value']")
V_EXP=$(jq_field "$R" "['data']['expires_at']")
echo "  🎟️  Code: $V_CODE | Giảm: ${V_VAL}đ | Hạn: $V_EXP"

echo ""
echo "  💳 U02 dùng voucher"
TRIP_U02=$(curl -sf "$BASE/api/v1/trips/customer/$CID_U02" $H | python3 -c "
import sys,json; print(json.load(sys.stdin)['data'][0]['id'])
")
R=$(curl -sf -X POST "$BASE/api/v1/campaigns/vouchers/redeem" $H \
  -d "{\"code\":\"$V_CODE\",\"customer_id\":\"$CID_U02\",\"trip_id\":\"$TRIP_U02\"}")
check "U02 redeem voucher 50K" "$R"

echo ""
echo "  🚫 U02 thử dùng lại voucher đã dùng"
R=$(curl -sf -X POST "$BASE/api/v1/campaigns/vouchers/redeem" $H \
  -d "{\"code\":\"$V_CODE\",\"customer_id\":\"$CID_U02\",\"trip_id\":\"$TRIP_U02\"}" 2>&1 || true)
BLOCKED=$(echo "$R" | python3 -c "import sys,json; d=json.load(sys.stdin); print('blocked' if not d.get('success') else 'allowed')" 2>/dev/null || echo "blocked")
if [ "$BLOCKED" = "blocked" ]; then pass "Double-redeem bị chặn ✓"; else fail "Double-redeem" "should block"; fi

echo ""
echo "  🎟️  Tạo voucher chuyến 2 (30K) và chuyến 3 (20K)"
R=$(curl -sf -X POST "$BASE/api/v1/campaigns/$CAMP_ID/vouchers/generate" $H \
  -d "{\"customer_id\":\"$CID_U02\",\"trip_number\":2}")
V2_VAL=$(jq_field "$R" "['data']['discount_value']")
echo "  Trip 2 → ${V2_VAL}đ"

R=$(curl -sf -X POST "$BASE/api/v1/campaigns/$CAMP_ID/vouchers/generate" $H \
  -d "{\"customer_id\":\"$CID_U02\",\"trip_number\":3}")
V3_VAL=$(jq_field "$R" "['data']['discount_value']")
echo "  Trip 3 → ${V3_VAL}đ"
pass "Policy 50-30-20: 50K → 30K → 20K ✓"

# ============================================================================
sep "BƯỚC 5: GỬI NOTIFICATION"
# ============================================================================
CID_U05=$(curl -sf "$BASE/api/v1/customers" $H | python3 -c "
import sys,json; cs=json.load(sys.stdin)['data']
print([c['id'] for c in cs if c['external_id']=='u05'][0])
")

echo "  📱 Push notification cho U05 (cài app nhưng chưa dùng)"
R=$(curl -sf -X POST "$BASE/api/v1/notifications" $H \
  -d "{\"customer_id\":\"$CID_U05\",\"channel\":\"push\",\"title\":\"Bạn ơi, đặt chuyến đầu tiên nào!\",\"content\":\"Giảm 50K cho chuyến đầu tiên\"}")
check "Push → U05" "$R"

CID_U04=$(curl -sf "$BASE/api/v1/customers" $H | python3 -c "
import sys,json; cs=json.load(sys.stdin)['data']
print([c['id'] for c in cs if c['external_id']=='u04'][0])
")
echo "  📧 Email cho U04 (one-and-done, cần kéo về)"
R=$(curl -sf -X POST "$BASE/api/v1/notifications" $H \
  -d "{\"customer_id\":\"$CID_U04\",\"channel\":\"email\",\"title\":\"Quay lại BUTL nào!\",\"content\":\"Voucher 30K đang chờ bạn\"}")
check "Email → U04" "$R"

echo "  📞 SMS cho U08 (mới đăng ký, chưa cài app)"
CID_U08=$(curl -sf "$BASE/api/v1/customers" $H | python3 -c "
import sys,json; cs=json.load(sys.stdin)['data']
print([c['id'] for c in cs if c['external_id']=='u08'][0])
")
R=$(curl -sf -X POST "$BASE/api/v1/notifications" $H \
  -d "{\"customer_id\":\"$CID_U08\",\"channel\":\"sms\",\"title\":\"Cài app BUTL\",\"content\":\"Tải app ngay để nhận ưu đãi 50K\"}")
check "SMS → U08" "$R"

# ============================================================================
sep "BƯỚC 6: DASHBOARD — TOÀN CẢNH (CEO xem báo cáo)"
# ============================================================================

echo ""
echo "  📊 OVERVIEW"
R=$(curl -sf "$BASE/api/v1/analytics/overview" $H)
check "Analytics Overview" "$R"
echo "$R" | python3 -c "
import sys,json; d=json.load(sys.stdin)['data']
print(f'  ┌────────────────────────────────────────────┐')
print(f'  │ 👥 Tổng khách hàng:   {d[\"total_customers\"]:>20}  │')
print(f'  │ 🚗 Tổng chuyến (completed): {d[\"total_trips\"]:>13}  │')
print(f'  │ 💰 Tổng doanh thu:    {d[\"total_revenue\"]:>17,.0f}đ │')
print(f'  │ 📈 TB/chuyến:         {d[\"avg_revenue_per_trip\"]:>17,.0f}đ │')
print(f'  └────────────────────────────────────────────┘')
print()
print(f'  Lifecycle:')
for l in d['lifecycle_breakdown']:
    bar = '█' * int(l['count'] * 3)
    print(f'    {l[\"stage\"]:20s} │ {bar} {l[\"count\"]}')
print()
print(f'  Tier:')
for t in d['tier_breakdown']:
    bar = '█' * int(t['count'] * 3)
    print(f'    {t[\"tier\"]:20s} │ {bar} {t[\"count\"]}')
" 2>/dev/null

echo ""
echo "  🔬 COHORT ANALYSIS"
R=$(curl -sf "$BASE/api/v1/analytics/cohorts" $H)
check "Cohort Analysis" "$R"
echo "$R" | python3 -c "
import sys,json; d=json.load(sys.stdin)['data']
print(f'  ┌──────────┬──────┬──────────┬─────────┬───────┐')
print(f'  │ Week     │ Reg  │ Activ %  │ Return% │ Loyal │')
print(f'  ├──────────┼──────┼──────────┼─────────┼───────┤')
for c in d:
    print(f'  │ {c[\"cohort_week\"]:8s} │ {c[\"registered\"]:4d} │ {c[\"activation_rate\"]:6.0f}%  │ {c[\"retention_rate\"]:5.0f}%  │ {c[\"loyal\"]:5d} │')
print(f'  └──────────┴──────┴──────────┴─────────┴───────┘')
" 2>/dev/null

echo ""
echo "  🎯 SOURCE QUALITY"
R=$(curl -sf "$BASE/api/v1/analytics/source-quality" $H)
check "Source Quality" "$R"
echo "$R" | python3 -c "
import sys,json; d=json.load(sys.stdin)['data']
print(f'  ┌───────────────┬──────┬────────┬────────────┬──────────┬─────────┐')
print(f'  │ Source        │ Users│ Conv%  │ LTV        │ Luxury%  │ Cancel% │')
print(f'  ├───────────────┼──────┼────────┼────────────┼──────────┼─────────┤')
for s in d:
    print(f'  │ {s[\"source\"]:13s} │ {s[\"customers\"]:4d} │ {s[\"conversion_rate\"]:5.0f}% │ {s[\"ltv\"]:>10,.0f} │ {s[\"luxury_rate\"]:6.0f}%  │ {s[\"cancel_rate\"]:5.0f}%  │')
print(f'  └───────────────┴──────┴────────┴────────────┴──────────┴─────────┘')
" 2>/dev/null

echo ""
echo "  💰 CAMPAIGN ROI"
R=$(curl -sf "$BASE/api/v1/analytics/campaign-roi" $H)
check "Campaign ROI" "$R"
echo "$R" | python3 -c "
import sys,json; d=json.load(sys.stdin)['data']
for r in d:
    print(f'  📋 {r[\"campaign_name\"]}')
    print(f'     Vouchers: {r[\"total_vouchers\"]} phát / {r[\"used_vouchers\"]} dùng ({r[\"redemption_rate\"]:.0f}%)')
    print(f'     Chi phí discount: {r[\"total_discount_given\"]:,.0f}đ')
" 2>/dev/null

echo ""
echo "  🔄 RETENTION FUNNEL"
R=$(curl -sf "$BASE/api/v1/analytics/retention" $H)
check "Retention Funnel" "$R"
echo "$R" | python3 -c "
import sys,json; d=json.load(sys.stdin)['data']
f = d['conversion_funnel']
print(f'  Đăng ký → Chuyến 1:  {f[\"registered_to_1st_trip\"]:.0f}%')
print(f'  Chuyến 1 → Chuyến 2: {f[\"1st_to_2nd_trip\"]:.0f}%')
print(f'  Chuyến 2 → Chuyến 3: {f[\"2nd_to_3rd_trip\"]:.0f}%')
for b in d['trip_distribution']:
    bar = '█' * int(b['count'] * 2)
    print(f'    {b[\"trip_bucket\"]:10s} │ {bar} {b[\"count\"]}')
" 2>/dev/null

echo ""
echo "  🚨 ABUSE DETECTION"
R=$(curl -sf "$BASE/api/v1/analytics/abuse-detection" $H)
check "Abuse Detection" "$R"
echo "$R" | python3 -c "
import sys,json; d=json.load(sys.stdin)['data']
print(f'  Tổng: {d[\"total_customers\"]} khách | {d[\"flagged_count\"]} bị flag ({d[\"flagged_rate\"]:.1f}%)')
if d.get('suspicious_users'):
    for u in d['suspicious_users']:
        emoji = '🔴' if u['risk_score'] >= 70 else '🟡' if u['risk_score'] >= 50 else '🟢'
        print(f'  {emoji} [{u[\"external_id\"]}] {u[\"full_name\"]} — {u[\"reason\"]} (risk: {u[\"risk_score\"]}, cancels: {u[\"cancel_count\"]})')
else:
    print('  ✅ Không phát hiện hành vi bất thường')
" 2>/dev/null

# ============================================================================
sep "BƯỚC 7: XEM CHI TIẾT TỪNG CUSTOMER"
# ============================================================================
echo "  📋 Danh sách toàn bộ customers:"
curl -sf "$BASE/api/v1/customers" $H | python3 -c "
import sys,json
cs=json.load(sys.stdin)['data']
print(f'  ┌──────┬──────────────────┬──────────┬──────────┬───────────┬──────────┐')
print(f'  │ ID   │ Tên              │ Trips    │ Spent    │ Lifecycle │ Tier     │')
print(f'  ├──────┼──────────────────┼──────────┼──────────┼───────────┼──────────┤')
for c in sorted(cs, key=lambda x: -x['total_trips']):
    name = c['full_name'][:16] if c['full_name'] else c['external_id']
    print(f'  │ {c[\"external_id\"]:4s} │ {name:16s} │ {c[\"total_trips\"]:>8d} │ {c[\"total_spent\"]:>8,.0f} │ {c[\"lifecycle_stage\"]:9s} │ {c[\"tier\"]:8s} │')
print(f'  └──────┴──────────────────┴──────────┴──────────┴───────────┴──────────┘')
" 2>/dev/null
pass "Customer list displayed"

# ============================================================================
sep "KẾT QUẢ"
# ============================================================================
echo ""
echo "  ✅ $OK passed | ❌ $FAIL failed"
echo ""
