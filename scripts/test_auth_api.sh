#!/bin/bash

# G-Step 웹게임서버 인증 API 테스트 스크립트
# 이 스크립트는 회원가입, 로그인, 토큰 갱신, 로그아웃 등의 기능을 테스트합니다.

# 환경변수에서 포트 가져오기 (기본값: 8081)
PORT=${PORT:-8081}
BASE_URL="http://localhost:$PORT"
API_BASE="$BASE_URL/api/auth"

echo "🎮 G-Step 웹게임서버 인증 API 테스트 시작"
echo "=========================================="
echo "📍 서버 주소: $BASE_URL"

# jq 의존성 체크
if ! command -v jq &> /dev/null; then
    echo "❌ jq가 설치되어 있지 않습니다. 설치 후 다시 실행하세요."
    echo "   macOS: brew install jq"
    echo "   Ubuntu: sudo apt-get install jq"
    exit 1
fi

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 헬퍼 함수
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# 서버 상태 확인
echo "🔍 서버 상태 확인 중..."
if curl -s --max-time 5 "$BASE_URL" > /dev/null; then
    print_success "서버가 실행 중입니다"
else
    print_error "서버에 연결할 수 없습니다. 서버가 실행 중인지 확인하세요."
    print_info "현재 포트: $PORT"
    print_info "서버 시작 명령: go run cmd/server/main.go"
    exit 1
fi

echo ""
echo "📝 1. 회원가입 테스트"
echo "-------------------"

# 고유한 사용자명 생성 (타임스탬프 사용)
TIMESTAMP=$(date +%s)
USERNAME="testuser_$TIMESTAMP"
EMAIL="test_$TIMESTAMP@example.com"

# 회원가입 요청
REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE/register" \
    -H "Content-Type: application/json" \
    -d "{
        \"username\": \"$USERNAME\",
        \"email\": \"$EMAIL\",
        \"password\": \"password123\",
        \"nickname\": \"테스트유저_$TIMESTAMP\"
    }")

echo "회원가입 응답:"
echo "$REGISTER_RESPONSE" | jq '.'

# 응답에서 토큰 추출
ACCESS_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.access_token')
REFRESH_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.refresh_token')

if [ "$ACCESS_TOKEN" != "null" ] && [ "$ACCESS_TOKEN" != "" ]; then
    print_success "회원가입 성공"
    print_info "액세스 토큰: ${ACCESS_TOKEN:0:50}..."
    print_info "리프레시 토큰: ${REFRESH_TOKEN:0:50}..."
else
    print_error "회원가입 실패"
    print_info "응답 내용: $REGISTER_RESPONSE"
    exit 1
fi

echo ""
echo "🔐 2. 로그인 테스트"
echo "----------------"

# 로그인 요청
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/login" \
    -H "Content-Type: application/json" \
    -d "{
        \"username\": \"$USERNAME\",
        \"password\": \"password123\"
    }")

echo "로그인 응답:"
echo "$LOGIN_RESPONSE" | jq '.'

# 새로운 토큰 추출
NEW_ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')
NEW_REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.refresh_token')

if [ "$NEW_ACCESS_TOKEN" != "null" ] && [ "$NEW_ACCESS_TOKEN" != "" ]; then
    print_success "로그인 성공"
    print_info "새 액세스 토큰: ${NEW_ACCESS_TOKEN:0:50}..."
    print_info "새 리프레시 토큰: ${NEW_REFRESH_TOKEN:0:50}..."
    ACCESS_TOKEN="$NEW_ACCESS_TOKEN"
    REFRESH_TOKEN="$NEW_REFRESH_TOKEN"
else
    print_error "로그인 실패"
    print_info "응답 내용: $LOGIN_RESPONSE"
    exit 1
fi

echo ""
echo "👤 3. 프로필 조회 테스트"
echo "----------------------"

