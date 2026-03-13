/*
Package bark 是 Bark iOS 推送通知服务的 Go SDK 封装。

Bark 是一款专为 iOS 设计的自定义推送工具，支持通过 HTTP API 向 iPhone/iPad 发送通知。

# 快速开始

使用包级别函数（官方服务端）：

	package main

	import bark "github.com/lkzc19/bark.sdk.go"

	func main() {
		err := bark.Notify(bark.Req{
			DeviceKey: "your-device-key",
			Title:     "服务器告警",
			Body:      "CPU 使用率超过 90%",
		})
		if err != nil {
			panic(err)
		}
	}

# 使用自建服务端

	client := bark.NewWithURL("https://your-bark-server.com")
	err := client.Notify(bark.Req{
		DeviceKey: "your-device-key",
		Body:      "来自私有服务端的推送",
	})

# 高级用法

重要警告（忽略静音和勿扰模式）：

	bark.Notify(bark.Req{
		DeviceKey: "your-device-key",
		Title:     "紧急告警",
		Body:      "数据库连接失败",
		Level:     bark.Critical,
	})

时效性通知（可在专注模式下显示）：

	bark.Notify(bark.Req{
		DeviceKey: "your-device-key",
		Title:     "任务完成",
		Body:      "数据处理完毕，点击查看结果",
		Level:     bark.TimeSensitive,
		URL:       "https://your-app.com/result",
	})

不存档消息：

	notArchive := false
	bark.Notify(bark.Req{
		DeviceKey: "your-device-key",
		Body:      "此消息不会保存到 Bark",
		IsArchive: &notArchive,
	})

端到端加密推送：

	bark.Notify(bark.Req{
		DeviceKey: "your-device-key",
		Title:     "验证码",
		Body:      "你的验证码是：123456",
		Encrypt:   &bark.EncryptConfig{Key: "your-16-byte-key"},
		AutoCopy:  true,
	})
*/
package bark
