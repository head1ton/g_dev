// Package service는 G-Step 웹게임서버의 비즈니스 로직을 담당.
// 데이터베이스와 상호작용하여 사용자 관리, 게임 로직 등을 처리.
package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"g_dev/internal/model"

	"gorm.io/gorm"
)

// UserService는 사용자 관련 비즈니스 로직을 처리하는 서비스.
// 사용자 생성, 조회, 수정, 삭제 및 인증 기능을 제공.
type UserService struct {
	// 데이터베이스 연결
	db *gorm.DB
}

// NewUserService는 새로운 UserService 인스턴스를 생성.
// 데이터베이스 연결을 받아서 서비스를 초기화.
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

// CreateUser는 새로운 사용자를 생성.
// 사용자 정보를 검증하고 데이터베이스에 저장.
func (s *UserService) CreateUser(user *model.User) error {
	// 사용자 정보 검증
	if err := user.Validate(); err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}

	// 사용자명 중복 확인
	var existingUser model.User
	if err := s.db.Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
		return fmt.Errorf("username already exists: %s", user.Username)
	}

	// 이메일 중복 확인
	if user.Email != "" {
		if err := s.db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
			return fmt.Errorf("email already exists: %s", user.Email)
		}
	}

	// 기본값 설정
	if user.Level == 0 {
		user.Level = 1
	}
	if user.Experience == 0 {
		user.Experience = 0
	}
	if user.Gold == 0 {
		user.Gold = 1000
	}
	if user.Diamond == 0 {
		user.Diamond = 10
	}
	if user.Status == "" {
		user.Status = model.UserStatusActive
	}
	if user.Role == "" {
		user.Role = model.UserRoleUser
	}

	// 사용자 생성
	if err := s.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByID는 ID로 사용자를 조회.
