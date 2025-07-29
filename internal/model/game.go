package model

import (
	"errors"
	"strings"
	"time"
)

// 게임 정보
type Game struct {
	BaseModel

	// 게임 이름
	Name string `json:"name" gorm:"size:100;uniqueIndex;not null"`

	// 게임 설명
	Description string `json:"description" gorm:"size:1000"`

	// 게임 카테고리 (puzzle, action, strategy, rpg 등)
	Category GameCategory `json:"category" gorm:"size:50;not null"`

	// 게임 난이도(easy, normal, hard, expert)
	Difficulty GameDifficulty `json:"difficulty" gorm:"size:20;not null"`

	// 게임 상태(active, inactive, maintenance)
	Status GameStatus `json:"status" gorm:"default:'active';not null"`

	// 게임 버전
	Version string `json:"version" gorm:"size:20;default:'1.0.0'"`

	// 최소 레벨 요구사항
	MinLevel int `json:"min_level" gorm:"default:1;not null"`

	// 최대 플레이어 수
	MaxPlayers int `json:"max_players" gorm:"default:1;not null"`

	// 예상 플레이 시간(분)
	EstimatedPlayTime int `json:"estimated_play_time" gorm:"default:10"`

	// 게임 이미지 URL
	ImageURL string `json:"image_url" gorm:"size:500"`

	// 게임 아이콘 URL
	IconURL string `json:"icon_url" gorm:"size:500"`

	// 게임 설정 (JSON 형태로 저장)
	GameSettings string `json:"game_settings" gorm:"size:2000;default:'{}'"`

	// 게임 규칙 (JSON 형태로 저장)
	GameRules string `json:"game_rules" gorm:"size:2000;default:'{}'"`

	// 보상 설정 (JSON 형태로 저장)
	RewardSettings string `json:"reward_settings" gorm:"size:1000;default:'{}'"`

	// 게임 태그 (쉼표로 구분)
	Tags string `json:"tags" gorm:"size:500"`

	// 평균 평점 (1~5점)
	AverageRating float64 `json:"average_rating" gorm:"default:0"`

	// 총 평가 수
	TotalRatings int `json:"total_ratings" gorm:"default:0"`

	// 총 플레이 수
	TotalPlays int `json:"total_plays" gorm:"default:0"`

	// 총 플레이 시간 (분)
	TotalPlayTime int `json:"total_play_time" gorm:"default:0"`

	// 게임 출시일
	ReleaseDate *time.Time `json:"release_date"`

	// 마지막 업데이트일
	LastUpdatedAt *time.Time `json:"last_updated_at"`

	// 개발자 정보
	Developer string `json:"developer" gorm:"size:100"`

	// 게임 언어 지원 (쉼표로 구분)
	SupportedLanguages string `json:"supported_languages" gorm:"size:200;default:ko,en"`

	// 게임 플랫폼 지원 (쉼표로 구분)
	SupportedPlatforms string `json:"supported_platforms" gorm:"size:200;default:'web'"`

	// 게임 파일 크기 (MB)
	FileSize int `json:"file_size" gorm:"default:0"`

	// 게임 다운로드 URL
	DownloadURL string `json:"download_url" gorm:"size:500"`

	// 게임 실행 URL
	PlayURL string `json:"play_url" gorm:"size:500"`

	// 게임 튜토리얼 URL
	TutorialURL string `json:"tutorial_url" gorm:"size:500"`

	// 게임 FAQ URL
	FAQURL string `json:"faq_url" gorm:"size:500"`

	// 게임 커뮤니티 URL
	CommunityURL string `json:"community_url" gorm:"size:500"`

	// 게임 공식 사이트 URL
	OfficialURL string `json:"official_url" gorm:"size:500"`

	// 게임 라이센스
	License string `json:"license" gorm:"size:100"`

	// 게임 가격 (0이면 무료)
	Price int `json:"price" gorm:"default:0"`

	// 게임 통화 (KRW, USD 등)
	Currency string `json:"currency" gorm:"size:10;default:'KRW'"`

	// 게임 할인율 (0-100%)
	DiscountRate int `json:"discount_rate" gorm:"default:0"`

	// 게임 할인 종료일
	DiscountEndDate *time.Time `json:"discount_end_date"`

	// 게임 인기도 점수
	PopularityScore float64 `json:"popularity_score" gorm:"default:0"`

	// 게임 트렌딩 점수
	TrendingScore float64 `json:"trending_score" gorm:"default:0"`

	// 게임 추천 점수
	RecommendationScore float64 `json:"recommendation_score" gorm:"default:0"`

	// 게임 메타데이터 (JSON 형태로 저장)
	Metadata string `json:"metadata" gorm:"size:2000;default:'{}'"`
}

