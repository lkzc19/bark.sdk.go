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
	// Sound 铃声
	Sound string
	// Call	持续响铃 持续重复30秒 默认不持续响铃
	Call bool
	// NotArchive 是否不保存消息(保存在Bark中) 默认保存
	NotArchive bool
	// 推送图标 错误图标不会显示
	Icon      string
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
	url += "?"
	if req.Sound != "" {
		url = fmt.Sprintf("%ssound=%s&", url, req.Sound)
	}
	if req.Call {
		url = fmt.Sprintf("%scall=%o&", url, 1)
	}
	if req.NotArchive {
		url = fmt.Sprintf("%sisArchive=%o&", url, 0)
	}
	if req.Icon != "" {
		url = fmt.Sprintf("%sicon=%s&", url, req.Icon)
	}
	if req.GroupName != "" {
		url = fmt.Sprintf("%sgroup=%s&", url, req.GroupName)
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
