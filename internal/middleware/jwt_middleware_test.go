package middleware

import (
	"context"
	"encoding/json"
	"g_dev/internal/auth"
	"g_dev/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// JWT 인증 시스템 생성
func setupTestJWT(t *testing.T) *auth.JWTAuth {
	// 환경변수 설정
	os.Setenv("JWT_SECRET_KEY", "test-secret-key-2024")
	os.Setenv("JWT_ACCESS_TOKEN_EXPIRY", "15m")
	os.Setenv("JWT_REFRESH_TOKEN_EXPIRY", "168h")

	// 설정 로드
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// JWT 설정 생성
	jwtConfig := auth.NewJWTConfig(cfg)

	// 테스트용 Redis 클라이언트 생성
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       15,
	})

	// 연결 테스트
	ctx := context.Background()
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis server not available, skipping Redis-dependent tests")
	}

	redisClient.FlushDB(ctx)

	jwtAuth, err := auth.NewJWTAuth(jwtConfig, redisClient)
	if err != nil {
		t.Fatalf("failed to create JWT auth: %v", err)
	}

	return jwtAuth
}

// JWT 미들웨어 생성자 테스트
func TestNewJWTMiddleware(t *testing.T) {
	jwtAuth := setupTestJWT(t)
	middleware := NewJWTMiddleware(jwtAuth)

	//t.Log(middleware)
	//log.Printf("%v", middleware)
	assert.NotNil(t, middleware)
	assert.Equal(t, jwtAuth, middleware.jwtAuth)
}

// 인증 미들웨어 테스트
func TestJWTMiddleware_Authenticate(t *testing.T) {
	jwtAuth := setupTestJWT(t)
	middleware := NewJWTMiddleware(jwtAuth)

	// 테스트용 토큰 생성
	userID := uint(123)
	username := "testuser"
	role := "user"
	accessToken, err := jwtAuth.GenerateAccessToken(userID, username, role)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "유효한 토큰",
			authHeader:     "Bearer " + accessToken,
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "토큰 없음",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "토큰이 필요합니다",
		},
		{
			name:           "잘못된 형식",
			authHeader:     "Invalid " + accessToken,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "토큰이 필요합니다",
		},
		{
			name:           "빈 토큰",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "토큰이 필요합니다",
		},
		{
			name:           "잘못된 토큰",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "유효하지 않은 토큰입니다",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 테스트 핸들러 생성
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// 사용자 정보 확인
				userInfo, ok := GetUserFromContext(r.Context())
				//log.Printf("%v", userInfo)
				if ok {
					assert.Equal(t, userID, userInfo.UserID)
					assert.Equal(t, username, userInfo.Username)
					assert.Equal(t, role, userInfo.Role)
					assert.Equal(t, accessToken, userInfo.Token)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			// 미들웨어 적용
			middlewareHandler := middleware.Authenticate(handler)

			// 요청 생성
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// 응답 기록
			w := httptest.NewRecorder()
			middlewareHandler.ServeHTTP(w, req)

			// 결과 확인
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["message"], tt.expectedBody)
			}
		})
	}
}

// 역할 기반 접근 제어 테스트
func TestJWTMiddleware_RequireRole(t *testing.T) {
	jwtAuth := setupTestJWT(t)
	middleware := NewJWTMiddleware(jwtAuth)

	// 테스트용 토큰 생성
	userToken, err := jwtAuth.GenerateAccessToken(123, "user", "user")
	assert.NoError(t, err)
	adminToken, err := jwtAuth.GenerateAccessToken(456, "admin", "admin")
	assert.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		requiredRole   string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "올바른 역할",
			authHeader:     "Bearer " + adminToken,
			requiredRole:   "admin",
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "잘못된 역할",
			authHeader:     "Bearer " + userToken,
			requiredRole:   "admin",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "'admin' 역할이 필요합니다",
		},
		{
			name:           "토큰 없음",
			authHeader:     "",
			requiredRole:   "admin",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "토큰이 필요합니다",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 테스트 핸들러 생성
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			// 미들웨어 체인 적용
			middlewareHandler := middleware.Authenticate(middleware.RequireRole(tt.requiredRole)(handler))
			//log.Printf("%v", middlewareHandler)
			// 요청 생성
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// 응답 기록
			w := httptest.NewRecorder()
			middlewareHandler.ServeHTTP(w, req)

			// 결과 확인
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["message"], tt.expectedBody)
			}
		})
	}
}

