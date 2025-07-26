package utils

type Calculator struct {
	Name    string
	History []Calculator
}

type Calculation struct {
	Operation string  `json:"operation"`
	Operand1  float64 `json:"operand1"`
	Operand2  float64 `json:"operand2"`
	Result    float64 `json:"result"`
	Error     error   `json:"error"`
}

func NewCalculator(name string) *Calculator {
	return &Calculator{
		Name:    name,
		History: make([]Calculator, 0),
	}
}
