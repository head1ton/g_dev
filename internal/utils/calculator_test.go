package utils

import (
	"testing"
)

// TestNewCalculator는 Calculator 생성자 함수를 테스트
func TestNewCalculator(t *testing.T) {
	// test case 1: 정상
	calc := NewCalculator("TestCalc")

	// 구조체 필드 검증
	if calc.Name != "TestCalc" {
		t.Errorf("Expected name 'TestCalc', got '%s'", calc.Name)
	}

	// 히스토리 슬라이스가 초기화되었는지 확인
	if calc.History == nil {
		t.Error("History slice should be initialized, not nil")
	}

	if len(calc.History) != 0 {
		t.Errorf("Expected empty history, got length %d", len(calc.History))
	}

	// test case 2: 빈 이름으로 생성
	calc2 := NewCalculator("")
	if calc2.Name != "" {
		t.Errorf("Expected empty name, got '%s'", calc2.Name)
	}
}

func TestCalculator_Add(t *testing.T) {
	calc := NewCalculator("AddTest")

	testCases := []struct {
		name     string
		a, b     float64
		expected float64
	}{
		{"positive numbers", 5.0, 3.0, 8.0},
		{"negative numbers", -5.0, -3.0, -8.0},
		{"mixed numbers", 5.0, -3.0, 2.0},
		{"zero values", 0.0, 0.0, 0.0},
		{"large numbers", 1e10, 2e10, 3e10},
		{"decimal numbers", 3.14, 2.86, 6.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := calc.Add(tc.a, tc.b)

			if err != nil {
				t.Errorf("Unexpected error for case '%s': %v", tc.name, err)
			}

			if result != tc.expected {
				t.Errorf("Expected %f, got %f for case '%s'", tc.expected, result, tc.name)
			}
		})
	}

	history := calc.History
	if len(history) != len(testCases) {
		t.Errorf("Expected %d history entries, got %d", len(testCases), len(history))
	}

	lastCalc := history[len(history)-1]
	if lastCalc.Operation != "add" {
		t.Errorf("Expected operation 'add', got '%s'", lastCalc.Operation)
	}
	if lastCalc.Operand1 != 3.14 || lastCalc.Operand2 != 2.86 {
		t.Errorf("Expected operands 3.14, 2.86, got %f, %f", lastCalc.Operand1, lastCalc.Operand2)
	}
}

func TestCalculator_Subtract(t *testing.T) {
	calc := NewCalculator("SubtractTest")

	testCases := []struct {
		name     string
		a, b     float64
		expected float64
	}{
		{"positive result", 5.0, 3.0, 2.0},
		{"negative result", 3.0, 5.0, -2.0},
		{"zero result", 5.0, 5.0, 0.0},
		{"negative numbers", -5.0, -3.0, -2.0},
		{"large numbers", 1e10, 5e9, 5e9},
		{"decimal numbers", 3.14, 1.14, 2.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := calc.Subtract(tc.a, tc.b)

			if err != nil {
				t.Errorf("Unexpected error for case '%s': %v", tc.name, err)
			}

			if result != tc.expected {
				t.Errorf("Expected %f, got %f for case '%s'", tc.expected, result, tc.name)
			}
		})
	}

	history := calc.History
	if len(history) != len(testCases) {
		t.Errorf("Expected %d history entries, got %d", len(testCases), len(history))
	}

	lastCalc := history[len(history)-1]
	if lastCalc.Operation != "subtract" {
		t.Errorf("Expected operation 'subtract', got '%s'", lastCalc.Operation)
	}
}

func TestCalculator_Multiply(t *testing.T) {
	calc := NewCalculator("MultiplyTest")

	testCases := []struct {
		name     string
		a, b     float64
		expected float64
	}{
		{"positive numbers", 5.0, 3.0, 15.0},
		{"negative numbers", -5.0, -3.0, 15.0},
		{"mixed numbers", 5.0, -3.0, -15.0},
		{"zero multiplication", 5.0, 0.0, 0.0},
		{"large numbers", 1e6, 2e6, 2e12},
		{"decimal numbers", 3.14, 2.0, 6.28},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := calc.Multiply(tc.a, tc.b)

			if err != nil {
				t.Errorf("Unexpected error for case '%s': %v", tc.name, err)
			}

			if result != tc.expected {
				t.Errorf("Expected %f, got %f for case '%s'", tc.expected, result, tc.name)
			}
		})
	}

	history := calc.History
	if len(history) != len(testCases) {
		t.Errorf("Expected %d history entries, got %d", len(testCases), len(history))
	}

	lastCalc := history[len(history)-1]
	if lastCalc.Operation != "multiply" {
		t.Errorf("Expected operation 'multiply', got '%s'", lastCalc.Operation)
	}
}

