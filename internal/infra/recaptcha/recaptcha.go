package recaptcha

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CaptchaResponse struct {
	Success bool
	Score  float32
}


func Get(token string)(CaptchaResponse, error){
	//
	url := fmt.Sprintf("https://www.google.com/recaptcha/api/siteverify?secret=6LeUMCMaAAAAAJoU0YrCr8u_2KqARZvW-bRTmjzw&response=%s", token)

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return CaptchaResponse{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	res := CaptchaResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return CaptchaResponse{}, err
	}

	//time.Sleep(5 * time.Second)
	return res, nil
}
