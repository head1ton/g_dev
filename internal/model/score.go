package model

import (
	"errors"
	"time"
)

// Score는 게임 점수 기록을 담는 모델
// 사용자의 게임 플레이 결과와 점수를 관리
type Score struct {
	BaseModel

	// 사용자 ID (외래키)
	UserID uint `json:"user_id" gorm:"not null;index"`

	// 게임 ID (외래키)
	GameID uint `json:"game_id" gorm:"not null;index"`

	// 점수 (게임별로 다른 점수 체계 사용 가능)
	Score int `json:"score" gorm:"not null"`

	// 최고 점수 여부 (해당 게임에서의 최고 점수인지)
	IsHighScore bool `json:"is_high_score" gorm:"default:false"`

	// 플레이 시간 (초)
	PlayTime int `json:"play_time" gorm:"default:0"`

	// 게임 완료 여부
	Completed bool `json:"completed" gorm:"default:false"`

	// 게임 난이도
	Difficulty GameDifficulty `json:"difficulty" gorm:"size:20;default:'normal'"`

	// 게임 모드 (single, multiplayer, tournament 등)
	GameMode string `json:"game_mode" gorm:"size:50;default:'single'"`

	// 게임 설정 (JSON 형태로 저장)
	GameSettings string `json:"game_settings" gorm:"size:1000;default:'{}'"`

	// 게임 결과 데이터 (JSON 형태로 저장)
	GameData string `json:"game_data" gorm:"size:2000;default:'{}'"`

	// 획득한 골드
	EarnedGold int `json:"earned_gold" gorm:"default:0"`

	// 획득한 경험치
	EarnedExperience int `json:"earned_experience" gorm:"default:0"`

	// 획득한 다이아몬드
	EarnedDiamond int `json:"earned_diamond" gorm:"default:0"`

	// 게임 시작 시간
	StartedAt *time.Time `json:"started_at"`

	// 게임 종료 시간
	EndedAt *time.Time `json:"ended_at"`

	// 게임 플랫폼 (web, mobile, desktop)
	Platform string `json:"platform" gorm:"size:20;default:'web'"`

	// 게임 버전
	GameVersion string `json:"game_version" gorm:"size:20"`

	// IP 주소 (보안 및 분석용)
	IPAddress string `json:"ip_address" gorm:"size:45"`

	// 사용자 에이전트 (브라우저 정보)
	UserAgent string `json:"user_agent" gorm:"size:500"`

	// 게임 세션 ID
	SessionID string `json:"session_id" gorm:"size:100"`

	// 게임 재시작 횟수
	RestartCount int `json:"restart_count" gorm:"default:0"`

	// 게임 일시정지 횟수
	PauseCount int `json:"pause_count" gorm:"default:0"`

	// 게임 저장 횟수
	SaveCount int `json:"save_count" gorm:"default:0"`

	// 게임 로드 횟수
	LoadCount int `json:"load_count" gorm:"default:0"`

	// 게임 에러 발생 여부
	HasError bool `json:"has_error" gorm:"default:false"`

	// 게임 에러 메시지
	ErrorMessage string `json:"error_message" gorm:"size:500"`

	// 게임 성공률 (0-100%)
	SuccessRate float64 `json:"success_rate" gorm:"default:0"`

	// 게임 정확도 (0-100%)
	Accuracy float64 `json:"accuracy" gorm:"default:0"`

	// 게임 속도 (초당 액션 수)
	Speed float64 `json:"speed" gorm:"default:0"`

	// 게임 효율성 (점수/시간 비율)
	Efficiency float64 `json:"efficiency" gorm:"default:0"`

	// 게임 난이도 점수 (게임별로 다른 계산 방식)
	DifficultyScore float64 `json:"difficulty_score" gorm:"default:0"`

	// 게임 보너스 점수
	BonusScore int `json:"bonus_score" gorm:"default:0"`

	// 게임 페널티 점수
	PenaltyScore int `json:"penalty_score" gorm:"default:0"`

	// 게임 메타데이터 (JSON 형태로 저장)
	Metadata string `json:"metadata" gorm:"size:2000;default:'{}'"`

	// 관계 정의
	User User `json:"user" gorm:"foreignKey:UserID"`
	Game Game `json:"game" gorm:"foreignKey:GameID"`
}

