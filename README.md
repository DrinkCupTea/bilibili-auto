# Bilibili Auto Downloader

一个自动下载bilibili收藏夹视频的Go程序。

## 功能特性

- 🎯 **批量下载**: 支持批量下载指定收藏夹中的所有视频
- 📂 **分类管理**: 自动按收藏夹名称创建子目录，视频分类存储
- ⚡ **并发下载**: 可配置的并发下载数量，提高下载效率
- 🔐 **自动登录**: 支持cookie保存，避免重复登录
- 📁 **智能管理**: 自动创建下载目录，避免重复下载
- 🎥 **多质量**: 支持多种视频质量选择（4k/1080p/720p/480p/360p）
- 📊 **实时进度**: 显示下载进度和状态信息
- ⚙️ **灵活配置**: JSON配置文件，易于修改和管理

## 快速开始

### 1. 编译程序

```bash
go build -o bilibili-auto.exe
```

### 2. 配置设置

首次运行程序会自动创建 `config.json` 配置文件，或者你可以复制 `config.example.json` 并修改：

```json
{
  "favorite_ids": [],
  "download_path": "./downloads",
  "cookie_file": "cookie",
  "max_concurrent": 3,
  "video_quality": "1080p"
}
```

**配置说明：**
- `favorite_ids`: 收藏夹ID列表（必填）
- `download_path`: 下载目录路径
- `cookie_file`: cookie文件路径
- `max_concurrent`: 最大并发下载数
- `video_quality`: 视频质量偏好

### 3. 获取收藏夹ID

1. 打开bilibili网页版
2. 进入你的收藏夹页面
3. 从URL中获取ID，例如：`https://space.bilibili.com/xxx/favlist?fid=3469786402`
4. 将ID添加到配置文件的 `favorite_ids` 数组中

### 4. 运行程序

```bash
./bilibili-auto.exe
```

首次运行需要扫描二维码登录，后续运行会自动使用保存的cookie。

## 项目结构

```
├── main.go              # 主程序入口
├── config.go            # 配置管理
├── downloader.go        # 下载器实现
├── config.json          # 配置文件（运行时生成）
├── config.example.json  # 配置文件示例
├── cookie              # cookie文件（登录后生成）
└── downloads/          # 默认下载目录
    ├── 收藏夹1/
    │   ├── 视频1.mp4
    │   ├── 视频1.json
    │   └── 视频2.mp4
    ├── 收藏夹2/
    │   ├── 视频3.mp4
    │   └── 视频3.json
    └── ...
```

每个收藏夹的视频会自动保存到对应的子目录中，便于管理和查找。

## 注意事项

- 请确保收藏夹是公开的或者你有访问权限
- 程序已实现完整的视频下载功能，包括获取视频流URL和文件下载
- 下载的视频文件会保存在 `downloads/videos/` 目录下
- 每个视频会同时生成信息文件（.txt），包含视频详细信息
- 请遵守bilibili的使用条款和相关法律法规
- 建议合理设置并发数，避免对服务器造成过大压力
- 下载进度会每5秒显示一次，避免日志过于频繁

## API参考

- [bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect)
- [哔哩哔哩API的Go版本SDK](https://github.com/CuteReimu/bilibili?tab=readme-ov-file)

## 许可证

本项目采用MIT许可证，详见LICENSE文件。
