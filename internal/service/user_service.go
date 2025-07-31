package service

import (
	"errors"
	"fmt"
	"g_dev/internal/model"
	"gorm.io/gorm"
	"time"
)

// 사용자 관리 비지니스 로직
type UserService struct {
	db *gorm.DB
}

// 새로운 UserService 인스턴스를 생성
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

// 새로운 사용자 생성
func (s *UserService) CreateUser(user *model.User) error {
	// 입력 데이터 검증
	if err := user.Validate(); err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}

	// 사용자명 중복 검사
	existingUser, err := s.GetUserByUsername(user.Username)
	if err == nil && existingUser != nil {
		return errors.New("username already exists")
	}

	// 이메일 중복 검사
	existingUser, err = s.GetUserByEmail(user.Email)
	if err == nil && existingUser != nil {
		return errors.New("email already exists")
	}

	// 기본값 설정
	if user.Status == "" {
		user.Status = model.UserStatusActive
	}
	if user.Role == "" {
		user.Role = model.UserRoleUser
	}
	if user.Language == "" {
		user.Language = "ko"
	}
	if user.TimeZone == "" {
		user.TimeZone = "Asia/Seoul"
	}
	if user.Level == 0 {
		user.Level = 1
	}
	if user.Experience == 0 {
		user.Experience = 0
	}
	if user.Gold == 0 {
		user.Gold = 0
	}
	if user.Diamond == 0 {
		user.Diamond = 0
	}

	// 데이터 저장
	if err := s.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// 사용자 정보 업데이트
func (s *UserService) UpdateUser(user *model.User) error {
	// 입력 데이터 검증
	if err := user.Validate(); err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}

	// 기존 사용자 조회
	existingUser, err := s.GetUserByID(user.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing user: %w", err)
	}
	if existingUser == nil {
		return errors.New("user not found")
	}

	// 사용자명 변경 시 중복 검사
	if user.Username != existingUser.Username {
		duplicateUser, err := s.GetUserByUsername(user.Username)
		if err == nil && duplicateUser != nil {
			return errors.New("username already exists")
		}
	}

	// 이메일 변경 시 중복 검사
	if user.Email != existingUser.Email {
		duplicateUser, err := s.GetUserByEmail(user.Email)
		if err == nil && duplicateUser != nil {
			return errors.New("email already exists")
		}
	}

	// 데이터베이스 업데이트
	if err := s.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// 사용자명으로 사용자를 조회
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

// 이메일로 사용자 조회
func (s *UserService) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

// ID로 사용자 조회
func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

