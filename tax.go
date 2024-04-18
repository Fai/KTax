package main

type Allowance struct {
	AllowanceType string
	Amount        float64
}

func CalculateTax(totalIncome, wht float64, allowances []Allowance) (float64, error) {
	totalAllowance := 60000.0
	grossIncome := totalIncome - totalAllowance
	totalTax := 0.0

	if grossIncome <= 150000 {
		return totalTax, nil
	}
	grossIncome -= 150000

	if grossIncome <= 350000 {
		totalTax += grossIncome * 0.1
		return totalTax, nil
	}
	totalTax += 350000 * 0.1
	grossIncome -= 350000

	if grossIncome <= 500000 {
		totalTax += grossIncome * 0.15
		return totalTax, nil
	}
	totalTax += 500000 * 0.15
	grossIncome -= 500000

	if grossIncome <= 1000000 {
		totalTax += grossIncome * 0.2
		return totalTax, nil
	}
	totalTax += 1000000 * 0.2
	grossIncome -= 1000000

	totalTax += grossIncome * 0.35
	return totalTax, nil
}
