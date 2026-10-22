# MC Launcher Kernel

跨平台 Minecraft TCP 服务器核心 - 为 Android APK 和远程客户端提供启动器 API。

## 功能特性

### TCP API 接口
通过 JSON over TCP 协议暴露完整启动器功能：

```json
// 连接示例
nc 127.0.0.1 9876
{"cmd":"ping"}
{"cmd":"info"}
{"cmd":"versions.list"}
{"cmd":"modrinth.search","args":{"query":"optifine","type":"mod"}}
```

### Modrinth API 集成
- **模组搜索**: `modrinth.search` - 按名称/类型搜索
- **项目详情**: `modrinth.project` - 获取项目信息
- **版本列表**: `modrinth.versions` - 获取兼容版本
- **一键安装**: 
  - `mod.pack.install` - 安装整合包
  - `mod.shader.install` - 安装光影
  - `mod.resourcepack.install` - 安装资源包
  - `mod.datapack.install` - 安装数据包

### 多语言支持
- `-lang en` - 英语
- `-lang zh` - 中文

### 命令行参数
```bash
./mc-launcher-kernel -addr 0.0.0.0:9876 -lang zh
./mc-launcher-kernel -version
```

## API 命令列表

### 通用
| 命令 | 描述 |
|------|------|
| `ping` | 心跳检测 |
| `info` | 启动器信息 |
| `settings.get` | 获取设置 |
| `settings.set` | 设置设置 |
| `lang.get` | 获取当前语言 |
| `lang.set` | 设置语言 |
| `lang.list` | 列出支持的语言 |

### 版本管理
| 命令 | 描述 |
|------|------|
| `versions.list` | 列出已安装版本 |
| `versions.available` | 获取可用版本 |
| `version.detail` | 版本详情 |
| `install` | 安装版本 |
| `launch` | 启动游戏 |
| `remove` | 删除版本 |
| `verify` | 验证版本完整性 |

### Java 管理
| 命令 | 描述 |
|------|------|
| `java.list` | 列出已安装 Java |
| `java.install` | 安装 Java |
| `java.pick` | 自动选择 Java |

### 认证
| 命令 | 描述 |
|------|------|
| `auth.offline` | 离线登录 |
| `auth.microsoft` | Microsoft 登录 |
| `auth.authlib` | Authlib 登录 |
| `account.list` | 列出账号 |
| `account.remove` | 删除账号 |

### Modrinth 内容
| 命令 | 描述 |
|------|------|
| `modrinth.search` | 搜索项目 |
| `modrinth.project` | 项目详情 |
| `modrinth.files` | 文件列表 |
| `modrinth.versions` | 版本列表 |
| `modrinth.featured` | 推荐项目 |
| `modrinth.categories` | 分类列表 |
| `mod.list` | 列出模组 |
| `mod.add` | 添加模组 |
| `mod.remove` | 删除模组 |
| `mod.pack.install` | 安装整合包 |
| `shader.list` | 列出光影 |
| `shader.install` | 安装光影 |
| `resourcepack.list` | 列出资源包 |
| `resourcepack.install` | 安装资源包 |
| `datapack.list` | 列出数据包 |
| `datapack.install` | 安装数据包 |

### 存档/地图
| 命令 | 描述 |
|------|------|
| `saves.list` | 列出存档 |
| `saves.backup` | 备份存档 |
| `saves.delete` | 删除存档 |
| `map.install` | 从 Modrinth 安装地图 |

### 服务器/崩溃
| 命令 | 描述 |
|------|------|
| `servers.list` | 列出服务器 |
| `servers.add` | 添加服务器 |
| `servers.remove` | 删除服务器 |
| `crash.list` | 列出崩溃报告 |
| `crash.analyze` | 分析崩溃 |

## 构建

### 服务端二进制
```bash
CGO_ENABLED=0 go build -o mc-launcher-kernel .
```

### Android JNI 共享库
```bash
CGO_ENABLED=1 GOOS=android GOARCH=arm64 \
  go build -buildmode=c-shared -o libmc_launcher.so \
  android.go *.go
```

## 目录结构
```
~/.mc-launcher/
├── settings.json          # 全局设置
├── accounts.json          # 账号数据
├── versions/              # 已安装版本
├── libraries/             # 依赖库
├── assets/                # 资源缓存
│   ├── indexes/
│   └── objects/
├── java/                  # 内置 Java
├── logs/                  # 日志
├── sessions/              # 会话
├── mods/                  # 模组
│   ├── modrinth/          # Modrinth 缓存
│   │   ├── cache/
│   │   ├── installed/
│   │   └── projects/
│   ├── curseforge/
│   └── *.jar
├── shaderpacks/           # 光影包
├── resourcepacks/         # 资源包
├── data_packs/            # 数据包
└── saves/                 # 存档
```

## License
MIT