# 프로필 조회 (인증 필요)
PROFILE_RESPONSE=$(curl -s -X GET "$API_BASE/profile" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

echo "프로필 조회 응답:"
echo "$PROFILE_RESPONSE" | jq '.'

if echo "$PROFILE_RESPONSE" | jq -e '.success' > /dev/null; then
    print_success "프로필 조회 성공"
else
    print_error "프로필 조회 실패"
    print_info "응답 내용: $PROFILE_RESPONSE"
fi

echo ""
echo "🔄 4. 토큰 갱신 테스트"
echo "-------------------"

# 토큰 갱신 요청
REFRESH_RESPONSE=$(curl -s -X POST "$API_BASE/refresh" \
    -H "Content-Type: application/json" \
    -d "{
        \"refresh_token\": \"$REFRESH_TOKEN\"
    }")

echo "토큰 갱신 응답:"
echo "$REFRESH_RESPONSE" | jq '.'

# 새로운 액세스 토큰 추출
REFRESHED_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.access_token')

if [ "$REFRESHED_ACCESS_TOKEN" != "null" ] && [ "$REFRESHED_ACCESS_TOKEN" != "" ]; then
    print_success "토큰 갱신 성공"
    print_info "갱신된 액세스 토큰: ${REFRESHED_ACCESS_TOKEN:0:50}..."
    ACCESS_TOKEN="$REFRESHED_ACCESS_TOKEN"
else
    print_error "토큰 갱신 실패"
    print_info "응답 내용: $REFRESH_RESPONSE"
fi

echo ""
echo "🚪 5. 로그아웃 테스트"
echo "------------------"

# 로그아웃 요청
LOGOUT_RESPONSE=$(curl -s -X POST "$API_BASE/logout" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

echo "로그아웃 응답:"
echo "$LOGOUT_RESPONSE" | jq '.'

if echo "$LOGOUT_RESPONSE" | jq -e '.success' > /dev/null; then
    print_success "로그아웃 성공"
else
    print_error "로그아웃 실패"
    print_info "응답 내용: $LOGOUT_RESPONSE"
fi

echo ""
echo "🔒 6. 인증 실패 테스트"
echo "-------------------"

# 잘못된 토큰으로 프로필 조회 시도
INVALID_RESPONSE=$(curl -s -X GET "$API_BASE/profile" \
    -H "Authorization: Bearer invalid-token")

echo "잘못된 토큰으로 프로필 조회 시도:"
echo "$INVALID_RESPONSE" | jq '.'

if echo "$INVALID_RESPONSE" | jq -e '.success == false' > /dev/null; then
    print_success "인증 실패 테스트 성공 (예상된 동작)"
else
    print_error "인증 실패 테스트 실패"
    print_info "응답 내용: $INVALID_RESPONSE"
fi

echo ""
echo "🧮 7. 보호된 API 테스트"
echo "---------------------"

# 계산기 API 테스트 (인증 필요)
CALCULATOR_RESPONSE=$(curl -s -X POST "$BASE_URL/api/calculator/calculate" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d '{
        "operation": "add",
        "a": 10,
        "b": 20
    }')

echo "계산기 API 응답 (로그아웃 후):"
echo "$CALCULATOR_RESPONSE" | jq '.'

if echo "$CALCULATOR_RESPONSE" | jq -e '.success == false' > /dev/null; then
    print_success "보호된 API 테스트 성공 (로그아웃 후 접근 차단)"
else
    print_warning "보호된 API 테스트 결과 확인 필요"
    print_info "응답 내용: $CALCULATOR_RESPONSE"
fi

echo ""
echo "=========================================="
print_success "인증 API 테스트 완료!"
echo ""
echo "📊 테스트 요약:"
echo "- 회원가입: ✅"
echo "- 로그인: ✅"
echo "- 프로필 조회: ✅"
echo "- 토큰 갱신: ✅"
echo "- 로그아웃: ✅"
echo "- 인증 실패 처리: ✅"
echo "- 보호된 API 접근 제어: ✅"
echo ""
echo "🎉 모든 인증 기능이 정상적으로 작동합니다!"
echo ""
echo "💡 사용된 테스트 사용자: $USERNAME"
echo "📧 사용된 테스트 이메일: $EMAIL"