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
	// URL 点击跳转
	URL string
	// 时效性
	Level Level
	// Copy 下拉等出现复制按钮时点击复制[Copy]的值
	Copy string
	// AutoCopy 自动复制 iOS14.5之后长按或下拉可触发自动复制，iOS14.5之前无需任何操作即可触发自动复制
	AutoCopy bool
	// Badge 角标
	Badge int
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

	if req.Level != "" && req.Level != Active && req.Level != TimeSensitive && req.Level != Passive {
		return errors.New("参数[Level]错误")
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
	if req.Level != "" {
		url = fmt.Sprintf("%slevel=%s&", url, req.Level)
	}
	if req.URL != "" {
		url = fmt.Sprintf("%surl=%s&", url, req.URL)
	}
	if req.Copy != "" {
		url = fmt.Sprintf("%scopy=%s&", url, req.Copy)
	}
	if req.AutoCopy {
		url = fmt.Sprintf("%sautoCopy=%o&", url, 1)
	}
	if req.Badge != 0 {
		url = fmt.Sprintf("%sbadge=%o&", url, req.Badge)
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

type Level string

const (
	// Active 默认值，系统会立即亮屏显示通知
	Active Level = "active"
	// TimeSensitive 时效性通知，可在专注状态下显示通知
	TimeSensitive Level = "timeSensitive"
	// Passive 仅将通知添加到通知列表，不会亮屏提醒
	Passive Level = "passive"
)
