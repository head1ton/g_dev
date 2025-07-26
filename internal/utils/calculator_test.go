package utils

import "testing"

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
