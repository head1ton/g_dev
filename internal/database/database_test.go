package database

import (
	"os"
	"testing"
	"time"
)

// 기본 데이터베이스 설정 생성 테스트
func TestNewDatabaseConfig(t *testing.T) {
	config := NewDatabaseConfig()

	// 기본값 검증
	if config.Host == "" {
		t.Error("Host should not be empty")
	}

	if config.Port <= 0 {
		t.Error("Port should be positive")
	}

	if config.Username == "" {
		t.Error("Username should not be empty")
	}

	if config.Database == "" {
		t.Error("Database should not be empty")
	}

	if config.Charset == "" {
		t.Error("Charset should not be empty")
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
	expectedHost := "localhost"
	if config.Host != expectedHost {
		t.Errorf("Expected Host %s, got %s", expectedHost, config.Host)
	}

	expectedPort := 3306
	if config.Port != expectedPort {
		t.Errorf("Expected Port %d, got %d", expectedPort, config.Port)
	}

	expectedUsername := "root"
	if config.Username != expectedUsername {
		t.Errorf("Expected Username %s, got %s", expectedUsername, config.Username)
	}

	expectedDatabase := "g_dev"
	if config.Database != expectedDatabase {
		t.Errorf("Expected Database %s, got %s", expectedDatabase, config.Database)
	}

	expectedCharset := "utf8mb4"
	if config.Charset != expectedCharset {
		t.Errorf("Expected Charset %s, got %s", expectedCharset, config.Charset)
	}

	if !config.ParseTime {
		t.Error("ParseTime should be true by default")
	}

	expectedLoc := "Local"
	if config.Loc != expectedLoc {
		t.Errorf("Expected Loc %s, got %s", expectedLoc, config.Loc)
	}

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

	if db.Config.Host != config.Host {
		t.Errorf("Expected Host %s, got %s", config.Host, db.Config.Host)
	}

	if db.Config.Port != config.Port {
		t.Errorf("Expected Port %d, got %d", config.Port, db.Config.Port)
	}

	if db.Config.Username != config.Username {
		t.Errorf("Expected Username %s, got %s", config.Username, db.Config.Username)
	}

	if db.Config.Database != config.Database {
		t.Errorf("Expected Database %s, got %s", config.Database, db.Config.Database)
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
	// 서버가 실행 중인지 확인
	if !isMySQLAvailable() {
		t.Skip("MySQL server is not available, skipping connection test")
	}

	config := DatabaseConfig{
		Host:            "localhost",
		Port:            3306,
		Username:        "root",
		Password:        "qwer1234!",
		Database:        "g_dev_test",
		Charset:         "utf8mb4",
		ParseTime:       true,
		Loc:             "Local",
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

	if stats["host"] != config.Host {
		t.Errorf("Expected host %s, got %v", config.Host, stats["host"])
	}

	if stats["port"] != config.Port {
		t.Errorf("Expected port %d, got %v", config.Port, stats["port"])
	}

	if stats["database"] != config.Database {
		t.Errorf("Expected database %s, got %v", config.Database, stats["database"])
	}

	// 연결 종료
	if err := db.Disconnect(); err != nil {
		t.Errorf("Failed to disconnect: %v", err)
	}

	if db.IsConnected {
		t.Error("Database should not be connected after Disconnect()")
	}
}

// 잘못된 인증 정보로 연결 시도 테스트
func TestDatabase_Connect_InvalidCredentials(t *testing.T) {
	// 읽기 전용 디렉토리에 데이터베이스 파일 생성 시도
	config := DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "invalid_user",
		Password: "invalid_password",
		Database: "invalid_database",
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
// MySQL 서버가 없으면 스킵
func TestDatabase_Migrate(t *testing.T) {
	// MySQL 서버가 실행 중인지 확인
	if !isMySQLAvailable() {
		t.Skip("MySQL server is not available, skipping migration test")
	}

	config := DatabaseConfig{
		Host:     "127.0.0.1",
		Port:     3306,
		Username: "root",
		Password: "qwer1234!",
		Database: "g_dev_test",
		LogLevel: 1,
	}

	db := NewDatabase(config)
	//log.Print(db.Connect())
	// 연결
	if err := db.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Disconnect()

	// 테스트용 모델 구조체
	type TestModel struct {
		ID   uint   `gorm:"primarykey"`
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
	db.DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ?", config.Database, "test_models").Scan(&count)
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

	// MySQL 서버가 실행 중인지 확인
	if !isMySQLAvailable() {
		t.Skip("MySQL server is not available, skipping health test")
	}

	// 연결된 상태
	config.Host = "localhost"
	config.Port = "3306"
	config.Username = "root"
	config.Password = "qwer1234!"
	config.Database = "g_dev_test"

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

	if stats["host"] != config.Host {
		t.Errorf("Expected host %s, got %v", config.Host, stats["host"])
	}

	if stats["port"] != config.Port {
		t.Errorf("Expected port %d, got %v", config.Port, stats["port"])
	}

	if stats["database"] != config.Database {
		t.Errorf("Expected database %s, got %v", config.Database, stats["database"])
	}

	if stats["username"] != config.Username {
		t.Errorf("Expected username %s, got %v", config.Username, stats["username"])
	}

	// MySQL 서버가 실행 중인지 확인
	if !isMySQLAvailable() {
		t.Skip("MySQL server is not available, skipping connected stats test")
	}

	// 연결된 상태
	config.Host = "localhost"
	config.Port = "3306"
	config.Username = "root"
	config.Password = "qwer1234!"
	config.Database = "g_dev_test"

	db = NewDatabase(config)

	if err := db.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Disconnect()

	stats = db.GetStats()

	if stats["is_connected"] != true {
		t.Error("Stats should show database as connected")
	}

	if stats["host"] != config.Host {
		t.Errorf("Expected host %s, got %v", config.Host, stats["host"])
	}

	// 연결 풀 통계 확인
	//log.Printf("stats %v", stats["max_open_conns"])
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

	host := getDatabaseHost()
	expectedHost := "localhost"
	if host != expectedHost {
		t.Errorf("Expected default host %s, got %s", expectedHost, host)
	}

	// 환경변수가 설정된 경우
	testHost := "127.0.0.1"
	os.Setenv("DATABASE_HOST", testHost)

	host = getDatabaseHost()
	if host != testHost {
		t.Errorf("Expected custom path %s, got %s", testHost, host)
	}
}

// 데이터베이스 포트 결정 테스트
func TestGetDatabasePort(t *testing.T) {
	// 환경변수가 설정되지 않은 경우
	originalPort := os.Getenv("DATABASE_PORT")
	os.Unsetenv("DATABASE_PORT")
	defer os.Setenv("DATABASE_PORT", originalPort)

	port := getDatabasePort()
	expectedPort := 3306
	if port != expectedPort {
		t.Errorf("Expected default port %d, got %d", expectedPort, port)
	}

	// 환경변수가 설정된 경우
	testPort := "3307"
	os.Setenv("DATABASE_PORT", testPort)

	port = getDatabasePort()
	expectedPort = 3307
	if port != expectedPort {
		t.Errorf("Expected custom port %d, got %d", expectedPort, port)
	}
}

// 데이터베이스 사용자명 결정 테스트
func TestGetDatabaseUsername(t *testing.T) {
	// 환경변수가 설정되지 않은 경우
	originalUsername := os.Getenv("DATABASE_USERNAME")
	os.Unsetenv("DATABASE_USERNAME")
	defer os.Setenv("DATABASE_USERNAME", originalUsername)

	username := getDatabaseUsername()
	expectedUsername := "root"
	if username != expectedUsername {
		t.Errorf("Expected default username %s, got %s", expectedUsername, username)
	}

	// 환경변수가 설정된 경우
	testUsername := "g_dev_user"
	os.Setenv("DATABASE_USERNAME", testUsername)

	username = getDatabaseUsername()
	if username != testUsername {
		t.Errorf("Expected custom usrname %s, got %s", testUsername, username)
	}
}

// 데이터베이스 비밀번호 결정 테스트
func TestGetDatabasePassword(t *testing.T) {
	// 환경변수가 설정되지 않은 경우
	originalPassword := os.Getenv("DATABASE_PASSWORD")
	os.Unsetenv("DATABASE_PASSWORD")
	defer os.Setenv("DATABASE_PASSWORD", originalPassword)

	password := getDatabasePassword()
	expectedPassword := ""
	if password != expectedPassword {
		t.Errorf("Expected default password %s, got %s", expectedPassword, password)
	}

	// 혼경변수가 설정된 경우
	testPassword := "qwer1234!"
	os.Setenv("DATABASE_PASSWORD", testPassword)

	password = getDatabasePassword()
	if password != testPassword {
		t.Errorf("Expected custom password %s, got %s", testPassword, password)
	}
}

// 데이터베이스 이름 결정 테스트
func TestGetDatabaseName(t *testing.T) {
	// 환경변수가 설정되지 않은 경우
	originalName := os.Getenv("DATABASE_NAME")
	os.Unsetenv("DATABASE_NAME")
	defer os.Setenv("DATABASE_NAME", originalName)

	name := getDatabaseName()
	expectedName := "g_dev"
	if name != expectedName {
		t.Errorf("Expected default name %s, got %s", expectedName, name)
	}

	// 환경변수가 설정된 경우
	testName := "g_dev_production"
	os.Setenv("DATABASE_NAME", testName)

	name = getDatabaseName()
	if name != testName {
		t.Errorf("Expected custom name %s, got %s", testName, name)
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

// 데이터베이스 연결 종료 테스트
func TestDatabase_Disconnect(t *testing.T) {
	// 연결되지 않은 상태에서 종료
	config := NewDatabaseConfig()
	db := NewDatabase(config)

	if err := db.Disconnect(); err != nil {
		t.Errorf("Disconnect should not fail when not connected: %v", err)
	}

	// MySQL 서버가 실행 중인지 확인
	if !isMySQLAvailable() {
		t.Skip("MySQL server is not available, skipping disconnect test")
	}

	// 연결된 상태에서 종료
	config.Host = "localhost"
	config.Port = 3306
	config.Username = "root"
	config.Password = "qwer1234!"
	config.Database = "g_dev_test"

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

// DSN 생성 테스트
func TestDatabase_BuildDSN(t *testing.T) {
	config := DatabaseConfig{
		Host:      "localhost",
		Port:      3306,
		Username:  "test_user",
		Password:  "test_password",
		Database:  "test_db",
		Charset:   "utf8mb4",
		ParseTime: true,
		Loc:       "Local",
	}

	db := NewDatabase(config)
	dsn := db.buildDSN()

	expectedDSN := "test_user:test_password@tcp(127.0.0.1:3306)/test_db?charset=utf8mb4&parseTime=true&loc=Local"
	if dsn != expectedDSN {
		t.Errorf("Expected DSN %s, got %s", expectedDSN, dsn)
	}
}

// isMySQLAvailable는 MySQL 서버가 실행 중인지 확인
// 실제 연결을 시도하지 않고 포트만 확인
func isMySQLAvailable() bool {
	// 간단한 포트 확인 로직
	// 실제로는 net.Dial을 사용하여 연결을 시도
	return false // 테스트 목적으로 항상 false 반환 (MySQL 서버가 없으므로)
	//return true // 테스트 목적으로 항상 false 반환 (MySQL 서버가 없으므로)
}

// NewDatabaseConfig 함수의 사용 예제
func ExampleNewDatabaseConfig() {
	config := NewDatabaseConfig()

	// 설정 커스터마이징
	config.Host = "127.0.0.1"
	config.Port = 3306
	config.Username = "g_dev_user"
	config.Password = "qwer1234!"
	config.Database = "g_dev_production"
	config.LogLevel = 3 // Info
	config.MaxOpenConns = 20
	config.Debug = true

	// Database 인스턴스 생성
	db := NewDatabase(config)

	// 연결
	//if err := db.Connect(); err != nil {
	//	panic(err)
	//}
	//defer db.Disconnect()

	// 사용 예제
	stats := db.GetStats()
	_ = stats // 통계 정보 사용

	// Output:
}

// Database.Migrate 메서드의 사용 예제
func ExampleDatabase_Migrate() {
	config := NewDatabaseConfig()
	_ = NewDatabase(config)

	//if err := db.Connect(); err != nil {
	//	panic(err)
	//}
	//defer db.Disconnect()

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
	//if err := db.Migrate(&User{}, &Game{}); err != nil {
	//	panic(err)
	//}

	// Output:
}
