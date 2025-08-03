# 🧪 테스트 로깅 가이드

이 문서는 테스트 실행 시 API 로그와 데이터베이스 쿼리 로그를 활성화하는 방법을 설명합니다.

## 📋 목차

1. [개요](#개요)
2. [로깅 설정](#로깅-설정)
3. [테스트 실행 방법](#테스트-실행-방법)
4. [로그 해석](#로그-해석)
5. [문제 해결](#문제-해결)

## 🎯 개요

테스트 실행 시 다음과 같은 로그를 확인할 수 있습니다:

- **HTTP 요청/응답 로그**: API 엔드포인트 호출 시 요청과 응답 정보
- **데이터베이스 쿼리 로그**: GORM이 실행하는 SQL 쿼리
- **애플리케이션 로그**: 비즈니스 로직 실행 과정

## ⚙️ 로깅 설정

### 1. 환경변수 설정

`test.env` 파일을 사용하여 테스트용 로깅 설정을 관리합니다:

```bash
# 데이터베이스 로깅 설정
DATABASE_LOG_LEVEL=4        # 0: Silent, 1: Error, 2: Warn, 3: Info, 4: Debug
DATABASE_DEBUG=true         # GORM 디버그 모드 활성화

# 일반 로깅 설정
LOG_LEVEL=debug            # 로그 레벨 설정
LOG_FORMAT=json           # 로그 형식 (json 또는 text)
```

### 2. 로깅 레벨 설명

| 레벨 | 값 | 설명 |
|------|----|----|
| Silent | 0 | 로그 출력 안함 |
| Error | 1 | 에러만 출력 |
| Warn | 2 | 경고와 에러 출력 |
| Info | 3 | 정보, 경고, 에러 출력 |
| Debug | 4 | 모든 로그 출력 (쿼리 포함) |

## 🚀 테스트 실행 방법

### 1. 자동 스크립트 사용 (권장)

```bash
# 전체 테스트 실행
./scripts/test_with_logging.sh

# 특정 패키지 테스트 실행
./scripts/test_with_logging.sh ./internal/handler

# 특정 테스트 함수 실행
./scripts/test_with_logging.sh ./internal/handler -run TestInventoryHandler
```

### 2. 수동 환경변수 설정

```bash
# 환경변수 설정
export DATABASE_LOG_LEVEL=4
export DATABASE_DEBUG=true
export LOG_LEVEL=debug

# 테스트 실행
go test -v ./internal/handler
```

### 3. 테스트 실행 시 직접 설정

```bash
# 한 번에 실행
DATABASE_LOG_LEVEL=4 DATABASE_DEBUG=true LOG_LEVEL=debug go test -v ./internal/handler
```

## 📊 로그 해석

### HTTP 요청 로그 예시

```
[HTTP] POST /api/auth/login - 200 - 45.2ms
```

- `POST`: HTTP 메서드
- `/api/auth/login`: 요청 경로
- `200`: HTTP 상태 코드
- `45.2ms`: 요청 처리 시간

### 데이터베이스 쿼리 로그 예시

```
[2024-01-01 12:00:00] [INFO] SELECT * FROM users WHERE username = 'testuser' LIMIT 1
[2024-01-01 12:00:01] [INFO] INSERT INTO users (username, email, password_hash) VALUES ('testuser', 'test@example.com', 'hash')
```

### 상세 JSON 로그 예시

```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "method": "POST",
  "path": "/api/auth/login",
  "status_code": 200,
  "duration_ms": 45,
  "request_body": {
    "username": "testuser",
    "password": "password123"
  },
  "response_body": {
    "success": true,
    "access_token": "eyJ...",
    "user": {
      "id": 1,
      "username": "testuser"
    }
  }
}
```

## 🔧 문제 해결

### 로그가 보이지 않는 경우

1. **환경변수 확인**
   ```bash
   echo $DATABASE_LOG_LEVEL
   echo $DATABASE_DEBUG
   ```

2. **test.env 파일 확인**
   ```bash
   cat test.env
   ```

3. **스크립트 권한 확인**
   ```bash
   ls -la scripts/test_with_logging.sh
   ```

### 로그가 너무 많은 경우

`test.env` 파일에서 로깅 레벨을 조정하세요:

```bash
# 경고와 에러만 출력
DATABASE_LOG_LEVEL=2
LOG_LEVEL=warn

# 에러만 출력
DATABASE_LOG_LEVEL=1
LOG_LEVEL=error
```

### 특정 테스트만 로깅

```bash
# 특정 테스트 함수만 실행
go test -v ./internal/handler -run TestInventoryHandler_CreateInventory

# 특정 패키지만 실행
go test -v ./internal/model
```

## 📝 추가 팁

### 1. 로그 파일로 저장

```bash
# 로그를 파일로 저장
./scripts/test_with_logging.sh ./internal/handler > test.log 2>&1

# 실시간으로 파일과 화면에 출력
./scripts/test_with_logging.sh ./internal/handler 2>&1 | tee test.log
```

### 2. 특정 로그만 필터링

```bash
# HTTP 요청만 필터링
./scripts/test_with_logging.sh | grep "\[HTTP\]"

# 데이터베이스 쿼리만 필터링
./scripts/test_with_logging.sh | grep "SELECT\|INSERT\|UPDATE\|DELETE"

# 에러만 필터링
./scripts/test_with_logging.sh | grep -i "error"
```

### 3. 성능 측정

```bash
# 테스트 실행 시간 측정
time ./scripts/test_with_logging.sh

# 특정 테스트의 실행 시간 측정
time go test -v ./internal/handler -run TestInventoryHandler
```

## 🎯 결론

이 가이드를 통해 테스트 실행 시 상세한 로그를 확인할 수 있습니다. 로깅을 통해 API 동작과 데이터베이스 쿼리를 정확히 파악하여 테스트 디버깅과 성능 최적화에 도움을 받을 수 있습니다. 