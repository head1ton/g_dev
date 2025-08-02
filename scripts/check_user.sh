#!/bin/bash

# MySQL 데이터베이스에서 사용자 정보를 확인하는 스크립트

echo "🔍 데이터베이스 사용자 정보 확인"
echo "=================================="

# MySQL 연결 정보 (환경변수에서 가져오기, 기본값 설정)
MYSQL_HOST=${MYSQL_HOST:-"127.0.0.1"}
MYSQL_PORT=${MYSQL_PORT:-"3306"}
MYSQL_USER=${MYSQL_USER:-"root"}
MYSQL_PASSWORD=${MYSQL_PASSWORD:-"qwer1234!"}
MYSQL_DATABASE=${MYSQL_DATABASE:-"g_dev"}

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

# MySQL 연결 테스트
echo "🔍 MySQL 연결 테스트 중..."
if mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "SELECT 1;" "$MYSQL_DATABASE" > /dev/null 2>&1; then
    print_success "MySQL 연결 성공"
else
    print_error "MySQL 연결 실패"
    print_info "연결 정보:"
    print_info "  호스트: $MYSQL_HOST"
    print_info "  포트: $MYSQL_PORT"
    print_info "  사용자: $MYSQL_USER"
    print_info "  데이터베이스: $MYSQL_DATABASE"
    print_info ""
    print_info "환경변수 설정 예시:"
    print_info "  export MYSQL_HOST=127.0.0.1"
    print_info "  export MYSQL_PORT=3306"
    print_info "  export MYSQL_USER=root"
    print_info "  export MYSQL_PASSWORD=qwer1234!"
    print_info "  export MYSQL_DATABASE=g_dev"
    exit 1
fi

echo ""
echo "📊 사용자 테이블 구조:"
mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE" -e "DESCRIBE users;" 2>/dev/null || {
    print_error "users 테이블을 찾을 수 없습니다"
    print_info "데이터베이스 마이그레이션이 필요할 수 있습니다"
    exit 1
}

echo ""
echo "👥 모든 사용자 목록:"
USER_COUNT=$(mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE" -e "SELECT COUNT(*) as user_count FROM users;" 2>/dev/null | tail -n 1)
if [ "$USER_COUNT" -gt 0 ]; then
    mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE" -e "SELECT id, username, email, nickname, status, role, email_verified, created_at FROM users ORDER BY created_at DESC;"
    print_info "총 사용자 수: $USER_COUNT"
else
    print_warning "사용자가 없습니다"
fi

echo ""
echo "🔐 특정 사용자 상세 정보 (debuguser):"
DEBUG_USER=$(mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE" -e "SELECT id, username, email, nickname, status, role, email_verified, created_at FROM users WHERE username = 'debuguser';" 2>/dev/null)
if [ -n "$DEBUG_USER" ] && [ "$DEBUG_USER" != "id	username	email	nickname	status	role	email_verified	created_at" ]; then
    echo "$DEBUG_USER"
    print_success "debuguser를 찾았습니다"
else
    print_warning "debuguser를 찾을 수 없습니다"
fi

echo ""
echo "🔍 최근 생성된 사용자 5명:"
mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE" -e "SELECT id, username, email, nickname, status, role, email_verified, created_at FROM users ORDER BY created_at DESC LIMIT 5;" 2>/dev/null || print_error "사용자 조회 실패"

echo ""
echo "📈 사용자 통계:"
mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE" -e "
SELECT
    COUNT(*) as total_users,
    SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END) as active_users,
    SUM(CASE WHEN email_verified = 1 THEN 1 ELSE 0 END) as verified_users,
    SUM(CASE WHEN role = 'admin' THEN 1 ELSE 0 END) as admin_users
FROM users;" 2>/dev/null || print_error "통계 조회 실패"

echo ""
print_success "데이터베이스 확인 완료!"