// 사용자 삭제 (소프트 삭제)
func (s *UserService) DeleteUser(id uint) error {
	if err := s.db.Delete(&model.User{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// 사용자 인증
func (s *UserService) AuthenticateUser(username, password, ipAddress string) (*model.User, error) {
	// 사용자 조회
	user, err := s.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("authentification failed: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 계정 상태 확인
	if !user.IsActive() {
		return nil, errors.New("account is not active")
	}

	// 계정 잠금 확인
	if user.IsLocked() {
		return nil, errors.New("account is locked")
	}

	// 이메일 인증 확인
	if !user.EmailVerified {
		return nil, errors.New("email not verified")
	}

	// 비밀번호 확인
	if !user.CheckPassword(password) {
		// 로그인 실패 횟수 증가
		user.IncrementLoginAttempts()
		if err := s.UpdateUser(user); err != nil {
			return nil, fmt.Errorf("failed to update login attempts: %w", err)
		}
		return nil, errors.New("invalid password")
	}

	// 로그인 성공 시 정보 업데이트
	user.UpdateLastLogin(ipAddress)
	if err := s.UpdateUser(user); err != nil {
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}

	return user, nil
}

// 사용자 비밀번호 변경
func (s *UserService) ChangePassword(userID uint, currentPassword, newPassword string) error {
	// 사용자 조회
	user, err := s.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 현재 비밀번호 확인
	if !user.CheckPassword(currentPassword) {
		return errors.New("current password is incorrect")
	}

	// 새 비밀번호 설정
	if err := user.SetPassword(newPassword); err != nil {
		return fmt.Errorf("failed to set new password: %w", err)
	}

	// 데이터베이스 업데이트
	if err := s.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// 비밀번호 재설정 토큰을 생성
func (s *UserService) ResetPassword(email string) (string, error) {
	// 사용자 조회
	user, err := s.GetUserByEmail(email)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return "", errors.New("user not found")
	}

	// 토큰 생성
	token := generateResetToken()
	expiresAt := time.Now().Add(24 * time.Hour)

	user.PasswordResetToken = token
	user.PasswordResetExpiresAt = &expiresAt

	// 데이터베이스 업데이트
	if err := s.UpdateUser(user); err != nil {
		return "", fmt.Errorf("failed to update user: %w", err)
	}

	return token, nil
}

// 비밀번호 재설정을 확인하고 새 비밀번호를 설정
func (s *UserService) ConfirmPasswordReset(token, newPassword string) error {
	// 토큰으로 사용자 조회
	var user model.User
	if err := s.db.Where("password_reset_token = ? AND password_reset_expires_at > ?", token, time.Now()).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid or expired reset token")
		}
		return fmt.Errorf("failed to get user by reset token: %w", err)
	}

	// 새 비밀번호 설정
	if err := user.SetPassword(newPassword); err != nil {
		return fmt.Errorf("failed to set new password: %w", err)
	}

	// 토큰 초기화
	user.PasswordResetToken = ""
	user.PasswordResetExpiresAt = nil

	// 데이터 업데이트
	if err := s.UpdateUser(&user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// 이메일 인증 확인
func (s *UserService) VerifyEmail(token string) error {
	// 토큰으로 사용자 조회
	var user model.User
	if err := s.db.Where("email_verification_token = ?", token).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid verification token")
		}
		return fmt.Errorf("failed to get user by verification token: %w", err)
	}

	// 이메일 인증 완료
	user.EmailVerified = true
	user.EmailVerificationToken = ""

	if err := s.UpdateUser(&user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// 사용자에게 경험치 추가
func (s *UserService) AddExperience(userID uint, experience int) error {
	// 사용자 조회
	user, err := s.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 경험치 추가
	user.AddExperience(experience)

	if err := s.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// 사용자에게 골드 추가
func (s *UserService) AddGold(userID uint, gold int) error {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 골드 추가
	if err := user.AddGold(gold); err != nil {
		return fmt.Errorf("failed to add gold: %w", err)
	}

	if err := s.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// 다이아몬드 추가
func (s *UserService) AddDiamond(userID uint, diamond int) error {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 다이아몬드 추가
	if err := user.AddDiamond(diamond); err != nil {
		return fmt.Errorf("failed to add diamond: %w", err)
	}

	if err := s.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// 사용자 통계 정보 반환
func (s *UserService) GetUserStats(userID uint) (map[string]interface{}, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 점수 통계 조회
	var totalScores int64
	if err := s.db.Model(&model.Score{}).Where("user_id = ?", userID).Count(&totalScores).Error; err != nil {
		return nil, fmt.Errorf("failed to count scores: %w", err)
	}

	var highScores int64
	if err := s.db.Model(&model.Score{}).Where("user_id = ? AND is_high_score = ?", userID, true).Count(&highScores).Error; err != nil {
		return nil, fmt.Errorf("failed to count high scores: %w", err)
	}

	var totalPlayTime int64
	if err := s.db.Model(&model.Score{}).Where("user_id = ?", userID).Select("COALESCE(SUM(play_time), 0)").Scan(&totalPlayTime).Error; err != nil {
		return nil, fmt.Errorf("failed to sum play time: %w", err)
	}

	// 통계 정보 반환
	stats := map[string]interface{}{
		"user_id":         userID,
		"username":        user.Username,
		"nickname":        user.Nickname,
		"level":           user.Level,
		"experience":      user.Experience,
		"gold":            user.Gold,
		"diamond":         user.Diamond,
		"total_scores":    totalScores,
		"high_scores":     highScores,
		"total_play_time": totalPlayTime,
		"created_at":      user.CreatedAt,
		"last_login_at":   user.LastLoginAt,
	}

	return stats, nil
}

// 사용자 검색
func (s *UserService) SearchUsers(query string, limit, offset int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	// 검색 조건 설정
	searchQuery := s.db.Model(&model.User{})
	if query != "" {
		searchQuery = searchQuery.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%")
	}

	// 총 개수 조회
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// 사용자 목록 조회
	if err := searchQuery.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search users: %w", err)
	}

	return users, total, nil
}

func generateResetToken() string {
	return fmt.Sprintf("reset_%d", time.Now().UnixNano())
}
