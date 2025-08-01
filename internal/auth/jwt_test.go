package auth

import (
	"context"
	"fmt"
	"g_dev/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"
)

// 테스트용 설정
func setupTestConfig(t *testing.T) *config.Config {
	// 환경변수 설정
	os.Setenv("JWT_SECRET_KEY", "test-secret-key-2024")
	os.Setenv("JWT_ACCESS_TOKEN_EXPIRY", "15m")
	os.Setenv("JWT_REFRESH_TOKEN_EXPIRY", "168h")

	// 설정 로드
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	return cfg
}

// 테스트용 Redis 클라이언트를 생성
func setupTestRedisClient(t *testing.T) *redis.Client {
	// 테스트용 Redis 클라이언트 생성 (메모리 기반)
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       15, // 테스트용 DB 사용
	})

	// 연결 테스트
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		// Redis가 없으면 테스트 스킵
		t.Skip("Redis server not available, skipping Redis-dependent tests")
	}

	// 테스트 DB 초기화
	rdb.FlushDB(ctx)

	return rdb
}

// JWT 설정 생성자를 테스트
func TestNewJWTConfig(t *testing.T) {
	cfg := setupTestConfig(t)
	jwtConfig := NewJWTConfig(cfg)

	// 기본값 확인
	assert.NotEmpty(t, jwtConfig.SecretKey)
	assert.Equal(t, 15*time.Minute, jwtConfig.AccessTokenExpiry)
	assert.Equal(t, 168*time.Hour, jwtConfig.RefreshTokenExpiry)
	assert.Equal(t, "g_dev", jwtConfig.Issuer)
	assert.Equal(t, "g_dev_users", jwtConfig.Audience)
	assert.Equal(t, "HS256", jwtConfig.Algorithm)
}

// JWT 인증 시스템 생성자를 테스트
func TestNewJWTAuth(t *testing.T) {
	cfg := setupTestConfig(t)
	jwtConfig := NewJWTConfig(cfg)
	redisClient := setupTestRedisClient(t)

	tests := []struct {
		name        string
		config      JWTConfig
		redisClient *redis.Client
		wantErr     bool
	}{
		{
			name:        "정상적인 설정",
			config:      jwtConfig,
			redisClient: redisClient,
			wantErr:     false,
		},
		{
			name: "빈 시크릿 키",
			config: JWTConfig{
				SecretKey:          "",
				AccessTokenExpiry:  15 * time.Minute,
				RefreshTokenExpiry: 7 * 24 * time.Hour,
				Issuer:             "g_dev",
				Audience:           "g_dev_users",
				Algorithm:          "HS256",
			},
			redisClient: redisClient,
			wantErr:     true,
		},
		{
			name:        "Redis 클라이언트 없음",
			config:      jwtConfig,
			redisClient: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwtAuth, err := NewJWTAuth(tt.config, tt.redisClient)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, jwtAuth)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, jwtAuth)
				assert.Equal(t, tt.config.SecretKey, jwtAuth.Config.SecretKey)
			}
		})
	}
}

// 액세스 토큰 생성 기능을 테스트
func TestJWTAuth_GenerateAccessToken(t *testing.T) {
	cfg := setupTestConfig(t)
	jwtConfig := NewJWTConfig(cfg)
	redisClient := setupTestRedisClient(t)
	jwtAuth, err := NewJWTAuth(jwtConfig, redisClient)
	log.Printf("%v", jwtAuth)
	assert.NoError(t, err)

	// 액세스 토큰 생성
	userID := uint(123)
	username := "testuser"
	role := "user"

	token, err := jwtAuth.GenerateAccessToken(userID, username, role)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 토큰 검증
	claims, err := jwtAuth.ValidateAccessToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, role, claims.Role)
	assert.Equal(t, "access", claims.TokenType)
	assert.Equal(t, "g_dev", claims.Issuer)
	assert.Equal(t, "g_dev_users", claims.Audience[0])
}