// 여러 역할 중 하나를 요구하는 기능 테스트
func TestJWTMiddleware_RequireAnyRole(t *testing.T) {
	jwtAuth := setupTestJWT(t)
	middleware := NewJWTMiddleware(jwtAuth)

	// 테스트용 토큰 생성
	userToken, err := jwtAuth.GenerateAccessToken(123, "user", "user")
	assert.NoError(t, err)
	adminToken, err := jwtAuth.GenerateAccessToken(456, "admin", "admin")
	assert.NoError(t, err)
	moderatorToken, err := jwtAuth.GenerateAccessToken(789, "moderator", "moderator")
	assert.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		requiredRoles  []string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "첫 번째 역할 일치",
			authHeader:     "Bearer " + adminToken,
			requiredRoles:  []string{"admin", "moderator"},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "두 번째 역할 일치",
			authHeader:     "Bearer " + moderatorToken,
			requiredRoles:  []string{"admin", "moderator"},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "역할 불일치",
			authHeader:     "Bearer " + userToken,
			requiredRoles:  []string{"admin", "moderator"},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "다음 역할 중 하나가 필요합니다: admin,moderator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 테스트 핸들러 생성
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			// 미들웨어 체인 적용
			middlewareHandler := middleware.Authenticate(middleware.RequireAnyRole(tt.requiredRoles...)(handler))

			// 요청 생성
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// 응답 기록
			w := httptest.NewRecorder()
			middlewareHandler.ServeHTTP(w, req)

			// 결과 확인
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["message"], tt.expectedBody)
			}
		})
	}
}

// 선택적 인증을 테스트
func TestJWTMiddleware_OptionalAuth(t *testing.T) {
	jwtAuth := setupTestJWT(t)
	middleware := NewJWTMiddleware(jwtAuth)

	// 테스트용 토큰 생성
	accessToken, err := jwtAuth.GenerateAccessToken(123, "testuser", "user")
	assert.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
		expectUserInfo bool
	}{
		{
			name:           "유효한 토큰",
			authHeader:     "Bearer " + accessToken,
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			expectUserInfo: true,
		},
		{
			name:           "토큰 없음",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			expectUserInfo: false,
		},
		{
			name:           "잘못된 토큰",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			expectUserInfo: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 테스트 핸들러 생성
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// 사용자 정보 확인
				userInfo, ok := GetUserFromContext(r.Context())
				if tt.expectUserInfo {
					assert.True(t, ok)
					assert.NotNil(t, userInfo)
					assert.Equal(t, uint(123), userInfo.UserID)
					assert.Equal(t, "testuser", userInfo.Username)
					assert.Equal(t, "user", userInfo.Role)
				} else {
					assert.False(t, ok)
					assert.Nil(t, userInfo)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			// 미들웨어 적용
			middlewareHandler := middleware.OptionalAuth(handler)

			// 요청 생성
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// 응답 기록
			w := httptest.NewRecorder()
			middlewareHandler.ServeHTTP(w, req)

			// 결과 확인
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, w.Body.String(), tt.expectedBody)
		})
	}
}

// 컨텍스트에서 사용자 정보를 가져오는 기능 테스트
func TestGetUserFromContext(t *testing.T) {
	// 빈 컨텍스트
	ctx := context.Background()
	userInfo, ok := GetUserFromContext(ctx)
	log.Printf("%v", userInfo)
	log.Printf("%v", ok)
	assert.False(t, ok)
	assert.Nil(t, userInfo)

	// 토큰이 있는 컨텍스트
	testToken := "test-token"
	ctx = context.WithValue(ctx, TokenContextKey, testToken)

	retrievedToken, ok := GetTokenFromContext(ctx)
	log.Printf("%v", retrievedToken)
	log.Printf("%v", ok)
	assert.True(t, ok)
	assert.Equal(t, testToken, retrievedToken)
}

// 헬퍼 함수들을 테스트
func TestHelperFunctions(t *testing.T) {
	jwtAuth := setupTestJWT(t)

	// RequireAuth 테스트
	requireAuth := RequireAuth(jwtAuth)
	assert.NotNil(t, requireAuth)

	// RequireRole 테스트
	requireRole := RequireRole(jwtAuth, "admin")
	assert.NotNil(t, requireRole)

	// RequireAnyRole 테스트
	requireAnyRole := RequireAnyRole(jwtAuth, "admin", "moderator")
	assert.NotNil(t, requireAnyRole)

	// OptionalAuth 테스트
	optionalAuth := OptionalAuth(jwtAuth)
	assert.NotNil(t, optionalAuth)
}
