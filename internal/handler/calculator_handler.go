package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// 계산기 계산 API 엔드포인트 처리
// @Summary 계산기 계산 수행
// @Description 두 숫자에 대한 사칙연산을 수행합니다.
// @Tags Calculator
// @Accept json
// @Produce json
// @Param request body CalculatorRequest true "계산 요청"
// @Success 200 {object} APIResponse{data=CalculatorResponse} "계산 성공"
// @Failure 400 {object} APIResponse "잘못된 요청"
// @Failure 405 {object} APIResponse "허용되지 않는 HTTP 메서드"
// @Router /api/calculator/calculate [post]
// POST /api/calculator/calculate
// 요청 예시: {"operation": "add", "operand1": 10, "operand2": 5}
// 응답 예시: {"success": true, "message": "계산 완료", "data": {"result": 15, "operation": "add", "operand1": 10, "operand2": 5}}
func (h *APIHandler) HandleCalculatorCalculate(w http.ResponseWriter, r *http.Request) {
	// HTTP 메서드 검증
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed. Use POST")
		return
	}

	// 요청 본문 파싱
	var request CalculatorRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON request: "+err.Error())
	}

	// 요청 유효성 검사
	if err := h.validateCalculatorRequest(&request); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// 계산 수행
	result, err := h.performCalculation(&request)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Calculation error: "+err.Error())
		return
	}

	// 성공 응답
	response := CalculatorResponse{
		Result:    result,
		Operation: request.Operation,
		Operand1:  request.Operand1,
		Operand2:  request.Operand2,
	}

	h.writeSuccessResponse(w, response, "계산이 완료되었습니다.")
}

func (h *APIHandler) validateCalculatorRequest(request *CalculatorRequest) error {
	// 연산 타입 검사
	validOperations := map[string]bool{
		"add":      true,
		"subtract": true,
		"multiply": true,
		"divide":   true,
	}

	if !validOperations[request.Operation] {
		return fmt.Errorf("Invalid operation '%s'. Valid operations: add, subtract, multiply, divide", request.Operation)
	}

	// 0으로 나누기 검사 (divide 연산인 경우)
	if request.Operation == "divide" && request.Operand2 == 0 {
		return fmt.Errorf("Division by zero is not allowed")
	}

	return nil
}

func (h *APIHandler) performCalculation(request *CalculatorRequest) (float64, error) {
	switch request.Operation {
	case "add":
		return h.calculator.Add(request.Operand1, request.Operand2)
	case "subtract":
		return h.calculator.Subtract(request.Operand1, request.Operand2)
	case "multiply":
		return h.calculator.Multiply(request.Operand1, request.Operand2)
	case "divide":
		return h.calculator.Divide(request.Operand1, request.Operand2)
	default:
		return 0, fmt.Errorf("Unknown operation: %s", request.Operation)
	}
}

// 계산기 히스토리 조회 API 엔드포인트
// @Summary 계산기 히스토리 조회
// @Description 수행된 모든 계산의 히스토리를 조회합니다.
// @Tags Calculator
// @Produce json
// @Success 200 {object} APIResponse{data=[]map[string]interface{}} "히스토리 조회 성공"
// @Failure 405 {object} APIResponse "허용되지 않는 HTTP 메서드"
// @Router /api/calculator/history [get]
// GET /api/calculator/history
// 응답 예시 : {"success": true, "message": "히스토리 조회 완료", "data": [{"operation": "add", "operand1": 10, "operand2": 5, "result": 15}]}
func (h *APIHandler) HandleCalculatorHistory(w http.ResponseWriter, r *http.Request) {
	// HTTP 메서드 검증
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed. Use GET")
		return
	}

	// 히스토리 조회
	history := h.calculator.GetHistory()

	// 히스토리가 비어있는 경우
	if len(history) == 0 {
		h.writeSuccessResponse(w, []interface{}{}, "계산 히스토리가 비어있습니다.")
		return
	}

	// 히스토리 데이터 변환
	historyData := make([]map[string]interface{}, len(history))
	for i, calc := range history {
		historyData[i] = map[string]interface{}{
			"operation": calc.Operation,
			"operand1":  calc.Operand1,
			"operand2":  calc.Operand2,
			"result":    calc.Result,
			"error":     calc.Error,
		}
	}

	h.writeSuccessResponse(w, historyData, "히스토리 조회가 완료되었습니다.")
}

// 계산기 히스토리 초기화 API
// @Summary 계산기 히스토리 초기화
// @Description 모든 계산 히스토리를 삭제합니다.
// @Tags Calculator
// @Produce json
// @Success 200 {object} APIResponse "히스토리 초기화 성공"
// @Failure 405 {object} APIResponse "허용되지 않는 HTTP 메서드"
// @Router /api/calculator/history [delete]
// DELETE /api/calculator/history
// 응답 예시 : {"success": true, "message": "히스토리가 초기화 되었습니다.", "data": null}
func (h *APIHandler) HandleCalculatorHistoryClear(w http.ResponseWriter, r *http.Request) {
	// HTTP 메서드 검증
	if r.Method != http.MethodDelete {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed. Use DELETE")
		return
	}

	// 히스토리 초기화
	h.calculator.ClearHistory()

	h.writeSuccessResponse(w, nil, "히스토리가 초기화되었습니다.")
}

// 계산기 통계 조회 API 엔드 포인트
// @Summary 계산기 통계 조회
// @Description 계산기 사용 통계 정보를 조회합니다.
// @Tags Calculator
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]interface{}} "통계 조회 성공"
// @Failure 405 {object} APIResponse "허용되지 않는 HTTP 메서드"
// @Router /api/calculator/stats [get]
// GET /api/calculator/stats
// 응답 예시 : {"success": true, "message": "통계 조회 완료", "data": {"total_calculations": 10, "error_count": 2}}
func (h *APIHandler) HandleCalculatorStats(w http.ResponseWriter, r *http.Request) {
	// HTTP 메서드 검증
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed. Use GET")
		return
	}

	// 통계 조회
	stats := h.calculator.GetHistorySummary()

	h.writeSuccessResponse(w, stats, "통계 조회가 완료되었습니다.")
}
