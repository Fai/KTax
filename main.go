package main

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
)

// @title tax API
// @version 1.0
// @description This is a k-tax API
// @host localhost:8080
func main() {
	port := os.Getenv("PORT")

	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Go Bootcamp!")
	})

	e.Logger.Fatal(e.Start(":" + port))
}
