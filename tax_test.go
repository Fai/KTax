package main

import (
	"testing"
)

func TestCalculateTax(t *testing.T) {
	totalIncome := 500000.0
	wht := 0.0
	allowances := []Allowance{
		{
			AllowanceType: "donation",
			Amount:        0.0,
		},
	}

	tax, err := CalculateTax(totalIncome, wht, allowances)
	if err != nil {
		t.Errorf("Failed to calculate tax with error: %v", err)
	}

	expectedTax := 29000.0
	if tax != expectedTax {
		t.Errorf("Expected tax to be %v but got %v", expectedTax, tax)
	}
}
