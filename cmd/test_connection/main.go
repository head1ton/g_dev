package main

import (
	"fmt"
	"time"

	"g_dev/internal/database"
	"g_dev/internal/redis"
)

// main 함수는 프로그램의 진입점
// MySQL과 Redis 연결을 순차적으로 테스트
func main() {
	fmt.Println("=== G-Dev 연결 테스트 시작 ===")
	fmt.Println()

	// MySQL 연결 테스트
	fmt.Println("1. MySQL 연결 테스트")
	if err := testMySQLConnection(); err != nil {
		fmt.Printf("MySQL 연결 실패: %v\n", err)
	} else {
		fmt.Println("MySQL 연결 성공")
	}
	fmt.Println()

	// 잠시 대기
	time.Sleep(1 * time.Second)

	// Redis 연결 테스트
	fmt.Println("2. Redis 연결 테스트")
	if err := testRedisConnection(); err != nil {
		fmt.Printf("Redis 연결 실패: %v\n", err)
	} else {
		fmt.Println("Redis 연결 성공")
	}
	fmt.Println()

	fmt.Println("=== 연결 테스트 완료 ===")
}

// testMySQLConnection은 MySQL 데이터베이스 연결을 테스트합니다.
// 데이터베이스 설정을 생성하고 연결을 시도한 후 통계 정보를 출력합니다.
func testMySQLConnection() error {
	// 데이터베이스 설정 생성
	config := database.NewDatabaseConfig()
	fmt.Printf("   - 호스트: %s\n", config.Host)
	fmt.Printf("   - 포트: %s\n", config.Port)
	fmt.Printf("   - 사용자: %s\n", config.Username)
	fmt.Printf("   - 데이터베이스: %s\n", config.Database)

	// 데이터베이스 인스턴스 생성
	db := database.NewDatabase(config)

	// 연결 시도
	fmt.Println("   - 연결 시도 중...")
	if err := db.Connect(); err != nil {
		return fmt.Errorf("데이터베이스 연결 실패: %w", err)
	}

	// 연결 상태 확인
	if !db.IsHealthy() {
		return fmt.Errorf("데이터베이스 헬스체크 실패")
	}

	// 통계 정보 출력
	stats := db.GetStats()
	fmt.Printf("   - 연결 상태: %v\n", stats["connected"])
	fmt.Printf("   - SQL 통계: %v\n", stats["sql_stats"])

	// 연결 종료
	if err := db.Disconnect(); err != nil {
		return fmt.Errorf("데이터베이스 연결 종료 실패: %w", err)
	}

	return nil
}

// testRedisConnection은 Redis 서버 연결을 테스트합니다.
// Redis 설정을 생성하고 연결을 시도한 후 기본적인 Redis 작업을 테스트합니다.
func testRedisConnection() error {
	// Redis 설정 생성
	config := redis.NewRedisConfig()
	fmt.Printf("   - 호스트: %s\n", config.Host)
	fmt.Printf("   - 포트: %s\n", config.Port)
	fmt.Printf("   - 데이터베이스: %d\n", config.Database)

	// Redis 클라이언트 생성
	redisClient := redis.NewRedisClient(config)

	// 연결 시도
	fmt.Println("   - 연결 시도 중...")
	if err := redisClient.Connect(); err != nil {
		return fmt.Errorf("Redis 연결 실패: %w", err)
	}

	// 연결 상태 확인
	if !redisClient.IsHealthy() {
		return fmt.Errorf("Redis 헬스체크 실패")
	}

	// 기본 Redis 작업 테스트
	fmt.Println("   - 기본 작업 테스트 중...")

	// 키-값 저장 테스트
	testKey := "g_dev_test_connection"
	testValue := "Hello Redis!"

	if err := redisClient.Set(testKey, testValue, 30*time.Second); err != nil {
		return fmt.Errorf("Redis SET 작업 실패: %w", err)
	}

	// 키-값 조회 테스트
	retrievedValue, err := redisClient.Get(testKey)
	if err != nil {
		return fmt.Errorf("Redis GET 작업 실패: %w", err)
	}

	if retrievedValue != testValue {
		return fmt.Errorf("Redis 값 불일치: 예상=%s, 실제=%s", testValue, retrievedValue)
	}

	// 키 존재 확인 테스트
	exists, err := redisClient.Exists(testKey)
	if err != nil {
		return fmt.Errorf("Redis EXISTS 작업 실패: %w", err)
	}

	if !exists {
		return fmt.Errorf("Redis 키가 존재하지 않음: %s", testKey)
	}

	// TTL 확인 테스트
	ttl, err := redisClient.TTL(testKey)
	if err != nil {
		return fmt.Errorf("Redis TTL 작업 실패: %w", err)
	}

	fmt.Printf("   - TTL: %v\n", ttl)

	// 테스트 키 삭제
	if err := redisClient.Del(testKey); err != nil {
		return fmt.Errorf("Redis DEL 작업 실패: %w", err)
	}

	// 통계 정보 출력
	stats := redisClient.GetStats()
	fmt.Printf("   - 연결 상태: %v\n", stats["connected"])
	if poolStats, ok := stats["pool_stats"].(map[string]interface{}); ok {
		fmt.Printf("   - 연결 풀 통계: %v\n", poolStats)
	}

	// 연결 종료
	if err := redisClient.Disconnect(); err != nil {
		return fmt.Errorf("Redis 연결 종료 실패: %w", err)
	}

	return nil
}