// 게임 카테고리
type GameCategory string

const (
	GameCategoryPuzzle      GameCategory = "puzzle"      // 퍼즐
	GameCategoryAction      GameCategory = "action"      //액션
	GameCategoryStrategy    GameCategory = "strategy"    // 전략
	GameCategoryRPG         GameCategory = "rpg"         // 롤플레잉
	GameCategoryAdventure   GameCategory = "adventure"   // 어드벤처
	GameCategoryRacing      GameCategory = "racing"      // 레이싱
	GameCategorySports      GameCategory = "sports"      // 스포츠
	GameCategorySimulation  GameCategory = "simulation"  // 시뮬레이션
	GameCategoryCasual      GameCategory = "casual"      // 캐주얼
	GameCategoryEducational GameCategory = "educational" // 교육
	GameCategoryMusic       GameCategory = "music"       // 음악
	GameCategoryBoard       GameCategory = "board"       // 보드게임
	GameCategoryCard        GameCategory = "card"        // 카드게임
	GameCategoryArcade      GameCategory = "arcade"      // 아케이드
	GameCategoryOther       GameCategory = "other"       // 기타
)

// 게임 난이도
type GameDifficulty string

const (
	GameDifficultyEasy   GameDifficulty = "easy"   // 쉬움
	GameDifficultyNormal GameDifficulty = "normal" // 보통
	GameDifficultyHard   GameDifficulty = "hard"   // 어려움
	GameDifficultyExpert GameDifficulty = "expert" // 전문가
)

// 게임 상태
type GameStatus string

const (
	GameStatusActive      GameStatus = "active"      // 활성
	GameStatusInactive    GameStatus = "inactive"    // 비활성
	GameStatusMaintenance GameStatus = "maintenance" // 점검 중
	GameStatusBeta        GameStatus = "beta"        // 베타
	GameStatusAlpha       GameStatus = "alpha"       // 알파
)

// Game 모델의 테이블 이름을 반환
func (Game) TableName() string {
	return "games"
}

// 데이터 유효성 검사
func (g *Game) Validate() error {
	// 게임 이름 검증
	if strings.TrimSpace(g.Name) == "" {
		return errors.New("game name cannot be empty")
	}
	if len(g.Name) < 2 || len(g.Name) > 100 {
		return errors.New("game name must be between 2 and 100 characters")
	}

	// 게임 설명 검증
	if len(g.Description) > 1000 {
		return errors.New("game description cannot exceed 1000 characters")
	}

	// 게임 카테고리 검증
	if !isValidGameCategory(g.Category) {
		return errors.New("invalid game category")
	}

	// 게임 난이도 검증
	if !isValidGameDifficulty(g.Difficulty) {
		return errors.New("invalid game difficulty")
	}

	// 게임 상태 검증
	if !isValidGameStatus(g.Status) {
		return errors.New("invalid game status")
	}

	// 최소 레벨 검증
	if g.MinLevel < 1 {
		return errors.New("minimum level must be at least 1")
	}

	// 최대 플레이어 수 검증
	if g.MaxPlayers < 1 {
		return errors.New("maximum players must be at least 1")
	}

	// 예상 플레이 시간 검증
	if g.EstimatedPlayTime < 0 {
		return errors.New("estimated play time cannot be negative")
	}

	// 평균 평점 검증
	if g.AverageRating < 0 || g.AverageRating > 5 {
		return errors.New("average rating must be between 0 and 5")
	}

	// 총 평가 수 검증
	if g.TotalRatings < 0 {
		return errors.New("total ratings cannot be negative")
	}

	// 총 플레이 수 검증
	if g.TotalPlays < 0 {
		return errors.New("total plays cannot be negative")
	}

	// 총 플레이 시간 검증
	if g.TotalPlayTime < 0 {
		return errors.New("total play time cannot be negative")
	}

	// 게임 가격 검증
	if g.Price < 0 {
		return errors.New("game price cannot be negative")
	}

	// 할인율 검증
	if g.DiscountRate < 0 || g.DiscountRate > 100 {
		return errors.New("discount rate must be between 0 and 100")
	}

	return nil
}

