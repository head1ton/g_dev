package handler

import (
	"encoding/json"
	"g_dev/internal/utils"
	"net/http"
)

// 계산기와 파일 처리 기능을 웹 API
type APIHandler struct {
	// 계산기 인스턴스 (API 요청별로 독립적인 계산기 사용)
	calculator *utils.Calculator
	// 파일 처리기 인스턴스
	fileProcessor *utils.FileProcessor
}

// APIResponse는 API 응답의 공통 구조체
type APIResponse struct {
	Success bool        `json:"success"` // 요청 성공 여부
	Message string      `json:"message"` // 응답 메시지
	Data    interface{} `json:"data"`    // 응답 데이터
	Error   string      `json:"error"`   // 에러 메시지
}

// 계산기 API 요청 구조체
type CalculationRequest struct {
	Operation string  `json:"operation"` // 수행할 연산 (add, subtract, multiply, divide)
	Operand1  float64 `json:"operand1"`  // 첫번째 피연산자
	Operand2  float64 `json:"operand2"`  // 두번째 피연산자
}

// 계산기 API 응답 구조체
type CalculationResponse struct {
	Result    float64 `json:"result"`    // 계산 결과
	Operation string  `json:"operation"` // 수행된 연산
	Operand1  float64 `json:"operand1"`  // 첫 번째 피연산자
	Operand2  float64 `json:"operand2"`  // 두 번째 피연산자
}

// 파일 검색 API 요청 구조체
type FileSearchRequest struct {
	Pattern       string `json:"pattern"`        // 검색 패턴 (와일드카드)
	RegexPattern  string `json:"regex_pattern"`  // 정규표현식 패턴
	Extension     string `json:"extension"`      // 확장자
	Content       string `json:"content"`        // 파일 내용 검색 텍스트
	CaseSensitive bool   `json:"case_sensitive"` // 대소문자 구분 여부
}

// 새로운 APIHandler 인스턴스 생성
func NewAPIHandler() *APIHandler {
	return &APIHandler{
		calculator:    utils.NewCalculator("API_Calculator"),
		fileProcessor: utils.NewFileProcessor("."),
	}
}

// JSON 응답을 작성하는 헬퍼 메서드. 모든 API 핸들러에서 공통으로 사용.
func (h *APIHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, response APIResponse) {
	// 응답 헤더 설정
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// JSON 인코딩 및 응답 작성
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// JSON 인코딩 실패 시 기본 에러 응답
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// 에러 응답을 작성하는 헬퍼 메서드
func (h *APIHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := APIResponse{
		Success: false,
		Message: message,
		Error:   message,
	}
	h.writeJSONResponse(w, statusCode, response)
}

// 성공 응답을 작성하는 헬퍼 메서드
func (h *APIHandler) writeSuccessResponse(w http.ResponseWriter, data interface{}, message string) {
	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	h.writeJSONResponse(w, http.StatusOK, response)
}
