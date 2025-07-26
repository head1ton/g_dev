package utils

type Calculator struct {
	Name    string
	History []Calculation
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