// TableName은 Score 모델의 테이블 이름을 반환
func (Score) TableName() string {
	return "scores"
}

// Validate는 Score 모델의 데이터 유효성을 검사
func (s *Score) Validate() error {
	// 사용자 ID 검증
	if s.UserID == 0 {
		return errors.New("user ID cannot be zero")
	}

	// 게임 ID 검증
	if s.GameID == 0 {
		return errors.New("game ID cannot be zero")
	}

	// 점수 검증
	if s.Score < 0 {
		return errors.New("score cannot be negative")
	}

	// 플레이 시간 검증
	if s.PlayTime < 0 {
		return errors.New("play time cannot be negative")
	}

	// 게임 난이도 검증
	if !isValidGameDifficulty(s.Difficulty) {
		return errors.New("invalid game difficulty")
	}

	// 게임 모드 검증
	if s.GameMode == "" {
		return errors.New("game mode cannot be empty")
	}

	// 획득한 골드 검증
	if s.EarnedGold < 0 {
		return errors.New("earned gold cannot be negative")
	}

	// 획득한 경험치 검증
	if s.EarnedExperience < 0 {
		return errors.New("earned experience cannot be negative")
	}

	// 획득한 다이아몬드 검증
	if s.EarnedDiamond < 0 {
		return errors.New("earned diamond cannot be negative")
	}

	// 게임 재시작 횟수 검증
	if s.RestartCount < 0 {
		return errors.New("restart count cannot be negative")
	}

	// 게임 일시정지 횟수 검증
	if s.PauseCount < 0 {
		return errors.New("pause count cannot be negative")
	}

	// 게임 저장 횟수 검증
	if s.SaveCount < 0 {
		return errors.New("save count cannot be negative")
	}

	// 게임 로드 횟수 검증
	if s.LoadCount < 0 {
		return errors.New("load count cannot be negative")
	}

	// 성공률 검증
	if s.SuccessRate < 0 || s.SuccessRate > 100 {
		return errors.New("success rate must be between 0 and 100")
	}

	// 정확도 검증
	if s.Accuracy < 0 || s.Accuracy > 100 {
		return errors.New("accuracy must be between 0 and 100")
	}

	// 게임 속도 검증
	if s.Speed < 0 {
		return errors.New("speed cannot be negative")
	}

	// 게임 효율성 검증
	if s.Efficiency < 0 {
		return errors.New("efficiency cannot be negative")
	}

	// 게임 난이도 점수 검증
	if s.DifficultyScore < 0 {
		return errors.New("difficulty score cannot be negative")
	}

	// 게임 보너스 점수 검증
	if s.BonusScore < 0 {
		return errors.New("bonus score cannot be negative")
	}

	// 게임 페널티 점수 검증
	if s.PenaltyScore < 0 {
		return errors.New("penalty score cannot be negative")
	}

	return nil
}

// GetTotalScore는 총 점수를 반환
// 기본 점수 + 보너스 점수 - 페널티 점수
func (s *Score) GetTotalScore() int {
	return s.Score + s.BonusScore - s.PenaltyScore
}

// GetPlayDuration는 실제 플레이 시간을 반환
// 시작 시간과 종료 시간이 있으면 그 차이를, 없으면 PlayTime을 반환
func (s *Score) GetPlayDuration() time.Duration {
	if s.StartedAt != nil && s.EndedAt != nil {
		return s.EndedAt.Sub(*s.StartedAt)
	}
	return time.Duration(s.PlayTime) * time.Second
}

// GetScorePerMinute는 분당 점수를 반환
func (s *Score) GetScorePerMinute() float64 {
	duration := s.GetPlayDuration()
	if duration == 0 {
		return 0
	}

	minutes := duration.Minutes()
	if minutes == 0 {
		return 0
	}

	return float64(s.GetTotalScore()) / minutes
}

