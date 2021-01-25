package main

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"
)

func main() {
	// Echo instance
	e := echo.New()
	e.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
		fmt.Println(string(reqBody))
		fmt.Println(string(resBody))
	}))
	// Routes
	e.GET("/", handler)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

// Handler
func handler(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}
