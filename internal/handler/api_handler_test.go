package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// APIHandler 생성자 테스트
func TestNewAPIHandler(t *testing.T) {
	handler := NewAPIHandler()

	// 핸들러가 nil이 아닌지 확인
	if handler == nil {
		t.Error("Expected non-nil APIHandler")
	}

	// 계산기가 초기화되었는지 확인
	if handler.calculator == nil {
		t.Errorf("Expected calculator to be initailized")
	}

	// 파일 처리기가 초기화 되었는지 확인
	if handler.fileProcessor == nil {
		t.Error("Expected fileProcessor to be initialized")
	}

	// 계산기 이름 확인
	if handler.calculator.Name != "API_Calculator" {
		t.Errorf("Expected calculator name 'API_Calculator', got '%s'", handler.calculator.Name)
	}
}

// JSON 응답 작성 기능을 테스트
func TestAPIHandler_writeJSONResponse(t *testing.T) {
	handler := NewAPIHandler()

	testResponse := APIResponse{
		Success: true,
		Message: "Test message",
		Data:    map[string]string{"key": "value"},
		Error:   "",
	}

	// HTTP 응답 레코더 생성
	recorder := httptest.NewRecorder()

	// JSON 응답 작성
	handler.writeJSONResponse(recorder, http.StatusOK, testResponse)

	// 응답 코드 확인
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	// Content-Type 헤더 확인
	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	// 응답 본문 파싱
	var response APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	// 응답 내용 확인
	if response.Success != testResponse.Success {
		t.Errorf("Expected Success %v, got %v", testResponse.Success, response.Success)
	}
	if response.Message != testResponse.Message {
		t.Errorf("Expected Message '%s', got '%s'", testResponse.Message, response.Message)
	}
}

// 에러 응답 작성 기능을 테스트
func TestAPIHandler_writeErrorResponse(t *testing.T) {
	handler := NewAPIHandler()

	// HTTP 응답 레코더 생성
	recorder := httptest.NewRecorder()

	// 에러 응답 작성
	errorMessage := "Test error message"
	handler.writeErrorResponse(recorder, http.StatusBadRequest, errorMessage)

	// 응답 코드 확인
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	// 응답 본문 파싱
	var response APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	// 에러 응답 내용 확인
	if response.Success {
		t.Error("Expected Success to be false for error response")
	}

	if response.Message != errorMessage {
		t.Errorf("Expected Message '%s', got '%s'", errorMessage, response.Message)
	}

	if response.Error != errorMessage {
		t.Errorf("Expected Error '%s', got '%s'", errorMessage, response.Error)
	}
}

// 성공 응답 작성 기능 테스트
func TestAPIHandler_writeSuccessResponse(t *testing.T) {
	handler := NewAPIHandler()

	// HTTP 응답 레코더 생성
	recorder := httptest.NewRecorder()

	// 성공 응답 작성
	testData := map[string]int{"count": 53}
	successMessage := "Operation completed successfully"
	handler.writeSuccessResponse(recorder, testData, successMessage)

	// 응답 코드 확인
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	// 응답 본문 파싱
	var response APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	// 성공 응답 내용 확인
	if !response.Success {
		t.Error("Expected Success to be true for success response")
	}
	if response.Message != successMessage {
		t.Errorf("Expected Message '%s', got '%s'", successMessage, response.Message)
	}
	if response.Error != "" {
		t.Errorf("Expected empty Error, got '%s'", response.Error)
	}

	// 데이터 확인
	if response.Data == nil {
		t.Error("Expected non-nil Data")
	}
}

// APIResponse 구조체의 JSON 태그를 테스트
func TestAPIResponse_JSONTags(t *testing.T) {
	response := APIResponse{
		Success: true,
		Message: "Test message",
		Data:    "test data",
		Error:   "",
	}

	// JSON 마살링
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal response: %v", err)
	}

	// JSON 문자열에 필드가 포함되어 있는지 확인
	jsonStr := string(jsonData)
	expectedFields := []string{"success", "message", "data", "error"}

	for _, field := range expectedFields {
		if !contains(jsonStr, field) {
			t.Errorf("Expected JSON to contain field '%s'", field)
		}
	}
}

// 문자열이 다른 문자열에 포함되어 있는지 확인하는 헬퍼 함수
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
