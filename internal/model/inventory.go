package model

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

// 인벤토리 아이템 모델
type Inventory struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null;index"`
	ItemID    string         `json:"item_id" gorm:"not null;size:50"`
	ItemName  string         `json:"item_name" gorm:"not null;size:100"`
	Quantity  int            `json:"quantity" gorm:"not null;default:1"`
	ItemType  string         `json:"item_type" gorm:"not null;size:20"` // weapon, armor, consumable, etc
	Rarity    string         `json:"rarity" gorm:"not null;size:20"`    // common, rare, epic, legendary
	Level     int            `json:"level" gorm:"not null;default:1"`
	IsActive  bool           `json:"is_active" gorm:"not null;default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// 관계 설정
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// GORM에서 사용할 테이블 이름을 지정
func (Inventory) TableName() string {
	return "inventories"
}

// GORM 훅으로, 레코드 생성 전에 실행
func (i *Inventory) BeforeCreate(tx *gorm.DB) error {
	if i.CreatedAt.IsZero() {
		i.CreatedAt = time.Now()
	}
	if i.UpdatedAt.IsZero() {
		i.UpdatedAt = time.Now()
	}
	return nil
}

// GORM 훅으로, 레코드 업데이트 전에 실행
func (i *Inventory) BeforeUpdate(tx *gorm.DB) error {
	i.UpdatedAt = time.Now()
	return nil
}

// 인벤토리 아이템 유효성 검사
func (i *Inventory) Validate() error {
	if i.UserID == 0 {
		return ErrInvalidUserID
	}
	if i.ItemID == "" {
		return ErrInvalidItemID
	}
	if i.ItemName == "" {
		return ErrInvalidItemName
	}
	if i.Quantity <= 0 {
		return ErrInvalidQuantity
	}
	if i.ItemType == "" {
		return ErrInvalidItemType
	}
	if i.Rarity == "" {
		return ErrInvalidRarity
	}
	if i.Level <= 0 {
		return ErrInvalidLevel
	}
	return nil
}

// 아이템 사용
func (i *Inventory) UseItem() error {
	if i.Quantity <= 0 {
		return ErrInsufficientQuantity
	}
	i.Quantity--
	return nil
}

// 아이템 추가
func (i *Inventory) AddItem(quantity int) {
	i.Quantity += quantity
}

// 아이템이 전설 등급인지 확인
func (i *Inventory) IsLegendary() bool {
	return i.Rarity == "legendary"
}

// 에픽 등급인지 확인
func (i *Inventory) IsEpic() bool {
	return i.Rarity == "epic"
}

// 레어 등급인지 확인
func (i *Inventory) IsRare() bool {
	return i.Rarity == "rare"
}

// 일반 등급인지 확인
func (i *Inventory) IsCommon() bool {
	return i.Rarity == "common"
}

// 등급에 따른 색상 반환
func (i *Inventory) GetRarityColor() string {
	switch i.Rarity {
	case "legendary":
		return "#FFD700" // 골드
	case "epic":
		return "#9932CC" // 보라
	case "rare":
		return "#0070DD" // 파랑
	case "common":
		return "#9D9D9D" // 회색
	default:
		return "#FFFFFF" // 흰색
	}
}

// 에러 정의
var (
	ErrInvalidUserID        = errors.New("사용자 ID가 유효하지 않습니다")
	ErrInvalidItemID        = errors.New("아이템 ID가 유효하지 않습니다")
	ErrInvalidItemName      = errors.New("아이템 이름이 유효하지 않습니다")
	ErrInvalidQuantity      = errors.New("수량이 유효하지 않습니다")
	ErrInvalidItemType      = errors.New("아이템 타입이 유효하지 않습니다")
	ErrInvalidRarity        = errors.New("등급이 유효하지 않습니다")
	ErrInvalidLevel         = errors.New("레벨이 유효하지 않습니다")
	ErrInsufficientQuantity = errors.New("수량이 부족합니다")
)