func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// GetUserByUsername은 사용자명으로 사용자를 조회.
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// GetUserByEmail은 이메일로 사용자를 조회.
func (s *UserService) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// UpdateUser는 사용자 정보를 업데이트.
func (s *UserService) UpdateUser(user *model.User) error {
	// 사용자 정보 검증
	if err := user.Validate(); err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}

	// 사용자 존재 확인
	var existingUser model.User
	if err := s.db.First(&existingUser, user.ID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 사용자명 중복 확인 (자신 제외)
	if err := s.db.Where("username = ? AND id != ?", user.Username, user.ID).First(&existingUser).Error; err == nil {
		return fmt.Errorf("username already exists: %s", user.Username)
	}

	// 이메일 중복 확인 (자신 제외)
	if user.Email != "" {
		if err := s.db.Where("email = ? AND id != ?", user.Email, user.ID).First(&existingUser).Error; err == nil {
			return fmt.Errorf("email already exists: %s", user.Email)
		}
	}

	// 사용자 업데이트
	if err := s.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser는 사용자를 삭제.
func (s *UserService) DeleteUser(id uint) error {
	if err := s.db.Delete(&model.User{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// AuthenticateUser는 사용자 인증을 수행.
// 사용자명과 비밀번호를 확인하여 인증 성공 여부를 반환.
func (s *UserService) AuthenticateUser(username, password string) (*model.User, error) {
	// 사용자 조회
	user, err := s.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// 계정 상태 확인
	if !user.CanLogin() {
		return nil, fmt.Errorf("account is locked or inactive")
	}

	// 비밀번호 확인
	if !user.CheckPassword(password) {
		// 로그인 시도 횟수 증가
		user.IncrementLoginAttempts()
		s.db.Save(user)
		return nil, fmt.Errorf("invalid password")
	}

	// 로그인 성공 시 시도 횟수 초기화 및 마지막 로그인 시간 업데이트
	user.UpdateLastLogin("") // IP 주소는 나중에 구현
	s.db.Save(user)

	return user, nil
}

// ChangePassword는 사용자 비밀번호를 변경.
func (s *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	// 사용자 조회
	user, err := s.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 기존 비밀번호 확인
	if !user.CheckPassword(oldPassword) {
		return fmt.Errorf("invalid old password")
	}

	// 새 비밀번호 설정
	if err := user.SetPassword(newPassword); err != nil {
		return fmt.Errorf("failed to set new password: %w", err)
	}

	// 사용자 업데이트
	if err := s.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// ResetPassword는 사용자 비밀번호를 재설정.
// 이메일 인증을 통해 비밀번호를 재설정하는 기능.
func (s *UserService) ResetPassword(email string) error {
	// 사용자 조회
	user, err := s.GetUserByEmail(email)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 재설정 토큰 생성
	token := s.generateResetToken()
	expiresAt := time.Now().Add(24 * time.Hour)
	user.PasswordResetToken = token
	user.PasswordResetExpiresAt = &expiresAt

	// 사용자 업데이트
	if err := s.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// TODO: 이메일 발송 로직 구현
	// 실제 프로덕션에서는 이메일 서비스를 통해 재설정 링크를 발송해야 .

	return nil
}

// ConfirmPasswordReset는 비밀번호 재설정을 확인하고 새 비밀번호를 설정.
func (s *UserService) ConfirmPasswordReset(token, newPassword string) error {
	// 토큰으로 사용자 조회
	var user model.User
	if err := s.db.Where("password_reset_token = ? AND password_reset_expires_at > ?", token, time.Now()).First(&user).Error; err != nil {
		return fmt.Errorf("invalid or expired reset token: %w", err)
	}

	// 새 비밀번호 설정
	if err := user.SetPassword(newPassword); err != nil {
		return fmt.Errorf("failed to set new password: %w", err)
	}

	// 토큰 초기화
	user.PasswordResetToken = ""
	user.PasswordResetExpiresAt = nil

	// 사용자 업데이트
	if err := s.db.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// VerifyEmail은 이메일 인증을 확인.
func (s *UserService) VerifyEmail(token string) error {
	// 토큰으로 사용자 조회
	var user model.User
	if err := s.db.Where("email_verification_token = ?", token).First(&user).Error; err != nil {
		return fmt.Errorf("invalid verification token: %w", err)
	}

	// 이메일 인증 완료
	user.EmailVerified = true
	user.EmailVerificationToken = ""

	// 사용자 업데이트
	if err := s.db.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// AddExperience는 사용자에게 경험치를 추가.
// 레벨업 로직도 함께 처리.
func (s *UserService) AddExperience(userID uint, experience int) error {
	// 사용자 조회
	user, err := s.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 경험치 추가
	user.AddExperience(experience)

	// 사용자 업데이트
	if err := s.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// AddGold는 사용자에게 골드를 추가.
func (s *UserService) AddGold(userID uint, gold int) error {
	// 사용자 조회
	user, err := s.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 골드 추가
	user.AddGold(gold)

	// 사용자 업데이트
	if err := s.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// AddDiamond는 사용자에게 다이아몬드를 추가.
func (s *UserService) AddDiamond(userID uint, diamond int) error {
	// 사용자 조회
	user, err := s.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 다이아몬드 추가
	user.AddDiamond(diamond)

	// 사용자 업데이트
	if err := s.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// GetUserStats는 사용자 통계 정보를 반환.
func (s *UserService) GetUserStats(userID uint) (map[string]interface{}, error) {
	// 사용자 조회
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 기본 통계 정보
	stats := map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"nickname":   user.Nickname,
		"level":      user.Level,
		"experience": user.Experience,
		"gold":       user.Gold,
		"diamond":    user.Diamond,
		"status":     user.Status,
		"role":       user.Role,
		"created_at": user.CreatedAt,
		"last_login": user.LastLoginAt,
	}

	// 게임 관련 통계 (향후 구현)
	// 예: 총 게임 플레이 수, 최고 점수, 평균 점수 등

	return stats, nil
}

// SearchUsers는 사용자를 검색.
// 사용자명, 닉네임, 이메일 등으로 검색할 수 있습니다.
func (s *UserService) SearchUsers(query string, limit, offset int) ([]*model.User, error) {
	var users []*model.User

	// 검색 쿼리 실행
	if err := s.db.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?",
		"%"+query+"%", "%"+query+"%", "%"+query+"%").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return users, nil
}

// generateResetToken는 비밀번호 재설정 토큰을 생성.
// 32바이트 랜덤 토큰을 생성하여 16진수 문자열로 반환.
func (s *UserService) generateResetToken() string {
	token := make([]byte, 32)
	rand.Read(token)
	return hex.EncodeToString(token)
}
