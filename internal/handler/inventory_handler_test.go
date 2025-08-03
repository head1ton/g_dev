package handler

import (
	"bytes"
	"encoding/json"
	"g_dev/internal/model"
	"g_dev/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 테스트용 Mock 인벤토리 서비스
type MockInventoryService struct {
	mock.Mock
}

func (m *MockInventoryService) CreateInventory(inventory *model.Inventory) error {
	args := m.Called(inventory)
	return args.Error(0)
}

func (m *MockInventoryService) GetInventoryByID(id uint) (*model.Inventory, error) {
	args := m.Called(id)
	return args.Get(0).(*model.Inventory), args.Error(1)
}

func (m *MockInventoryService) GetUserInventory(userID uint) ([]model.Inventory, error) {
	args := m.Called(userID)
	return args.Get(0).([]model.Inventory), args.Error(1)
}

func (m *MockInventoryService) GetUserInventoryByType(userID uint, itemType string) ([]model.Inventory, error) {
	args := m.Called(userID, itemType)
	return args.Get(0).([]model.Inventory), args.Error(1)
}

func (m *MockInventoryService) GetUserInventoryByRarity(userID uint, rarity string) ([]model.Inventory, error) {
	args := m.Called(userID, rarity)
	return args.Get(0).([]model.Inventory), args.Error(1)
}

func (m *MockInventoryService) UpdateInventory(inventory *model.Inventory) error {
	args := m.Called(inventory)
	return args.Error(0)
}

func (m *MockInventoryService) DeleteInventory(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockInventoryService) AddItemQuantity(userID uint, itemID string, quantity int) error {
	args := m.Called(userID, itemID, quantity)
	return args.Error(0)
}

func (m *MockInventoryService) UseItem(userID uint, itemID string) error {
	args := m.Called(userID, itemID)
	return args.Error(0)
}

func (m *MockInventoryService) GetUserInventoryStats(userID uint) (*service.InventoryStats, error) {
	args := m.Called(userID)
	return args.Get(0).(*service.InventoryStats), args.Error(1)
}

func (m *MockInventoryService) ActivateItem(userID uint, itemID string) error {
	args := m.Called(userID, itemID)
	return args.Error(0)
}

func (m *MockInventoryService) DeactivateItem(userID uint, itemID string) error {
	args := m.Called(userID, itemID)
	return args.Error(0)
}

func (m *MockInventoryService) GetActiveItems(userID uint) ([]model.Inventory, error) {
	args := m.Called(userID)
	return args.Get(0).([]model.Inventory), args.Error(1)
}

// 테스트용 라우터 설정
func setupTestRouter() (*gin.Engine, *MockInventoryService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := &MockInventoryService{}
	handler := NewInventoryHandler(mockService)

	// 인벤토리 라우트 설정
	inventory := router.Group("/inventory")
	{
		inventory.POST("", handler.CreateInventory)
		inventory.GET("/:id", handler.GetInventoryByID)
		inventory.PUT("/:id", handler.UpdateInventory)
		inventory.DELETE("/:id", handler.DeleteInventory)

		userInventory := inventory.Group("/user/:user_id")
		{
			userInventory.GET("", handler.GetUserInventory)
			userInventory.GET("/type", handler.GetUserInventoryByType)
			userInventory.GET("/rarity", handler.GetUserInventoryByRarity)
			userInventory.GET("/stats", handler.GetUserInventoryStats)
			userInventory.GET("/active", handler.GetActiveItems)

			item := userInventory.Group("/item/:item_id")
			{
				item.POST("/add", handler.AddItemQuantity)
				item.POST("/use", handler.UseItem)
				item.POST("/activate", handler.ActivateItem)
				item.POST("/deactivate", handler.DeactivateItem)
			}
		}
	}

	return router, mockService
}

// CreateInventory 핸들러 테스트
func TestInventoryHandler_CreateInventory(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    CreateInventoryRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "정상적인 인벤토리 생성",
			requestBody: CreateInventoryRequest{
				UserID:   1,
				ItemID:   "sword_001",
				ItemName: "강화된 검",
				Quantity: 1,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    5,
				IsActive: false,
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "필수 필드 누락",
			requestBody: CreateInventoryRequest{
				UserID:   1,
				ItemID:   "",
				ItemName: "강화된 검",
				Quantity: 1,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    5,
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "수량이 0인 경우",
			requestBody: CreateInventoryRequest{
				UserID:   1,
				ItemID:   "sword_001",
				ItemName: "강화된 검",
				Quantity: 0,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    5,
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService := setupTestRouter()

			// Mock 설정
			if !tt.expectError {
				mockService.On("CreateInventory", mock.AnythingOfType("*model.Inventory")).Return(nil)
			}

			// 요청 생성
			jsonData, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/inventory", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// 응답 기록
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 결과 검증
			assert.Equal(t, tt.expectedStatus, w.Code, "HTTP 상태 코드가 일치해야 합니다")

			if !tt.expectError {
				var response InventoryResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "응답을 파싱할 수 있어야 합니다")
				assert.Equal(t, tt.requestBody.UserID, response.UserID, "UserID가 일치해야 합니다")
				assert.Equal(t, tt.requestBody.ItemID, response.ItemID, "ItemID가 일치해야 합니다")
			} else {
				var errorResponse ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err, "에러 응답을 파싱할 수 있어야 합니다")
				assert.NotEmptyf(t, errorResponse.Error, "에러 메시지가 있어야 합니다")
			}

			mockService.AssertExpectations(t)
		})
	}
}

