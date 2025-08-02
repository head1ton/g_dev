package service

import (
	"errors"
	"fmt"
	"g_dev/internal/model"
	"gorm.io/gorm"
)

type InventoryService struct {
	db *gorm.DB
}

func NewInventoryService(db *gorm.DB) *InventoryService {
	return &InventoryService{
		db: db,
	}
}

// 새로운 인벤토리 아이템 생성
func (s *InventoryService) CreateInventory(inventory *model.Inventory) error {
	// 유효성 검사
	if err := inventory.Validate(); err != nil {
		return fmt.Errorf("인벤토리 유효성 검사 실패: %w", err)
	}

	// 사용자 존재 여부 확인
	var userCount int64
	if err := s.db.Model(&model.User{}).Where("id = ?", inventory.UserID).Count(&userCount).Error; err != nil {
		return fmt.Errorf("사용자 확인 중 오류 발생: %w", err)
	}
	if userCount == 0 {
		return model.ErrInvalidUserID
	}

	// 동일한 아이템이 이미 존재하는지 확인
	var existingCount int64
	if err := s.db.Model(&model.Inventory{}).
		Where("user_id = ? AND item_id = ?", inventory.UserID, inventory.ItemID).
		Count(&existingCount).Error; err != nil {
		return fmt.Errorf("기존 아이템 확인 중 오류 발생: %w", err)
	}

	if existingCount > 0 {
		// 기존 아이템이 있으면 수량만 증가
		return s.AddItemQuantity(inventory.UserID, inventory.ItemID, inventory.Quantity)
	}

	// 새 아이템 생성
	if err := s.db.Create(inventory).Error; err != nil {
		return fmt.Errorf("인벤토리 생성 중 오류 발생: %w", err)
	}

	return nil
}

// ID로 인벤토리 아이템을 조회
func (s *InventoryService) GetInventoryByID(id uint) (*model.Inventory, error) {
	var inventory model.Inventory
	if err := s.db.Preload("User").First(&inventory, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("인벤토리 아이템을 찾을 수 없습니다: %d", id)
		}
		return nil, fmt.Errorf("인벤토리 조회 중 오류 발생: %w", err)
	}
	return &inventory, nil
}

// 사용자의 모든 인벤토리 아이템을 조회
func (s *InventoryService) GetUserInventory(userID uint) ([]model.Inventory, error) {
	var inventories []model.Inventory
	if err := s.db.Where("user_id = ?", userID).Find(&inventories).Error; err != nil {
		return nil, fmt.Errorf("사용자 인벤토리 조회 중 오류 발생: %w", err)
	}
	return inventories, nil
}

// 사용자의 특정 타입 인벤토리 아이템 조회
func (s *InventoryService) GetUserInventoryByType(userID uint, itemType string) ([]model.Inventory, error) {
	var inventories []model.Inventory
	if err := s.db.Where("user_id = ? AND item_type = ?", userID, itemType).Find(&inventories).Error; err != nil {
		return nil, fmt.Errorf("사용자 인벤토리 타입별 조회 중 오류 발생: %w", err)
	}
	return inventories, nil
}

// 사용자의 특정 등급 인벤토리 아이템을 조회
func (s *InventoryService) GetUserInventoryByRarity(userID uint, rarity string) ([]model.Inventory, error) {
	var inventories []model.Inventory
	if err := s.db.Where("user_id = ? AND rarity = ?", userID, rarity).Find(&inventories).Error; err != nil {
		return nil, fmt.Errorf("사용자 인벤토리 등급별 조회 중 오류 발생: %w", err)
	}
	return inventories, nil
}

// 인벤토리 아이템을 업데이트
func (s *InventoryService) UpdateInventory(inventory *model.Inventory) error {
	// 유효성 검사
	if err := inventory.Validate(); err != nil {
		return fmt.Errorf("인벤토리 유효성 검사 실패: %w", err)
	}

	// 기존 아이템 존재 여부 확인
	var existingInventory model.Inventory
	if err := s.db.First(&existingInventory, inventory.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("업데이트할 인벤토리 아이템을 찾을 수 없습니다: %d", inventory.ID)
		}
		return fmt.Errorf("기존 인벤토리 조회 중 오류 발생: %w", err)
	}

	// 업데이트
	if err := s.db.Save(inventory).Error; err != nil {
		return fmt.Errorf("인벤토리 업데이트 중 오류 발생: %w", err)
	}

	return nil
}

// 인벤토리 아이템을 삭제
func (s *InventoryService) DeleteInventory(id uint) error {
	var inventory model.Inventory
	if err := s.db.First(&inventory, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("삭제할 인벤토리 아이템을 찾을 수 없습니다: %d", id)
		}
		return fmt.Errorf("인벤토리 조회 중 오류 발생: %w", err)
	}

	if err := s.db.Delete(&inventory).Error; err != nil {
		return fmt.Errorf("인벤토리 삭제 중 오류 발생: %w", err)
	}

	return nil
}

// 특정 아이템의 수량을 증가
func (s *InventoryService) AddItemQuantity(userID uint, itemID string, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("추가할 수량은 0보다 커야 합니다: %d", quantity)
	}

	var inventory model.Inventory
	if err := s.db.Where("user_id = ? AND item_id = ?", userID, itemID).First(&inventory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("아이템을 찾을 수 없습니다: user_id=%d, item_id=%s", userID, itemID)
		}
		return fmt.Errorf("아이템 조회 중 오류 발생: %w", err)
	}

	inventory.AddItem(quantity)
	if err := s.db.Save(&inventory).Error; err != nil {
		return fmt.Errorf("아이템 수량 업데이트 중 오류 발생: %w", err)
	}

	return nil
}

