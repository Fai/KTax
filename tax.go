package main

type Allowance struct {
	AllowanceType string
	Amount        float64
}

func CalculateTax(totalIncome, wht float64, allowances []Allowance) (float64, error) {
	// TODO: Implement the tax calculation logic here
	return 0, nil
}
