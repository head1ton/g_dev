package handler

import (
	"g_dev/internal/model"
	"g_dev/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type InventoryServiceInterface interface {
	CreateInventory(inventory *model.Inventory) error
	GetInventoryByID(id uint) (*model.Inventory, error)
	GetUserInventory(userID uint) ([]model.Inventory, error)
	GetUserInventoryByType(userID uint, itemType string) ([]model.Inventory, error)
	GetUserInventoryByRarity(userID uint, rarity string) ([]model.Inventory, error)
	UpdateInventory(inventory *model.Inventory) error
	DeleteInventory(id uint) error
	AddItemQuantity(userID uint, itemID string, quantity int) error
	UseItem(userID uint, itemID string) error
	GetUserInventoryStats(userID uint) (*service.InventoryStats, error)
	ActivateItem(userID uint, itemID string) error
	DeactivateItem(userID uint, itemID string) error
	GetActiveItems(userID uint) ([]model.Inventory, error)
}

// 인벤토리 관련 HTTP 요청을 처리하는 핸들러
type InventoryHandler struct {
	inventoryService InventoryServiceInterface
}

// 새로운 InventoryHandler 인스턴스를 생성
func NewInventoryHandler(inventoryService InventoryServiceInterface) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
	}
}

// 인벤토리 생성 요청
type CreateInventoryRequest struct {
	UserID   uint   `json:"user_id" binding:"required"`
	ItemID   string `json:"Item_id" binding:"required"`
	ItemName string `json:"item_name" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
	ItemType string `json:"item_type" binding:"required"`
	Rarity   string `json:"rarity" binding:"required"`
	Level    int    `json:"level" binding:"required,min=1"`
	IsActive bool   `json:"is_active"`
}

// 인벤토리 업데이터 요청
type UpdateInventoryRequest struct {
	ItemName string `json:"item_name"`
	Quantity int    `json:"quantity" binding:"min=1"`
	ItemType string `json:"item_type"`
	Rarity   string `json:"rarity"`
	Level    int    `json:"level" binding:"min=1"`
	IsActive bool   `json:"is_active"`
}

// 아이템 수량 추가 요청
type AddItemQuantityRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

// 인벤토리 응답
type InventoryResponse struct {
	ID          uint   `json:"id"`
	UserID      uint   `json:"user_id"`
	ItemID      string `json:"item_id"`
	ItemName    string `json:"item_name"`
	Quantity    int    `json:"quantity"`
	ItemType    string `json:"item_type"`
	Rarity      string `json:"rarity"`
	Level       int    `json:"level"`
	IsActive    bool   `json:"is_active"`
	RarityColor string `json:"rarity_color"`
}

// 인벤토리 목록 응답
type InventoryListResponse struct {
	Inventories []InventoryResponse `json:"inventories"`
	Total       int                 `json:"total"`
}

// 에러 응답
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// 성공 응답
type SuccessResponse struct {
	Message  string      `json:"message"`
	UserID   uint        `json:"user_id,omitempty"`
	ItemID   string      `json:"item_id,omitempty"`
	Quantity int         `json:"quantity,omitempty"`
	Data     interface{} `json:"data,omitempty"`
}

// 새로운 인벤토리 아이템을 생성
// @Summary 인벤토리 아이템 생성
// @Description 새로운 인벤토리 아이템을 생성합니다.
// @Tags Inventory
// @Accept json
// @Produce json
// @Param inventory body CreateInventoryRequest true "인벤토리 정보"
// @Success 201 {object} InventoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /inventory [post]
func (h *InventoryHandler) CreateInventory(c *gin.Context) {
	var req CreateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "잘못된 요청 형식입니다",
			Message: err.Error(),
		})
		return
	}

	inventory := &model.Inventory{
		UserID:   req.UserID,
		ItemID:   req.ItemID,
		ItemName: req.ItemName,
		Quantity: req.Quantity,
		ItemType: req.ItemType,
		Rarity:   req.Rarity,
		Level:    req.Level,
		IsActive: req.IsActive,
	}

	if err := h.inventoryService.CreateInventory(inventory); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "인벤토리 생성에 실패했습니다",
			Message: err.Error(),
		})
		return
	}

	response := InventoryResponse{
		ID:          inventory.ID,
		UserID:      inventory.UserID,
		ItemID:      inventory.ItemID,
		ItemName:    inventory.ItemName,
		Quantity:    inventory.Quantity,
		ItemType:    inventory.ItemType,
		Rarity:      inventory.Rarity,
		Level:       inventory.Level,
		IsActive:    inventory.IsActive,
		RarityColor: inventory.GetRarityColor(),
	}

	c.JSON(http.StatusCreated, response)
}

// ID로 인벤토리 아이템을 조회
// @Summary 인벤토리 아이템 조회
// @Description ID로 인벤토리 아이템을 조회합니다.
// @Tags Inventory
// @Accept json
// @Produce json
// @Param id path int true "인벤토리 ID"
// @Success 200 {object} InventoryResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /inventory/{id} [get]
func (h *InventoryHandler) GetInventoryByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "잘못된 ID 형식입니다",
			Message: err.Error(),
		})
		return
	}

	inventory, err := h.inventoryService.GetInventoryByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "인벤토리 아이템을 찾을 수 없습니다",
			Message: err.Error(),
		})
		return
	}

	response := InventoryResponse{
		ID:          inventory.ID,
		UserID:      inventory.UserID,
		ItemID:      inventory.ItemID,
		ItemName:    inventory.ItemName,
		Quantity:    inventory.Quantity,
		ItemType:    inventory.ItemType,
		Rarity:      inventory.Rarity,
		Level:       inventory.Level,
		IsActive:    inventory.IsActive,
		RarityColor: inventory.GetRarityColor(),
	}

	c.JSON(http.StatusOK, response)
}

// 사용자의 모든 인벤토리 아이템을 조회
// @Summary 사용자 인벤토리 조회
// @Description 사용자의 모든 인벤토리 아이템을 조회합니다.
// @Tags Inventory
// @Accept json
// @Produce json
// @Param user_id path int true "사용자 ID"
// @Success 200 {object} InventoryListResponse
// @Failure 500 {object} ErrorResponse
// @Router /inventory/user/{user_id} [get]
func (h *InventoryHandler) GetUserInventory(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "잘못된 사용자 ID 형식입니다",
			Message: err.Error(),
		})
		return
	}

	inventories, err := h.inventoryService.GetUserInventory(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "사용자 인벤토리 조회에 실패했습니다",
			Message: err.Error(),
		})
		return
	}

	response := InventoryListResponse{
		Inventories: make([]InventoryResponse, len(inventories)),
		Total:       len(inventories),
	}

	for i, inventory := range inventories {
		response.Inventories[i] = InventoryResponse{
			ID:          inventory.ID,
			UserID:      inventory.UserID,
			ItemID:      inventory.ItemID,
			ItemName:    inventory.ItemName,
			Quantity:    inventory.Quantity,
			ItemType:    inventory.ItemType,
			Rarity:      inventory.Rarity,
			Level:       inventory.Level,
			IsActive:    inventory.IsActive,
			RarityColor: inventory.GetRarityColor(),
		}
	}

	c.JSON(http.StatusOK, response)
}

// 사용자의 특정 타입 인벤토리 아이템을 조회
// @Summary 사용자 인벤토리 타입별 조회
// @Description 사용자의 특정 타입 인벤토리 아이템을 조회합니다.
// @Tags Inventory
// @Accept json
// @Produce json
// @Param user_id path int true "사용자 ID"
// @Param type query string true "아이템 타입"
// @Success 200 {object} InventoryListResponse
// @Failure 500 {object} ErrorResponse
// @Router /inventory/user/{user_id}/type [get]
func (h *InventoryHandler) GetUserInventoryByType(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "잘못된 사용자 ID 형식입니다",
			Message: err.Error(),
		})
		return
	}

	itemType := c.Query("type")
	if itemType == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "아이템 타입이 필요합니다",
			Message: "type 쿼리 파라미터가 필요합니다",
		})
		return
	}

	inventories, err := h.inventoryService.GetUserInventoryByType(uint(userID), itemType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "사용자 인벤토리 타입별 조회에 실패했습니다",
			Message: err.Error(),
		})
		return
	}

	response := InventoryListResponse{
		Inventories: make([]InventoryResponse, len(inventories)),
		Total:       len(inventories),
	}

	for i, inventory := range inventories {
		response.Inventories[i] = InventoryResponse{
			ID:          inventory.ID,
			UserID:      inventory.UserID,
			ItemID:      inventory.ItemID,
			ItemName:    inventory.ItemName,
			Quantity:    inventory.Quantity,
			ItemType:    inventory.ItemType,
			Rarity:      inventory.Rarity,
			Level:       inventory.Level,
			IsActive:    inventory.IsActive,
			RarityColor: inventory.GetRarityColor(),
		}
	}

	c.JSON(http.StatusOK, response)
}

// 사용자의 특정 등급 인벤토리 아이템을 조회
// @Summary 사용자 인벤토리 등급별 조회
// @Description 사용자의 특정 등급 인벤토리 아이템을 조회합니다.
// @Tags Inventory
// @Accept json
// @Produce json
// @Param user_id path int true "사용자 ID"
// @Param rarity query string true "아이템 등급"
// @Success 200 {object} InventoryListResponse
// @Failure 500 {object} ErrorResponse
// @Router /inventory/user/{user_id}/rarity [get]
func (h *InventoryHandler) GetUserInventoryByRarity(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "잘못된 사용자 ID 형식입니다",
			Message: err.Error(),
		})
		return
	}

	rarity := c.Query("rarity")
	if rarity == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "아이템 등급이 필요합니다",
			Message: "rarity 쿼리 파라미터가 필요합니다",
		})
		return
	}

	inventories, err := h.inventoryService.GetUserInventoryByRarity(uint(userID), rarity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "사용자 인벤토리 등급별 조회에 실패했습니다",
			Message: err.Error(),
		})
		return
	}

	response := InventoryListResponse{
		Inventories: make([]InventoryResponse, len(inventories)),
		Total:       len(inventories),
	}

	for i, inventory := range inventories {
		response.Inventories[i] = InventoryResponse{
			ID:          inventory.ID,
			UserID:      inventory.UserID,
			ItemID:      inventory.ItemID,
			ItemName:    inventory.ItemName,
			Quantity:    inventory.Quantity,
			ItemType:    inventory.ItemType,
			Rarity:      inventory.Rarity,
			Level:       inventory.Level,
			IsActive:    inventory.IsActive,
			RarityColor: inventory.GetRarityColor(),
		}
	}

	c.JSON(http.StatusOK, response)
}

// 인벤토리 아이템을 업데이트
// @Summary 인벤토리 아이템 업데이트
// @Description 인벤토리 아이템을 업데이트합니다.
// @Tags Inventory
// @Accept json
// @Produce json
// @Param id path int true "인벤토리 ID"
// @Param inventory body UpdateInventoryRequest true "업데이트할 인벤토리 정보"
// @Success 200 {object} InventoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /inventory/{id} [put]
func (h *InventoryHandler) UpdateInventory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "잘못된 ID 형식입니다",
			Message: err.Error(),
		})
		return
	}

	var req UpdateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "잘못된 요청 형식입니다",
			Message: err.Error(),
		})
		return
	}

	// 기존 인벤토리 조회
	existingInventory, err := h.inventoryService.GetInventoryByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "업데이트할 인벤토리 아이템을 찾을 수 없습니다",
			Message: err.Error(),
		})
		return
	}

	// 업데이트할 필드만 변경
	if req.ItemName != "" {
		existingInventory.ItemName = req.ItemName
	}
	if req.Quantity > 0 {
		existingInventory.Quantity = req.Quantity
	}
	if req.ItemType != "" {
		existingInventory.ItemType = req.ItemType
	}
	if req.Rarity != "" {
		existingInventory.Rarity = req.Rarity
	}
	if req.Level > 0 {
		existingInventory.Level = req.Level
	}
	existingInventory.IsActive = req.IsActive

	if err := h.inventoryService.UpdateInventory(existingInventory); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "인벤토리 업데이트에 실패했습니다",
			Message: err.Error(),
		})
		return
	}

	response := InventoryResponse{
		ID:          existingInventory.ID,
		UserID:      existingInventory.UserID,
		ItemID:      existingInventory.ItemID,
		ItemName:    existingInventory.ItemName,
		Quantity:    existingInventory.Quantity,
		ItemType:    existingInventory.ItemType,
		Rarity:      existingInventory.Rarity,
		Level:       existingInventory.Level,
		IsActive:    existingInventory.IsActive,
		RarityColor: existingInventory.GetRarityColor(),
	}

	c.JSON(http.StatusOK, response)
}

// 인벤토리 아이템을 삭제
// @Summary 인벤토리 아이템 삭제
// @Description 인벤토리 아이템을 삭제합니다.
// @Tags Inventory
// @Accept json
// @Produce json
// @Param id path int true "인벤토리 ID"
// @Success 204 "No Content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /inventory/{id} [delete]
func (h *InventoryHandler) DeleteInventory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "잘못된 ID 형식입니다",
			Message: err.Error(),
		})
		return
	}

	if err := h.inventoryService.DeleteInventory(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "삭제할 인벤토리 아이템을 찾을 수 없습니다",
			Message: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// 특정 아이템의 수량을 증가
// @Summary 아이템 수량 추가
// @Description 특정 아이템의 수량을 증가시킵니다.
// @Tags Inventory
// @Accept json
// @Produce json
// @Param user_id path int true "사용자 ID"
// @Param item_id path string true "아이템 ID"
// @Param request body AddItemQuantityRequest true "추가할 수량"
// @Success 200 {object} InventoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /inventory/user/{user_id}/item/{item_id}/add [post]
func (h *InventoryHandler) AddItemQuantity(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "잘못된 사용자 ID 형식입니다",
			Message: err.Error(),
		})
		return
	}

	itemID := c.Param("item_id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "아이템 ID가 필요합니다",
			Message: "item_id 파라미터가 필요합니다",
		})
		return
	}

	var req AddItemQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "잘못된 요청 형식입니다",
			Message: err.Error(),
		})
		return
	}

	if err := h.inventoryService.AddItemQuantity(uint(userID), itemID, req.Quantity); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "아이템 수량 추가에 실패했습니다",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message:  "아이템 수량이 성공적으로 추가되었습니다",
		UserID:   uint(userID),
		ItemID:   itemID,
		Quantity: req.Quantity,
	})
}

// 특정 아이템을 사용
// @Summary 아이템 사용
// @Description 특정 아이템을 사용합니다.
// @Tags inventory
// @Accept json
// @Produce json
// @Param user_id path int true "사용자 ID"
// @Param item_id path string true "아이템 ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /inventory/user/{user_id}/item/{item_id}/use [post]
func (h *InventoryHandler) UseItem(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "잘못된 사용자 ID 형식입니다",
			Message: err.Error(),
		})
		return
	}

	itemID := c.Param("item_id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "아이템 ID가 필요합니다",
			Message: "item_id 파라미터가 필요합니다",
		})
		return
	}

	if err := h.inventoryService.UseItem(uint(userID), itemID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "아이템 사용에 실패했습니다",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "아이템이 성공적으로 사용되었습니다",
		UserID:  uint(userID),
		ItemID:  itemID,
	})
}

// 사용자의 인벤토리 통계를 반환
// @Summary 사용자 인벤토리 통계
// @Description 사용자의 인벤토리 통계를 반환합니다.
// @Tags Inventory
// @Accept json
// @Produce json
// @Param user_id path int true "사용자 ID"
// @Success 200 {object} service.InventoryStats
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /inventory/user/{user_id}/stats [get]
func (h *InventoryHandler) GetUserInventoryStats(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "잘못된 사용자 ID 형식입니다",
			Message: err.Error(),
		})
		return
	}

	stats, err := h.inventoryService.GetUserInventoryStats(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "인벤토리 통계 조회에 실패했습니다",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// 아이템을 활성화
// @Summary 아이템 활성화
// @Description 아이템을 활성화합니다.
// @Tags Inventory
// @Accept json
// @Produce json
// @Param user_id path int true "사용자 ID"
// @Param item_id path string true "아이템 ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /inventory/user/{user_id}/item/{item_id}/activate [post]
func (h *InventoryHandler) ActivateItem(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "잘못된 사용자 ID 형식입니다",
			Message: err.Error(),
		})
		return
	}

	itemID := c.Param("item_id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "아이템 ID가 필요합니다",
			Message: "item_id 파라미터가 필요합니다",
		})
		return
	}

	if err := h.inventoryService.ActivateItem(uint(userID), itemID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "아이템 활성화에 실패했습니다",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "아이템이 성공적으로 활성화되었습니다",
		UserID:  uint(userID),
		ItemID:  itemID,
	})
}

// 아이템을 비활성화
// @Summary 아이템 비활성화
// @Description 아이템을 비활성화합니다.
// @Tags Inventory
// @Accept json
// @Produce json
// @Param user_id path int true "사용자 ID"
// @Param item_id path string true "아이템 ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /inventory/user/{user_id}/item/{item_id}/deactivate [post]
func (h *InventoryHandler) DeactivateItem(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "잘못된 사용자 ID 형식입니다",
			Message: err.Error(),
		})
		return
	}

	itemID := c.Param("item_id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "아이템 ID가 필요합니다",
			Message: "item_id 파라미터가 필요합니다",
		})
		return
	}

	if err := h.inventoryService.DeactivateItem(uint(userID), itemID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "아이템 비활성화에 실패했습니다",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "아이템이 성공적으로 비활성화되었습니다",
		UserID:  uint(userID),
		ItemID:  itemID,
	})
}

// 사용자의 활성화된 아이템들을 조회
// @Summary 활성화된 아이템 조회
// @Description 사용자의 활성화된 아이템들을 조회합니다.
// @Tags Inventory
// @Accept json
// @Produce json
// @Param user_id path int true "사용자 ID"
// @Success 200 {object} InventoryListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /inventory/user/{user_id}/active [get]
func (h *InventoryHandler) GetActiveItems(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "잘못된 사용자 ID 형식입니다",
			Message: err.Error(),
		})
		return
	}

	inventories, err := h.inventoryService.GetActiveItems(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "활성화된 아이템 조회에 실패했습니다",
			Message: err.Error(),
		})
		return
	}

	response := InventoryListResponse{
		Inventories: make([]InventoryResponse, len(inventories)),
		Total:       len(inventories),
	}

	for i, inventory := range inventories {
		response.Inventories[i] = InventoryResponse{
			ID:          inventory.ID,
			UserID:      inventory.UserID,
			ItemID:      inventory.ItemID,
			ItemName:    inventory.ItemName,
			Quantity:    inventory.Quantity,
			ItemType:    inventory.ItemType,
			Rarity:      inventory.Rarity,
			Level:       inventory.Level,
			IsActive:    inventory.IsActive,
			RarityColor: inventory.GetRarityColor(),
		}
	}

	c.JSON(http.StatusOK, response)
}
