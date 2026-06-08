#!/bin/bash
# ChillCat 服务端接口自测脚本
# 使用方式: bash test_api.sh [base_url]
# 默认 base_url=http://localhost:8080

BASE_URL="${1:-http://localhost:8080}"
PASS=0
FAIL=0

green() { echo -e "\033[32m$1\033[0m"; }
red() { echo -e "\033[31m$1\033[0m"; }
bold() { echo -e "\033[1m$1\033[0m"; }

assert_eq() {
    local expected=$1
    local actual=$2
    local msg=$3
    if [ "$expected" = "$actual" ]; then
        green "  ✅ $msg"
        PASS=$((PASS + 1))
    else
        red "  ❌ $msg (期望: $expected, 实际: $actual)"
        FAIL=$((FAIL + 1))
    fi
}

assert_contain() {
    local haystack=$1
    local needle=$2
    local msg=$3
    if echo "$haystack" | grep -q "$needle"; then
        green "  ✅ $msg"
        PASS=$((PASS + 1))
    else
        red "  ❌ $msg (未包含: $needle)"
        FAIL=$((FAIL + 1))
    fi
}

echo ""
bold "═══════════════════════════════════════════"
bold "   ChillCat API 自测脚本"
bold "   目标: $BASE_URL"
bold "═══════════════════════════════════════════"
echo ""

# =============================================
# 1. 健康检查
# =============================================
bold "📋 1. 健康检查"
echo ""

RESP=$(curl -s "$BASE_URL/health")
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
assert_eq "0" "$CODE" "健康检查返回 code=0"

echo ""

# =============================================
# 2. 注册
# =============================================
bold "📋 2. 注册"
echo ""

TIMESTAMP=$(date +%s)
USERNAME="test_$TIMESTAMP"
EMAIL="test_$TIMESTAMP@chillcat.app"
PASSWORD="123456"

RESP=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USERNAME\",\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
TOKEN=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null)
USER_ID=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['user_id'])" 2>/dev/null)

assert_eq "0" "$CODE" "注册新用户"
assert_contain "$TOKEN" "." "注册返回 Token 有效"

echo "  用户: $USERNAME | ID: $USER_ID"

# 重复注册
RESP=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USERNAME\",\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
assert_eq "20002" "$CODE" "重复注册返回 20002(用户已存在)"

echo ""

# =============================================
# 3. 登录
# =============================================
bold "📋 3. 登录"
echo ""

# 正常登录
RESP=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
LOGIN_TOKEN=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null)
assert_eq "0" "$CODE" "正常登录返回 code=0"
assert_contain "$LOGIN_TOKEN" "." "登录返回 Token 有效"

# 密码错误
RESP=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USERNAME\",\"password\":\"wrong_pass\"}")
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
assert_eq "20003" "$CODE" "密码错误返回 20003"

# 用户不存在
RESP=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"not_exist_user\",\"password\":\"123456\"}")
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
assert_eq "20001" "$CODE" "用户不存在返回 20001"

echo ""

# =============================================
# 4. 获取用户信息（需认证）
# =============================================
bold "📋 4. 用户信息"
echo ""

# 带 Token 请求
RESP=$(curl -s "$BASE_URL/api/v1/user/profile" \
    -H "Authorization: Bearer $LOGIN_TOKEN")
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
USERNAME_RESP=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['username'])" 2>/dev/null)
assert_eq "0" "$CODE" "获取用户信息返回 code=0"
assert_eq "$USERNAME" "$USERNAME_RESP" "用户名正确"

# 不带 Token
RESP=$(curl -s "$BASE_URL/api/v1/user/profile")
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
assert_eq "10002" "$CODE" "未认证请求返回 10002(未授权)"

echo ""

# =============================================
# 5. 会员信息
# =============================================
bold "📋 5. 会员信息"
echo ""

# 非会员
RESP=$(curl -s "$BASE_URL/api/v1/member/info" \
    -H "Authorization: Bearer $LOGIN_TOKEN")
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
STATUS=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['status'])" 2>/dev/null)
IS_VALID=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['is_valid'])" 2>/dev/null)
assert_eq "0" "$CODE" "查询会员信息返回 code=0"
assert_eq "none" "$STATUS" "非会员状态为 none"
assert_eq "false" "$IS_VALID" "非会员 is_valid 为 false"

echo ""

# =============================================
# 6. 商品列表
# =============================================
bold "📋 6. 商品列表"
echo ""

RESP=$(curl -s "$BASE_URL/api/v1/member/products" \
    -H "Authorization: Bearer $LOGIN_TOKEN")
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
PRODUCT_COUNT=$(echo "$RESP" | python3 -c "import sys,json; print(len(json.load(sys.stdin)['data']))" 2>/dev/null)
assert_eq "0" "$CODE" "商品列表返回 code=0"
assert_eq "4" "$PRODUCT_COUNT" "返回 4 个商品"

echo ""

# =============================================
# 7. 权益列表
# =============================================
bold "📋 7. 权益列表"
echo ""

RESP=$(curl -s "$BASE_URL/api/v1/member/privileges" \
    -H "Authorization: Bearer $LOGIN_TOKEN")
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
PRIVILEGE_COUNT=$(echo "$RESP" | python3 -c "import sys,json; print(len(json.load(sys.stdin)['data']))" 2>/dev/null)
assert_eq "0" "$CODE" "权益列表返回 code=0"
assert_eq "5" "$PRIVILEGE_COUNT" "返回 5 个权益"

echo ""

# =============================================
# 8. 购买会员
# =============================================
bold "📋 8. 购买会员"
echo ""