// GetEfficiency는 게임 효율성을 계산
// 점수/시간 비율로 계산
func (s *Score) GetEfficiency() float64 {
	duration := s.GetPlayDuration()
	if duration == 0 {
		return 0
	}

	minutes := duration.Minutes()
	if minutes == 0 {
		return 0
	}

	return float64(s.GetTotalScore()) / minutes
}

// IsHighScoreCandidate는 최고 점수 후보인지 확인
// 점수가 양수이고 게임이 완료된 경우 true를 반환
func (s *Score) IsHighScoreCandidate() bool {
	return s.Score > 0 && s.Completed
}

// GetTotalEarnings는 총 획득 보상을 반환
func (s *Score) GetTotalEarnings() map[string]int {
	return map[string]int{
		"gold":       s.EarnedGold,
		"experience": s.EarnedExperience,
		"diamond":    s.EarnedDiamond,
	}
}

// SetGameStart는 게임 시작 시간을 설정
func (s *Score) SetGameStart() {
	now := time.Now()
	s.StartedAt = &now
}

// SetGameEnd는 게임 종료 시간을 설정
func (s *Score) SetGameEnd() {
	now := time.Now()
	s.EndedAt = &now

	// 플레이 시간 계산
	if s.StartedAt != nil {
		s.PlayTime = int(s.EndedAt.Sub(*s.StartedAt).Seconds())
	}
}

// AddBonus는 보너스 점수를 추가
func (s *Score) AddBonus(bonus int) {
	if bonus > 0 {
		s.BonusScore += bonus
	}
}

// AddPenalty는 페널티 점수를 추가
func (s *Score) AddPenalty(penalty int) {
	if penalty > 0 {
		s.PenaltyScore += penalty
	}
}

// IncrementRestartCount는 재시작 횟수를 증가시킵니다.
func (s *Score) IncrementRestartCount() {
	s.RestartCount++
}

// IncrementPauseCount는 일시정지 횟수를 증가시킵니다.
func (s *Score) IncrementPauseCount() {
	s.PauseCount++
}

// IncrementSaveCount는 저장 횟수를 증가시킵니다.
func (s *Score) IncrementSaveCount() {
	s.SaveCount++
}

// IncrementLoadCount는 로드 횟수를 증가시킵니다.
func (s *Score) IncrementLoadCount() {
	s.LoadCount++
}

// SetError는 게임 에러 정보를 설정
func (s *Score) SetError(message string) {
	s.HasError = true
	s.ErrorMessage = message
}

// CalculateSuccessRate는 성공률을 계산
// 게임별로 다른 계산 방식을 사용할 수 있습니다.
func (s *Score) CalculateSuccessRate(totalActions, successfulActions int) {
	if totalActions == 0 {
		s.SuccessRate = 0
		return
	}

	s.SuccessRate = float64(successfulActions) / float64(totalActions) * 100
}

// CalculateAccuracy는 정확도를 계산
// 게임별로 다른 계산 방식을 사용할 수 있습니다.
func (s *Score) CalculateAccuracy(totalShots, hits int) {
	if totalShots == 0 {
		s.Accuracy = 0
		return
	}

	s.Accuracy = float64(hits) / float64(totalShots) * 100
}

// CalculateSpeed는 게임 속도를 계산
// 초당 액션 수로 계산
func (s *Score) CalculateSpeed(totalActions int) {
	duration := s.GetPlayDuration()
	if duration == 0 {
		s.Speed = 0
		return
	}

	seconds := duration.Seconds()
	if seconds == 0 {
		s.Speed = 0
		return
	}

	s.Speed = float64(totalActions) / seconds
}

// UpdateEfficiency는 게임 효율성을 업데이트
func (s *Score) UpdateEfficiency() {
	s.Efficiency = s.GetEfficiency()
}

// IsValid는 점수 기록이 유효한지 확인
func (s *Score) IsValid() bool {
	return s.UserID > 0 && s.GameID > 0 && s.Score >= 0 && s.Completed
}
