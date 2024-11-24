/*
Package bark provides a service for sending messages to bark.

Usage:

	package main

	import "github.com/lkzc19/bark.sdk.go"

	func main() {
		req := bark.Req{
			DeviceKey: "wKcg9jK8Z8h4JMLJyeCDBc",
			Title:     "bark.example.go",
			Content:   "Bark API 请求简单封装测试",
		}
		bark.Notify(req)
	}
*/
package bark
