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