func TestCalculator_Divide(t *testing.T) {
	calc := NewCalculator("DivideTest")

	testCases := []struct {
		name     string
		a, b     float64
		expected float64
		hasError bool
	}{
		{"normal division", 10.0, 2.0, 5.0, false},
		{"decimal result", 10.0, 3.0, 10.0 / 3.0, false},
		{"negative numbers", -10.0, -2.0, 5.0, false},
		{"mixed numbers", -10.0, 2.0, -5.0, false},
		{"zero dividend", 0.0, 5.0, 0.0, false},
		{"division by zero", 10.0, 0.0, 0.0, true},
		{"large numbers", 1e10, 1e5, 1e5, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := calc.Divide(tc.a, tc.b)

			if tc.hasError {
				if err == nil {
					t.Errorf("Expected error for case '%s', but got none", tc.name)
				}

				if tc.name == "division by zero" {
					if err.Error() != "division by zero is not allowed" {
						t.Errorf("Expected 'division by zero' error, got '%s'", err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for case '%s': %v", tc.name, err)
				}

				if result != tc.expected {
					t.Errorf("Expected %f, got %f for case '%s", tc.expected, result, tc.name)
				}
			}
		})
	}

	history := calc.History
	if len(history) != len(testCases) {
		t.Errorf("Expected %d history entries, got %d", len(testCases), len(history))
	}

	lastCalc := history[len(history)-1]
	if lastCalc.Operation != "divide" {
		t.Errorf("Expected operation 'divide', got '%s'", lastCalc.Operation)
	}
}

func TestCalculator_GetHistory(t *testing.T) {
	calc := NewCalculator("HistoryTest")

	// 초기 히스토리는 비어있어야 함
	history := calc.GetHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d entries", len(history))
	}

	calc.Add(5, 3)
	calc.Subtract(10, 4)

	history = calc.GetHistory()
	if len(history) != 2 {
		t.Errorf("Expected 2 history entries, got %d", len(history))
	}

	if history[0].Operation != "add" {
		t.Errorf("Expected operation 'add', got '%s'", history[0].Operation)
	}

	if history[1].Operation != "subtract" {
		t.Errorf("Expected operation 'subtract', got '%s'", history[1].Operation)
	}
}

// 히스토리 초기화 기능 테스트
func TestCalculator_ClearHistory(t *testing.T) {
	calc := NewCalculator("ClearHistoryTest")

	calc.Add(1, 2)
	calc.Multiply(3, 4)

	if len(calc.History) != 2 {
		t.Errorf("Expected 2 history entries before clear, got %d", len(calc.History))
	}

	calc.ClearHistory()

	if len(calc.History) != 0 {
		t.Errorf("Expected empty history after clear, got %d entries", len(calc.History))
	}
}

// 마지막 계산 조회 기능 테스트
func TestCalculator_GetLastCalculation(t *testing.T) {
	calc := NewCalculator("LastCalcTest")

	lastCalc, err := calc.GetLastCalculation()
	if err == nil {
		t.Error("Expected error for empty history, but got none")
	}

	if lastCalc != nil {
		t.Error("Expected nil result for empty history")
	}

	calc.Add(10, 5)
	calc.Multiply(3, 7)

	lastCalc, err = calc.GetLastCalculation()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if lastCalc == nil {
		t.Error("Expected last calculation, got nil")
	}

	// 마지막 계산 결과 확인
	if lastCalc.Operation != "multiply" {
		t.Errorf("Expected last operation 'multiply', got '%s'", lastCalc.Operation)
	}
	if lastCalc.Result != 21.0 {
		t.Errorf("Expected last result 21.0, got %f", lastCalc.Result)
	}
}

