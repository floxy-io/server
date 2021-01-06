package main

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"io/ioutil"
	"net/http"
	"time"
)

//6LeUMCMaAAAAABmP3FZtGTOFcGDgpGR0Z0pI7j2R
// server = 6LeUMCMaAAAAAJoU0YrCr8u_2KqARZvW-bRTmjzw

func main() {
	e := echo.New()
	e.Static("/home", "assets")
	e.POST("/burn", burn)
	e.Logger.Fatal(e.Start(":8080"))
}

type googleCaptchaResponse struct {
	Success bool
	Score  float32
}

func burn(c echo.Context) error{
	token := c.QueryParam("token")
	//
	url := fmt.Sprintf("https://www.google.com/recaptcha/api/siteverify?secret=6LeUMCMaAAAAAJoU0YrCr8u_2KqARZvW-bRTmjzw&response=%s", token)

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	res := googleCaptchaResponse{}
	json.Unmarshal(body, &res)

	time.Sleep(5 * time.Second)

	if !res.Success {
		return c.String(200, "nok")
	}
	if res.Score < 0.5 {
		return c.String(200, "nok")
	}
	if res.Score < 0.9 {
		return c.String(200, "challenge")
	}
	return c.String(200, "ok")
}