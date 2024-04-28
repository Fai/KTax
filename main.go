package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/lib/pq"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var TaxLevelToggle bool = true

var db *sql.DB

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

type TaxLevelDetail struct {
	Level string  `json:"level"`
	Tax   float64 `json:"tax"`
}

type DetailedTaxResult struct {
	Tax    float64          `json:"tax"`
	Levels []TaxLevelDetail `json:"levels"`
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
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}
	calculatedTax, err := CalculateTotalTax(i.TotalIncome, i.Wht, i.Allowances)
	if calculatedTax < 0 {
		calculatedTax *= -1
		return c.JSON(http.StatusOK, TaxRefund{Refund: calculatedTax})
	}
	if TaxLevelToggle == true {
		return c.JSON(http.StatusOK, DetailedTaxResult{Tax: calculatedTax, Levels: CalculateTaxLevel(calculatedTax, i.Wht)})
	}
	return c.JSON(http.StatusOK, TaxResult{Tax: calculatedTax})
}

func CalculateAllowance(allowances []Allowance) float64 {
	donationAllowance := 0.0
	kReceiptAllowance := 0.0
	for _, allowance := range allowances {
		if allowance.Amount < 0 {

		}
		if allowance.AllowanceType == "donation" {
			donationAllowance += allowance.Amount
		}
		if allowance.AllowanceType == "k-receipt" {
			kReceiptAllowance += allowance.Amount
		}
	}
	if donationAllowance > 100000 {
		donationAllowance = 100000
	}
	if kReceiptAllowance > KReceiptDeductionLimit {
		kReceiptAllowance = KReceiptDeductionLimit
	}
	return donationAllowance + kReceiptAllowance
}

func CalculateTaxLevel(tax float64, wht float64) []TaxLevelDetail {
	var CalculatedTaxLevels []TaxLevelDetail
	CalculatedTaxLevels = []TaxLevelDetail{
		{"0-150,000", 0.0},
		{"150,001-500,000", 0.0},
		{"500,001-1,000,000", 0.0},
		{"1,000,001-2,000,000", 0.0},
		{"2,000,001 ขึ้นไป", 0.0},
	}
	base := tax + wht
	if base <= 35000.0 {
		CalculatedTaxLevels[1].Tax = base
		return CalculatedTaxLevels
	}
	CalculatedTaxLevels[1].Tax = 35000.0
	base -= 35000.0
	if base <= 75000.0 {
		CalculatedTaxLevels[2].Tax = base
		return CalculatedTaxLevels
	}
	CalculatedTaxLevels[2].Tax = 75000.0
	base -= 75000.0
	if base <= 200000.0 {
		CalculatedTaxLevels[3].Tax = base
		return CalculatedTaxLevels
	}
	CalculatedTaxLevels[3].Tax = 200000.0
	base -= 200000.0
	CalculatedTaxLevels[4].Tax = base
	return CalculatedTaxLevels
}

func CalculateTotalTax(totalIncome float64, wht float64, allowances []Allowance) (float64, error) {
	totalAllowance := PersonalDeduction + CalculateAllowance(allowances)
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

type Deduction struct {
	Amount float64 `json:"amount"`
}

type PersonalResponse struct {
	PersonalDeduction float64 `json:"personalDeduction"`
}

func PersonalDeductionsHandler(c echo.Context) error {
	var d Deduction
	err := c.Bind(&d)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}
	if d.Amount > 100000 {
		return c.JSON(http.StatusBadRequest, Err{Message: "Personal deduction must not exceed 100,000"})
	}
	if d.Amount < 60000 {
		return c.JSON(http.StatusBadRequest, Err{Message: "Personal deduction must start from 60000"})
	}

	PersonalDeduction = d.Amount

	_, err = db.Exec("UPDATE deductions SET personal = $1 WHERE id = (SELECT MAX(id) FROM deductions)", PersonalDeduction)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, PersonalResponse{PersonalDeduction: PersonalDeduction})
}

func KReceiptDeductionsHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, "K-Receipt Deductions Adjustment")
}

type CSVTaxResult struct {
	TotalIncome float64 `json:"totalIncome"`
	Tax         float64 `json:"tax,omitempty"`
	TaxRefund   float64 `json:"taxRefund,omitempty"`
}

func CSVTaxCalculationsHandler(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}
	defer src.Close()

	reader := csv.NewReader(src)
	taxRecords, err := reader.ReadAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}
	taxRecords = taxRecords[1:]

	var csvResult []CSVTaxResult
	var csvA []Allowance
	for _, taxRecord := range taxRecords {
		totalIncome, err := strconv.ParseFloat(taxRecord[0], 64)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
		}
		wht, err := strconv.ParseFloat(taxRecord[1], 64)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
		}
		donation, err := strconv.ParseFloat(taxRecord[2], 64)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
		}
		csvA = append(csvA, Allowance{AllowanceType: "donation", Amount: donation})
		calculatedTax, err := CalculateTotalTax(totalIncome, wht, csvA)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
		}
		if calculatedTax < 0 {
			csvResult = append(csvResult, CSVTaxResult{TotalIncome: totalIncome, TaxRefund: calculatedTax * -1})
		} else {
			csvResult = append(csvResult, CSVTaxResult{TotalIncome: totalIncome, Tax: calculatedTax})
		}
	}

	return c.JSON(http.StatusOK, csvResult)
}

func AuthMiddleware(username, password string, c echo.Context) (bool, error) {
	if username == os.Getenv("ADMIN_USERNAME") && password == os.Getenv("ADMIN_PASSWORD") {
		return true, nil
	}
	return false, nil
}

var PersonalDeduction float64
var KReceiptDeductionLimit float64

func loadDeductions() error {
	row := db.QueryRow("SELECT personal, kreceipt FROM deductions ORDER BY id DESC LIMIT 1")
	err := row.Scan(&PersonalDeduction, &KReceiptDeductionLimit)
	if err != nil {
		if err == sql.ErrNoRows {
			PersonalDeduction = 60000.0
			KReceiptDeductionLimit = 50000.0
		} else {
			return err
		}
	}
	return nil
}

func main() {

	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Connect to database error", err)
	}
	defer db.Close()

	createTb := `
		CREATE TABLE IF NOT EXISTS deductions (
			id SERIAL PRIMARY KEY,
			personal FLOAT,
			kreceipt FLOAT
		);
	`

	_, err = db.Exec(createTb)
	if err != nil {
		log.Fatal("can't create table", err)
	}

	err = loadDeductions()
	if err != nil {
		log.Fatal("Failed to load deductions", err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	e.GET("/", HealthCheckHandler)

	t := e.Group("/tax/calculations")
	t.POST("", TaxCalculationsHandler)
	t.POST("/upload-csv", CSVTaxCalculationsHandler)

	ad := e.Group("/admin/deductions")
	ad.Use(middleware.BasicAuth(AuthMiddleware))
	ad.POST("/personal", PersonalDeductionsHandler)
	ad.POST("/k-receipt", KReceiptDeductionsHandler)

	port := os.Getenv("PORT")
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
