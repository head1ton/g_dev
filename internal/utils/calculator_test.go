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
