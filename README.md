# MC Launcher WebUI

基于 Web 技术的 Minecraft 启动器，支持 Windows、Linux、Android。

## 架构

```
mc-launcher-webui/
├── Core/           # 编译产物目录（自动更新）
│   └── dist/       # 各平台二进制文件
├── backend/        # Go HTTP API 服务器
├── frontend/       # Web UI (Vue 3)
├── desktop/        # 桌面壳工程 (Tauri)
│   ├── windows/    # Windows 打包配置
│   └── linux/      # Linux 打包配置
└── android/        # Android WebView 壳
```

## 功能特性

- ✅ 版本管理 (安装/删除/验证)
- ✅ 账号系统 (离线/Microsoft/Authlib)
- ✅ Java 管理 (自动检测/下载)
- ✅ Modrinth 集成 (搜索/安装模组/光影/资源包)
- ✅ 服务器管理
- ✅ 崩溃分析
- ✅ 多语言支持 (EN/ZH)

## API 文档

访问内置文档：
- API 文档: http://localhost:9876/docs/api.html
- 使用教程: http://localhost:9876/docs/tutorial.html
- FAQ: http://localhost:9876/docs/faq.html

## 构建

### 后端编译
```bash
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o mc-launcher-linux-amd64 .
```

### 桌面版 (Tauri)
```bash
cd desktop/windows
npm install
npm run tauri build
```

### Android
```bash
cd android
./gradlew assembleDebug
```

## GitHub Actions

- **mc-launcher**: 编译核心代码，产物到 Core/dist
- **mc-launcher-webui**: 使用 Core/dist 进行打包

## 许可证

MIT
