#!/usr/bin/env bash
set -euo pipefail

API_BASE="${API_BASE:-http://127.0.0.1:8080/api/v1}"
BUYER_TOKEN="${BUYER_TOKEN:-}"
SELLER_TOKEN="${SELLER_TOKEN:-}"
WAIT_SECONDS="${WAIT_SECONDS:-8}"

if [[ -z "$BUYER_TOKEN" || -z "$SELLER_TOKEN" ]]; then
  echo "Usage: BUYER_TOKEN=... SELLER_TOKEN=... bash scripts/verify-bidding-flow.sh"
  echo "Optional: API_BASE=http://127.0.0.1:8080/api/v1 WAIT_SECONDS=8"
  echo "BUYER_TOKEN and SELLER_TOKEN only mean initiator/quoter here; they must belong to different certified enterprises."
  exit 2
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

json_get() {
  node -e "const fs=require('fs'); const obj=JSON.parse(fs.readFileSync(process.argv[1],'utf8')); const path=process.argv[2].split('.'); let cur=obj; for (const key of path) cur=cur?.[key]; if (cur === undefined || cur === null) process.exit(3); process.stdout.write(String(cur));" "$1" "$2"
}

assert_success() {
  local file="$1"
  local step="$2"
  local code
  if ! code="$(json_get "$file" code 2>/dev/null)"; then
    echo "   $step failed: response is not a standard API envelope"
    cat "$file"
    exit 1
  fi
  if [[ "$code" != "0" ]]; then
    echo "   $step failed: API code=$code"
    cat "$file"
    exit 1
  fi
}

request() {
  local method="$1"
  local url="$2"
  local token="$3"
  local body="${4:-}"
  local output="$5"

  if [[ -n "$body" ]]; then
    curl -sS --max-time 20 -X "$method" "$url" \
      -H "Authorization: Bearer $token" \
      -H "Content-Type: application/json" \
      -d "$body" \
      -o "$output"
  else
    curl -sS --max-time 20 -X "$method" "$url" \
      -H "Authorization: Bearer $token" \
      -H "Content-Type: application/json" \
      -o "$output"
  fi
}

future_time="$(node -e "console.log(new Date(Date.now() + (${WAIT_SECONDS} - 2) * 1000).toISOString())")"
title="验收竞价-$(date +%Y%m%d%H%M%S)"

create_body="$(node -e "console.log(JSON.stringify({title: process.argv[1], category: '热轧板卷', quantity: 120, unit: '吨', start_price: 4200, delivery_region: '上海指定仓库', ended_at: process.argv[2]}))" "$title" "$future_time")"

echo "1. 买家发起竞价"
request POST "$API_BASE/biddings" "$BUYER_TOKEN" "$create_body" "$tmpdir/create.json"
assert_success "$tmpdir/create.json" "create bidding"
bidding_id="$(json_get "$tmpdir/create.json" data.id)"
echo "   bidding_id=$bidding_id"

echo "2. 卖家提交报价"
quote_body='{"unit_price":3980,"quantity":100,"delivery_days":5,"remark":"验收报价"}'
request POST "$API_BASE/biddings/$bidding_id/quotes" "$SELLER_TOKEN" "$quote_body" "$tmpdir/quote.json"
assert_success "$tmpdir/quote.json" "submit quote"
quote_id="$(json_get "$tmpdir/quote.json" data.quotes.0.id)"
echo "   quote_id=$quote_id"

echo "3. 验证竞价中不能定标"
request POST "$API_BASE/biddings/$bidding_id/award" "$BUYER_TOKEN" "{\"quote_id\":$quote_id}" "$tmpdir/award_before_deadline.json"
award_before_code="$(json_get "$tmpdir/award_before_deadline.json" code)"
if [[ "$award_before_code" == "0" ]]; then
  echo "   expected award before deadline to fail, but it succeeded"
  cat "$tmpdir/award_before_deadline.json"
  exit 1
fi
echo "   blocked as expected"

echo "4. 等待截止并触发状态校正"
sleep "$WAIT_SECONDS"
request GET "$API_BASE/biddings/$bidding_id" "$BUYER_TOKEN" "" "$tmpdir/detail_pending_award.json"
assert_success "$tmpdir/detail_pending_award.json" "get pending award detail"
status="$(json_get "$tmpdir/detail_pending_award.json" data.status)"
if [[ "$status" != "pending_award" ]]; then
  echo "   expected status=pending_award, got $status"
  cat "$tmpdir/detail_pending_award.json"
  exit 1
fi
echo "   status=$status"

echo "5. 买家定标"
request POST "$API_BASE/biddings/$bidding_id/award" "$BUYER_TOKEN" "{\"quote_id\":$quote_id}" "$tmpdir/award.json"
assert_success "$tmpdir/award.json" "award bidding"
status="$(json_get "$tmpdir/award.json" data.status)"
if [[ "$status" != "awarded" ]]; then
  echo "   expected status=awarded, got $status"
  cat "$tmpdir/award.json"
  exit 1
fi
echo "   status=$status"

echo "6. 买家生成订单"
request POST "$API_BASE/biddings/$bidding_id/orders" "$BUYER_TOKEN" "" "$tmpdir/order.json"
assert_success "$tmpdir/order.json" "create bidding order"
order_id="$(json_get "$tmpdir/order.json" data.order_id)"
order_no="$(json_get "$tmpdir/order.json" data.order_no)"
final_status="$(json_get "$tmpdir/order.json" data.bidding.status)"
if [[ "$final_status" != "order_created" ]]; then
  echo "   expected status=order_created, got $final_status"
  cat "$tmpdir/order.json"
  exit 1
fi
echo "   order_id=$order_id order_no=$order_no status=$final_status"

echo "Bidding flow verified."
