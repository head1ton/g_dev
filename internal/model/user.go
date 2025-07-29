package model

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

type User struct {
	BaseModel

	// 사용자명
	Username string `json:"username" gorm:"size:50;uniqueIndex;not null"`

	// 이메일 주소
	Email string `json:"email" gorm:"size:255;uniqueIndex;not null"`

	// 비밀번호 해시 (bcrypt로 암호화)
	PasswordHash string `json:"-" gorm:"size:255;not null"`

	// 사용자 닉네임 (게임 내 표시명)
	Nickname string `json:"nickname" gorm:"size:100;not null"`

	// 사용자 레벨
	Level int `json:"level" gorm:"default:1;not null"`

	/// 경험치
	Experience int `json:"experience" gorm:"default:0;not null"`

	// 골드 (게임 내 화폐)
	Gold int `json:"gold" gorm:"default:0;not null"`

	// 다이아몬드 (프리미엄 화폐)
	Diamond int `json:"diamond" gorm:"default:0;not null"`

	// 마지막 로그인 시간
	LastLoginAt *time.Time `json:"last_login_at"`

	// 계정 상태 (active, suspended, banned)
	Status UserStatus `json:"status" gorm:"default:'active';not null"`

	// 사용자 역할 (user, admin, moderator)
	Role UserRole `json:"role" gorm:"default;'user';not null"`

	// 프로필 이미지 URL
	ProfileImageURL string `json:"profile_image_url" gorm:"size:500"`

	// 자기 소개
	Bio string `json:"bio" gorm:"size:1000"`

	// 생년월일
	BirthDate *time.Time `json:"birth_date"`

	// 성별 (male, female, other)
	Gender Gender `json:"gender"`

	// 국가/지역
	Country string `json:"country" gorm:"size:100"`

	// 언어 설정
	Language string `json:"language" gorm:"size:10;default:'ko'"`

	// 시간대
	TimeZone string `json:"time_zone" gorm:"size:50;default:'Asia/Seoul'"`

	// 알림 설정 (JSON 형태로 저장)
	NotificationSettings string `json:"notification_settings" gorm:"size:1000;default:'{}'"`

	// 개인 정보 설정 (JSON 형태로 저장)
	PrivacySettings string `json:"privacy_settings" gorm:"size:1000;default:'{}'"`

	// 계정 생성 IP 주소
	CreateIP string `json:"created_ip" gorm:"size:45"`

	// 마지막 로그인 IP 주소
	LastLoginIP string `json:"last_login_ip" gorm:"size:45"`

	// 로그인 시도 횟수
	LoginAttempts int `json:"login_attempts" gorm:"default:0"`

	// 계정 잠금 해제 시간
	LockedUntil *time.Time `json:"locked_until"`

	// 이메일 인증 여부
	EmailVerified bool `json:"email_verified" gorm:"default:false"`

	// 이메일 인증 토큰
	EmailVerificationToken string `json:"-" gorm:"size:255"`

	// 비밀번호 재설정 토큰
	PasswordResetToken string `json:"-" gorm:"size:255"`

	// 비밀번호 재설정 토큰 만료 시간
	PasswordResetExpiresAt *time.Time `json:"-"`
}

// 사용자 계정 상태
type UserStatus string

const (
	UserStatusActive     UserStatus = "active"    // 활성
	UserStatusSupspended UserStatus = "suspended" // 일시 정지
	UserStatusBanned     UserStatus = "banned"    // 영구 정지
)

// 사용자 역할을 나타내는 열거형
type UserRole string

const (
	UserRoleUser      UserRole = "user"      // 일반 사용자
	UserRoleModerator UserRole = "moderator" // 중재자
	UserRoleAdmin     UserRole = "admin"     // 관리자
)

type Gender string

const (
	GenderMale   Gender = "male"   // 남성
	GenderFemale Gender = "female" // 여성
	GenderOther  Gender = "other"  // 기타
)

// User 모델의 테이블 이름 반환
func (User) TableName() string {
	return "users"
}

