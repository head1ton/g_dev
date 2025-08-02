package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// 테이블 이름이 올바르게 반환되는지 테스트
func TestInventory_TableName(t *testing.T) {
	inventory := &Inventory{}
	expected := "inventories"
	actual := inventory.TableName()

	assert.Equal(t, expected, actual, "테이블 이름이 'inventories'여야 합니다.")
}

// BeforeCreate 훅이 올바르게 작동하는지 테스트
func TestInventory_BeforeCreate(t *testing.T) {
	tests := []struct {
		name           string
		inventory      *Inventory
		expectedResult bool
	}{
		{
			name: "정상적인 인벤토리 생성",
			inventory: &Inventory{
				UserID:   1,
				ItemID:   "sword_001",
				ItemName: "강화된 검",
				Quantity: 1,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    5,
			},
			expectedResult: true,
		},
		{
			name: "이미 시간이 설정된 인벤토리",
			inventory: &Inventory{
				UserID:    1,
				ItemID:    "sword_001",
				ItemName:  "강화된 검",
				Quantity:  1,
				ItemType:  "weapon",
				Rarity:    "rare",
				Level:     5,
				CreatedAt: time.Now().Add(-time.Hour),
				UpdatedAt: time.Now().Add(-time.Hour),
			},
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalCreatedAt := tt.inventory.CreatedAt
			originalUpdatedAt := tt.inventory.UpdatedAt

			err := tt.inventory.BeforeCreate(nil)

			assert.NoError(t, err, "BeforeCreate에서 에러가 발생하면 안됩니다")

			if originalCreatedAt.IsZero() {
				assert.False(t, tt.inventory.CreatedAt.IsZero(), "CreatedAt이 설정되어야 합니다")
			} else {
				assert.Equal(t, originalCreatedAt, tt.inventory.CreatedAt, "기존 CreatedAt이 유지되어야 합니다")
			}

			if originalUpdatedAt.IsZero() {
				assert.False(t, tt.inventory.UpdatedAt.IsZero(), "UpdatedAt이 설정되어야 합니다")
			} else {
				assert.Equal(t, originalUpdatedAt, tt.inventory.UpdatedAt, "기존 UpdatedAt이 유지되어야 합니다")
			}
		})
	}
}

// BeforeUpdate 훅이 올바르게 작동하는지 테스트
func TestInventory_BeforeUpdate(t *testing.T) {
	originalTime := time.Now().Add(-time.Hour)
	inventory := &Inventory{
		UserID:    1,
		ItemID:    "sword_001",
		ItemName:  "강화된 검",
		Quantity:  1,
		ItemType:  "weapon",
		Rarity:    "rare",
		Level:     5,
		UpdatedAt: originalTime,
	}

	err := inventory.BeforeUpdate(nil)

	assert.NoError(t, err, "BeforeUpdate에서 에러가 발생하면 안됩니다")
	assert.True(t, inventory.UpdatedAt.After(originalTime), "UpdatedAt이 업데이트되어야 합니다")
}

