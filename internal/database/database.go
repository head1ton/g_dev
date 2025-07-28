package database

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

type DatabaseConfig struct {
	FilePath string `json:"file_path"`

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
		FilePath:        getDatabaseFilePath(),
		LogLevel:        getLogLevel(),
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		AutoMigrate:     true,
		Debug:           false,
	}
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
	if err := d.ensureDatabaseDirectory(); err != nil {
		return fmt.Errorf("failed to ensure database directory: %w", err)
	}

	// GORM Logger
	gormLogger := d.createGormLogger()

	// SQLite 연결 설정
	dsn := fmt.Sprintf("%s?cache=shared&_foreign_keys=on", d.Config.FilePath)

	// GORM DB 인스턴스 생성
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
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

	log.Printf("Connected to database at %s", d.Config.FilePath)
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
	log.Printf("데이터베이스 연결 종료: %s", d.Config.FilePath)
	return nil
}

// 데이터베이스 파일이 위치할 디렉토리가 존재하는지 확인
func (d *Database) ensureDatabaseDirectory() error {
	dir := getDirectoryFromPath(d.Config.FilePath)
	if dir == "" {
		return nil // 현재 디렉토리에 생성하는 경우
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory %s: %w", dir, err)
	}

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
		"file_path":    d.Config.FilePath,
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
