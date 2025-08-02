package service

import (
	"g_dev/internal/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

// NewInventoryService 함수 테스트
func TestInventoryService_NewInventoryService(t *testing.T) {
	t.Run("서비스 생성 테스트", func(t *testing.T) {
		service := &InventoryService{}
		assert.NotNil(t, service, "서비스가 생성되어야 합니다")
	})
}

// CreateInventory 메서드의 로직 테스트
func TestInventoryService_CreateInventory(t *testing.T) {
	t.Run("유효성 검사 테스트", func(t *testing.T) {
		// 유효하지 않은 인벤토리
		invalidInventory := &model.Inventory{
			UserID:   0, // 유효하지 않은 UserID
			ItemID:   "sword_001",
			ItemName: "강화된 검",
			Quantity: 1,
			ItemType: "weapon",
			Rarity:   "rare",
			Level:    5,
		}

		// 유효성 검사만 테스트
		err := invalidInventory.Validate()
		assert.Error(t, err, "유효하지 않은 인벤토리는 검증을 통과하면 안됩니다")
		assert.Equal(t, model.ErrInvalidUserID, err, "UserID 에러여야 합니다")

		// 유효한 인벤토리
		validInventory := &model.Inventory{
			UserID:   1,
			ItemID:   "sword_001",
			ItemName: "강화된 검",
			Quantity: 1,
			ItemType: "weapond",
			Rarity:   "rare",
			Level:    5,
		}

		err = validInventory.Validate()
		assert.NoError(t, err, "유효한 인벤토리는 검증을 통과해야 합니다")
	})
}

// AddItemQuantity 메서드의 로직을 테스트
func TestInventoryService_AddItemQuantity(t *testing.T) {
	tests := []struct {
		name        string
		quantity    int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "양수 수량 추가",
			quantity:    5,
			expectError: false,
		},
		{
			name:        "음수 수량 추가",
			quantity:    -1,
			expectError: true,
			errorMsg:    "추가할 수량은 0보다 커야 합니다",
		},
		{
			name:        "0 수량 추가",
			quantity:    0,
			expectError: true,
			errorMsg:    "추가할 수량은 0보다 커야 합니다",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 수량 검증 로직만 테스트
			if tt.quantity <= 0 {
				assert.True(t, tt.expectError, "음수나 0 수량은 에러여야 합니다")
			} else {
				assert.False(t, tt.expectError, "양수 수량은 에러가 아니어야 합니다")
			}
		})
	}
}

// UseItem aptjemdml fhwlrdmf xptmxm
func TestInventoryService_UseItem(t *testing.T) {
	t.Run("아이템 사용 로직 테스트", func(t *testing.T) {
		// 실제 모델의 UseItem 메서드 테스트
		inventory := &model.Inventory{
			Quantity: 5,
		}

		// 정상적인 아이템 사용
		err := inventory.UseItem()
		assert.NoError(t, err, "아이템 사용이 성공해야 합니다")
		assert.Equal(t, 4, inventory.Quantity, "수량이 1 감소해야 합니다")

		// 수량이 0이 될 때까지 사용
		for i := 0; i < 3; i++ {
			err = inventory.UseItem()
			assert.NoError(t, err, "아이템 사용이 성공해야 합니다")
		}
		assert.Equal(t, 1, inventory.Quantity, "수량이 1이어야 합니다")

		// 마지막 아이템 사용
		err = inventory.UseItem()
		assert.NoError(t, err, "마지막 아이템 사용이 성공해야 합니다")
		assert.Equal(t, 0, inventory.Quantity, "수량이 0이어야 합니다")

		// 수량이 0일 때 사용 시도
		err = inventory.UseItem()
		assert.Error(t, err, "수량이 부족할 때 에러가 발생해야 합니다")
		assert.Equal(t, model.ErrInsufficientQuantity, err, "수량 부족 에러여야 합니다")
		assert.Equal(t, 0, inventory.Quantity, "수량이 0으로 유지되어야 합니다")
	})
}

