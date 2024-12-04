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

	req.Title = "Notify"
	req.Content = "TestNotify"

	err := Notify(req)
	if err != nil {
		log.Fatal(err)
	}
}

func TestSound(t *testing.T) {
	before()

	req.Title = "Sound"
	req.Content = "TestSound"
	req.Sound = "paymentsuccess"
	req.Call = true

	err := Notify(req)
	if err != nil {
		log.Fatal(err)
	}
}

func TestIcon(t *testing.T) {
	before()

	req.Title = "Icon"
	req.Content = "TestIcon"
	req.Icon = "https://day.app/assets/images/avatar.jpg"

	err := Notify(req)
	if err != nil {
		log.Fatal(err)
	}
}

func TestLevel(t *testing.T) {
	before()

	req.Title = "Level"
	req.Content = "TestLevel"
	req.Level = "zxc"

	err := Notify(req)
	if err != nil {
		log.Fatal(err)
	}
}

func TestURL(t *testing.T) {
	before()

	req.Title = "URL"
	req.Content = "TestURL"
	req.URL = "https://github.com/lkzc19/bark.sdk.go"

	err := Notify(req)
	if err != nil {
		log.Fatal(err)
	}
}

func TestCopy(t *testing.T) {
	before()

	req.Title = "Copy"
	req.Content = "TestCopy"
	req.Copy = "https://pkg.go.dev/github.com/lkzc19/bark.sdk.go"
	req.AutoCopy = true

	err := Notify(req)
	if err != nil {
		log.Fatal(err)
	}
}

func TestBadge(t *testing.T) {
	before()

	req.Title = "Badge"
	req.Content = "TestBadge"
	req.Badge = 42

	err := Notify(req)
	if err != nil {
		log.Fatal(err)
	}
}