# 购买月度会员
RESP=$(curl -s -X POST "$BASE_URL/api/v1/member/purchase" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $LOGIN_TOKEN" \
    -d '{"product_id": "monthly_001"}')
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
ORDER_NO=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['order_no'])" 2>/dev/null)
MEMBER_TYPE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['member_info']['member_type'])" 2>/dev/null)
assert_eq "0" "$CODE" "购买月度会员返回 code=0"
assert_contain "$ORDER_NO" "CC" "订单号以 CC 开头"
assert_eq "monthly" "$MEMBER_TYPE" "会员类型为 monthly"

echo "  订单号: $ORDER_NO"

# 购买后查询会员信息
RESP=$(curl -s "$BASE_URL/api/v1/member/info" \
    -H "Authorization: Bearer $LOGIN_TOKEN")
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
STATUS=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['status'])" 2>/dev/null)
IS_VALID=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['is_valid'])" 2>/dev/null)
assert_eq "active" "$STATUS" "购买后会员状态为 active"
assert_eq "true" "$IS_VALID" "购买后 is_valid 为 true"

# 购买不存在的商品
RESP=$(curl -s -X POST "$BASE_URL/api/v1/member/purchase" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $LOGIN_TOKEN" \
    -d '{"product_id": "not_exist"}')
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
assert_eq "30004" "$CODE" "购买不存在商品返回 30004"

echo ""

# =============================================
# 9. 购买历史
# =============================================
bold "📋 9. 购买历史"
echo ""

RESP=$(curl -s "$BASE_URL/api/v1/member/history?page=1&page_size=10" \
    -H "Authorization: Bearer $LOGIN_TOKEN")
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
TOTAL=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total'])" 2>/dev/null)
assert_eq "0" "$CODE" "购买历史返回 code=0"
assert_eq "1" "$TOTAL" "购买历史有 1 条记录"

echo ""

# =============================================
# 10. 购买永久会员（另一个用户）
# =============================================
bold "📋 10. 购买永久会员"
echo ""

# 注册一个新用户
TIMESTAMP2=$(date +%s)000
USERNAME2="perm_$TIMESTAMP2"
RESP=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USERNAME2\",\"email\":\"$USERNAME2@chillcat.app\",\"password\":\"123456\"}")
TOKEN2=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null)

# 购买永久会员
RESP=$(curl -s -X POST "$BASE_URL/api/v1/member/purchase" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN2" \
    -d '{"product_id": "permanent_001"}')
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
MEMBER_TYPE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['member_info']['member_type'])" 2>/dev/null)
IS_VALID=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['member_info']['is_valid'])" 2>/dev/null)
REMAINING=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['member_info']['remaining_days'])" 2>/dev/null)
assert_eq "0" "$CODE" "购买永久会员返回 code=0"
assert_eq "permanent" "$MEMBER_TYPE" "会员类型为 permanent"
assert_eq "true" "$IS_VALID" "永久会员 is_valid 为 true"
assert_eq "-1" "$REMAINING" "永久会员 remaining_days 为 -1"

echo ""

# =============================================
# 11. Token 刷新
# =============================================
bold "📋 11. Token 刷新"
echo ""

RESP=$(curl -s -X POST "$BASE_URL/api/v1/auth/refresh" \
    -H "Authorization: Bearer $LOGIN_TOKEN")
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
NEW_TOKEN=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null)
assert_eq "0" "$CODE" "刷新Token返回 code=0"
assert_contain "$NEW_TOKEN" "." "刷新Token有效"

echo ""

# =============================================
# 12. Feed 内容流
# =============================================
bold "📋 12. Feed 内容流"
echo ""

RESP=$(curl -s "$BASE_URL/api/v1/feeds?page=1&page_size=5" \
    -H "Authorization: Bearer $LOGIN_TOKEN")
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
FEED_COUNT=$(echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin)['data']; print(len(d['list']))" 2>/dev/null)
assert_eq "0" "$CODE" "Feed列表返回 code=0"
assert_contain "$RESP" "list" "Feed响应包含 list 字段"
echo "  返回 $FEED_COUNT 条内容"

echo ""

# =============================================
# 13. 搜索
# =============================================
bold "📋 13. 搜索"
echo ""

RESP=$(curl -s "$BASE_URL/api/v1/search?q=ChillCat" \
    -H "Authorization: Bearer $LOGIN_TOKEN")
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
assert_eq "0" "$CODE" "搜索返回 code=0"

echo ""

# =============================================
# 14. 消息列表 & 未读数
# =============================================
bold "📋 14. 消息中心"
echo ""

RESP=$(curl -s "$BASE_URL/api/v1/messages?page=1&page_size=10" \
    -H "Authorization: Bearer $LOGIN_TOKEN")
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
assert_eq "0" "$CODE" "消息列表返回 code=0"

RESP=$(curl -s "$BASE_URL/api/v1/messages/unread" \
    -H "Authorization: Bearer $LOGIN_TOKEN")
CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['code'])")
UNREAD=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['count'])" 2>/dev/null)
assert_eq "0" "$CODE" "未读消息数返回 code=0"
echo "  未读消息: $UNREAD 条"

echo ""

# =============================================
# 汇总
# =============================================
bold "═══════════════════════════════════════════"
bold "   测试结果"
bold "═══════════════════════════════════════════"
echo ""
green "  通过: $PASS"
if [ $FAIL -gt 0 ]; then
    red "  失败: $FAIL"
    echo ""
    red "  ⚠️  有测试未通过，请检查服务端日志"
else
    green "  失败: $FAIL"
    echo ""
    green "  🎉 所有接口测试通过！"
fi
echo ""
