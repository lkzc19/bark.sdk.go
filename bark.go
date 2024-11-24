package bark

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Req struct {
	DeviceKey string
	Title     string
	Content   string
	GroupName string
}

type _resp struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

var baseURL = "https://api.day.app"

func Notify(req Req) error {
	if req.DeviceKey == "" {
		return errors.New("参数[DeviceKey]不可为空")
	}

	if req.Title == "" && req.Content == "" {
		return errors.New("参数[Title Content]至少需要一个")
	}

	url := fmt.Sprintf("%s/%s", baseURL, req.DeviceKey)
	if req.Title != "" {
		url = fmt.Sprintf("%s/%s", url, req.Title)
	}
	if req.Content != "" {
		url = fmt.Sprintf("%s/%s", url, req.Content)
	}
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
		barkResp := _resp{}
		err = json.Unmarshal(respBody, &barkResp)
		if err != nil {
			return err
		}
		return errors.New(fmt.Sprintf("[bark]请求报错: %s", barkResp.Message))
	}
	return nil
}
