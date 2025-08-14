package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 应用配置结构
type Config struct {
	// FavoriteIDs 收藏夹ID列表
	FavoriteIDs []int `json:"favorite_ids"`
	// DownloadPath 下载路径
	DownloadPath string `json:"download_path"`
	// CookieFile cookie文件路径
	CookieFile string `json:"cookie_file"`
	// MaxConcurrent 最大并发下载数
	MaxConcurrent int `json:"max_concurrent"`
	// VideoQuality 视频质量偏好 (1080p, 720p, 480p等)
	VideoQuality string `json:"video_quality"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		FavoriteIDs:   []int{},
		DownloadPath:  "./downloads",
		CookieFile:    "cookie",
		MaxConcurrent: 3,
		VideoQuality:  "1080p",
	}
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := DefaultConfig()
		if err := config.Save(configPath); err != nil {
			return nil, fmt.Errorf("创建默认配置文件失败: %w", err)
		}
		return config, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// Save 保存配置到文件
func (c *Config) Save(configPath string) error {
	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 序列化配置
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	if len(c.FavoriteIDs) == 0 {
		return fmt.Errorf("收藏夹ID列表不能为空")
	}

	for _, id := range c.FavoriteIDs {
		if id <= 0 {
			return fmt.Errorf("收藏夹ID必须大于0: %d", id)
		}
	}

	if c.DownloadPath == "" {
		c.DownloadPath = "./downloads"
	}

	if c.CookieFile == "" {
		c.CookieFile = "cookie"
	}

	if c.MaxConcurrent <= 0 {
		c.MaxConcurrent = 3
	}

	if c.VideoQuality == "" {
		c.VideoQuality = "1080p"
	}

	return nil
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() string {
	return "config.json"
}
