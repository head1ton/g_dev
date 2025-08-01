package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"time"
)

// Redis 연결 설정
type RedisConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	Database int    `json:"database"` // 데이터베이스 번호 (기본값: 0)

	PoolSize     int `json:"pool_size"`      // 연결 풀 크기 (기본값: 10)
	MinIdleConns int `json:"min_idle_conns"` // 최소 유휴 연결 수 (기본값 : 5)

	DialTimeout  time.Duration `json:"dial_timeout"`  // 연결 타임아웃 (기본값 ; 5초)
	ReadTimeout  time.Duration `json:"read_timeout"`  // 읽기 타임아웃 (기본값: 3초)
	WriteTimeout time.Duration `json:"write_timeout"` // 쓰기 타임아웃 (기본값: 3초)

	PoolTimeout time.Duration `json:"pool_timeout"` // 풀 타임아웃 (기본값: 4초)
	IdleTimeout time.Duration `json:"idle_timeout"` // 유휴 타임아웃 (기본값: 5초)

	// 재시도 설정
	MaxRetries      int           `json:"max_retries"`       // 최대 재시도 횟수 (기본값: 3)
	MinRetryBackoff time.Duration `json:"min_retry_backoff"` // 최소 재시도 간격 (기본값: 8ms)
	MaxRetryBackoff time.Duration `json:"max_retry_backoff"` // 최대 재시도 간격 (기본값: 512ms)
}

// 클라이언트 래핑한 구조체
type RedisClient struct {
	// Redis 클라이언트 인스턴스
	Client *redis.Client

	// Redis 설정
	Config RedisConfig

	// 연결 상태
	IsConnected bool

	// 컨텍스트
	Ctx context.Context
}

// 기본 Redis 설정을 생성
// 환경변수나 사용자 설정에 따라 커스터마이징
func NewRedisConfig() RedisConfig {
	return RedisConfig{
		Host:            getRedisHost(),
		Port:            getRedisPort(),
		Password:        getRedisPassword(),
		Database:        0,
		PoolSize:        10,
		MinIdleConns:    5,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolTimeout:     4 * time.Second,
		IdleTimeout:     5 * time.Minute,
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
	}
}

// 새로운 RedisClient 인스턴스를 생성
// 설정을 받아서 Redis 클라이언트를 초기화
func NewRedisClient(config RedisConfig) *RedisClient {
	return &RedisClient{
		Config:      config,
		IsConnected: false,
		Ctx:         context.Background(),
	}
}

// Redis 서버에 연결을 시도
// Redis 클라이언트를 생성하고 연결을 확인
func (r *RedisClient) Connect() error {
	// Redis 클라이언트 생성
	r.Client = redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%s", r.Config.Host, r.Config.Port),
		Password:        r.Config.Password,
		DB:              r.Config.Database,
		PoolSize:        r.Config.PoolSize,
		MinIdleConns:    r.Config.MinIdleConns,
		DialTimeout:     r.Config.DialTimeout,
		ReadTimeout:     r.Config.ReadTimeout,
		WriteTimeout:    r.Config.WriteTimeout,
		PoolTimeout:     r.Config.PoolTimeout,
		MaxRetries:      r.Config.MaxRetries,
		MinRetryBackoff: r.Config.MinRetryBackoff,
		MaxRetryBackoff: r.Config.MaxRetryBackoff,
	})

	// 연결 테스트
	if err := r.Client.Ping(r.Ctx).Err(); err != nil {
		log.Printf("Redis 연결 실패: %v", err)
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	r.IsConnected = true
	log.Printf("Redis 연결 성공: %s:%s", r.Config.Host, r.Config.Port)
	return nil
}

// Redis 연결을 종료
// 클라이언트를 정리하고 연결 상태를 업데이트
func (r *RedisClient) Disconnect() error {
	if r.Client != nil {
		if err := r.Client.Close(); err != nil {
			log.Printf("Redis 연결 종료 실패: %v", err)
			return fmt.Errorf("failed to disconnect from Redis: %w", err)
		}
	}

	r.IsConnected = false
	log.Printf("Redis 연결 종료: %s:%s", r.Config.Host, r.Config.Port)
	return nil
}