// 토큰 쌍 생성 기능을 테스트
func TestJWTAuth_GenerateTokenPair(t *testing.T) {
	cfg := setupTestConfig(t)
	jwtConfig := NewJWTConfig(cfg)
	redisClient := setupTestRedisClient(t)
	jwtAuth, err := NewJWTAuth(jwtConfig, redisClient)
	assert.NoError(t, err)

	// 토큰 쌍 생성
	userID := uint(123)
	username := "testuser"
	role := "user"

	accessToken, refreshToken, err := jwtAuth.GenerateTokenPair(userID, username, role)
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)

	// 액세스 토큰 검증
	accessClaims, err := jwtAuth.ValidateAccessToken(accessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, accessClaims.UserID)
	assert.Equal(t, "access", accessClaims.TokenType)

	// 리프레시 토큰 검증
	refreshClaims, err := jwtAuth.ValidateRefreshToken(refreshToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, refreshClaims.UserID)
	assert.Equal(t, "refresh", refreshClaims.TokenType)
}

// 토큰 검증 기능을 테스트
func TestJWTAuth_ValidateToken(t *testing.T) {
	cfg := setupTestConfig(t)
	jwtConfig := NewJWTConfig(cfg)
	redisClient := setupTestRedisClient(t)
	jwtAuth, err := NewJWTAuth(jwtConfig, redisClient)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "빈 토큰",
			token:   "",
			wantErr: true,
		},
		{
			name:    "잘못된 형식의 토큰",
			token:   "invalid-token",
			wantErr: true,
		},
		{
			name:    "다른 시크릿 키로 서명된 토큰",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := jwtAuth.ValidateToken(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// 액세스 토큰 검증 기능을 테스트
func TestJWTAuth_ValidateAccessToken(t *testing.T) {
	cfg := setupTestConfig(t)
	jwtConfig := NewJWTConfig(cfg)
	redisClient := setupTestRedisClient(t)
	jwtAuth, err := NewJWTAuth(jwtConfig, redisClient)
	assert.NoError(t, err)

	// 액세스 토큰 생성
	accessToken, err := jwtAuth.GenerateAccessToken(123, "testuser", "user")
	assert.NoError(t, err)

	// 리프레시 토큰 생성
	refreshToken, err := jwtAuth.GenerateRefreshToken(123, "testuser", "user")
	assert.NoError(t, err)

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "유효한 액세스 토큰",
			token:   accessToken,
			wantErr: false,
		},
		{
			name:    "리프레시 토큰으로 액세스 토큰 검증 시도",
			token:   refreshToken,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := jwtAuth.ValidateAccessToken(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// 리프레시 토큰 검증 기능을 테스트
func TestJWTAuth_ValidateRefreshToken(t *testing.T) {
	cfg := setupTestConfig(t)
	jwtConfig := NewJWTConfig(cfg)
	redisClient := setupTestRedisClient(t)
	jwtAuth, err := NewJWTAuth(jwtConfig, redisClient)
	assert.NoError(t, err)

	// 액세스 토큰 생성
	accessToken, err := jwtAuth.GenerateAccessToken(123, "testuser", "user")
	assert.NoError(t, err)

	// 리프레시 토큰 생성
	refreshToken, err := jwtAuth.GenerateRefreshToken(123, "testuser", "user")
	assert.NoError(t, err)

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "유효한 리프레시 토큰",
			token:   refreshToken,
			wantErr: false,
		},
		{
			name:    "액세스 토큰으로 리프레시 토큰 검증 시도",
			token:   accessToken,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := jwtAuth.ValidateRefreshToken(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// 액세스 토큰 갱신 기능을 테스트
func TestJWTAuth_RefreshAccessToken(t *testing.T) {
	cfg := setupTestConfig(t)
	jwtConfig := NewJWTConfig(cfg)
	redisClient := setupTestRedisClient(t)
	jwtAuth, err := NewJWTAuth(jwtConfig, redisClient)
	assert.NoError(t, err)

	// 토큰 쌍 생성
	userID := uint(123)
	username := "testuser"
	role := "user"

	_, refreshToken, err := jwtAuth.GenerateTokenPair(userID, username, role)
	assert.NoError(t, err)

	// 액세스 토큰 갱신
	newAccessToken, err := jwtAuth.RefreshAccessToken(refreshToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)

	// 새로운 액세스 토큰 검증
	claims, err := jwtAuth.ValidateAccessToken(newAccessToken)
	assert.NoError(t, err)
	assert.Equal(t, uint(123), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "user", claims.Role)
	assert.Equal(t, "access", claims.TokenType)
}

// 잘못된 리프레시 토큰으로 액세스 토큰 갱신을 테스트
func TestJWTAuth_RefreshAccessToken_InvalidToken(t *testing.T) {
	cfg := setupTestConfig(t)
	jwtConfig := NewJWTConfig(cfg)
	redisClient := setupTestRedisClient(t)
	jwtAuth, err := NewJWTAuth(jwtConfig, redisClient)
	assert.NoError(t, err)

	// 잘못된 토큰으로 갱신 시도
	_, err = jwtAuth.RefreshAccessToken("invalid-token")
	assert.Error(t, err)
}

// 토큰 블랙리스트 기능을 테스트
func TestJWTAuth_BlacklistToken(t *testing.T) {
	cfg := setupTestConfig(t)
	jwtConfig := NewJWTConfig(cfg)
	redisClient := setupTestRedisClient(t)
	jwtAuth, err := NewJWTAuth(jwtConfig, redisClient)
	assert.NoError(t, err)

	// 액세스 토큰 생성
	accessToken, err := jwtAuth.GenerateAccessToken(123, "testuser", "user")
	assert.NoError(t, err)

	// 토큰을 블랙리스트에 추가
	err = jwtAuth.BlacklistToken(accessToken)
	assert.NoError(t, err)

	// 블랙리스트된 토큰 검증 시도
	_, err = jwtAuth.ValidateToken(accessToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token is revoked")
}

// 블랙리스트에서 토큰 제거 기능을 테스트
func TestJWTAuth_RemoveFromBlacklist(t *testing.T) {
	cfg := setupTestConfig(t)
	jwtConfig := NewJWTConfig(cfg)
	redisClient := setupTestRedisClient(t)
	jwtAuth, err := NewJWTAuth(jwtConfig, redisClient)
	assert.NoError(t, err)

	// 액세스 토큰 생성
	accessToken, err := jwtAuth.GenerateAccessToken(123, "testuser", "user")
	assert.NoError(t, err)

	// 토큰을 블랙리스트에 추가
	err = jwtAuth.BlacklistToken(accessToken)
	assert.NoError(t, err)

	// 블랙리스트에서 토큰 제거
	err = jwtAuth.RemoveFromBlacklist(accessToken)
	assert.NoError(t, err)

	// 제거된 토큰 검증 시도 (성공해야 함)
	_, err = jwtAuth.ValidateToken(accessToken)
	assert.NoError(t, err)
}

// 토큰 무효화 기능을 테스트
func TestJWTAuth_RevokeToken(t *testing.T) {
	cfg := setupTestConfig(t)
	jwtConfig := NewJWTConfig(cfg)
	redisClient := setupTestRedisClient(t)
	jwtAuth, err := NewJWTAuth(jwtConfig, redisClient)
	assert.NoError(t, err)

	// 액세스 토큰 생성
	accessToken, err := jwtAuth.GenerateAccessToken(123, "testuser", "user")
	assert.NoError(t, err)

	// 토큰 무효화
	err = jwtAuth.RevokeToken(accessToken)
	assert.NoError(t, err)

	// 무효화된 토큰 검증 시도
	_, err = jwtAuth.ValidateToken(accessToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token is revoked")
}

// 로그아웃 기능을 테스트
func TestJWTAuth_Logout(t *testing.T) {
	cfg := setupTestConfig(t)
	jwtConfig := NewJWTConfig(cfg)
	redisClient := setupTestRedisClient(t)
	jwtAuth, err := NewJWTAuth(jwtConfig, redisClient)
	assert.NoError(t, err)

	// 토큰 쌍 생성
	userID := uint(123)
	username := "testuser"
	role := "user"

	accessToken, refreshToken, err := jwtAuth.GenerateTokenPair(userID, username, role)
	assert.NoError(t, err)

	// 로그아웃
	err = jwtAuth.Logout(accessToken)
	assert.NoError(t, err)

	// 액세스 토큰이 유효해야 함
	_, err = jwtAuth.ValidateToken(accessToken)
	assert.NoError(t, err)

	// 리프레시 토큰으로 갱신 시도 (실패해야 함)
	_, err = jwtAuth.RefreshAccessToken(refreshToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refresh token not found in storage")
}

// 랜덤 시크릿 키 생성 기능을 테스트
func TestGenerateRandomSecretKey(t *testing.T) {
	// 첫 번째 키 생성
	key1, err := GenerateRandomSecretKey()
	assert.NoError(t, err)
	assert.NotEmpty(t, key1)

	// 두 번째 키 생성
	key2, err := GenerateRandomSecretKey()
	assert.NoError(t, err)
	assert.NotEmpty(t, key2)

	// 두 키가 다른지 확인 (랜덤성 테스트)
	assert.NotEqual(t, key1, key2)
}

// JWT 인증 시스템의 통합 테스트를 수행
func TestJWTAuth_Integration(t *testing.T) {
	cfg := setupTestConfig(t)
	jwtConfig := NewJWTConfig(cfg)
	redisClient := setupTestRedisClient(t)
	jwtAuth, err := NewJWTAuth(jwtConfig, redisClient)
	assert.NoError(t, err)

	// 1. 토큰 쌍 생성
	userID := uint(456)
	username := "integrationuser"
	role := "admin"

	accessToken, refreshToken, err := jwtAuth.GenerateTokenPair(userID, username, role)
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)

	// 2. 액세스 토큰 검증
	accessClaims, err := jwtAuth.ValidateAccessToken(accessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, accessClaims.UserID)
	assert.Equal(t, username, accessClaims.Username)
	assert.Equal(t, role, accessClaims.Role)
	assert.Equal(t, "access", accessClaims.TokenType)

	// 3. 리프레시 토큰 검증
	refreshClaims, err := jwtAuth.ValidateRefreshToken(refreshToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, refreshClaims.UserID)
	assert.Equal(t, username, refreshClaims.Username)
	assert.Equal(t, role, refreshClaims.Role)
	assert.Equal(t, "refresh", refreshClaims.TokenType)

	// 4. 액세스 토큰 갱신
	newAccessToken, err := jwtAuth.RefreshAccessToken(refreshToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)

	// 5. 새로운 액세스 토큰 검증
	newAccessClaims, err := jwtAuth.ValidateAccessToken(newAccessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, newAccessClaims.UserID)
	assert.Equal(t, username, newAccessClaims.Username)
	assert.Equal(t, role, newAccessClaims.Role)
	assert.Equal(t, "access", newAccessClaims.TokenType)

	// 6. 기존 액세스 토큰도 여전히 유효한지 확인
	oldAccessClaims, err := jwtAuth.ValidateAccessToken(accessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, oldAccessClaims.UserID)

	// 7. 로그아웃
	err = jwtAuth.Logout(accessToken)
	assert.NoError(t, err)

	// 8. 로그아웃 후 액세스 토큰은 여전히 유효
	_, err = jwtAuth.ValidateToken(accessToken)
	assert.NoError(t, err)

	// 9. 리프레시 토큰으로 갱신 시도 (실패해야 함 - Redis에서 삭제됨)
	_, err = jwtAuth.RefreshAccessToken(refreshToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refresh token not found in storage")

	// 10. 보안상 문제가 있는 토큰을 블랙리스트에 추가
	err = jwtAuth.BlacklistToken(accessToken)
	assert.NoError(t, err)

	// 11. 블랙리스트된 토큰 검증 시도 (실패해야 함)
	_, err = jwtAuth.ValidateToken(accessToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token is revoked")
}

// JWTAuth 생성자의 사용 예시를 제공
func ExampleNewJWTAuth() {
	// 환경변수 설정 (실제로는 서버 시작 시 설정됨)
	os.Setenv("JWT_SECRET_KEY", "example-secret-key")
	os.Setenv("JWT_ACCESS_TOKEN_EXPIRY", "15m")
	os.Setenv("JWT_REFRESH_TOKEN_EXPIRY", "168h")

	// 설정 로드
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// JWT 설정 생성
	jwtConfig := NewJWTConfig(cfg)

	// Redis 클라이언트 생성
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// JWTAuth 인스턴스 생성
	jwtAuth, err := NewJWTAuth(jwtConfig, redisClient)
	if err != nil {
		panic(err)
	}

	// 토큰 쌍 생성
	accessToken, refreshToken, err := jwtAuth.GenerateTokenPair(123, "testuser", "user")
	if err != nil {
		panic(err)
	}

	// 토큰 검증
	claims, err := jwtAuth.ValidateAccessToken(accessToken)
	if err != nil {
		panic(err)
	}

	fmt.Printf("User ID: %d, Username: %s, Role: %s\n", claims.UserID, claims.Username, claims.Role)
	fmt.Printf("Access Token: %s\n", accessToken[:50]+"...")
	fmt.Printf("Refresh Token: %s\n", refreshToken[:50]+"...")

	// Output:
	// Warning: .env file not found, using environment variables only
	// User ID: 123, Username: testuser, Role: user
	// Access Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiO...
	// Refresh Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiO...
}
