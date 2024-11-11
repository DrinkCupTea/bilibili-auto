package main

import (
	"fmt"
	"log"
	"os"

	"github.com/CuteReimu/bilibili/v2"
)

func main() {
	fmt.Println("")
	var client = bilibili.New()

	qrCode, _ := client.GetQRCode()
	qrCode.Print()

	result, err := client.LoginWithQRCode(bilibili.LoginWithQRCodeParam{
		QrcodeKey: qrCode.QrcodeKey,
	})

	if err == nil && result.Code == 0 {
		log.Println("登录成功")
	}

	cookieString := client.GetCookiesString()
	os.WriteFile("cookie.txt", []byte(cookieString), 0644)

	// if file exist
	if _, err := os.Stat("cookie.txt"); err == nil {
		cookieData, _ := os.ReadFile("cookie.txt")
		cookieString := string(cookieData)
		client.SetCookiesString(string(cookieString))
	}

	client.SetCookiesString(cookieString)
}