// Redis 연결 상태를 확인
// Ping 명령어를 사용하여 연결이 정상인지 확인.
func (r *RedisClient) IsHealthy() bool {
	if r.Client == nil || !r.IsConnected {
		return false
	}

	if err := r.Client.Ping(r.Ctx).Err(); err != nil {
		log.Printf("Redis 헬스체크 실패: %v", err)
		return false
	}

	return true
}

// Redis 서버의 통계 정보를 반환
// 연결 상태, 메모리 사용량, 명령어 통계 등을 포함
func (r *RedisClient) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"connected": r.IsConnected,
		"host":      r.Config.Host,
		"port":      r.Config.Port,
		"database":  r.Config.Database,
	}

	if r.Client != nil && r.IsConnected {
		// Redis INFO 명령어로 서버 정보 가져오기
		if info, err := r.Client.Info(r.Ctx).Result(); err == nil {
			stats["info"] = info
		}

		// 연결 풀 통계
		poolStats := r.Client.PoolStats()
		stats["pool_stats"] = map[string]interface{}{
			"total_connections": poolStats.TotalConns,
			"idle_connections":  poolStats.IdleConns,
			"stale_connections": poolStats.StaleConns,
		}
	}

	return stats
}

// Redis에 키-값 쌍을 저장
// 만료 시간을 설정
func (r *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	if !r.IsHealthy() {
		return fmt.Errorf("Redis is not connected")
	}

	return r.Client.Set(r.Ctx, key, value, expiration).Err()
}

// Redis에서 키에 해당하는 값을 가져옴
func (r *RedisClient) Get(key string) (string, error) {
	if !r.IsHealthy() {
		return "", fmt.Errorf("Redis is not connected")
	}

	return r.Client.Get(r.Ctx, key).Result()
}

// Redis에서 지정된 키들을 삭제
func (r *RedisClient) Del(keys ...string) error {
	if !r.IsHealthy() {
		return fmt.Errorf("Redis is not connected")
	}

	return r.Client.Del(r.Ctx, keys...).Err()
}

// 지정된 키가 존재하는지 확인
func (r *RedisClient) Exists(key string) (bool, error) {
	if !r.IsHealthy() {
		return false, fmt.Errorf("Redis is not connected")
	}

	result, err := r.Client.Exists(r.Ctx, key).Result()
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

// 키에 만료 시간을 설정함
func (r *RedisClient) Expire(key string, expiration time.Duration) error {
	if !r.IsHealthy() {
		return fmt.Errorf("Redis is not connected")
	}

	return r.Client.Expire(r.Ctx, key, expiration).Err()
}

// 키의 남은 만료 시간을 반환
func (r *RedisClient) TTL(key string) (time.Duration, error) {
	if !r.IsHealthy() {
		return 0, fmt.Errorf("Redis is not connected")
	}

	return r.Client.TTL(r.Ctx, key).Result()
}

// Redis 호스트를 결정
// 환경변수 REDIS_HOST가 설정되어 있으면 그 값을 사용하고,
// 없으면 기본값 "localhost"를 사용
func getRedisHost() string {
	if host := os.Getenv("REDIS_HOST"); host != "" {
		return host
	}
	return "localhost"
}

// Redis 포트를 결정
// 환경변수 REDIS_PORT가 설정되어 있으면 그 값을 사용하고,
// 없으면 기본값 "6379"를 사용
func getRedisPort() string {
	if port := os.Getenv("REDIS_PORT"); port != "" {
		return port
	}
	return "6379"
}

// Redis 비밀번호를 결정
// 환경변수 REDIS_PASSWORD가 설정되어 있으면 그 값을 사용하고,
// 없으면 기본값 ""(빈 문자열)을 사용
func getRedisPassword() string {
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		return password
	}
	return ""
}