// GetInventoryByID 핸들러 테스트
func TestInventoryHandler_GetInventoryByID(t *testing.T) {
	tests := []struct {
		name           string
		inventoryID    string
		mockInventory  *model.Inventory
		mockError      error
		expectedStatus int
	}{
		{
			name:        "정상적인 인벤토리 조회",
			inventoryID: "1",
			mockInventory: &model.Inventory{
				ID:       1,
				UserID:   1,
				ItemID:   "sword_001",
				ItemName: "강화된 검",
				Quantity: 1,
				ItemType: "weapon",
				Rarity:   "rare",
				Level:    5,
				IsActive: false,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "잘못된 ID 형식",
			inventoryID:    "invalid",
			mockInventory:  nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "존재하지 않는 인벤토리",
			inventoryID:    "999",
			mockInventory:  nil,
			mockError:      assert.AnError,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService := setupTestRouter()

			// Mock 설정
			if tt.inventoryID == "1" {
				mockService.On("GetInventoryByID", uint(1)).Return(tt.mockInventory, tt.mockError)
			} else if tt.inventoryID == "999" {
				mockService.On("GetInventoryByID", uint(999)).Return(tt.mockInventory, tt.mockError)
			}

			// 요청 생성
			req, _ := http.NewRequest("GET", "/inventory/"+tt.inventoryID, nil)

			// 응답 기록
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 결과 검증
			assert.Equal(t, tt.expectedStatus, w.Code, "HTTP 상태 코드가 일치해야 합니다")

			if tt.mockInventory != nil && tt.mockError == nil {
				var response InventoryResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "응답을 파싱할 수 있어야 합니다")
				assert.Equal(t, tt.mockInventory.ID, response.ID, "ID가 일치해야 합니다")
				assert.Equal(t, tt.mockInventory.ItemName, response.ItemName, "ItemName이 일치해야 합니다")
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestInventoryHandler_GetUserInventory(t *testing.T) {
	tests := []struct {
		name            string
		userID          string
		mockInventories []model.Inventory
		mockError       error
		expectedStatus  int
	}{
		{
			name:   "정상적인 사용자 인벤토리 조회",
			userID: "1",
			mockInventories: []model.Inventory{
				{
					ID:       1,
					UserID:   1,
					ItemID:   "sword_001",
					ItemName: "강화된 검",
					Quantity: 1,
					ItemType: "weapon",
					Rarity:   "rare",
					Level:    5,
				},
				{
					ID:       2,
					UserID:   1,
					ItemID:   "shield_001",
					ItemName: "강화된 방패",
					Quantity: 1,
					ItemType: "armor",
					Rarity:   "common",
					Level:    3,
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:            "잘못된 사용자 ID 형식",
			userID:          "invalid",
			mockInventories: nil,
			mockError:       nil,
			expectedStatus:  http.StatusBadRequest,
		},
		{
			name:            "빈 인벤토리",
			userID:          "999",
			mockInventories: []model.Inventory{},
			mockError:       nil,
			expectedStatus:  http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService := setupTestRouter()

			// Mock 설정
			if tt.userID == "1" {
				mockService.On("GetUserInventory", uint(1)).Return(tt.mockInventories, tt.mockError)
			} else if tt.userID == "999" {
				mockService.On("GetUserInventory", uint(999)).Return(tt.mockInventories, tt.mockError)
			}

			// 요청 생성
			req, _ := http.NewRequest("GET", "/inventory/user/"+tt.userID, nil)

			// 응답 기록
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 결과 검증
			assert.Equal(t, tt.expectedStatus, w.Code, "HTTP 상태 코드가 일치해야 합니다")

			if tt.mockInventories != nil && tt.mockError == nil {
				var response InventoryListResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "응답을 파싱할 수 있어야 합니다")
				assert.Equal(t, len(tt.mockInventories), response.Total, "총 아이템 수가 일치해야 합니다")
				assert.Equal(t, len(tt.mockInventories), len(response.Inventories), "인벤토리 목록 길이가 일치해야 합니다")
			}

			mockService.AssertExpectations(t)
		})
	}
}

// AddItemQuantity 테스트 수행
func TestInventoryHandler_AddItemQuantity(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		itemID         string
		requestBody    AddItemQuantityRequest
		mockError      error
		expectedStatus int
	}{
		{
			name:   "정상적인 아이템 수량 추가",
			userID: "1",
			itemID: "sword_001",
			requestBody: AddItemQuantityRequest{
				Quantity: 5,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "잘못된 사용자 ID",
			userID: "invalid",
			itemID: "sword_001",
			requestBody: AddItemQuantityRequest{
				Quantity: 5,
			},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "수량이 0인 경우",
			userID: "1",
			itemID: "sword_001",
			requestBody: AddItemQuantityRequest{
				Quantity: 0,
			},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService := setupTestRouter()

			// Mock 설정
			if tt.mockError == nil && tt.expectedStatus == http.StatusOK {
				mockService.On("AddItemQuantity", uint(1), tt.itemID, tt.requestBody.Quantity).Return(tt.mockError)
			}

			// 요청 생성
			jsonData, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/inventory/user/"+tt.userID+"/item/"+tt.itemID+"/add", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// 응답 기록
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 결과 검증
			assert.Equal(t, tt.expectedStatus, w.Code, "HTTP 상태 코드가 일치해야 합니다")

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "응답을 파싱할 수 있어야 합니다")
				assert.Equal(t, "아이템 수량이 성공적으로 추가되었습니다", response["message"], "메시지가 일치해야 합니다")
			}

			mockService.AssertExpectations(t)
		})
	}
}

// UseItem 핸들러를 테스트
func TestInventoryHandler_UseItem(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		itemID         string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "정상적인 아이템 사용",
			userID:         "1",
			itemID:         "sword_001",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "잘못된 사용자 ID",
			userID:         "invalid",
			itemID:         "sword_001",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "아이템 사용 실패",
			userID:         "1",
			itemID:         "nonexistent_item",
			mockError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService := setupTestRouter()

			// Mock 설정
			if tt.userID == "1" {
				if tt.itemID == "sword_001" {
					mockService.On("UseItem", uint(1), "sword_001").Return(tt.mockError)
				} else if tt.itemID == "nonexistent_item" {
					mockService.On("UseItem", uint(1), "nonexistent_item").Return(tt.mockError)
				}
			}

			// 요청 생성
			req, _ := http.NewRequest("POST", "/inventory/user/"+tt.userID+"/item/"+tt.itemID+"/use", nil)

			// 응답 기록
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 결과 검증
			assert.Equal(t, tt.expectedStatus, w.Code, "HTTP 상태 코드가 일치해야 합니다")

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "응답을 파싱할 수 있어야 합니다")
				assert.Equal(t, "아이템이 성공적으로 사용되었습니다", response["message"], "메시지가 일치해야 합니다")
			}

			mockService.AssertExpectations(t)
		})
	}
}

