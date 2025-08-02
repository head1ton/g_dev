package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"
)

// 설정 로딩 기능을 테스트
func TestLoadConfig(t *testing.T) {
	// 테스트용 환경변수 설정
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	os.Setenv("PORT", "9090")
	os.Setenv("DATABASE_HOST", "test-db-host")
	os.Setenv("REDIS_PORT", "6380")

	// 설정 로드
	config, err := LoadConfig()
	log.Printf("config: %v", config)
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// 서버 설정 확인
	assert.Equal(t, "9090", config.Server.Port)
	assert.Equal(t, "localhost", config.Server.Host)

	// 데이터베이스 설정 확인
	assert.Equal(t, "test-db-host", config.Database.Host)
	assert.Equal(t, "3306", config.Database.Port)
	assert.Equal(t, "root", config.Database.Username)
	assert.Equal(t, "qwer1234!", config.Database.Password)
	assert.Equal(t, "g_dev", config.Database.Database)
	assert.Equal(t, 2, config.Database.LogLevel)

	// Redis 설정 확인
	assert.Equal(t, "localhost", config.Redis.Host)
	assert.Equal(t, "6380", config.Redis.Port)
	assert.Equal(t, "", config.Redis.Password)
	assert.Equal(t, 0, config.Redis.Database)

	// JWT 설정 확인
	assert.Equal(t, "test-secret-key", config.JWT.SecretKey)
	assert.Equal(t, 15*time.Minute, config.JWT.AccessTokenExpiry)
	assert.Equal(t, 168*time.Hour, config.JWT.RefreshTokenExpiry)

	// 보안 설정 확인
	assert.Equal(t, "http://localhost:3000,http://localhost:8081", config.Security.CORSAllowedOrigins)
	assert.Equal(t, 100, config.Security.RateLimitRequests)
	assert.Equal(t, time.Minute, config.Security.RateLimitWindow)

	// 로깅 설정 확인
	assert.Equal(t, "info", config.Log.Level)
	assert.Equal(t, "json", config.Log.Format)

	// 게임 설정 확인
	assert.Equal(t, 1, config.Game.DefaultLevel)
	assert.Equal(t, 1000, config.Game.DefaultGold)
	assert.Equal(t, 10, config.Game.DefaultDiamond)
}

// JWT 시크릿 키가 없을 때의 에러를 테스트
func TestLoadConfig_MissingJWTSecret(t *testing.T) {
	// JWT_SECRET_KEY 환경변수 제거
	os.Unsetenv("JWT_SECRET_KEY")

	// 설정 로드 시도
	config, err := LoadConfig()
	log.Printf("config: %v", config)
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "JWT_SECRET_KEY가 설정되어 있지 않습니다")
}

// 잘못된 시간 형식에 대한 에러를 테스트
func TestLoadConfig_InvalidDuration(t *testing.T) {
	// 테스트용 환경변수 설정 (잘못된 시간 형식)
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	os.Setenv("JWT_ACCESS_TOKEN_EXPIRY", "invalid-duration")

	// 설정 로드 시도
	config, err := LoadConfig()
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "잘못된 JWT_ACCESS_TOKEN_EXPIRY 형식")
}

// 기본값 설정을 테스트
func TestLoadConfig_DefaultValues(t *testing.T) {
	// 모든 환경변수 제거 후 최소한의 환경변수만 설정
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_HOST")
	os.Unsetenv("REDIS_PORT")
	os.Unsetenv("JWT_ACCESS_TOKEN_EXPIRY")
	os.Unsetenv("JWT_REFRESH_TOKEN_EXPIRY")
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")

	// 설정 로드
	config, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// 기본값 확인
	assert.Equal(t, "8081", config.Server.Port)
	assert.Equal(t, "127.0.0.1", config.Database.Host)
	assert.Equal(t, "6379", config.Redis.Port)
}

// 설정 검증 기능을 테스트
func TestValidateConfig(t *testing.T) {
	// 유효한 설정
	validConfig := &Config{
		JWT: JWTConfig{
			SecretKey: "valid-secret",
		},
		Database: DatabaseConfig{
			Host: "localhost",
			Port: "3306",
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: "6379",
		},
	}

	err := ValidateConfig(validConfig)
	assert.NoError(t, err)
}

// 잘못된 설정에 대한 검증을 테스트
func TestValidateConfig_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "빈 JWT 시크릿 키",
			config: &Config{
				JWT: JWTConfig{
					SecretKey: "",
				},
				Database: DatabaseConfig{
					Host: "localhost",
					Port: "3306",
				},
				Redis: RedisConfig{
					Host: "localhost",
					Port: "6379",
				},
			},
			wantErr: true,
		},
		{
			name: "빈 데이터베이스 호스트",
			config: &Config{
				JWT: JWTConfig{
					SecretKey: "valid-secret",
				},
				Database: DatabaseConfig{
					Host: "",
					Port: "3306",
				},
				Redis: RedisConfig{
					Host: "localhost",
					Port: "6379",
				},
			},
			wantErr: true,
		},
		{
			name: "빈 Redis 포트",
			config: &Config{
				JWT: JWTConfig{
					SecretKey: "valid-secret",
				},
				Database: DatabaseConfig{
					Host: "localhost",
					Port: "3306",
				},
				Redis: RedisConfig{
					Host: "localhost",
					Port: "",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// 환경변수 기본값 함수를 테스트
func TestGetEnvOrDefault(t *testing.T) {
	// 환경변수가 설정된 경우
	os.Setenv("TEST_KEY", "test-value")
	value := getEnvOrDefault("TEST_KEY", "default-value")
	assert.Equal(t, "test-value", value)

	// 환경변수가 설정되지 않은 경우
	os.Unsetenv("TEST_KEY")
	value = getEnvOrDefault("TEST_KEY", "default-value")
	assert.Equal(t, "default-value", value)
}

// 정수 환경변수 기본값 함수를 테스트
func TestGetEnvAsIntOrDefault(t *testing.T) {
	// 유효한 정수 환경변수
	os.Setenv("TEST_INT", "123")
	value := getEnvAsIntOrDefault("TEST_INT", 0)
	assert.Equal(t, 123, value)

	// 잘못된 정수 환경변수
	os.Setenv("TEST_INT", "invalid")
	value = getEnvAsIntOrDefault("TEST_INT", 456)
	assert.Equal(t, 456, value)

	// 환경변수가 설정되지 않은 경우
	os.Unsetenv("TEST_INT")
	value = getEnvAsIntOrDefault("TEST_INT", 789)
	assert.Equal(t, 789, value)
}

// 설정 로딩의 사용 예시를 제공
func ExampleLoadConfig() {
	// 환경변수 설정
	os.Setenv("JWT_SECRET_KEY", "example-secret-key")
	os.Setenv("PORT", "8081")
	os.Setenv("DATABASE_HOST", "localhost")

	// 설정 로드
	config, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	// 설정 검증
	err = ValidateConfig(config)
	if err != nil {
		panic(err)
	}

	// 설정 사용
	fmt.Printf("Server Port: %s\n", config.Server.Port)
	fmt.Printf("Database Host: %s\n", config.Database.Host)
	fmt.Printf("JWT Secret Key: %s\n", config.JWT.SecretKey[:10]+"...")

	// Output:
	// Warning: .env file not found, using environment variables only
	// Server Port: 8081
	// Database Host: localhost
	// JWT Secret Key: example-se...
}
