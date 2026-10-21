# MC Launcher
跨平台 Minecraft 命令行启动器 - 复刻 PCL 核心功能

## 功能特性

### 版本管理
- `mc_launcher -install <version>` - 安装版本（自动下载依赖+资源+原生）
- `mc_launcher -list` - 列出已安装版本
- `mc_launcher -remove <version>` - 删除版本

### Java 管理
- `mc_launcher -java list` - 列出已安装 Java
- `mc_launcher -java install [版本]` - 自动下载 Temurin JDK
- `mc_launcher -java set <path>` - 设置 Java 路径

### 认证系统
- `mc_launcher -login ms` - Microsoft 设备码登录
- `mc_launcher -login offline [用户名]` - 离线登录
- `mc_launcher -login authlib <服务器> <用户> <密码>` - Authlib Injector

### 模组管理
- `mc_launcher -mods <版本> list` - 列出模组
- `mc_launcher -mods <版本> add <路径>` - 添加模组
- `mc_launcher -mods <版本> rm <名称>` - 删除模组

### 加载器支持
- Forge / Fabric / Quilt / NeoForge
- `mc_launcher -loaders <版本> install <type> <ver>`

### 资源管理
- `mc_launcher -assets <版本> download` - 下载资源
- `mc_launcher -assets <版本> verify` - 验证完整性

### 崩溃分析
- `mc_launcher -crash list` - 列出崩溃报告
- `mc_launcher -crash analyze <文件>` - 分析崩溃

### 存档管理
- `mc_launcher -saves <版本> list` - 列出存档
- `mc_launcher -saves <版本> rename <旧> <新>` - 重命名
- `mc_launcher -saves <版本> delete <名称>` - 删除

### 服务器列表
- `mc_launcher -servers add <地址> [名称]`
- `mc_launcher -servers list`
- `mc_launcher -servers rm <地址>`

### 配置
- `mc_launcher -config mem <MB>` - 设置内存
- `mc_launcher -config jvmarg <参数>` - 添加 JVM 参数
- `mc_launcher -config show` - 显示配置

## 构建

### 依赖
- CMake ≥ 3.20
- C++17 编译器
- libcurl (静态链接)
- nlohmann_json (header-only)
- miniz (zip 处理)

### Linux
```bash
sudo apt install cmake build-essential libcurl4-openssl-dev zlib1g-dev
mkdir build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
make -j$(nproc)
```

### Windows
```powershell
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build --config Release
```

## 使用

```bash
# 安装 Minecraft 1.20.1
./mc_launcher -install 1.20.1

# 启动游戏
./mc_launcher -launch 1.20.1 --username Player --offline

# Microsoft 登录
./mc_launcher -login ms
# 然后在浏览器中登录

# 启动游戏（使用已登录账号）
./mc_launcher -launch 1.20.1
```

## 目录结构
```
launcher/                 # 启动器目录（与可执行文件同级）
├── settings.json        # 全局设置
├── versions/            # 已安装版本
│   └── 1.20.1/
│       ├── 1.20.1.json  # 版本 manifest
│       ├── 1.20.1.jar   # 游戏核心
│       ├── libraries/   # 依赖库
│       └── mods/        # 模组
├── java/                # 内置 Java
│   └── temurin-17/
├── sessions/            # 登录会话
├── assets/              # 资产缓存
│   ├── indexes/
│   └── objects/
└── crash-reports/       # 崩溃报告
```

## License
MIT
