package main

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/net/websocket"
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
	e.GET("/ws", wsserver)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

// Handler
func handler(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}

func wsserver(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		for {
			// Write
			err := websocket.Message.Send(ws, "Hello, Client!")
			if err != nil {
				c.Logger().Error(err)
			}

			// Read
			msg := ""
			err = websocket.Message.Receive(ws, &msg)
			if err != nil {
				c.Logger().Error(err)
				return
			}
			fmt.Printf("%s\n", msg)
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