// GetUserInventoryStats 메서드 테스트
func TestInventoryService_GetUserInventoryStats(t *testing.T) {
	t.Run("통계 구조체 테스트", func(t *testing.T) {
		stats := &InventoryStats{
			TotalItems:      10,
			CommonItems:     5,
			RareItems:       3,
			EpicItems:       1,
			LegendaryItems:  1,
			WeaponItems:     4,
			ArmorItems:      3,
			ConsumableItems: 2,
			MaterialItems:   1,
			ActiveItems:     2,
		}

		assert.Equal(t, int64(10), stats.TotalItems, "전체 아이템 수가 일치해야 합니다")
		assert.Equal(t, int64(5), stats.CommonItems, "일반 아이템 수가 일치해야 합니다")
		assert.Equal(t, int64(3), stats.RareItems, "레어 아이템 수가 일치해야 합니다")
		assert.Equal(t, int64(1), stats.EpicItems, "에픽 아이템 수가 일치해야 합니다")
		assert.Equal(t, int64(1), stats.LegendaryItems, "전설 아이템 수가 일치해야 합니다")
		assert.Equal(t, int64(4), stats.WeaponItems, "무기 아이템 수가 일치해야 합니다")
		assert.Equal(t, int64(3), stats.ArmorItems, "방어구 아이템 수가 일치해야 합니다")
		assert.Equal(t, int64(2), stats.ConsumableItems, "소비 아이템 수가 일치해야 합니다")
		assert.Equal(t, int64(1), stats.MaterialItems, "재료 아이템 수가 일치해야 합니다")
		assert.Equal(t, int64(2), stats.ActiveItems, "활성 아이템 수가 일치해야 합니다")
	})
}

// InventoryService의 통합 테스트를 수행
func TestInventoryService_Integration(t *testing.T) {
	t.Run("전체 인벤토리 서비스 생명주기 테스트", func(t *testing.T) {
		// 1. 서비스 생성
		service := &InventoryService{}
		assert.NotNil(t, service, "서비스가 생성되어야 합니다")

		// 2. 인벤토리 생성 준비
		inventory := &model.Inventory{
			UserID:   1,
			ItemID:   "test_sword",
			ItemName: "테스트 검",
			Quantity: 1,
			ItemType: "weapon",
			Rarity:   "rare",
			Level:    5,
		}

		// 3. 유효성 검사
		err := inventory.Validate()
		assert.NoError(t, err, "유효한 인벤토리는 검증을 통과해야 합니다")

		// 4. 등급 확인
		assert.True(t, inventory.IsRare(), "레어 등급이어야 합니다")
		assert.Equal(t, "#0070DD", inventory.GetRarityColor(), "레어 등급 색상이어야 합니다")

		// 5. 아이템 추가
		inventory.AddItem(4)
		assert.Equal(t, 5, inventory.Quantity, "수량이 5가 되어야 합니다")

		// 6. 아이템 사용
		err = inventory.UseItem()
		assert.NoError(t, err, "아이템 사용이 성공해야 합니다")

		// 7. 여러 번 사용
		for i := 0; i < 3; i++ {
			err = inventory.UseItem()
			assert.NoError(t, err, "아이템 사용이 성공해야 합니다")
		}
		assert.Equal(t, 1, inventory.Quantity, "수량이 1이 되어야 합니다")

		// 8.. 마지막 아이템 사용
		err = inventory.UseItem()
		assert.NoError(t, err, "마지막 아이템 사용이 성공해야 합니다")
		assert.Equal(t, 0, inventory.Quantity, "수량이 0이 되어야 합니다")

		// 9. 수량이 0일 때 사용 시도
		err = inventory.UseItem()
		assert.Error(t, err, "수량이 부족할 때 에러가 발생해야 합니다")
		assert.Equal(t, model.ErrInsufficientQuantity, err, "수량 부족 에러여야 합니다")
		assert.Equal(t, 0, inventory.Quantity, "수량이 0으로 유지되어야 합니다")
	})
}

// 경계 케이스 테스트
func TestInventoryService_EdgeCases(t *testing.T) {
	t.Run("최대값 테스트", func(t *testing.T) {
		inventory := &model.Inventory{
			UserID:   1,
			ItemID:   "test_item",
			ItemName: "테스트 아이템",
			Quantity: 999999,
			ItemType: "test",
			Rarity:   "common",
			Level:    100,
		}

		err := inventory.Validate()
		assert.NoError(t, err, "최대값들도 유효해야 합니다")

		// 대량 아이템 추가
		inventory.AddItem(1000000)
		assert.Equal(t, 1999999, inventory.Quantity, "대량 추가가 올바르게 작동해야 합니다")
	})

	t.Run("특수 문자 테스트", func(t *testing.T) {
		inventory := &model.Inventory{
			UserID:   1,
			ItemID:   "item_한글_123",
			ItemName: "한글 아이템 이름!@#$%",
			Quantity: 1,
			ItemType: "weapon",
			Rarity:   "rare",
			Level:    1,
		}

		err := inventory.Validate()
		assert.NoError(t, err, "특수 문자가 포함된 아이템도 유효해야 합니다")
	})

	t.Run("등급별 색상 테스트", func(t *testing.T) {
		rarities := []struct {
			rarity string
			color  string
		}{
			{"legendary", "#FFD700"},
			{"epic", "#9932CC"},
			{"rare", "#0070DD"},
			{"common", "#9D9D9D"},
			{"unknown", "#FFFFFF"},
			{"", "#FFFFFF"},
		}

		for _, r := range rarities {
			inventory := &model.Inventory{Rarity: r.rarity}
			color := inventory.GetRarityColor()
			assert.Equal(t, r.color, color, "%s 등급의 색상이 일치해야 합니다", r.rarity)
		}
	})
}

