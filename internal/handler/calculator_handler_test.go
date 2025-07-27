package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 계산기 계산 API 테스트
func TestHandleCalculatorCalculate(t *testing.T) {
	handler := NewAPIHandler()

	tests := []struct {
		name           string
		method         string
		requestBody    CalculatorRequest
		expectedStatus int
		expectedResult float64
		expectError    bool
	}{
		{name: "덧셈 계산",
			method: http.MethodPost,
			requestBody: CalculatorRequest{
				Operation: "add",
				Operand1:  10,
				Operand2:  5,
			},
			expectedStatus: http.StatusOK,
			expectedResult: 15,
			expectError:    false,
		},
		{
			name:   "뺄셈 계산",
			method: http.MethodPost,
			requestBody: CalculatorRequest{
				Operation: "subtract",
				Operand1:  10,
				Operand2:  3,
			},
			expectedStatus: http.StatusOK,
			expectedResult: 7,
			expectError:    false,
		},
		{
			name:   "곱셈 계산",
			method: http.MethodPost,
			requestBody: CalculatorRequest{
				Operation: "multiply",
				Operand1:  4,
				Operand2:  6,
			},
			expectedStatus: http.StatusOK,
			expectedResult: 24,
			expectError:    false,
		},
		{
			name:   "나눗셈 계산",
			method: http.MethodPost,
			requestBody: CalculatorRequest{
				Operation: "divide",
				Operand1:  15,
				Operand2:  3,
			},
			expectedStatus: http.StatusOK,
			expectedResult: 5,
			expectError:    false,
		},
		{
			name:   "잘못된 HTTP 메서드",
			method: http.MethodGet,
			requestBody: CalculatorRequest{
				Operation: "add",
				Operand1:  10,
				Operand2:  5,
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
		},
		{
			name:   "잘못된 연산",
			method: http.MethodPost,
			requestBody: CalculatorRequest{
				Operation: "invalid",
				Operand1:  10,
				Operand2:  5,
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:   "0으로 나누기",
			method: http.MethodPost,
			requestBody: CalculatorRequest{
				Operation: "divide",
				Operand1:  10,
				Operand2:  0,
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 요청 본문 생성
			requestBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(tt.method, "/api/calculator/calculate", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// 응답 레코더 생성
			recorder := httptest.NewRecorder()

			// 핸들러 호출
			handler.HandleCalculatorCalculate(recorder, req)

			// 상태 코드 확인
			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			// 성공 케이스인 경우 결과 확인
			if !tt.expectError && tt.expectedStatus == http.StatusOK {
				var response APIResponse
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
					return
				}

				if !response.Success {
					t.Errorf("Expected success response, got error: %s", response.Error)
					return
				}

				// 계산 결과 확인
				if calcData, ok := response.Data.(map[string]interface{}); ok {
					if result, ok := calcData["result"].(float64); ok {
						if result != tt.expectedResult {
							t.Errorf("Expected result %f, got %f", tt.expectedResult, result)
						}
					} else {
						t.Error("Result field not found or not a number")
					}
				} else {
					t.Error("Data field is not a map")
				}
			}
		})
	}
}

// 계산기 히스토리 조회 API 테스트
func TestHandleCalculatorHistory(t *testing.T) {
	handler := NewAPIHandler()

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "정상적인 히스토리 조회",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "잘못된 HTTP 메서드",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/calculator/history", nil)
			recorder := httptest.NewRecorder()

			handler.HandleCalculatorHistory(recorder, req)

			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			if !tt.expectError {
				var response APIResponse
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
					return
				}

				if !response.Success {
					t.Errorf("Expected success response, got error: %s", response.Error)
				}
			}
		})
	}
}

// 계산기 히스토리 초기화 API 테스트
func TestHandleCalculatorHistoryClear(t *testing.T) {
	handler := NewAPIHandler()

	// 먼저 몇 개의 계산을 수행하여 히스토리 생성
	handler.calculator.Add(10, 5)
	handler.calculator.Multiply(3, 4)

	// 히스토리가 있는지 확인
	if len(handler.calculator.GetHistory()) == 0 {
		t.Error("Expected history to have entries before clearing")
	}

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "정상적인 히스토리 초기화",
			method:         http.MethodDelete,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "잘못된 HTTP 메서드",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/calculator/history", nil)
			recorder := httptest.NewRecorder()

			handler.HandleCalculatorHistoryClear(recorder, req)

			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			if !tt.expectError {
				var response APIResponse
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
					return
				}

				if !response.Success {
					t.Errorf("Expected success response, got error: %s", response.Error)
				}

				// 히스토리가 실제로 초기화 되었는지 확인
				if len(handler.calculator.GetHistory()) != 0 {
					t.Error("Expected history to be cleared")
				}
			}
		})
	}
}

