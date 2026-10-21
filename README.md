# MC Launcher - 跨平台 Minecraft 命令行启动器

一款命令行版的 Minecraft 启动器，复刻 PCL (Plain Craft Launcher) 的核心功能。

## 功能特性

- ✅ **版本管理** - 安装/启动/删除/验证 Minecraft 版本
- ✅ **Java 管理** - 自动查找、安装 Temurin JDK
- ✅ **认证系统** - 离线登录、Microsoft 正版、Authlib Injector
- ✅ **资源管理** - 资产下载、完整性验证、虚拟资源
- ✅ **原生库** - 自动解压游戏原生库
- ✅ **模组加载器** - Forge、Fabric、Quilt、NeoForge
- ✅ **模组管理** - 添加/删除/启用/禁用模组
- ✅ **存档管理** - 列出、备份、删除存档
- ✅ **服务器列表** - 管理多人服务器
- ✅ **崩溃分析** - 分析崩溃报告和日志
- ✅ **文件校验** - 验证并修复版本文件

## 系统要求

- **操作系统**: Windows 10+ / Ubuntu 20.04+ / macOS 12+
- **Java**: 17 或更高版本（可自动安装）
- **CMake**: 3.20+（编译时需要）
- **磁盘空间**: 至少 500MB

## 快速开始

### 从 Release 安装

```bash
# 下载对应平台的版本
# Windows: mc-launcher-windows.zip
# Linux:   mc-launcher-linux.tar.gz

# 解压并运行
./mc_launcher -help
```

### 从源码编译

```bash
git clone https://github.com/NekoMymaomao/mc-launcher.git
cd mc-launcher
mkdir build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
cmake --build . --parallel
cmake --install . --prefix ../install
```

## 使用方法

### 基本命令

```bash
# 查看帮助
./mc_launcher -help

# 显示调试信息
./mc_launcher -info

# 列出已安装版本
./mc_launcher -list

# 安装 Minecraft 1.20.1
./mc_launcher -install 1.20.1

# 启动游戏
./mc_launcher -launch 1.20.1

# 删除版本
./mc_launcher -remove 1.20.1

# 验证版本完整性
./mc_launcher -verify 1.20.1
```

### Java 管理

```bash
# 列出已安装的 Java
./mc_launcher -java list

# 安装 Java 17
./mc_launcher -java install 17

# 设置 Java 路径
./mc_launcher -java set /path/to/java
```

### 登录认证

```bash
# 离线登录
./mc_launcher -login offline Steve

# Microsoft 正版登录
./mc_launcher -login ms

# Authlib Injector 登录
./mc_launcher -login authlib http://your-server.com
```

### 模组管理

```bash
# 列出模组
./mc_launcher -mods 1.20.1 list

# 添加模组
./mc_launcher -mods 1.20.1 add ./my_mod.jar

# 删除模组
./mc_launcher -mods 1.20.1 rm my_mod.jar
```

### 加载器安装

```bash
# 安装 Fabric
./mc_launcher -loader 1.20.1 fabric latest

# 安装 Forge
./mc_launcher -loader 1.20.1 forge 47.2.0

# 安装 Quilt
./mc_launcher -loader 1.20.1 quilt latest

# 安装 NeoForge
./mc_launcher -loader 1.20.1 neoforge latest
```

### 其他功能

```bash
# 账号管理
./mc_launcher -account list
./mc_launcher -account select microsoft

# 资源管理
./mc_launcher -assets 1.20.1 download
./mc_launcher -assets 1.20.1 verify

# 崩溃分析
./mc_launcher -crash list
./mc_launcher -crash view crash_report.txt

# 存档管理
./mc_launcher -saves 1.20.1 list
./mc_launcher -saves 1.20.1 backup myworld

# 服务器管理
./mc_launcher -servers list
./mc_launcher -servers add mc.example.com Example Server

# 设置
./mc_launcher -settings show
./mc_launcher -settings mem 4096
```

## 目录结构

```
mc_launcher/          # 启动器目录
├── versions/         # 已安装的版本
├── libraries/        # 依赖库缓存
├── assets/           # 游戏资源
│   ├── indexes/      # 资产索引
│   └── objects/      # 资产对象
├── sessions/         # 认证会话
├── logs/             # 日志文件
├── backups/          # 存档备份
└── natives/          # 原生库（运行时解压）

launcher/             # 启动器配置目录
├── settings.json     # 全局设置
└── servers.json      # 服务器列表
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `MC_LAUNCHER_DIR` | 启动器目录 | `~/.mc-launcher` |
| `JAVA_HOME` | Java 安装路径 | 自动检测 |
| `MINECRAFT_DIR` | 游戏目录 | `<launcher_dir>/.minecraft` |

## 构建系统

本项目使用 CMake 构建，所有依赖通过 FetchContent 自动下载：

- **nlohmann/json** - JSON 解析
- **curl** - HTTP 请求
- **miniz** - ZIP 压缩/解压
- **ZLIB** - 压缩支持

### GitHub Actions 工作流

每次推送或创建 Tag 时，GitHub Actions 会自动构建并生成 Release：

```bash
# 创建新版本标签
git tag v1.0.0
git push origin v1.0.0
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 致谢

- [Minecraft](https://www.minecraft.net/) - Mojang Studios
- [PCL](https://hmcl.eu/) - Plain Craft Launcher
- [Adoptium](https://adoptium.net/) - Temurin JDK
