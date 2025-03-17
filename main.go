package main

import (
	"fmt"
	"log"
	"os"

	"github.com/CuteReimu/bilibili/v2"
)

func Login(client *bilibili.Client) bool {
	qrCode, _ := client.GetQRCode()
	qrCode.Print()

	result, err := client.LoginWithQRCode(bilibili.LoginWithQRCodeParam{
		QrcodeKey: qrCode.QrcodeKey,
	})

	if err == nil && result.Code == 0 {
		log.Println("登录成功")
		return true
	} else {
		log.Println(err)
		log.Println("登录失败")
		return false
	}
}

func InitClient(client *bilibili.Client) bool {
	var cookieString string
	var cookie_file string = "cookie"

	// cookie存在则直接读取
	if _, err := os.Stat(cookie_file); err == nil {
		cookieData, _ := os.ReadFile(cookie_file)
		cookieString = string(cookieData)
	} else {
		Login(client)
		cookieString = client.GetCookiesString()
		os.WriteFile(cookie_file, []byte(cookieString), 0644)
	}

	client.SetCookiesString(cookieString)
	return true
}

func GetFavVideos(client *bilibili.Client, FavId int) []string {
	log.Println("开始获取", FavId, "下视频")
	var bvIds []string

	page := 1
	for {
		var favList *bilibili.FavourList
		var err error
		params := bilibili.GetFavourListParam{
			MediaId:  FavId,
			Tid:      0,
			Keyword:  "",
			Order:    "mtime",
			Type:     0,
			Ps:       20,
			Pn:       page,
			Platform: "web",
		}
		favList, err = client.GetFavourList(params)
		page += 1

		if err != nil {
			log.Println("GetFavourList error")
			println(err)
			return bvIds
		}
		for _, v := range (*favList).Medias {
			bvIds = append(bvIds, v.BvId)
		}

		if len((*favList).Medias) == 0 {
			break
		}
	}

	return bvIds
}

func main() {
	fmt.Println("")
	var client = bilibili.New()
	InitClient(client)

	favIds := []int{}

	for _, id := range favIds {
		bvIds := GetFavVideos(client, id)
		fmt.Println("============")
		fmt.Println(len(bvIds))
		for _, bvId := range bvIds {
			fmt.Println(bvId)
		}
	}
}
