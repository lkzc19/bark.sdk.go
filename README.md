# bark.sdk.go

Bark API 请求简单封装

现已支持:

- 标题/内容
- 分组

## 使用

引入依赖

```bash
go get -u github.com/lkzc19/bark.sdk.go
```

使用

```go
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
```

## 单元测试

单元测试需要将`.env.example`文件改名为`.env`，修改`DeviceKey`值。