func TestCalculator_GetHistorySummary(t *testing.T) {
	calc := NewCalculator("SummaryTest")

	summary := calc.GetHistorySummary()

	if summary["total_calculations"] != 0 {
		t.Errorf("Expected 0 total calculations, got %v", summary["total_calculations"])
	}

	operations, ok := summary["operations"].(map[string]int)
	if !ok {
		t.Error("Expected operations to be map[string]int")
	}
	if len(operations) != 0 {
		t.Errorf("Expected empty operations map, got %d entries", len(operations))
	}

	calc.Add(1, 2)
	calc.Subtract(5, 3)
	calc.Multiply(2, 4)
	calc.Divide(10, 2)
	calc.Divide(5, 0)

	summary = calc.GetHistorySummary()

	if summary["total_calculations"] != 5 {
		t.Errorf("Expected 5 total calculations, got %v", summary["total_calculations"])
	}

	if summary["error_count"] != 1 {
		t.Errorf("Expected 1 error, got %v", summary["error_count"])
	}

	operations, ok = summary["operations"].(map[string]int)
	if !ok {
		t.Error("Expected operations to be map[string]int")
	}

	expectedOps := map[string]int{
		"add":      1,
		"subtract": 1,
		"multiply": 1,
		"divide":   2,
	}

	for op, expected := range expectedOps {
		if operations[op] != expected {
			t.Errorf("Expected %d '%s' operations, got %d", expected, op, operations[op])
		}
	}
}

func TestCalculator_Integration(t *testing.T) {
	calc := NewCalculator("IntegrationTest")

	t.Run("complex calculation sequence", func(t *testing.T) {
		// 1. 기본 사칙연산 수행
		result1, err := calc.Add(10, 5)
		if err != nil || result1 != 15 {
			t.Errorf("Add failed: expected 15, got %f, err %v", result1, err)
		}

		result2, err := calc.Subtract(result1, 3)
		if err != nil || result2 != 12 {
			t.Errorf("Subtract failed: expected 12, got %f, err %v", result2, err)
		}

		result3, err := calc.Multiply(result2, 2)
		if err != nil || result3 != 24 {
			t.Errorf("Multiply failed: expected 24, got %f, err %v", result3, err)
		}

		result4, err := calc.Divide(result3, 4)
		if err != nil || result4 != 6 {
			t.Errorf("Divide failed: expected 6, got %f, err %v", result4, err)
		}

		// 2. 히스토리 검증
		history := calc.GetHistory()
		if len(history) != 4 {
			t.Errorf("Expected 4 history entries, got %d", len(history))
		}

		// 3. 마지막 계산 확인
		lastCalc, err := calc.GetLastCalculation()
		if err != nil || lastCalc.Result != 6 {
			t.Errorf("Last calculation failed: expected 6, got %f, err: %v", lastCalc.Result, err)
		}
	})

	t.Run("error handling and recovery", func(t *testing.T) {
		// 0으로 나누기 시도
		_, err := calc.Divide(10, 0)
		if err == nil {
			t.Error("Expected error for division by zero")
		}

		// 에러 후에도 정상 계산 가능
		result, err := calc.Add(5, 3)
		if err != nil || result != 8 {
			t.Errorf("Add after error failed: expected 8, got %f, err: %v", result, err)
		}

		// 통계 확인
		summary := calc.GetHistorySummary()
		if summary["error_count"] != 1 {
			t.Errorf("Expected 1 error, got %v", summary["error_count"])
		}
	})

	t.Run("history management", func(t *testing.T) {
		// 히스토리 초기화
		calc.ClearHistory()

		// 초기화 후 히스토리 확이
		history := calc.GetHistory()
		if len(history) != 0 {
			t.Errorf("Expected empty history after clear, got %d entries", len(history))
		}

		// 초기화 후 마지막 계산 조회 시 에러
		_, err := calc.GetLastCalculation()
		if err == nil {
			t.Error("Expected error when getting last calculation from empty history")
		}
	})
}

// 덧셈 연산의 성능을 측정
func BenchmarkCalculator_Add(b *testing.B) {
	calc := NewCalculator("BenchmarkTest")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		calc.Add(float64(i), float64(i+1))
	}
}

// 히스토리 관리 성능을 측정
func BenchmarkCalculator_History(b *testing.B) {
	calc := NewCalculator("HistoryBenchmarkTest")

	for i := 0; i < 1000; i++ {
		calc.Add(float64(i), float64(i+1))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		calc.GetHistory()
	}
}

// String 메서드 테스트
func TestCalculator_String(t *testing.T) {
	calc1 := NewCalculator("EmptyCalc")
	expected1 := "Calculator 'EmptyCalc' (no calculations performed)"
	if calc1.String() != expected1 {
		t.Errorf("Expected '%s', got '%s'", expected1, calc1.String())
	}

	calc2 := NewCalculator("TestCalc")
	calc2.Add(1, 2)
	calc2.Subtract(5, 3)
	calc2.Divide(10, 0) // 에러 발생

	expected2 := "Calculator 'TestCalc' (3 calculations, 1 errors)"
	if calc2.String() != expected2 {
		t.Errorf("Expected '%s', got '%s'", expected2, calc2.String())
	}
}