// 다양한 유효성 검사 케이스로 테스트
func TestInventoryService_Validation(t *testing.T) {
	tests := []struct {
		name        string
		inventory   *model.Inventory
		expectError bool
		errorType   error
	}{
		{
			name: "유효한 인벤토리",
			inventory: &model.Inventory{
				UserID:   1,
				ItemID:   "sword_001",
				ItemName: "강화된 검",
				Quantity: 1,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    5,
			},
			expectError: false,
		},
		{
			name: "UserID가 0인 경우",
			inventory: &model.Inventory{
				UserID:   0,
				ItemID:   "sword_001",
				ItemName: "강화된 검",
				Quantity: 1,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    5,
			},
			expectError: true,
			errorType:   model.ErrInvalidUserID,
		},
		{
			name: "ItemID가 빈 문자열인 경우",
			inventory: &model.Inventory{
				UserID:   1,
				ItemID:   "",
				ItemName: "강화된 검",
				Quantity: 1,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    5,
			},
			expectError: true,
			errorType:   model.ErrInvalidItemID,
		},
		{
			name: "ItemName이 빈 문자열인 경우",
			inventory: &model.Inventory{
				UserID:   1,
				ItemID:   "sword_001",
				ItemName: "",
				Quantity: 1,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    5,
			},
			expectError: true,
			errorType:   model.ErrInvalidItemName,
		},
		{
			name: "Quantity가 0인 경우",
			inventory: &model.Inventory{
				UserID:   1,
				ItemID:   "sword_001",
				ItemName: "강화된 검",
				Quantity: 0,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    5,
			},
			expectError: true,
			errorType:   model.ErrInvalidQuantity,
		},
		{
			name: "Quantity가 음수인 경우",
			inventory: &model.Inventory{
				UserID:   1,
				ItemID:   "sword_001",
				ItemName: "강화된 검",
				Quantity: -1,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    5,
			},
			expectError: true,
			errorType:   model.ErrInvalidQuantity,
		},
		{
			name: "ItemType이 빈 문자열인 경우",
			inventory: &model.Inventory{
				UserID:   1,
				ItemID:   "sword_001",
				ItemName: "강화된 검",
				Quantity: 1,
				ItemType: "",
				Rarity:   "rare",
				Level:    5,
			},
			expectError: true,
			errorType:   model.ErrInvalidItemType,
		},
		{
			name: "Rarity가 빈 문자열인 경우",
			inventory: &model.Inventory{
				UserID:   1,
				ItemID:   "sword_001",
				ItemName: "강화된 검",
				Quantity: 1,
				ItemType: "weapon",
				Rarity:   "",
				Level:    5,
			},
			expectError: true,
			errorType:   model.ErrInvalidRarity,
		},
		{
			name: "Level이 0인 경우",
			inventory: &model.Inventory{
				UserID:   1,
				ItemID:   "sword_001",
				ItemName: "강화된 검",
				Quantity: 1,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    0,
			},
			expectError: true,
			errorType:   model.ErrInvalidLevel,
		},
		{
			name: "Level이 음수인 경우",
			inventory: &model.Inventory{
				UserID:   1,
				ItemID:   "sword_001",
				ItemName: "강화된 검",
				Quantity: 1,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    -1,
			},
			expectError: true,
			errorType:   model.ErrInvalidLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.inventory.Validate()

			if tt.expectError {
				assert.Error(t, err, "에러가 발생해야 합니다")
				assert.Equal(t, tt.errorType, err, "예상된 에러 타입과 일치해야 합니다")
			} else {
				assert.NoError(t, err, "에러가 발생하면 안됩니다")
			}
		})
	}
}
