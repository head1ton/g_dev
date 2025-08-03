#!/bin/bash

# 테스트 실행 시 로깅을 활성화하는 스크립트
# 사용법: ./scripts/test_with_logging.sh [테스트 패키지 경로]

set -e

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 테스트 로깅 활성화 스크립트${NC}"
echo "=================================="

# 환경변수 파일 로드
if [ -f "test.env" ]; then
    echo -e "${YELLOW}📄 test.env 파일을 로드합니다...${NC}"
    export $(cat test.env | grep -v '^#' | xargs)
else
    echo -e "${RED} test.env 파일을 찾을 수 없습니다.${NC}"
    exit 1
fi

# 로깅 설정 확인
echo -e "${GREEN} 로깅 설정:${NC}"
echo "  - DATABASE_LOG_LEVEL: $DATABASE_LOG_LEVEL"
echo "  - DATABASE_DEBUG: $DATABASE_DEBUG"
echo "  - LOG_LEVEL: $LOG_LEVEL"
echo "  - LOG_FORMAT: $LOG_FORMAT"

# 테스트 패키지 경로 설정
TEST_PATH=${1:-"./..."}

echo -e "${YELLOW} 테스트를 실행합니다: $TEST_PATH${NC}"
echo "=================================="

# 테스트 실행 (상세 출력 + 로깅)
go test -v $TEST_PATH \
    -timeout=30s \
    -run=. \
    -count=1 \
    -failfast=false \
    -parallel=1

echo -e "${GREEN} 테스트 완료${NC}"