// 계산기 통계 조회 API 테스트
func TestHandleCalculatorStats(t *testing.T) {
	handler := NewAPIHandler()

	handler.calculator.Add(10, 5)
	handler.calculator.Multiply(3, 4)
	handler.calculator.Divide(10, 0) // 에러 발생

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "정상적인 통계 조회",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "잘못된 HTTP 메서드",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/calculator/stats", nil)
			recorder := httptest.NewRecorder()

			handler.HandleCalculatorStats(recorder, req)

			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			if !tt.expectError {
				var response APIResponse
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
					return
				}

				if !response.Success {
					t.Errorf("Expected success response, got error: %s", response.Error)
				}

				// 통계 데이터 확인
				if stats, ok := response.Data.(map[string]interface{}); ok {
					if total, ok := stats["total_calculations"].(float64); ok {
						if total < 2 {
							t.Errorf("Expected at least 2 calculations, got %f", total)
						}
					} else {
						t.Error("total_calculations field not found or not a number")
					}
				} else {
					t.Error("Data field is not a map")
				}
			}
		})
	}
}

// 요청 유효성 검사 테스트
func TestValidateCalculatorRequest(t *testing.T) {
	handler := NewAPIHandler()

	tests := []struct {
		name        string
		request     CalculatorRequest
		expectError bool
	}{
		{
			name: "유효한 덧셈 요청",
			request: CalculatorRequest{
				Operation: "add",
				Operand1:  10,
				Operand2:  5,
			},
			expectError: false,
		},
		{
			name: "유효한 나눗셈 요청",
			request: CalculatorRequest{
				Operation: "divide",
				Operand1:  10,
				Operand2:  2,
			},
			expectError: false,
		},
		{
			name: "잘못된 연산",
			request: CalculatorRequest{
				Operation: "invalid",
				Operand1:  10,
				Operand2:  5,
			},
			expectError: true,
		},
		{
			name: "0으로 나누기",
			request: CalculatorRequest{
				Operation: "divide",
				Operand1:  10,
				Operand2:  0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateCalculatorRequest(&tt.request)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// 계산 수행 기능 테스트
func TestPerformCalculation(t *testing.T) {
	handler := NewAPIHandler()

	tests := []struct {
		name           string
		request        CalculatorRequest
		expectedResult float64
		expectError    bool
	}{
		{
			name: "덧셈 계산",
			request: CalculatorRequest{
				Operation: "add",
				Operand1:  10,
				Operand2:  5,
			},
			expectedResult: 15,
			expectError:    false,
		},
		{
			name: "뺄셈 계산",
			request: CalculatorRequest{
				Operation: "subtract",
				Operand1:  10,
				Operand2:  3,
			},
			expectedResult: 7,
			expectError:    false,
		},
		{
			name: "곱셈 계산",
			request: CalculatorRequest{
				Operation: "multiply",
				Operand1:  4,
				Operand2:  6,
			},
			expectedResult: 24,
			expectError:    false,
		},
		{
			name: "나눗셈 계산",
			request: CalculatorRequest{
				Operation: "divide",
				Operand1:  15,
				Operand2:  3,
			},
			expectedResult: 5,
			expectError:    false,
		},
		{
			name: "0으로 나누기",
			request: CalculatorRequest{
				Operation: "divide",
				Operand1:  10,
				Operand2:  0,
			},
			expectError: true,
		},
		{
			name: "알 수 없는 연산",
			request: CalculatorRequest{
				Operation: "unknown",
				Operand1:  10,
				Operand2:  5,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.performCalculation(&tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if result != tt.expectedResult {
					t.Errorf("Expected result %f, got %f", tt.expectedResult, result)
				}
			}
		})
	}
}