// 특정 아이템을 사용
func (s *InventoryService) UseItem(userID uint, itemID string) error {
	var inventory model.Inventory
	if err := s.db.Where("user_id = ? AND item_id = ?", userID, itemID).First(&inventory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("아이템을 찾을 수 없습니다: user_id=%d, item_id=%s", userID, itemID)
		}
		return fmt.Errorf("아이템 조회 중 오류 발생: %w", err)
	}

	if err := inventory.UseItem(); err != nil {
		return fmt.Errorf("아이템 사용 실패: %w", err)
	}

	// 수량이 0이 되면 아이템 삭제
	if inventory.Quantity == 0 {
		if err := s.db.Delete(&inventory).Error; err != nil {
			return fmt.Errorf("빈 아이템 삭제 중 오류 발생: %w", err)
		}
	} else {
		if err := s.db.Save(&inventory).Error; err != nil {
			return fmt.Errorf("아이템 수량 업데이트 중 오류 발생: %w", err)
		}
	}

	return nil
}

// 사용자의 인벤토리 통계를 반환
func (s *InventoryService) GetUserInventoryStats(userID uint) (*InventoryStats, error) {
	var stats InventoryStats

	// 전체 아이템 수
	if err := s.db.Model(&model.Inventory{}).Where("user_id = ?", userID).Count(&stats.TotalItems).Error; err != nil {
		return nil, fmt.Errorf("전체 아이템 수 조회 중 오류 발생: %w", err)
	}

	// 등급별 아이템 수
	rarities := []string{"common", "rare", "epic", "legendary"}
	for _, rarity := range rarities {
		var count int64
		if err := s.db.Model(&model.Inventory{}).Where("user_id = ? AND rarity = ?", userID, rarity).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("%s 등급 아이템 수 조회 중 오류 발생: %w", rarity, err)
		}
		switch rarity {
		case "common":
			stats.CommonItems = count
		case "rare":
			stats.RareItems = count
		case "epic":
			stats.EpicItems = count
		case "legendary":
			stats.LegendaryItems = count
		}
	}

	// 타입별 아이템 수
	types := []string{"weapon", "armor", "consumable", "material"}
	for _, itemType := range types {
		var count int64
		if err := s.db.Model(&model.Inventory{}).Where("user_id = ? AND item_type = ?", userID, itemType).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("%s 타입 아이템 수 조회 중 오류 발생: %w", itemType, err)
		}
		switch itemType {
		case "weapon":
			stats.WeaponItems = count
		case "armor":
			stats.ArmorItems = count
		case "consumable":
			stats.ConsumableItems = count
		case "material":
			stats.MaterialItems = count
		}
	}

	// 활성화된 아이템 수
	if err := s.db.Model(&model.Inventory{}).Where("user_id = ? AND is_active = ?", userID, true).Count(&stats.ActiveItems).Error; err != nil {
		return nil, fmt.Errorf("활성화된 아이템 수 조회 중 오류 발생: %w", err)
	}

	return &stats, nil
}

// 아이템을 활성화
func (s *InventoryService) ActivateItem(userID uint, itemID string) error {
	var inventory model.Inventory
	if err := s.db.Where("user_id = ? AND item_id = ?", userID, itemID).First(&inventory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("아이템을 찾을 수 없습니다: user_id=%d, item_id=%s", userID, itemID)
		}
		return fmt.Errorf("아이템 조회 중 오류 발생: %w", err)
	}

	// 같은 타입의 다른 아이템들을 비활성화
	if err := s.db.Model(&model.Inventory{}).
		Where("user_id = ? AND item_type = ? AND is_active = ?", userID, inventory.ItemType, true).
		Update("is_active", false).Error; err != nil {
		return fmt.Errorf("기존 활성 아이템 비활성화 중 오류 발생: %w", err)
	}

	// 현재 아이템 활성화
	inventory.IsActive = true
	if err := s.db.Save(&inventory).Error; err != nil {
		return fmt.Errorf("아이템 활성화 중 오류 발생: %w", err)
	}

	return nil
}

// 아이템을 비활성화
func (s *InventoryService) DeactivateItem(userID uint, itemID string) error {
	var inventory model.Inventory
	if err := s.db.Where("user_id = ? AND item_id = ?", userID, itemID).First(&inventory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("아이템을 찾을 수 없습니다: user_id=%d, item_id=%s", userID, itemID)
		}
		return fmt.Errorf("아이템 조회 중 오류 발생: %w", err)
	}

	inventory.IsActive = false
	if err := s.db.Save(&inventory).Error; err != nil {
		return fmt.Errorf("아이템 비활성화 중 오류 발생: %w", err)
	}

	return nil
}

// 사용자의 활성화된 아이템들을 조회
func (s *InventoryService) GetActiveItems(userID uint) ([]model.Inventory, error) {
	var inventories []model.Inventory
	if err := s.db.Where("user_id = ? AND is_active = ?", userID, true).Find(&inventories).Error; err != nil {
		return nil, fmt.Errorf("활성화된 아이템 조회 중 오류 발생: %w", err)
	}
	return inventories, nil
}

// 인벤토리 통계 정보를 담는 구조체
type InventoryStats struct {
	TotalItems      int64 `json:"total_items"`
	CommonItems     int64 `json:"common_items"`
	RareItems       int64 `json:"rare_items"`
	EpicItems       int64 `json:"epic_items"`
	LegendaryItems  int64 `json:"legendary_items"`
	WeaponItems     int64 `json:"weapon_items"`
	ArmorItems      int64 `json:"armor_items"`
	ConsumableItems int64 `json:"consumable_items"`
	MaterialItems   int64 `json:"material_items"`
	ActiveItems     int64 `json:"active_items"`
}
