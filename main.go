package main

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	e.GET("/", HealthCheckHandler)
	e.POST("/tax/calculations", TaxCalculationsHandler)

	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown
	fmt.Println("\nshutting down the server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

func HealthCheckHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, "Hello, Go Bootcamp!")
}

type Err struct {
	Message string `json:"message"`
}

type Allowance struct {
	AllowanceType string
	Amount        float64
}

type IncomeStatement struct {
	TotalIncome float64     `json:"totalIncome"`
	Wht         float64     `json:"wht"`
	Allowances  []Allowance `json:"allowances"`
}

type TaxResult struct {
	Tax float64 `json:"tax"`
}

type TaxRefund struct {
	Refund float64 `json:"taxRefund"`
}

// TaxCalculationsHandler
//
//	@Summary		Handles tax calculation
//	@Description	Calculate tax based on total income, WHT, and allowances
//	@Tags			tax
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	Tax
//	@Failure		500	{object}	Err
//	@Router			/tax/calculations [post]
func TaxCalculationsHandler(c echo.Context) error {
	var i IncomeStatement
	err := c.Bind(&i)
	fmt.Println(i)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}
	calculatedTax, err := CalculateTax(i.TotalIncome, i.Wht, i.Allowances)
	if calculatedTax < 0 {
		calculatedTax *= -1
		return c.JSON(http.StatusOK, TaxRefund{Refund: calculatedTax})
	}
	return c.JSON(http.StatusOK, TaxResult{Tax: calculatedTax})
}

func CalculateAllowance(allowances []Allowance) float64 {
	calculatedAllowance := 0.0
	for _, allowance := range allowances {
		if allowance.AllowanceType == "donation" {
			calculatedAllowance += allowance.Amount
		}
	}
	if calculatedAllowance > 100000 {
		calculatedAllowance = 100000
	}
	return calculatedAllowance
}

func CalculateTax(totalIncome, wht float64, allowances []Allowance) (float64, error) {
	totalAllowance := 60000.0 + CalculateAllowance(allowances)
	grossIncome := totalIncome - totalAllowance
	totalTax := 0.0

	if grossIncome <= 150000 {
		return totalTax, nil
	}
	grossIncome -= 150000

	if grossIncome <= 350000 {
		totalTax += grossIncome * 0.1
		totalTax -= wht
		return totalTax, nil
	}
	totalTax += 350000 * 0.1
	grossIncome -= 350000

	if grossIncome <= 500000 {
		totalTax += grossIncome * 0.15
		totalTax -= wht
		return totalTax, nil
	}
	totalTax += 500000 * 0.15
	grossIncome -= 500000

	if grossIncome <= 1000000 {
		totalTax += grossIncome * 0.2
		totalTax -= wht
		return totalTax, nil
	}
	totalTax += 1000000 * 0.2
	grossIncome -= 1000000

	totalTax += grossIncome * 0.35
	totalTax -= wht
	return totalTax, nil
}
