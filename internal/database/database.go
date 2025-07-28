package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

type DatabaseConfig struct {
	Host     string `json:"host"` // 데이터베이스 호스트
	Port     int    `json:"port"` // 포트
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`

	// 연결 옵션
	Charset   string `json:"charset"`    // 문자셋 (기본값 : utf8mb4)
	ParseTime bool   `json:"parse_time"` // 시간 파싱 여부 (기본값 : true)
	Loc       string `json:"loc"`        // 시간대 (기본값: Local)

	// 로깅 레벨 (0: Silent, 1: Error, 2: Warn, 3: Info, 4: Debug)
	LogLevel int `json:"log_level"`

	// 연결 풀 설정
	MaxOpenConns    int           `json:"max_open_conns"`    // 최대 열린 연결 수
	MaxIdleConns    int           `json:"max_idel_conns"`    // 최대 유휴 연결 수
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"` // 연결 최대 수명

	// 자동 마이그레이션 설정
	AutoMigrate bool `json:"auto_migrate"` // 서버 시작 시 마이그레이션 여부

	// 디버그 모드 설정
	Debug bool `json:"debug"` // GORM 디버그 모드 활성화 여부
}

// GORM DB 인스턴스 래핑
type Database struct {
	// GORM DB 인스턴스
	DB *gorm.DB

	// 데이터베이스 설정
	Config DatabaseConfig

	// 연결 상태
	IsConnected bool

	// 마이그레이션 상태
	IsMigrated bool
}

func NewDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:            getDatabaseHost(),
		Port:            getDatabasePort(),
		Username:        getDatabaseUsername(),
		Password:        getDatabasePassword(),
		Database:        getDatabaseName(),
		Charset:         "utf8mb4",
		ParseTime:       true,
		Loc:             "Local",
		LogLevel:        getLogLevel(),
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		AutoMigrate:     true,
		Debug:           false,
	}
}

func getDatabaseName() string {
	if name := os.Getenv("DATABASE_NAME"); name != "" {
		return name
	}
	return "g_dev"
}

func getDatabasePassword() string {
	if password := os.Getenv("DATABASE_PASSWORD"); password != "" {
		return password
	}
	return ""
}

func getDatabaseUsername() string {
	if username := os.Getenv("DATABASE_USERNAME"); username != "" {
		return username
	}
	return "root"
}

func getDatabasePort() int {
	if port := os.Getenv("DATABASE_PORT"); port != "" {
		switch port {
		case "3306":
			return 3306
		case "3307":
			return 3307
		case "3308":
			return 3308
		default:
			return 3306
		}
	}
	return 3306
}

func getDatabaseHost() string {
	if host := os.Getenv("DATABASE_HOST"); host != "" {
		return host
	}
	return "localhost"
}

func getLogLevel() int {
	if level := os.Getenv("DATABASE_LOG_LEVEL"); level != "" {
		switch level {
		case "0", "silent":
			return 0
		case "1", "error":
			return 1
		case "2", "warn":
			return 2
		case "3", "info":
			return 3
		case "4", "debug":
			return 4
		}
	}
	return 2 // 기본값 : Warn
}

func getDatabaseFilePath() string {
	if path := os.Getenv("DATABASE_PATH"); path != "" {
		return path
	}
	return "data/g_dev.db"
}

// 새로운 데이터베이스 인스턴스 생성
func NewDatabase(config DatabaseConfig) *Database {
	return &Database{
		Config:      config,
		IsConnected: false,
		IsMigrated:  false,
	}
}

// 데이터베이스 연결
func (d *Database) Connect() error {
	// GORM Logger
	gormLogger := d.createGormLogger()

	dsn := d.buildDSN()

	// GORM DB 인스턴스 생성
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 연결 풀 설정
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(d.Config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(d.Config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(d.Config.ConnMaxLifetime)

	// 연결 테스트
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	d.DB = db
	d.IsConnected = true

	log.Printf("데이터베이스 연결 성공: %s:%d/%s", d.Config.Host, d.Config.Port, d.Config.Database)
	return nil
}

// 데이터베이스 연결 종료
func (d *Database) Disconnect() error {
	if d.DB == nil {
		return nil
	}

	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	d.IsConnected = false
	log.Printf("데이터베이스 연결 종료: %s:%d/%s", d.Config.Host, d.Config.Port, d.Config.Database)
	return nil
}

// 데이터베이스 스키마 마이그레이션
// 모델 구조체들을 받아서 테이블 생성 및 업데이트
func (d *Database) Migrate(models ...interface{}) error {
	if !d.IsConnected {
		return fmt.Errorf("database is not connected")
	}

	// 자동 마이그레이션 설정
	if err := d.DB.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	d.IsMigrated = true
	log.Printf("데이터베이스 마이그레이션 완료: %d개 모델", len(models))
	return nil
}

// GORM DB 인스턴스 반환
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

// 데이터베이스 연결 상태 확인
func (d *Database) IsHealthy() bool {
	if !d.IsConnected || d.DB == nil {
		return false
	}

	sqlDB, err := d.DB.DB()
	if err != nil {
		return false
	}

	if err := sqlDB.Ping(); err != nil {
		return false
	}

	return true
}

// 데이터베이스 통계 정보 반환
// 연결 수, 상태 등의 정보 포함
func (d *Database) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"is_connected": d.IsConnected,
		"is_migrated":  d.IsMigrated,
		"host":         d.Config.Host,
		"port":         d.Config.Port,
		"database":     d.Config.Database,
		"username":     d.Config.Username,
	}

	if d.IsConnected && d.DB != nil {
		sqlDB, err := d.DB.DB()
		if err == nil {
			stats["max_open_conns"] = sqlDB.Stats().MaxOpenConnections
			stats["open_conns"] = sqlDB.Stats().OpenConnections
			stats["in_use"] = sqlDB.Stats().InUse
			stats["idle"] = sqlDB.Stats().Idle
		}
	}

	return stats
}

// 데이터베이스 연결 문자열(DSN) 생성
func (d *Database) buildDSN() string {
	// 기본 DSN 형식: username:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		d.Config.Username,
		d.Config.Password,
		d.Config.Host,
		d.Config.Port,
		d.Config.Database,
		d.Config.Charset,
		d.Config.ParseTime,
		d.Config.Loc,
	)

	return dsn
}

// GORM 로거 생성
func (d *Database) createGormLogger() logger.Interface {
	var logLevel logger.LogLevel
	switch d.Config.LogLevel {
	case 0:
		logLevel = logger.Silent
	case 1:
		logLevel = logger.Error
	case 2:
		logLevel = logger.Warn
	case 3:
		logLevel = logger.Info
	case 4:
		logLevel = logger.Info
	default:
		logLevel = logger.Warn
	}

	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
}

// 파일 경로에서 디렉토리 부분 추출
// data/g_dev.db -> data
func getDirectoryFromPath(filePath string) string {
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '/' || filePath[i] == '\\' {
			return filePath[:i]
		}
	}
	return ""
}