// Validate 메서드가 올바르게 작동하는지 테스트
func TestInventory_Validate(t *testing.T) {
	tests := []struct {
		name        string
		inventory   *Inventory
		expectError bool
		errorType   error
	}{
		{
			name: "유효한 인벤토리",
			inventory: &Inventory{
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
			inventory: &Inventory{
				UserID:   0,
				ItemID:   "sword_001",
				ItemName: "강화된 검",
				Quantity: 1,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    5,
			},
			expectError: true,
			errorType:   ErrInvalidUserID,
		},
		{
			name: "ItemID가 빈 문자열인 경우",
			inventory: &Inventory{
				UserID:   1,
				ItemID:   "",
				ItemName: "강화된 검",
				Quantity: 1,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    5,
			},
			expectError: true,
			errorType:   ErrInvalidItemID,
		},
		{
			name: "ItemName이 빈 문자열인 경우",
			inventory: &Inventory{
				UserID:   1,
				ItemID:   "sword_001",
				ItemName: "",
				Quantity: 1,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    5,
			},
			expectError: true,
			errorType:   ErrInvalidItemName,
		},
		{
			name: "Quantity가 0인 경우",
			inventory: &Inventory{
				UserID:   1,
				ItemID:   "sword_001",
				ItemName: "강화된 검",
				Quantity: 0,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    5,
			},
			expectError: true,
			errorType:   ErrInvalidQuantity,
		},
		{
			name: "Quantity가 음수인 경우",
			inventory: &Inventory{
				UserID:   1,
				ItemID:   "sword_001",
				ItemName: "강화된 검",
				Quantity: -1,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    5,
			},
			expectError: true,
			errorType:   ErrInvalidQuantity,
		},
		{
			name: "ItemType이 빈 문자열인 경우",
			inventory: &Inventory{
				UserID:   1,
				ItemID:   "sword_001",
				ItemName: "강화된 검",
				Quantity: 1,
				ItemType: "",
				Rarity:   "rare",
				Level:    5,
			},
			expectError: true,
			errorType:   ErrInvalidItemType,
		},
		{
			name: "Rarity가 빈 문자열인 경우",
			inventory: &Inventory{
				UserID:   1,
				ItemID:   "sword_001",
				ItemName: "강화된 검",
				Quantity: 1,
				ItemType: "weapon",
				Rarity:   "",
				Level:    5,
			},
			expectError: true,
			errorType:   ErrInvalidRarity,
		},
		{
			name: "Level이 0인 경우",
			inventory: &Inventory{
				UserID:   1,
				ItemID:   "sword_001",
				ItemName: "강화된 검",
				Quantity: 1,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    0,
			},
			expectError: true,
			errorType:   ErrInvalidLevel,
		},
		{
			name: "Level이 음수인 경우",
			inventory: &Inventory{
				UserID:   1,
				ItemID:   "sword_001",
				ItemName: "강화된 검",
				Quantity: 1,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    -1,
			},
			expectError: true,
			errorType:   ErrInvalidLevel,
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

// UseItem 메서드가 올바르게 작동하는지 테스트
func TestInventory_UseItem(t *testing.T) {
	tests := []struct {
		name        string
		inventory   *Inventory
		expectError bool
		errorType   error
		expectedQty int
	}{
		{
			name: "수량이 1인 아이템 사용",
			inventory: &Inventory{
				Quantity: 1,
			},
			expectError: false,
			expectedQty: 0,
		},
		{
			name: "수량이 5인 아이템 사용",
			inventory: &Inventory{
				Quantity: 5,
			},
			expectError: false,
			expectedQty: 4,
		},
		{
			name: "수량이 0인 아이템 사용",
			inventory: &Inventory{
				Quantity: 0,
			},
			expectError: true,
			errorType:   ErrInsufficientQuantity,
			expectedQty: 0,
		},
		{
			name: "수량이 음수인 아이템 사용",
			inventory: &Inventory{
				Quantity: -1,
			},
			expectError: true,
			errorType:   ErrInsufficientQuantity,
			expectedQty: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalQty := tt.inventory.Quantity
			err := tt.inventory.UseItem()

			if tt.expectError {
				assert.Error(t, err, "에러가 발생해야 합니다")
				assert.Equal(t, tt.errorType, err, "예상된 에러 타입과 일치해야 합니다")
				assert.Equal(t, originalQty, tt.inventory.Quantity, "수량이 변경되면 안됩니다")
			} else {
				assert.NoError(t, err, "에러가 발생하면 안됩니다")
				assert.Equal(t, tt.expectedQty, tt.inventory.Quantity, "수량이 올바르게 감소해야 합니다")
			}
		})
	}
}

// AddItem 메서드 테스트
func TestInventory_AddItem(t *testing.T) {
	tests := []struct {
		name        string
		inventory   *Inventory
		addQuantity int
		expectedQty int
	}{
		{
			name: "양수 수량 추가",
			inventory: &Inventory{
				Quantity: 1,
			},
			addQuantity: 5,
			expectedQty: 6,
		},
		{
			name: "0 수량 추가",
			inventory: &Inventory{
				Quantity: 1,
			},
			addQuantity: 0,
			expectedQty: 1,
		},
		{
			name: "음수 수량 추가",
			inventory: &Inventory{
				Quantity: 10,
			},
			addQuantity: -3,
			expectedQty: 7,
		},
		{
			name: "초기 수량이 0인 경우",
			inventory: &Inventory{
				Quantity: 0,
			},
			addQuantity: 5,
			expectedQty: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.inventory.AddItem(tt.addQuantity)
			assert.Equal(t, tt.expectedQty, tt.inventory.Quantity, "수량이 올바르게 추가되어야 합니다")
		})
	}
}

// 등급 관련 메서드 테스트
func TestInventory_RarityMethods(t *testing.T) {
	tests := []struct {
		name              string
		rarity            string
		expectedLegendary bool
		expectedEpic      bool
		expectedRare      bool
		expectedCommon    bool
	}{
		{
			name:              "전설 등급",
			rarity:            "legendary",
			expectedLegendary: true,
			expectedEpic:      false,
			expectedRare:      false,
			expectedCommon:    false,
		},
		{
			name:              "에픽 등급",
			rarity:            "epic",
			expectedLegendary: false,
			expectedEpic:      true,
			expectedRare:      false,
			expectedCommon:    false,
		},
		{
			name:              "레어 등급",
			rarity:            "rare",
			expectedLegendary: false,
			expectedEpic:      false,
			expectedRare:      true,
			expectedCommon:    false,
		},
		{
			name:              "일반 등급",
			rarity:            "common",
			expectedLegendary: false,
			expectedEpic:      false,
			expectedRare:      false,
			expectedCommon:    true,
		},
		{
			name:              "알 수 없는 등급",
			rarity:            "unknown",
			expectedLegendary: false,
			expectedEpic:      false,
			expectedRare:      false,
			expectedCommon:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inventory := &Inventory{
				Rarity: tt.rarity,
			}

			assert.Equal(t, tt.expectedLegendary, inventory.IsLegendary(), "IsLegendary 결과가 일치해야 합니다")
			assert.Equal(t, tt.expectedEpic, inventory.IsEpic(), "IsEpic 결과가 일치해야 합니다")
			assert.Equal(t, tt.expectedRare, inventory.IsRare(), "IsRare 결과가 일치해야 합니다")
			assert.Equal(t, tt.expectedCommon, inventory.IsCommon(), "IsCommon 결과가 일치해야 합니다")
		})
	}
}

// 등급 색상 테스트
func TestInventory_GetRarityColor(t *testing.T) {
	tests := []struct {
		name     string
		rarity   string
		expected string
	}{
		{
			name:     "전설 등급 색상",
			rarity:   "legendary",
			expected: "#FFD700",
		},
		{
			name:     "에픽 등급 색상",
			rarity:   "epic",
			expected: "#9932CC",
		},
		{
			name:     "레어 등급 색상",
			rarity:   "rare",
			expected: "#0070DD",
		},
		{
			name:     "일반 등급 색상",
			rarity:   "common",
			expected: "#9D9D9D",
		},
		{
			name:     "알 수 없는 등급 색상",
			rarity:   "unknown",
			expected: "#FFFFFF",
		},
		{
			name:     "빈 문자열 등급 색상",
			rarity:   "",
			expected: "#FFFFFF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inventory := &Inventory{Rarity: tt.rarity}
			actual := inventory.GetRarityColor()
			assert.Equal(t, tt.expected, actual, "등급 색상이 일치해야 합니다")
		})
	}
}