// 게임이 활성 상태
func (g *Game) IsActive() bool {
	return g.Status == GameStatusActive
}

// 게임을 플레이할 수 있는지 확인
func (g *Game) IsPlayable() bool {
	return g.IsActive() && g.PlayURL != ""
}

// 게임이 무료인지 확인
func (g *Game) IsFree() bool {
	return g.Price == 0
}

// 게임이 할인 중인지 확인
func (g *Game) IsDiscounted() bool {
	if g.DiscountRate <= 0 {
		return false
	}

	if g.DiscountEndDate == nil {
		return true
	}

	return time.Now().Before(*g.DiscountEndDate)
}

// 할인된 가격을 반환
func (g *Game) GetDiscountedPrice() int {
	if !g.IsDiscounted() {
		return g.Price
	}

	discountAmount := int(float64(g.Price) * float64(g.DiscountRate) / 100.0)
	return g.Price - discountAmount
}

// 게임 플레이 기록을 추가
func (g *Game) AddPlay(playTime int) {
	g.TotalPlays++
	g.TotalPlayTime += playTime
	g.updatePopularityScore()
}

// 게임 평점을 추가
func (g *Game) AddRating(rating float64) error {
	if rating < 1 || rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}

	// 평균 평점 업데이트
	totalRating := g.AverageRating * float64(g.TotalRatings)
	g.TotalRatings++
	g.AverageRating = (totalRating + rating) / float64(g.TotalRatings)

	return nil
}

// 게임 태그를 슬라이스로 반환
func (g *Game) GetTags() []string {
	if g.Tags == "" {
		return []string{}
	}

	tags := strings.Split(g.Tags, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}

	return tags
}

// 게임 태그를 설정
func (g *Game) SetTags(tags []string) {
	g.Tags = strings.Join(tags, ", ")
}

// 지원 언어를 슬라이스로 반환
func (g *Game) GetSupportedLanguages() []string {
	if g.SupportedLanguages == "" {
		return []string{"ko"}
	}

	languages := strings.Split(g.SupportedLanguages, ",")
	for i, lang := range languages {
		languages[i] = strings.TrimSpace(lang)
	}

	return languages
}

// 지원 플랫폼을 슬라이스로 반환
func (g *Game) GetSupportedPlatforms() []string {
	if g.SupportedPlatforms == "" {
		return []string{"web"}
	}

	platforms := strings.Split(g.SupportedPlatforms, ",")
	for i, platform := range platforms {
		platforms[i] = strings.TrimSpace(platform)
	}

	return platforms
}

// 인기도 점수를 업데이트
func (g *Game) updatePopularityScore() {
	// 간단한 인기도 계산: 플레이 수 + 평점 + 최근 활동
	playScore := float64(g.TotalPlays) * 0.1
	ratingScore := g.AverageRating * float64(g.TotalRatings) * 0.2
	timeScore := float64(g.TotalPlayTime) * 0.001

	g.PopularityScore = playScore + ratingScore + timeScore
}

// 게임 카테고리가 유효한지 확인
func isValidGameCategory(category GameCategory) bool {
	validCategories := []GameCategory{
		GameCategoryPuzzle, GameCategoryAction, GameCategoryStrategy,
		GameCategoryRPG, GameCategoryAdventure, GameCategoryRacing,
		GameCategorySports, GameCategorySimulation, GameCategoryCasual,
		GameCategoryEducational, GameCategoryMusic, GameCategoryBoard,
		GameCategoryCard, GameCategoryArcade, GameCategoryOther,
	}

	for _, valid := range validCategories {
		if category == valid {
			return true
		}
	}
	return false
}

// 게임 난이도가 유효한지 확인
func isValidGameDifficulty(difficulty GameDifficulty) bool {
	validDifficulties := []GameDifficulty{
		GameDifficultyEasy, GameDifficultyNormal, GameDifficultyHard, GameDifficultyExpert,
	}

	for _, valid := range validDifficulties {
		if difficulty == valid {
			return true
		}
	}
	return false
}

// 게임 상태가 유효한지 확인
func isValidGameStatus(status GameStatus) bool {
	validStatuses := []GameStatus{
		GameStatusActive, GameStatusInactive, GameStatusMaintenance,
		GameStatusBeta, GameStatusAlpha,
	}

	for _, valid := range validStatuses {
		if status == valid {
			return true
		}
	}
	return false
}
