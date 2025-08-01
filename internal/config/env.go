package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
)

// 서버 관련 설정
type ServerConfig struct {
	Port string
	Host string
}

// 데이터베이스 관련 설정
type DatabaseConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
	LogLevel int
}

// Redis 관련 설정
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	Database int
}

// JWT 인증 관련 설정
type JWTConfig struct {
	SecretKey          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

// 보안 관련 설정
type SecurityConfig struct {
	CORSAllowedOrigins string
	RateLimitRequests  int
	RateLimitWindow    time.Duration
}

// 로깅 관련 설정
type LogConfig struct {
	Level  string
	Format string
}

// 게임 관련 설정
type GameConfig struct {
	DefaultLevel   int
	DefaultGold    int
	DefaultDiamond int
}

// 전체 애플리케이션 설정
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Security SecurityConfig
	Log      LogConfig
	Game     GameConfig
}

// LoadConfig는 환경변수에서 설정 로드
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, using environment variables only")
	}

	config := &Config{}

	// 서버 설정 로드
	config.Server = ServerConfig{
		Port: getEnvOrDefault("PORT", "8080"),
		Host: getEnvOrDefault("HOST", "localhost"),
	}

	// 데이터베이스 설정 로드
	config.Database = DatabaseConfig{
		Host:     getEnvOrDefault("DATABASE_HOST", "127.0.0.1"),
		Port:     getEnvOrDefault("DATABASE_PORT", "3306"),
		Username: getEnvOrDefault("DATABASE_USERNAME", "root"),
		Password: getEnvOrDefault("DATABASE_PASSWORD", "qwer1234!"),
		Database: getEnvOrDefault("DATABASE_NAME", "g_step"),
		LogLevel: getEnvAsIntOrDefault("DATABASE_LOG_LEVEL", 2),
	}

	// Redis 설정 로드
	config.Redis = RedisConfig{
		Host:     getEnvOrDefault("REDIS_HOST", "localhost"),
		Port:     getEnvOrDefault("REDIS_PORT", "6379"),
		Password: getEnvOrDefault("REDIS_PASSWORD", ""),
		Database: getEnvAsIntOrDefault("REDIS_DATABASE", 0),
	}

	// JWT 설정 로드 (필수)
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		return nil, fmt.Errorf("환경변수 JWT_SECRET_KEY가 설정되어 있지 않습니다")
	}

	accessExpiry, err := time.ParseDuration(getEnvOrDefault("JWT_ACCESS_TOKEN_EXPIRY", "15m"))
	if err != nil {
		return nil, fmt.Errorf("잘못된 JWT_ACCESS_TOKEN_EXPIRY 형식: %w", err)
	}

	refreshExpiry, err := time.ParseDuration(getEnvOrDefault("JWT_REFRESH_TOKEN_EXPIRY", "168h"))
	if err != nil {
		return nil, fmt.Errorf("잘못된 JWT_REFRESH_TOKEN_EXPIRY 형식: %w", err)
	}

	config.JWT = JWTConfig{
		SecretKey:          jwtSecret,
		AccessTokenExpiry:  accessExpiry,
		RefreshTokenExpiry: refreshExpiry,
	}

	// 보안 설정 로드
	rateLimitRequests := getEnvAsIntOrDefault("RATE_LIMIT_REQUESTS", 100)
	rateLimitWindow, err := time.ParseDuration(getEnvOrDefault("RATE_LIMIT_WINDOW", "1m"))
	if err != nil {
		return nil, fmt.Errorf("잘못된 RATE_LIMIT_WINDOW 형식: %w", err)
	}

	config.Security = SecurityConfig{
		CORSAllowedOrigins: getEnvOrDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:8080"),
		RateLimitRequests:  rateLimitRequests,
		RateLimitWindow:    rateLimitWindow,
	}

	// 로깅 설정 로드
	config.Log = LogConfig{
		Level:  getEnvOrDefault("LOG_LEVEL", "info"),
		Format: getEnvOrDefault("LOG_FORMAT", "json"),
	}

	// 게임 설정 로드
	config.Game = GameConfig{
		DefaultLevel:   getEnvAsIntOrDefault("GAME_DEFAULT_LEVEL", 1),
		DefaultGold:    getEnvAsIntOrDefault("GAME_DEFAULT_GOLD", 1000),
		DefaultDiamond: getEnvAsIntOrDefault("GAME_DEFAULT_DIAMOND", 10),
	}

	return config, nil
}

// 환경변수를 정수로 가져오거나 기본값을 반환
func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// 환경변수를 가져오거나 기본값을 반환
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 설정의 유효성 검사
func ValidateConfig(config *Config) error {
	// JWT 시크릿 키 검증
	if config.JWT.SecretKey == "" {
		return fmt.Errorf("JWT 시크릿 키가 설정되지 않음")
	}

	// 데이터베이스 설정 검증
	if config.Database.Host == "" || config.Database.Port == "" {
		return fmt.Errorf("데이터베이스 호스트 또는 포트가 설정되지 않았음")
	}

	// Redis 설정 검증
	if config.Redis.Host == "" || config.Redis.Port == "" {
		return fmt.Errorf("Redis 호스트 또는 포트가 설정되지 않았음")
	}

	return nil
}
