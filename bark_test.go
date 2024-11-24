package bark

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"testing"
)

var req = Req{}

func before() {
	var err = godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	req.DeviceKey = os.Getenv("DeviceKey")
}

func TestNotify(t *testing.T) {
	before()

	req.Title = "标题"
	req.Content = "内容"
	req.GroupName = "测试组3"

	err := Notify(req)
	if err != nil {
		log.Fatal(err)
	}
}
