package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"g_dev/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"os"
	"time"
)

// JWT 인증 설정
type JWTConfig struct {
	SecretKey          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Issuer             string
	Audience           string
	Algorithm          string
}

// JWT 토큰에 포함될 클레임 정보
type Claims struct {
	// 사용자 ID
	UserID uint
	// 사용자명
	Username string
	// 사용자 역할 (user, admin, moderator)
	Role string
	// 토큰 타입 (access, refresh)
	TokenType string
	// 표준 JWT 클레임
	jwt.RegisteredClaims
}

// JWT 인증 기능 제공
type JWTAuth struct {
	// JWT 설정
	Config JWTConfig
	// 토큰 검증을 위한 키
	key         []byte
	redisClient *redis.Client
}

// JWT 설정 생성
func NewJWTConfig(cfg *config.Config) JWTConfig {
	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		panic("환경변수 JWT_SECRET_KEY가 설정되어 있지 않습니다. 서비스 기동 불가.")
	}
	return JWTConfig{
		SecretKey:          secret,
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "g_dev",
		Audience:           "g_dev_users",
		Algorithm:          "HS256",
	}
}

// 새로운 JWTAuth 인스턴스 생성
func NewJWTAuth(jwtConfig JWTConfig, redisClient *redis.Client) (*JWTAuth, error) {
	if jwtConfig.SecretKey == "" {
		return nil, fmt.Errorf("JWT secret key is required")
	}

	if redisClient == nil {
		return nil, fmt.Errorf("Redis client is required for token management")
	}

	return &JWTAuth{
		Config:      jwtConfig,
		key:         []byte(jwtConfig.SecretKey),
		redisClient: redisClient,
	}, nil
}

// 액세스 토큰 생성
func (j *JWTAuth) GenerateAccessToken(userID uint, username, role string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(j.Config.AccessTokenExpiry)

	claims := Claims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.Config.Issuer,
			Subject:   fmt.Sprintf("%d", userID),
			Audience:  []string{j.Config.Audience},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.key)
}

// 리프레시 토큰을 생성
// 액세스 토큰 갱신을 위한 긴 만료 시간을 가진 토큰을 생성
func (j *JWTAuth) GenerateRefreshToken(userID uint, username, role string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(j.Config.RefreshTokenExpiry)

	claims := Claims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.Config.Issuer,
			Subject:   fmt.Sprintf("%d", userID),
			Audience:  []string{j.Config.Audience},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.key)
}

// 액세스 토큰과 리프레시 토큰 쌍을 생성
func (j *JWTAuth) GenerateTokenPair(userID uint, username, role string) (string, string, error) {
	accessToken, err := j.GenerateAccessToken(userID, username, role)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := j.GenerateRefreshToken(userID, username, role)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 리프레시 토큰을 Redis에 저장
	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%d", userID)

	err = j.redisClient.Set(ctx, key, refreshToken, j.Config.RefreshTokenExpiry).Err()
	if err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// JWT 토큰을 검증하고 클레임을 반환
func (j *JWTAuth) ValidateToken(tokenString string) (*Claims, error) {
	// 블랙리스트 확인
	if j.IsTokenRevoked(tokenString) {
		return nil, fmt.Errorf("token is revoked")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 알고리즘 검증
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.key, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// 액세스 토큰을 검증
// 토큰 타입이 "access"인지 확인하고 검증
func (j *JWTAuth) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "access" {
		return nil, fmt.Errorf("invalid token type: expected access, got %s", claims.TokenType)
	}

	return claims, nil
}

// 리프레시 토큰을 검증
// 토큰 타입이 "refresh"인지 확인하고 검증
func (j *JWTAuth) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type: expected refresh, got %s", claims.TokenType)
	}

	return claims, nil
}

// 리프레시 토큰을 사용하여 새로운 액세스 토큰을 생성
// 리프레시 토큰이 유효한 경우에만 새로운 액세스 토큰을 발급
func (j *JWTAuth) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := j.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Redis에 저장된 리프레시 토큰과 비교
	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%d", claims.UserID)

	storedToken, err := j.redisClient.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("refresh token not found in storage: %w", err)
	}

	if storedToken != refreshToken {
		return "", fmt.Errorf("refresh token mismatch")
	}

	// 새로운 액세스 토큰 생성
	accessToken, err := j.GenerateAccessToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		return "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	return accessToken, nil
}

// RevokeToken은 토큰을 무효화
// Redis를 사용하여 블랙리스트에 토큰을 추가
func (j *JWTAuth) RevokeToken(tokenString string) error {
	return j.BlacklistToken(tokenString)
}

// 토큰을 블랙리스트에 추가
func (j *JWTAuth) BlacklistToken(tokenString string) error {
	// 토큰 검증
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// 토큰 만료 시간까지 블랙리스트에 추가
	expiresAt := claims.ExpiresAt.Time
	now := time.Now()
	ttl := expiresAt.Sub(now)

	if ttl <= 0 {
		return fmt.Errorf("token already expired")
	}

	ctx := context.Background()
	key := fmt.Sprintf("blacklist:%s", tokenString)

	err = j.redisClient.Set(ctx, key, "revoked", ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	return nil
}

// 토큰이 무효화되었는지 확인
// Redis 블랙리스트에서 토큰을 확인
func (j *JWTAuth) IsTokenRevoked(tokenString string) bool {
	ctx := context.Background()
	key := fmt.Sprintf("blacklist:%s", tokenString)

	exists, err := j.redisClient.Exists(ctx, key).Result()
	if err != nil {
		// Redis 오류 시 보안상 무효화된 것으로 처리
		return true
	}

	return exists > 0
}

// 사용자 로그아웃을 처리
// 액세스 토큰을 블랙리스트에 추가하고 리프레시 토큰을 삭제
func (j *JWTAuth) Logout(accessToken string) error {
	// 액세스 토큰 검증
	claims, err := j.ValidateAccessToken(accessToken)
	if err != nil {
		return fmt.Errorf("invalid access token: %w", err)
	}

	// 리프레시 토큰 삭제
	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%d", claims.UserID)

	err = j.redisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

// 토큰을 블랙리스트에서 제거
func (j *JWTAuth) RemoveFromBlacklist(tokenString string) error {
	ctx := context.Background()
	key := fmt.Sprintf("blacklist:%s", tokenString)

	err := j.redisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to remove token from blacklist: %w", err)
	}

	return nil
}

// 랜덤한 시크릿 키를 생성
// 32바이트 랜덤 데이터를 base64로 인코딩하여 반환.
func GenerateRandomSecretKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}