// 인벤토리 모델의 통합 테스트
func TestInventory_Integration(t *testing.T) {
	t.Run("전체 인벤토리 생명주기 테스트", func(t *testing.T) {
		// 1. 인벤토리 생성
		inventory := &Inventory{
			UserID:   1,
			ItemID:   "sword_001",
			ItemName: "강화된 검",
			Quantity: 1,
			ItemType: "weapon",
			Rarity:   "legendary",
			Level:    10,
		}

		// 2. 유효성 검사
		err := inventory.Validate()
		assert.NoError(t, err, "유효한 인벤토리는 검증을 통과해야 합니다")

		// 3. 등급 확인
		assert.True(t, inventory.IsLegendary(), "전설 등급이어야 합니다")
		assert.Equal(t, "#FFD700", inventory.GetRarityColor(), "전설 등급 색상이어야 합니다")

		// 4. 아이템 추가
		inventory.AddItem(4)
		assert.Equal(t, 5, inventory.Quantity, "수량이 5가 되어야 합니다")

		// 5. 아이템 사용
		err = inventory.UseItem()
		assert.NoError(t, err, "아이템 사용이 성공해야 합니다")
		assert.Equal(t, 4, inventory.Quantity, "수량이 4가 되어야 합니다")

		// 6. 여러 번 사용
		for i := 0; i < 3; i++ {
			err = inventory.UseItem()
			assert.NoError(t, err, "아이템 사용이 성공해야 합니다")
		}
		assert.Equal(t, 1, inventory.Quantity, "수량이 1이 되어야 합니다")

		// 7. 마지막 아이템 사용
		err = inventory.UseItem()
		assert.NoError(t, err, "마지막 아이템 사용이 성공해야 합니다")
		assert.Equal(t, 0, inventory.Quantity, "수량이 0이 되어야 합니다")

		// 8. 수량이 0일 때 사용 시도
		err = inventory.UseItem()
		assert.Error(t, err, "수량이 부족할 때 에러가 발생해야 합니다")
		assert.Equal(t, ErrInsufficientQuantity, err, "수량 부족 에러여야 합니다")
		assert.Equal(t, 0, inventory.Quantity, "수량이 0으로 유지되어야 합니다")
	})
}

// 경계 케이스들을 테스트
func TestInventory_EdgeCases(t *testing.T) {
	t.Run("최대값 테스트", func(t *testing.T) {
		inventory := &Inventory{
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
		inventory := &Inventory{
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
}
