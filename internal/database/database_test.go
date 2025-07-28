package database

import (
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// 기본 데이터베이스 설정 생성 테스트
func TestNewDatabaseConfig(t *testing.T) {
	config := NewDatabaseConfig()

	// 기본값 검증
	if config.FilePath == "" {
		t.Error("FilePath should not be empty")
	}

	if config.LogLevel < 0 || config.LogLevel > 4 {
		t.Errorf("LogLevel should be between 0 and 4, got %d", config.LogLevel)
	}

	if config.MaxOpenConns <= 0 {
		t.Error("MaxOpenConns should be positive")
	}

	if config.MaxIdleConns <= 0 {
		t.Error("MaxIdleConns should be positive.")
	}

	if config.ConnMaxLifetime <= 0 {
		t.Error("ConnMaxLifetime should be positive")
	}

	// 기본값 확인
	expectedMaxOpenConns := 10
	if config.MaxOpenConns != expectedMaxOpenConns {
		t.Errorf("Expected MaxOpenConns %d, got %d", expectedMaxOpenConns, config.MaxOpenConns)
	}

	expectedMaxIdleConns := 5
	if config.MaxIdleConns != expectedMaxIdleConns {
		t.Errorf("Expected MaxIdleConns %d, got %d", expectedMaxIdleConns, config.MaxIdleConns)
	}

	expectedConnMaxLifetime := time.Hour
	if config.ConnMaxLifetime != expectedConnMaxLifetime {
		t.Errorf("Expected ConnMaxLifetime %v, got %v", expectedConnMaxLifetime, config.ConnMaxLifetime)
	}

	if !config.AutoMigrate {
		t.Error("AutoMigrate should be true by default")
	}

	if config.Debug {
		t.Error("Debug should be false by default")
	}
}

// 새로운 Database 인스턴스 생성 테스트
func TestNewDatabase(t *testing.T) {
	config := NewDatabaseConfig()
	db := NewDatabase(config)

	if db == nil {
		t.Fatal("NewDatabase should not return nil")
	}

	if db.Config.FilePath != config.FilePath {
		t.Errorf("Expected FilePath %s, got %s", config.FilePath, db.Config.FilePath)
	}

	if db.IsConnected {
		t.Error("New database should not be connected")
	}

	if db.IsMigrated {
		t.Error("New database should not be migrated")
	}

	//t.Log(db)
	if db.DB != nil {
		t.Error("New database should not have DB instance")
	}
}

// 데이터베이스 연결 테스트
func TestDatabase_Connect(t *testing.T) {
	// 임시 데이터베이스 파일 경로
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	config := DatabaseConfig{
		FilePath:        dbPath,
		LogLevel:        1,
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
		AutoMigrate:     false,
		Debug:           false,
	}

	db := NewDatabase(config)

	// 연결 테스트
	if err := db.Connect(); err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 연결 상태 테스트
	if !db.IsConnected {
		t.Error("Database should be connected after Connect()")
	}

	if db.DB == nil {
		t.Error("Database should have DB instance after Connect()")
	}

	// health check
	if !db.IsHealthy() {
		t.Error("Database should be healthy after Connect()")
	}

	// 통계 정보 확인
	stats := db.GetStats()
	if stats["is_connected"] != true {
		t.Error("Stats should show database as connected")
	}

	if stats["file_path"] != dbPath {
		t.Errorf("Expected file_path %s, got %v", dbPath, stats["file_path"])
	}

	// 연결 종료
	if err := db.Disconnect(); err != nil {
		t.Errorf("Failed to disconnect: %v", err)
	}

	if db.IsConnected {
		t.Error("Database should not be connected after Disconnect()")
	}
}

// 잘못된 경로로 연결 시도 테스트
func TestDatabase_Connect_InvalidPath(t *testing.T) {
	// 읽기 전용 디렉토리에 데이터베이스 파일 생성 시도
	config := DatabaseConfig{
		FilePath: "/root/test.db",
		LogLevel: 1,
	}

	db := NewDatabase(config)

	// 연결 실패 예상
	if err := db.Connect(); err == nil {
		t.Error("Expected error when connecting to invalid path")
	}

	if db.IsConnected {
		t.Error("Database should not be connected after failed Connect()")
	}
}

// 마이그레이션 테스트
func TestDatabase_Migrate(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "migrate_test.db")

	config := DatabaseConfig{
		FilePath: dbPath,
		LogLevel: 1,
	}

	db := NewDatabase(config)

	// 연결
	if err := db.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Disconnect()

	// 테스트용 모델 구조체
	type TestModel struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100;not null"`
		Age  int    `gorm:"default:0"`
	}

	// 마이그레이션 실행
	if err := db.Migrate(&TestModel{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// 마이그레이션 상태 확인
	if !db.IsMigrated {
		t.Error("Database should be marked as migrated")
	}

	// 테이블이 실제로 생성되었는지 확인
	var count int64
	db.DB.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='test_models'").Scan(&count)
	if count == 0 {
		t.Error("Test table should be created after migration")
	}
}

// 연결되지 않은 상태에서 마이그레이션 시도 테스트
func TestDatabase_Migrate_NotConnected(t *testing.T) {
	config := NewDatabaseConfig()
	db := NewDatabase(config)

	// 연결하지 않은 상태에서 마이그레이션 시도
	type TestModel struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}

	if err := db.Migrate(&TestModel{}); err == nil {
		t.Error("Expected error when migrating without connection")
	}

	if db.IsMigrated {
		t.Error("Database should not be marked as migrated after failed migration")
	}
}

