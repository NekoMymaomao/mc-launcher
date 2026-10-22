# MC Launcher 使用教程

## 快速开始

### 1. 安装启动器

下载对应平台的启动器：
- **Windows**: `mc-launcher-windows-amd64.exe`
- **Linux**: `mc-launcher-linux-amd64.tar.gz`
- **Android**: `mc-launcher.apk`

### 2. 首次运行

启动器会自动创建以下目录结构：
```
~/.mc-launcher/
├── versions/      # 游戏版本
├── libraries/     # 依赖库
├── assets/        # 游戏资源
├── java/          # Java 运行时
├── mods/          # 模组
├── shaderpacks/   # 光影包
├── resourcepacks/ # 资源包
├── data_packs/    # 数据包
├── saves/         # 存档
├── servers.json   # 服务器列表
├── accounts.json  # 账号数据
├── settings.json  # 启动器配置
└── logs/          # 日志文件
```

---

## 基本功能

### 安装游戏版本

1. 打开启动器
2. 点击「版本管理」
3. 点击「安装版本」
4. 输入版本号（如 `1.20.1`）
5. 等待下载完成

### 启动游戏

1. 在版本列表中选择版本
2. 点击「启动」按钮
3. 启动器会自动：
   - 检查 Java 环境
   - 下载必要依赖
   - 启动游戏

### 账号登录

#### 离线模式
1. 点击「账号管理」
2. 点击「离线登录」
3. 输入游戏名称

#### Microsoft 账号
1. 点击「账号管理」
2. 点击「Microsoft 登录」
3. 在浏览器中完成授权

---

## Modrinth 模组市场

### 搜索模组
1. 点击「模组市场」
2. 输入搜索关键词
3. 选择需要的模组

### 安装模组
1. 找到目标模组
2. 选择对应的 Minecraft 版本
3. 点击「安装」

### 安装整合包
1. 搜索整合包
2. 点击「安装」
3. 启动器会自动安装所有依赖

### 安装光影/资源包
1. 切换到「光影」或「资源包」标签
2. 搜索并安装

---

## 配置选项

### 内存设置
- **最大内存**: 建议 4096MB 以上
- **最小内存**: 建议 2048MB 以上

### Java 设置
- 启动器会自动检测系统 Java
- 也可手动指定 Java 路径
- 支持自动下载 Temurin JDK

### JVM 参数
可在设置中添加自定义 JVM 参数：
```
-Xmx4G
-Xms2G
-XX:+UseG1GC
```

### 游戏参数
可自定义游戏启动参数。

---

## 服务器管理

### 添加服务器
1. 点击「服务器」
2. 点击「添加服务器」
3. 输入服务器地址和名称

### 导入服务器
支持导入 Minecraft 的 `servers.json` 文件。

---

## 高级功能

### 崩溃分析
1. 点击「崩溃报告」
2. 选择崩溃日志
3. 启动器会分析并给出建议

### 存档管理
- 备份存档
- 恢复存档
- 管理多版本存档

### 多语言支持
启动器支持中英文切换：
- 启动参数: `-lang zh` 或 `-lang en`
- 环境变量: `MC_LAUNCHER_LANG=zh`

---

## 常见问题

### Q: 提示找不到 Java？
A: 点击「Java 管理」->「安装 Java」自动下载。

### Q: 游戏启动失败？
A: 检查崩溃报告，确认 Java 版本正确（建议 Java 17+）。

### Q: 模组不兼容？
A: 确认模组支持的 Minecraft 版本与已安装版本匹配。

### Q: 如何更新启动器？
A: 重新下载最新版本覆盖安装即可。

---

## API 接口

启动器内置 HTTP API，可通过 Web UI 或脚本调用：

```bash
# 获取版本列表
curl http://localhost:9876/api/versions

# 安装版本
curl -X POST http://localhost:9876/api/versions/install \
  -H "Content-Type: application/json" \
  -d '{"version":"1.20.1"}'

# 启动游戏
curl -X POST http://localhost:9876/api/launch \
  -H "Content-Type: application/json" \
  -d '{"version":"1.20.1","username":"Player"}'
```

详细 API 文档见 [API 文档](./api.html)
