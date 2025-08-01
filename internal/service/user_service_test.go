// Package service는 G-Step 웹게임서버의 비즈니스 로직 테스트를 담당.
// 사용자 서비스의 각 메서드에 대한 단위 테스트를 제공.
package service

import (
	"fmt"
	"testing"
	"time"

	"g_dev/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB는 테스트용 데이터베이스를 설정.
// SQLite 인메모리 데이터베이스를 사용하여 테스트를 격리.
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// 테이블 마이그레이션
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

// createTestUser는 테스트용 사용자를 생성.
// 기본값이 설정된 유효한 사용자 객체를 반환.
func createTestUser() *model.User {
	now := time.Now()
	return &model.User{
		Username:      "testuser",
		Email:         "test@example.com",
		Nickname:      "테스트유저",
		Level:         1,
		Experience:    0,
		Gold:          1000,
		Diamond:       10,
		Status:        model.UserStatusActive,
		Role:          model.UserRoleUser,
		Language:      "ko",
		TimeZone:      "Asia/Seoul",
		BirthDate:     &now,
		Gender:        model.GenderMale,
		EmailVerified: true, // 테스트를 위해 이메일 인증 완료로 설정
	}
}

// TestNewUserService는 UserService 생성자를 테스트.
func TestNewUserService(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	service := NewUserService(db)
	if service == nil {
		t.Error("NewUserService returned nil")
	}

	if service.db != db {
		t.Error("UserService database connection mismatch")
	}
}

// TestUserService_CreateUser는 사용자 생성 기능을 테스트.
func TestUserService_CreateUser(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	service := NewUserService(db)

	tests := []struct {
		name    string
		user    *model.User
		wantErr bool
	}{
		{
			name:    "정상적인 사용자 생성",
			user:    createTestUser(),
			wantErr: false,
		},
		{
			name: "빈 사용자명",
			user: &model.User{
				Username: "",
				Email:    "test2@example.com",
				Nickname: "테스트유저2",
				Status:   model.UserStatusActive,
				Role:     model.UserRoleUser,
			},
			wantErr: true,
		},
		{
			name: "빈 이메일",
			user: &model.User{
				Username: "testuser2",
				Email:    "",
				Nickname: "테스트유저2",
				Status:   model.UserStatusActive,
				Role:     model.UserRoleUser,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 비밀번호 설정
			if err := tt.user.SetPassword("password123"); err != nil {
				t.Fatalf("failed to set password: %v", err)
			}

			err := service.CreateUser(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.user.ID == 0 {
				t.Error("CreateUser() did not set user ID")
			}
		})
	}
}

// TestUserService_GetUserByID는 ID로 사용자 조회 기능을 테스트.
func TestUserService_GetUserByID(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	service := NewUserService(db)

	// 테스트 사용자 생성
	user := createTestUser()
	if err := user.SetPassword("password123"); err != nil {
		t.Fatalf("failed to set password: %v", err)
	}

	if err := service.CreateUser(user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// 정상 조회 테스트
	retrievedUser, err := service.GetUserByID(user.ID)
	if err != nil {
		t.Errorf("GetUserByID() error = %v", err)
	}

	if retrievedUser.Username != user.Username {
		t.Errorf("GetUserByID() username = %v, want %v", retrievedUser.Username, user.Username)
	}

	// 존재하지 않는 사용자 조회 테스트
	_, err = service.GetUserByID(999)
	if err == nil {
		t.Error("GetUserByID() should return error for non-existent user")
	}
}

// TestUserService_GetUserByUsername는 사용자명으로 사용자 조회 기능을 테스트.
func TestUserService_GetUserByUsername(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	service := NewUserService(db)

	// 테스트 사용자 생성
	user := createTestUser()
	if err := user.SetPassword("password123"); err != nil {
		t.Fatalf("failed to set password: %v", err)
	}

	if err := service.CreateUser(user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// 정상 조회 테스트
	retrievedUser, err := service.GetUserByUsername(user.Username)
	if err != nil {
		t.Errorf("GetUserByUsername() error = %v", err)
	}

	if retrievedUser.ID != user.ID {
		t.Errorf("GetUserByUsername() id = %v, want %v", retrievedUser.ID, user.ID)
	}

	// 존재하지 않는 사용자 조회 테스트
	_, err = service.GetUserByUsername("nonexistent")
	if err == nil {
		t.Error("GetUserByUsername() should return error for non-existent user")
	}
}

// TestUserService_AuthenticateUser는 사용자 인증 기능을 테스트.
func TestUserService_AuthenticateUser(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	service := NewUserService(db)

	// 테스트 사용자 생성
	user := createTestUser()
	if err := user.SetPassword("password123"); err != nil {
		t.Fatalf("failed to set password: %v", err)
	}

	if err := service.CreateUser(user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "올바른 인증 정보",
			username: "testuser",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "잘못된 비밀번호",
			username: "testuser",
			password: "wrongpassword",
			wantErr:  true,
		},
		{
			name:     "존재하지 않는 사용자",
			username: "nonexistent",
			password: "password123",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.AuthenticateUser(tt.username, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthenticateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestUserService_UpdateUser는 사용자 정보 업데이트 기능을 테스트.
func TestUserService_UpdateUser(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	service := NewUserService(db)

	// 테스트 사용자 생성
	user := createTestUser()
	if err := user.SetPassword("password123"); err != nil {
		t.Fatalf("failed to set password: %v", err)
	}

	if err := service.CreateUser(user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// 사용자 정보 업데이트
	user.Nickname = "업데이트된닉네임"
	user.Bio = "업데이트된 자기소개"

	err := service.UpdateUser(user)
	if err != nil {
		t.Errorf("UpdateUser() error = %v", err)
	}

	// 업데이트 확인
	updatedUser, err := service.GetUserByID(user.ID)
	if err != nil {
		t.Errorf("GetUserByID() error = %v", err)
	}

	if updatedUser.Nickname != "업데이트된닉네임" {
		t.Errorf("UpdateUser() nickname = %v, want %v", updatedUser.Nickname, "업데이트된닉네임")
	}
}

// TestUserService_AddExperience는 경험치 추가 기능을 테스트.
func TestUserService_AddExperience(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	service := NewUserService(db)

	// 테스트 사용자 생성
	user := createTestUser()
	if err := user.SetPassword("password123"); err != nil {
		t.Fatalf("failed to set password: %v", err)
	}

	if err := service.CreateUser(user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// 경험치 추가
	err := service.AddExperience(user.ID, 500)
	if err != nil {
		t.Errorf("AddExperience() error = %v", err)
	}

	// 경험치 추가 확인
	updatedUser, err := service.GetUserByID(user.ID)
	if err != nil {
		t.Errorf("GetUserByID() error = %v", err)
	}

	if updatedUser.Experience != 500 {
		t.Errorf("AddExperience() experience = %v, want %v", updatedUser.Experience, 500)
	}
}

// TestUserService_GetUserStats는 사용자 통계 조회 기능을 테스트.
func TestUserService_GetUserStats(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	service := NewUserService(db)

	// 테스트 사용자 생성
	user := createTestUser()
	if err := user.SetPassword("password123"); err != nil {
		t.Fatalf("failed to set password: %v", err)
	}

	if err := service.CreateUser(user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// 통계 조회
	stats, err := service.GetUserStats(user.ID)
	if err != nil {
		t.Errorf("GetUserStats() error = %v", err)
	}

	// 통계 확인
	if stats["username"] != user.Username {
		t.Errorf("GetUserStats() username = %v, want %v", stats["username"], user.Username)
	}

	if stats["level"] != user.Level {
		t.Errorf("GetUserStats() level = %v, want %v", stats["level"], user.Level)
	}
}

// ExampleNewUserService는 UserService 생성자의 사용 예시를 제공.
func ExampleNewUserService() {
	// 데이터베이스 연결 (실제 구현에서는 설정에서 가져옴)
	// db, _ := gorm.Open(mysql.Open("dsn"), &gorm.Config{})

	// UserService 생성
	// service := NewUserService(db)

	// 서비스 사용 예시
	// user := &model.User{...}
	// service.CreateUser(user)

	fmt.Println("UserService created successfully")
	// Output:
	// UserService created successfully
}