// 데이터베이스 헬스체크 테스트
func TestDatabase_IsHealthy(t *testing.T) {
	// 연결되지 않은 상태
	config := NewDatabaseConfig()
	db := NewDatabase(config)

	if db.IsHealthy() {
		t.Error("Unconnected database should not be healthy")
	}

	// 연결된 상태
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "health_test.db")

	config.FilePath = dbPath
	db = NewDatabase(config)

	if err := db.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Disconnect()

	if !db.IsHealthy() {
		t.Error("Connected database should be healthy")
	}
}

// 데이터베이스 통계 정보 테스트
func TestDatabase_GetStats(t *testing.T) {
	// 연결되지 않은 상태
	config := NewDatabaseConfig()
	db := NewDatabase(config)

	stats := db.GetStats()

	if stats["is_connected"] != false {
		t.Error("Stats should show database as not connected")
	}

	if stats["is_migrated"] != false {
		t.Error("Stats should show database as not migrated")
	}

	if stats["file_path"] != config.FilePath {
		t.Errorf("Expected file_path %s, got %v", config.FilePath, stats["file_path"])
	}

	// 연결된 상태
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "stats_test.db")

	config.FilePath = dbPath
	db = NewDatabase(config)

	if err := db.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Disconnect()

	stats = db.GetStats()

	if stats["is_connected"] != true {
		t.Error("Stats should show database as connected")
	}

	if stats["file_path"] != dbPath {
		t.Errorf("Expected file_path %s, got %v", dbPath, stats["file_path"])
	}

	// 연결 풀 통계 확인
	log.Printf("stats %v", stats["max_open_conns"])
	if stats["max_open_conns"] == nil {
		t.Error("Stats should include max_open_conns")
	}

	if stats["open_conns"] == nil {
		t.Error("Stats should include open_conns")
	}
}

// 데이터베이스 파일 경로 결정 테스트
func TestGetDatabaseFilePath(t *testing.T) {
	// 환경변수가 설정되지 않은 경우
	originalPath := os.Getenv("DATABASE_PATH")
	os.Unsetenv("DATABASE_PATH")
	defer os.Setenv("DATABASE_PATH", originalPath)

	path := getDatabaseFilePath()
	expectedPath := "data/g_dev.db"
	if path != expectedPath {
		t.Errorf("Expected default path %s, got %s", expectedPath, path)
	}

	// 환경변수가 설정된 경우
	testPath := "/custom/path/database.db"
	os.Setenv("DATABASE_PATH", testPath)

	path = getDatabaseFilePath()
	if path != testPath {
		t.Errorf("Expected custom path %s, got %s", testPath, path)
	}
}

