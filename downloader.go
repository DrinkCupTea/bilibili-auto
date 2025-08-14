package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	bilibili "github.com/CuteReimu/bilibili/v2"
)

// VideoInfo 视频信息结构
type VideoInfo struct {
	BVID        string `json:"bvid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Duration    int    `json:"duration"`
	CID         int    `json:"cid"`
	Quality     string `json:"quality"`
	DownloadURL string `json:"download_url"`
}

// Downloader 下载器结构
type Downloader struct {
	client        *bilibili.Client
	config        *Config
	downloadQueue chan VideoInfo
	wg            sync.WaitGroup
}

// NewDownloader 创建新的下载器
func NewDownloader(client *bilibili.Client, config *Config) *Downloader {
	return &Downloader{
		client:        client,
		config:        config,
		downloadQueue: make(chan VideoInfo, config.MaxConcurrent*2),
	}
}

// GetVideoInfo 获取视频详细信息
func (d *Downloader) GetVideoInfo(bvid string) (*VideoInfo, error) {
	// 获取视频基本信息
	videoInfo, err := d.client.GetVideoInfo(bilibili.VideoParam{
		Bvid: bvid,
	})
	if err != nil {
		return nil, fmt.Errorf("获取视频信息失败 %s: %w", bvid, err)
	}

	if len(videoInfo.Pages) == 0 {
		return nil, fmt.Errorf("视频 %s 没有可用的分页", bvid)
	}

	// 使用第一个分页的CID
	cid := videoInfo.Pages[0].Cid

	// 获取视频流信息
	streamInfo, err := d.client.GetVideoStream(bilibili.GetVideoStreamParam{
		Bvid: bvid,
		Cid:  cid,
		Qn:   d.getQualityNumber(d.config.VideoQuality),
	})
	if err != nil {
		return nil, fmt.Errorf("获取视频流失败 %s: %w", bvid, err)
	}

	// 获取最佳质量的下载URL
	downloadURL := ""
	if len(streamInfo.Durl) > 0 {
		downloadURL = streamInfo.Durl[0].Url
	} else if len(streamInfo.Dash.Video) > 0 {
		downloadURL = streamInfo.Dash.Video[0].BaseUrl
	}

	return &VideoInfo{
		BVID:        bvid,
		Title:       videoInfo.Title,
		Description: videoInfo.Desc,
		Author:      videoInfo.Owner.Name,
		Duration:    videoInfo.Duration,
		CID:         cid,
		Quality:     d.config.VideoQuality,
		DownloadURL: downloadURL,
	}, nil
}

// DownloadVideo 下载单个视频
func (d *Downloader) DownloadVideo(videoInfo VideoInfo) error {
	// 创建下载目录
	downloadDir := filepath.Join(d.config.DownloadPath, "videos")
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return fmt.Errorf("创建下载目录失败: %w", err)
	}

	// 生成安全的文件名
	filename := sanitizeFilename(fmt.Sprintf("%s_%s.mp4", videoInfo.BVID, videoInfo.Title))
	filePath := filepath.Join(downloadDir, filename)

	// 检查文件是否已存在
	if _, err := os.Stat(filePath); err == nil {
		log.Printf("文件已存在，跳过下载: %s", filename)
		return nil
	}

	log.Printf("开始下载视频: %s - %s", videoInfo.BVID, videoInfo.Title)

	// 保存视频信息到文件
	infoFile := filepath.Join(downloadDir, fmt.Sprintf("%s_info.txt", videoInfo.BVID))
	infoContent := fmt.Sprintf(`视频信息:
BVID: %s
标题: %s
作者: %s
时长: %d秒
下载URL: %s
下载时间: %s
`,
		videoInfo.BVID, videoInfo.Title, videoInfo.Author, videoInfo.Duration, videoInfo.DownloadURL, time.Now().Format("2006-01-02 15:04:05"))

	if err := os.WriteFile(infoFile, []byte(infoContent), 0644); err != nil {
		log.Printf("保存视频信息失败: %v", err)
	}

	// 实际下载视频文件
	if err := d.downloadVideoFile(videoInfo.DownloadURL, filePath); err != nil {
		return fmt.Errorf("下载视频文件失败: %w", err)
	}

	log.Printf("视频下载完成: %s", filename)
	return nil
}

// downloadVideoFile 下载视频文件
func (d *Downloader) downloadVideoFile(url, filePath string) error {
	if url == "" {
		return fmt.Errorf("下载URL为空")
	}

	// 创建HTTP请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置必要的请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com")

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Minute, // 设置较长的超时时间
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	// 创建输出文件
	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer outFile.Close()

	// 获取文件大小用于进度显示
	contentLength := resp.Header.Get("Content-Length")
	totalSize, _ := strconv.ParseInt(contentLength, 10, 64)

	// 复制数据并显示进度
	var downloaded int64
	buffer := make([]byte, 32*1024) // 32KB buffer
	lastProgressTime := time.Now()

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := outFile.Write(buffer[:n])
			if writeErr != nil {
				return fmt.Errorf("写入文件失败: %w", writeErr)
			}
			downloaded += int64(n)

			// 每5秒显示一次下载进度
			if totalSize > 0 && time.Since(lastProgressTime) > 5*time.Second {
				progress := float64(downloaded) / float64(totalSize) * 100
				log.Printf("下载进度: %.2f%% (%d/%d bytes)", progress, downloaded, totalSize)
				lastProgressTime = time.Now()
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取数据失败: %w", err)
		}
	}

	log.Printf("下载完成: %d bytes", downloaded)

	return nil
}

// StartDownloadWorkers 启动下载工作协程
func (d *Downloader) StartDownloadWorkers() {
	for i := 0; i < d.config.MaxConcurrent; i++ {
		d.wg.Add(1)
		go func(workerID int) {
			defer d.wg.Done()
			log.Printf("下载工作协程 %d 启动", workerID)

			for videoInfo := range d.downloadQueue {
				if err := d.DownloadVideo(videoInfo); err != nil {
					log.Printf("下载失败 %s: %v", videoInfo.BVID, err)
				} else {
					log.Printf("下载成功: %s - %s", videoInfo.BVID, videoInfo.Title)
				}
			}

			log.Printf("下载工作协程 %d 结束", workerID)
		}(i)
	}
}

// AddToQueue 添加视频到下载队列
func (d *Downloader) AddToQueue(videoInfo VideoInfo) {
	d.downloadQueue <- videoInfo
}

// Close 关闭下载器
func (d *Downloader) Close() {
	close(d.downloadQueue)
	d.wg.Wait()
}

// getQualityNumber 将质量字符串转换为API需要的数字
func (d *Downloader) getQualityNumber(quality string) int {
	switch strings.ToLower(quality) {
	case "4k", "2160p":
		return 120
	case "1080p60":
		return 116
	case "1080p+":
		return 112
	case "1080p":
		return 80
	case "720p60":
		return 74
	case "720p":
		return 64
	case "480p":
		return 32
	case "360p":
		return 16
	default:
		return 80 // 默认1080p
	}
}

// sanitizeFilename 清理文件名中的非法字符
func sanitizeFilename(filename string) string {
	// 定义非法字符映射
	illegalChars := map[rune]rune{
		'<':  '_',
		'>':  '_',
		':':  '_',
		'"':  '_',
		'/':  '_',
		'\\': '_',
		'|':  '_',
		'?':  '_',
		'*':  '_',
	}

	// 替换非法字符
	var result strings.Builder
	for _, char := range filename {
		if replacement, exists := illegalChars[char]; exists {
			result.WriteRune(replacement)
		} else {
			result.WriteRune(char)
		}
	}

	// 限制文件名长度
	cleanName := result.String()
	if len(cleanName) > 200 {
		cleanName = cleanName[:200]
	}

	return cleanName
}
