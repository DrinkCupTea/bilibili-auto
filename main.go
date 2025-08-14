package main

import (
	"fmt"
	"log"
	"os"

	"github.com/CuteReimu/bilibili/v2"
)

func main() {
	log.Println("=== Bilibili 自动下载器启动 ===")

	// 加载配置
	configPath := GetConfigPath()
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	log.Printf("配置加载成功，收藏夹数量: %d", len(config.FavoriteIDs))

	// 初始化bilibili客户端
	client := bilibili.New()
	if !initClient(client, config.CookieFile) {
		log.Fatal("客户端初始化失败")
	}
	log.Println("客户端初始化成功")

	// 创建下载器
	downloader := NewDownloader(client, config)
	log.Printf("下载器创建成功，最大并发数: %d", config.MaxConcurrent)

	// 启动下载工作协程
	downloader.StartDownloadWorkers()
	log.Println("下载工作协程已启动")

	// 处理每个收藏夹
	totalVideos := 0
	for _, favID := range config.FavoriteIDs {
		log.Printf("开始处理收藏夹: %d", favID)

		// 获取收藏夹中的所有视频
		bvids := getFavVideos(client, favID)
		log.Printf("收藏夹 %d 包含 %d 个视频", favID, len(bvids))

		// 处理每个视频
		for i, bvid := range bvids {
			log.Printf("处理视频 %d/%d: %s", i+1, len(bvids), bvid)

			// 获取视频信息
			videoInfo, err := downloader.GetVideoInfo(bvid)
			if err != nil {
				log.Printf("获取视频信息失败 %s: %v", bvid, err)
				continue
			}

			// 添加到下载队列
			downloader.AddToQueue(*videoInfo)
			totalVideos++
		}
	}

	log.Printf("所有视频已添加到下载队列，总计: %d 个视频", totalVideos)

	// 关闭下载器并等待所有下载完成
	downloader.Close()
	log.Println("=== 所有下载任务完成 ===")
}

// initClient 初始化bilibili客户端
func initClient(client *bilibili.Client, cookieFile string) bool {
	var cookieString string

	// 检查cookie文件是否存在
	if _, err := os.Stat(cookieFile); err == nil {
		// 读取现有cookie
		cookieData, err := os.ReadFile(cookieFile)
		if err != nil {
			log.Printf("读取cookie文件失败: %v", err)
			return false
		}
		cookieString = string(cookieData)
		log.Println("使用现有cookie文件")
	} else {
		// 需要重新登录
		log.Println("cookie文件不存在，需要重新登录")
		if !login(client) {
			return false
		}

		// 保存cookie
		cookieString = client.GetCookiesString()
		if err := os.WriteFile(cookieFile, []byte(cookieString), 0644); err != nil {
			log.Printf("保存cookie文件失败: %v", err)
			return false
		}
		log.Println("cookie已保存")
	}

	// 设置cookie
	client.SetCookiesString(cookieString)
	return true
}

// login 用户登录
func login(client *bilibili.Client) bool {
	log.Println("开始登录流程...")

	// 获取二维码
	qrCode, err := client.GetQRCode()
	if err != nil {
		log.Printf("获取二维码失败: %v", err)
		return false
	}

	// 显示二维码
	fmt.Println("请使用bilibili手机客户端扫描以下二维码:")
	qrCode.Print()

	// 等待扫码登录
	result, err := client.LoginWithQRCode(bilibili.LoginWithQRCodeParam{
		QrcodeKey: qrCode.QrcodeKey,
	})

	if err != nil {
		log.Printf("登录失败: %v", err)
		return false
	}

	if result.Code == 0 {
		log.Println("登录成功!")
		return true
	} else {
		log.Printf("登录失败，错误代码: %d", result.Code)
		return false
	}
}

// getFavVideos 获取收藏夹中的所有视频
func getFavVideos(client *bilibili.Client, favID int) []string {
	log.Printf("开始获取收藏夹 %d 的视频列表", favID)
	var bvids []string

	page := 1
	for {
		params := bilibili.GetFavourListParam{
			MediaId:  favID,
			Tid:      0,
			Keyword:  "",
			Order:    "mtime",
			Type:     0,
			Ps:       20, // 每页20个视频
			Pn:       page,
			Platform: "web",
		}

		favList, err := client.GetFavourList(params)
		if err != nil {
			log.Printf("获取收藏夹列表失败 (页面 %d): %v", page, err)
			break
		}

		// 检查是否有视频
		if len(favList.Medias) == 0 {
			log.Printf("收藏夹 %d 第 %d 页没有更多视频", favID, page)
			break
		}

		// 添加视频ID
		for _, media := range favList.Medias {
			bvids = append(bvids, media.BvId)
		}

		log.Printf("收藏夹 %d 第 %d 页获取到 %d 个视频", favID, page, len(favList.Medias))
		page++
	}

	log.Printf("收藏夹 %d 总共获取到 %d 个视频", favID, len(bvids))
	return bvids
}