// User 모델의 데이터 유효성 검사
func (u *User) Validate() error {
	// 사용자명 검증
	if strings.TrimSpace(u.Username) == "" {
		return errors.New("username must be between 3 and 50 characters")
	}
	if !isValidUsername(u.Username) {
		return errors.New("username contains invalid characters")
	}

	// 이메일 검증
	if strings.TrimSpace(u.Email) == "" {
		return errors.New("email cannot be empty")
	}
	if !isValidEmail(u.Email) {
		return errors.New("invalid email format")
	}

	// 닉네임 검증
	if strings.TrimSpace(u.Nickname) == "" {
		return errors.New("nickname cannot be empty")
	}
	if len(u.Nickname) < 2 || len(u.Nickname) > 100 {
		return errors.New("nickname must be between 2 and 100 characters")
	}

	// 레벨 검증
	if u.Level < 1 {
		return errors.New("level must be a least 1")
	}

	// 경험치 검증
	if u.Experience < 0 {
		return errors.New("experience cannot be negative")
	}

	// 골드 검증
	if u.Gold < 0 {
		return errors.New("gold cannot be negative")
	}

	// 다이아몬드 검증
	if u.Diamond < 0 {
		return errors.New("diamond cannot be negative")
	}

	// 상태 검증
	if u.Status != UserStatusActive && u.Status != UserStatusSupspended && u.Status != UserStatusBanned {
		return errors.New("invalid user status")
	}

	// 역할 검증
	if u.Role != UserRoleUser && u.Role != UserRoleModerator && u.Role != UserRoleAdmin {
		return errors.New("invalid user role")
	}

	return nil
}

// 사용자 비밀번호를 설정
// bcrypt 를 사용하여 비밀번호를 해시화
func (u *User) SetPassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// bcrypt로 비밀번호 해시화
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	u.PasswordHash = string(hash)
	return nil
}

// 입력된 비밀번호 검증
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// 사용자에게 경험치 추가
// 레벨 업 조건을 확인하고 필요 시 레벨을 증가
func (u *User) AddExperience(exp int) {
	if exp <= 0 {
		return
	}

	u.Experience += exp

	// 레벨 업 조건 확인 (간단한 계산: 레벨 * 1000 경험치 필요)
	requiredExp := u.Level * 1000
	for u.Experience >= requiredExp {
		u.Experience -= requiredExp
		u.Level++
		requiredExp = u.Level * 1000
	}
}

// 사용자에게 골드 추가
func (u *User) AddGold(amount int) error {
	if amount < 0 && u.Gold+amount < 0 {
		return errors.New("insufficient gold")
	}
	u.Gold += amount
	return nil
}

// 사용자에게 다이아몬드 추가
func (u *User) AddDiamond(amount int) error {
	if amount < 0 && u.Diamond+amount < 0 {
		return errors.New("insufficient diamond")
	}
	u.Diamond += amount
	return nil
}

// 마지막 로그인 정보 업데이트
func (u *User) UpdateLastLogin(ip string) {
	now := time.Now()
	u.LastLoginAt = &now
	u.LastLoginIP = ip
	u.LoginAttempts = 0 // 로그인 성공 시 시도 횟수 초기화
}

// 로그인 시도 횟수 증가
func (u *User) IncrementLoginAttempts() {
	u.LoginAttempts++

	// 5회 실패 시 30분간 계정 잠금
	if u.LoginAttempts >= 5 {
		lockUntil := time.Now().Add(30 * time.Minute)
		u.LockedUntil = &lockUntil
	}
}

// 계정이 잠겨있는지 확인
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

// 활성 상태인지 확인
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive && !u.IsLocked()
}

// 사용자가 로그인할 수 있는지 확인
func (u *User) CanLogin() bool {
	return u.IsActive() && u.EmailVerified
}

// 사용자 이름 반환
// 닉네임이 있으면 닉네임을, 없으면 사용자명 반환
func (u *User) GetDisplayName() string {
	if strings.TrimSpace(u.Nickname) != "" {
		return u.Nickname
	}
	return u.Username
}

// 사용자명이 유효한지 확인
// 영문, 숫자, 언더스코어만 허용
func isValidUsername(username string) bool {
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}
	return true
}

// 이메일 형식 유효한지 확인
func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
