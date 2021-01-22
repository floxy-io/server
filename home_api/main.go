package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"io/ioutil"
	"net/http"
	"time"
)

// 6LeUMCMaAAAAABmP3FZtGTOFcGDgpGR0Z0pI7j2R
// server = 6LeUMCMaAAAAAJoU0YrCr8u_2KqARZvW-bRTmjzw

var pubSubCli *pubsub.Client

func main() {
	var err error
	pubSubCli, err = pubsub.NewClient(context.Background(), "floxy-300919")
	if err != nil {
		panic(err)
	}

	e := echo.New()
	e.Static("/", "assets")
	e.POST("/burn", burn)
	e.Logger.Fatal(e.Start(":8080"))
}

type googleCaptchaResponse struct {
	Success bool
	Score  float32
}

func burn(c echo.Context) error{
	// recaptcha
	{
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
	}

	// send message to certificate system
	topic := pubSubCli.Topic("create-binary")
	result := topic.Publish(c.Request().Context(),&pubsub.Message{Data: []byte("{}")})

	_, err := result.Get(c.Request().Context())
	if err != nil {
		return c.String(400, "nok")
	}
	//result := topic.Publish(ctx, &pubsub.Message{Data: []byte("payload")})

	return c.String(200, "ok")
}

func callRecaptcha(token string) (bool,float32, error){
	url := fmt.Sprintf("https://www.google.com/recaptcha/api/siteverify?secret=6LeUMCMaAAAAAJoU0YrCr8u_2KqARZvW-bRTmjzw&response=%s", token)

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	res := googleCaptchaResponse{}
	json.Unmarshal(body, &res)

	time.Sleep(5 * time.Second)

	if !res.Success {
		return false, 0, err
	}
	return true, res.Score, nil
}
