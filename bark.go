package bark_sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type BarkReq struct {
	Token     string
	Title     string
	Content   string
	GroupName string
}

type barkResp struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

var baseURL = "https://api.day.app"

func Notify(req BarkReq) error {
	if req.Token == "" {
		return errors.New("参数[token]不可为空")
	}

	url := fmt.Sprintf("%s/%s", baseURL, req.Token)
	if req.Title != "" {
		url = fmt.Sprintf("%s/%s", url, req.Title)
	}
	url = fmt.Sprintf("%s/%s", url, req.Content)
	if req.GroupName != "" {
		url = fmt.Sprintf("%s?group=%s", url, req.GroupName)
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		barkResp := barkResp{}
		err = json.Unmarshal(respBody, &barkResp)
		if err != nil {
			return err
		}
		return errors.New(fmt.Sprintf("[bark]请求报错: %s", barkResp.Message))
	}
	return nil
}
