# MC Launcher WebUI

基于 Web 的 Minecraft 启动器，支持 Windows、Linux、Android。

## 快速开始

### 1. 编译核心 (Go)

```bash
cd Core/mc-launcher
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o mc-launcher-linux-amd64 .
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o mc-launcher-linux-arm64 .
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o mc-launcher-windows-amd64.exe .
```

### 2. Android APK (本地构建)

由于 GitHub Actions 构建限制，建议本地构建 APK：

```bash
cd android
./gradlew assembleDebug
# APK 位置: app/build/outputs/apk/debug/app-debug.apk
```

### 3. 运行

```bash
# 启动核心服务器
./mc-launcher-linux-amd64 -addr 0.0.0.0:9876

# 打开浏览器访问
open http://localhost:9876
```

## API 文档

- API: http://localhost:9876/docs/api.html
- 教程: http://localhost:9876/docs/tutorial.html

## 项目结构

```
├── Core/mc-launcher/   # Go 核心 (Git submodule)
├── frontend/           # Web UI
├── android/            # Android 项目
└── desktop/            # Tauri 桌面壳
```