// DeleteInventory 핸들러를 테스트
func TestInventoryHandler_DeleteInventory(t *testing.T) {
	tests := []struct {
		name           string
		inventoryID    string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "정상적인 인벤토리 삭제",
			inventoryID:    "1",
			mockError:      nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "잘못된 ID 형식",
			inventoryID:    "invalid",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "존재하지 않는 인벤토리",
			inventoryID:    "999",
			mockError:      assert.AnError,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService := setupTestRouter()

			// Mock 설정
			if tt.inventoryID == "1" {
				mockService.On("DeleteInventory", uint(1)).Return(tt.mockError)
			} else if tt.inventoryID == "999" {
				mockService.On("DeleteInventory", uint(999)).Return(tt.mockError)
			}

			// 요청 생성
			req, _ := http.NewRequest("DELETE", "/inventory/"+tt.inventoryID, nil)

			// 응답 기록
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 결과 검증
			assert.Equal(t, tt.expectedStatus, w.Code, "HTTP 상태 코드가 일치해야 합니다")

			mockService.AssertExpectations(t)
		})
	}
}

// 통합 테스트를 수행
func TestInventoryHandler_Integration(t *testing.T) {
	t.Run("전체 인벤토리 핸들러 생명주기 테스트", func(t *testing.T) {
		router, mockService := setupTestRouter()

		// 1. 인벤토리 생성
		createRequest := CreateInventoryRequest{
			UserID:   1,
			ItemID:   "test_sword",
			ItemName: "테스트 검",
			Quantity: 1,
			ItemType: "weapon",
			Rarity:   "rare",
			Level:    5,
			IsActive: false,
		}

		mockService.On("CreateInventory", mock.AnythingOfType("*model.Inventory")).Return(nil)

		jsonData, _ := json.Marshal(createRequest)
		req, _ := http.NewRequest("POST", "/inventory", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "인벤토리 생성이 성공해야 합니다")

		// 2. 인벤토리 조회
		mockInventory := &model.Inventory{
			ID:       1,
			UserID:   1,
			ItemID:   "test_sword",
			ItemName: "테스트 검",
			Quantity: 1,
			ItemType: "weapon",
			Rarity:   "rare",
			Level:    5,
			IsActive: false,
		}

		mockService.On("GetInventoryByID", uint(1)).Return(mockInventory, nil)

		req, _ = http.NewRequest("GET", "/inventory/1", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "인벤토리 조회가 성공해야 합니다")

		// 3. 아이템 수량 추가
		addRequest := AddItemQuantityRequest{Quantity: 5}
		mockService.On("AddItemQuantity", uint(1), "test_sword", 5).Return(nil)

		jsonData, _ = json.Marshal(addRequest)
		req, _ = http.NewRequest("POST", "/inventory/user/1/item/test_sword/add", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "아이템 수량 추가가 성공해야 합니다")

		// 4. 아이템 사용
		mockService.On("UseItem", uint(1), "test_sword").Return(nil)

		req, _ = http.NewRequest("POST", "/inventory/user/1/item/test_sword/use", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "아이템 사용이 성공해야 합니다")

		// 5. 인벤토리 삭제
		mockService.On("DeleteInventory", uint(1)).Return(nil)

		req, _ = http.NewRequest("DELETE", "/inventory/1", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code, "인벤토리 삭제가 성공해야 합니다")

		mockService.AssertExpectations(t)
	})
}
