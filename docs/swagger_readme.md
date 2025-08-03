# 📚 Swagger/OpenAPI 문서화 가이드

## 🎯 개요

G-Dev 웹게임서버는 **Swagger/OpenAPI 2.0** 스펙을 사용하여 REST API를 문서화합니다. 이 문서는 Swagger 설정, 사용법, 그리고 API 문서화 방법에 대해 설명합니다.

## 📋 목차

1. [Swagger란?](#swagger란)
2. [설치 및 설정](#설치-및-설정)
3. [API 문서화 방법](#api-문서화-방법)
4. [생성된 파일들](#생성된-파일들)
5. [사용법](#사용법)
6. [주석 작성 가이드](#주석-작성-가이드)
7. [문제 해결](#문제-해결)

## 🤔 Swagger란?

### 정의
- **Swagger/OpenAPI**: REST API를 문서화하는 표준 스펙
- **자동 문서 생성**: Go 코드의 주석을 분석하여 API 문서 자동 생성
- **인터랙티브 문서**: 웹 브라우저에서 API를 직접 테스트 가능

### 장점
- ✅ **자동화**: 코드 변경 시 문서 자동 업데이트
- ✅ **표준화**: OpenAPI 2.0/3.0 표준 준수
- ✅ **테스트**: 문서에서 직접 API 호출 가능
- ✅ **협업**: 개발자 간 API 명세 공유 용이

## 🛠️ 설치 및 설정

### 1. 의존성 설치

```bash
# Swagger 관련 패키지 설치
go get -u github.com/swaggo/swag/cmd/swag
go get -u github.com/swaggo/http-swagger
go get -u github.com/swaggo/files

# Swagger CLI 도구 설치
go install github.com/swaggo/swag/cmd/swag@latest
```

### 2. 프로젝트 구조

```
g_step/
├── cmd/server/
│   └── main.go          # 메인 애플리케이션 (Swagger 주석 포함)
├── internal/handler/
│   ├── api_handler.go   # API 핸들러 구조체
│   └── calculator_handler.go  # 계산기 API (Swagger 주석 포함)
├── docs/
│   ├── docs.go          # 자동 생성된 Swagger 문서
│   ├── swagger.json     # JSON 형태 API 스펙
│   └── swagger.yaml     # YAML 형태 API 스펙
└── go.mod
```

## 📝 API 문서화 방법

### 1. 메인 애플리케이션 주석

`cmd/server/main.go` 파일 상단에 API 정보 주석을 추가합니다:

```go
// @title G-Dev 웹게임서버 API
// @version 1.0
// @description G-Dev 웹게임서버의 REST API 문서입니다.
// @description 이 API는 계산기 기능과 파일 처리 기능을 제공합니다.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support Team
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /

// @tag.name Calculator
// @tag.description 계산기 관련 API 엔드포인트

// @tag.name FileProcessor
// @tag.description 파일 처리 관련 API 엔드포인트
```

### 2. API 핸들러 주석

각 API 엔드포인트 함수에 Swagger 주석을 추가합니다:

```go
// HandleCalculatorCalculate는 계산기 계산 API 엔드포인트를 처리합니다.
// @Summary 계산기 계산 수행
// @Description 두 숫자에 대한 사칙연산을 수행합니다.
// @Tags Calculator
// @Accept json
// @Produce json
// @Param request body CalculatorRequest true "계산 요청"
// @Success 200 {object} APIResponse{data=CalculatorResponse} "계산 성공"
// @Failure 400 {object} APIResponse "잘못된 요청"
// @Failure 405 {object} APIResponse "허용되지 않는 HTTP 메서드"
// @Router /api/calculator/calculate [post]
func (h *APIHandler) HandleCalculatorCalculate(w http.ResponseWriter, r *http.Request) {
    // 함수 구현...
}
```

### 3. 구조체 정의

API 요청/응답 구조체를 정의합니다:

```go
// CalculatorRequest는 계산기 API 요청 구조체입니다.
type CalculatorRequest struct {
    Operation string  `json:"operation" example:"add"`     // 수행할 연산
    Operand1  float64 `json:"operand1" example:"10"`       // 첫 번째 피연산자
    Operand2  float64 `json:"operand2" example:"5"`        // 두 번째 피연산자
}

// CalculatorResponse는 계산기 API 응답 구조체입니다.
type CalculatorResponse struct {
    Result    float64 `json:"result" example:"15"`         // 계산 결과
    Operation string  `json:"operation" example:"add"`     // 수행된 연산
    Operand1  float64 `json:"operand1" example:"10"`       // 첫 번째 피연산자
    Operand2  float64 `json:"operand2" example:"5"`        // 두 번째 피연산자
}
```

## 📁 생성된 파일들

### 1. docs/docs.go
- Swagger 문서를 Go 코드로 생성
- 서버에서 Swagger UI 서빙을 위한 설정 포함

### 2. docs/swagger.json
- JSON 형태의 OpenAPI 2.0 스펙
- API 엔드포인트, 스키마, 예시 등 포함

### 3. docs/swagger.yaml
- YAML 형태의 OpenAPI 2.0 스펙
- JSON과 동일한 내용을 YAML 형식으로 제공

## 🚀 사용법

### 1. 문서 생성

```bash
# 프로젝트 루트에서 실행
swag init -g cmd/server/main.go

# 또는 전체 경로 사용
~/go/bin/swag init -g cmd/server/main.go
```

### 2. 서버 실행

```bash
go run cmd/server/main.go
```

### 3. 문서 접근

- **홈페이지**: http://localhost:8081/
- **Swagger UI**: http://localhost:8081/swagger/index.html
- **API JSON**: http://localhost:8081/swagger/doc.json

## 📖 주석 작성 가이드

### 1. 메인 애플리케이션 주석

| 태그 | 설명 | 예시 |
|------|------|------|
| `@title` | API 제목 | `@title G-Step 웹게임서버 API` |
| `@version` | API 버전 | `@version 1.0` |
| `@description` | API 설명 | `@description G-Step 웹게임서버의 REST API` |
| `@host` | 서버 호스트 | `@host localhost:8080` |
| `@BasePath` | 기본 경로 | `@BasePath /` |
| `@tag.name` | 태그 이름 | `@tag.name Calculator` |
| `@tag.description` | 태그 설명 | `@tag.description 계산기 관련 API` |

### 2. API 엔드포인트 주석

| 태그 | 설명 | 예시 |
|------|------|------|
| `@Summary` | API 요약 | `@Summary 계산기 계산 수행` |
| `@Description` | 상세 설명 | `@Description 두 숫자에 대한 사칙연산` |
| `@Tags` | 태그 분류 | `@Tags Calculator` |
| `@Accept` | 요청 Content-Type | `@Accept json` |
| `@Produce` | 응답 Content-Type | `@Produce json` |
| `@Param` | 매개변수 정의 | `@Param request body CalculatorRequest true "계산 요청"` |
| `@Success` | 성공 응답 | `@Success 200 {object} APIResponse` |
| `@Failure` | 실패 응답 | `@Failure 400 {object} APIResponse` |
| `@Router` | 라우터 경로 | `@Router /api/calculator/calculate [post]` |

### 3. 구조체 필드 주석

```go
type CalculatorRequest struct {
    Operation string  `json:"operation" example:"add" binding:"required"`     // 연산 타입
    Operand1  float64 `json:"operand1" example:"10" binding:"required"`       // 첫 번째 피연산자
    Operand2  float64 `json:"operand2" example:"5" binding:"required"`        // 두 번째 피연산자
}
```

| 태그 | 설명 | 예시 |
|------|------|------|
| `json` | JSON 필드명 | `json:"operation"` |
| `example` | 예시 값 | `example:"add"` |
| `binding` | 유효성 검사 | `binding:"required"` |

## 🔧 문제 해결

### 1. swag 명령어를 찾을 수 없는 경우

```bash
# PATH에 Go bin 디렉토리 추가
export PATH=$PATH:$(go env GOPATH)/bin

# 또는 전체 경로 사용
~/go/bin/swag init -g cmd/server/main.go
```

### 2. 문서가 업데이트되지 않는 경우

```bash
# 기존 docs 폴더 삭제 후 재생성
rm -rf docs/
swag init -g cmd/server/main.go
```

### 3. Swagger UI가 로드되지 않는 경우

```bash
# 의존성 재설치
go mod tidy
go get -u github.com/swaggo/http-swagger
```

### 4. 주석이 인식되지 않는 경우

- 주석이 함수 바로 위에 있는지 확인
- 주석 형식이 정확한지 확인 (`// @` 로 시작)
- 함수가 export된 함수인지 확인 (대문자로 시작)

## 📋 현재 구현된 API

### Calculator API

| 엔드포인트 | 메서드 | 설명 |
|-----------|--------|------|
| `/api/calculator/calculate` | POST | 계산 수행 |
| `/api/calculator/history` | GET | 히스토리 조회 |
| `/api/calculator/history` | DELETE | 히스토리 초기화 |
| `/api/calculator/stats` | GET | 통계 조회 |

### 예시 요청/응답

#### 계산 수행
```bash
curl -X POST http://localhost:8081/api/calculator/calculate \
  -H "Content-Type: application/json" \
  -d '{"operation": "add", "operand1": 10, "operand2": 5}'
```

```json
{
  "success": true,
  "message": "계산이 완료되었습니다",
  "data": {
    "result": 15,
    "operation": "add",
    "operand1": 10,
    "operand2": 5
  }
}
```

## 🔄 문서 업데이트 워크플로우

1. **코드 수정**: API 핸들러나 구조체 수정
2. **주석 업데이트**: Swagger 주석 추가/수정
3. **문서 재생성**: `swag init -g cmd/server/main.go` 실행
4. **서버 재시작**: `go run cmd/server/main.go`
5. **문서 확인**: Swagger UI에서 변경사항 확인

## 📚 추가 자료

- [Swaggo 공식 문서](https://github.com/swaggo/swag)
- [OpenAPI 2.0 스펙](https://swagger.io/specification/v2/)
- [Swagger 주석 가이드](https://github.com/swaggo/swag#declarative-comments-format)

--- 