// 로깅 레벨 결정 테스트
func TestGetLogLevel(t *testing.T) {
	// 환경변수가 설정되지 않은 경우
	originalLevel := os.Getenv("DATABASE_LOG_LEVEL")
	os.Unsetenv("DATABASE_LOG_LEVEL")
	defer os.Setenv("DATABASE_LOG_LEVEL", originalLevel)

	level := getLogLevel()
	expectedLevel := 2 // 기본값: Warn
	if level != expectedLevel {
		t.Errorf("Expected default level %d, got %d", expectedLevel, level)
	}

	// 다양한 환경변수 값 테스트
	testCases := []struct {
		envValue      string
		expectedLevel int
	}{
		{"0", 0},
		{"silent", 0},
		{"1", 1},
		{"error", 1},
		{"2", 2},
		{"warn", 2},
		{"3", 3},
		{"info", 3},
		{"4", 4},
		{"debug", 4},
		{"invalid", 2}, // 잘못된 값은 기본값 사용
	}

	for _, tc := range testCases {
		os.Setenv("DATABASE_LOG_LEVEL", tc.envValue)
		level = getLogLevel()
		if level != tc.expectedLevel {
			t.Errorf("For env value '%s', expected level %d, got %d", tc.envValue, tc.expectedLevel, level)
		}
	}
}

// 파일 경로에서 디렉토리 추출 테스트
func TestGetDirectoryFromPath(t *testing.T) {
	testCases := []struct {
		path     string
		expected string
	}{
		{"data/g_dev.db", "data"},
		{"/absolute/path/database.db", "/absolute/path"},
		{"database.db", ""}, // 현재 디렉토리
		{"", ""},            // 빈 경로
		{"data/subdir/file.db", "data/subdir"},
		{"C:\\Windows\\System32\\file.db", "C:\\Windows\\System32"},
	}

	for _, tc := range testCases {
		result := getDirectoryFromPath(tc.path)
		if result != tc.expected {
			t.Errorf("For path '%s', expected directory '%s', got '%s'", tc.path, tc.expected, result)
		}
	}
}

// 데이터베이스 연결 종료 테스트
func TestDatabase_Disconnect(t *testing.T) {
	// 연결되지 않은 상태에서 종료
	config := NewDatabaseConfig()
	db := NewDatabase(config)

	if err := db.Disconnect(); err != nil {
		t.Errorf("Disconnect should not fail when not connected: %v", err)
	}

	// 연결된 상태에서 종료
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "disconnect_test.db")

	config.FilePath = dbPath
	db = NewDatabase(config)

	if err := db.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	if err := db.Disconnect(); err != nil {
		t.Errorf("Failed to disconnect: %v", err)
	}

	if db.IsConnected {
		t.Error("Database should not be connected after Disconnect()")
	}
}

// NewDatabaseConfig 함수의 사용 예제
func ExampleNewDatabaseConfig() {
	config := NewDatabaseConfig()

	// 설정 커스터마이징
	config.FilePath = "custom/path/database.db"
	config.LogLevel = 3 // Info
	config.MaxOpenConns = 20
	config.Debug = true

	// Database 인스턴스 생성
	db := NewDatabase(config)

	// 연결
	if err := db.Connect(); err != nil {
		panic(err)
	}
	defer db.Disconnect()

	// 사용 예제
	stats := db.GetStats()
	_ = stats // 통계 정보 사용

	// Output:
}

// Database.Migrate 메서드의 사용 예제
func ExampleDatabase_Migrate() {
	config := NewDatabaseConfig()
	db := NewDatabase(config)

	if err := db.Connect(); err != nil {
		panic(err)
	}
	defer db.Disconnect()

	// 모델 정의
	type User struct {
		ID       uint   `gorm:"primarykey"`
		Username string `gorm:"size:100;uniqueIndex;not null"`
		Email    string `gorm:"size:255;uniqueIndex;not null"`
		Age      int    `gorm:"default:0"`
	}

	type Game struct {
		ID          uint   `gorm:"primarykey"`
		Name        string `gorm:"size:100;not null"`
		Description string `gorm:"size:500"`
		Score       int    `gorm:"default:0"`
	}

	// 마이그레이션 실행
	if err := db.Migrate(&User{}, &Game{}); err != nil {
		panic(err)
	}

	// Output:
}
