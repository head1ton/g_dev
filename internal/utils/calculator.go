package utils

import (
	"errors"
)

type Calculator struct {
	Name    string
	History []Calculation
}

type Calculation struct {
	Operation string  `json:"operation"`
	Operand1  float64 `json:"operand1"`
	Operand2  float64 `json:"operand2"`
	Result    float64 `json:"result"`
	Error     string  `json:"error"`
}

func NewCalculator(name string) *Calculator {
	return &Calculator{
		Name:    name,
		History: make([]Calculation, 0),
	}
}

func (c *Calculator) Add(a, b float64) (float64, error) {
	if a == 0 && b == 0 {
		result := 0.0
		calc := Calculation{
			Operation: "add",
			Operand1:  a,
			Operand2:  b,
			Result:    result,
		}
		c.History = append(c.History, calc)
		return result, nil
	}

	result := a + b

	calc := Calculation{
		Operation: "add",
		Operand1:  a,
		Operand2:  b,
		Result:    result,
	}
	c.History = append(c.History, calc)

	return result, nil
}

func (c *Calculator) Subtract(a, b float64) (float64, error) {
	result := a - b

	calc := Calculation{
		Operation: "subtract",
		Operand1:  a,
		Operand2:  b,
		Result:    result,
	}
	c.History = append(c.History, calc)

	return result, nil
}

func (c *Calculator) Multiply(a, b float64) (float64, error) {
	result := a * b

	calc := Calculation{
		Operation: "multiply",
		Operand1:  a,
		Operand2:  b,
		Result:    result,
	}
	c.History = append(c.History, calc)

	return result, nil
}

func (c *Calculator) Divide(a, b float64) (float64, error) {
	if b == 0 {
		err := errors.New("division by zero is not allowed")
		calc := Calculation{
			Operation: "divide",
			Operand1:  a,
			Operand2:  b,
			Result:    0,
			Error:     err.Error(),
		}
		c.History = append(c.History, calc)
		return 0, err
	}

	result := a / b

	calc := Calculation{
		Operation: "divide",
		Operand1:  a,
		Operand2:  b,
		Result:    result,
	}
	c.History = append(c.History, calc)

	return result, nil
}

// 모든 계산 히스토리 반환
func (c *Calculator) GetHistory() []Calculation {
	history := make([]Calculation, len(c.History))
	copy(history, c.History)
	return history
}

// 계산 히스토리 초기화
func (c *Calculator) ClearHistory() {
	c.History = c.History[:0] // 슬라이스 초기화
}

// 마지막 계산 결과 반환
func (c *Calculator) GetLastCalculation() (*Calculation, error) {
	if len(c.History) == 0 {
		return nil, errors.New("no calculation history available")
	}

	lastCalc := c.History[len(c.History)-1]
	return &lastCalc, nil
}

// 히스토리 통계 정보 반환
func (c *Calculator) GetHistorySummary() map[string]interface{} {
	summary := make(map[string]interface{})

	totalCalculations := len(c.History)
	summary["total_calculations"] = totalCalculations

	if totalCalculations == 0 {
		summary["operations"] = make(map[string]int)
		summary["error_count"] = 0
		return summary
	}

	operations := make(map[string]int)
	errorCount := 0

	for _, calc := range c.History {
		operations[calc.Operation]++
		if calc.Error != "" {
			errorCount++
		}
	}

	summary["operations"] = operations
	summary["error_count"] = errorCount

	return summary
}
