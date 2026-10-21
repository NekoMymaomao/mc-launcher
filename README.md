# MC Launcher

全平台通用的Minecraft命令行启动器

## 特性

- 跨平台支持 (Windows / Linux)
- 单一源码，条件编译
- 命令行界面
- 自动JVM查找
- 版本管理

## 依赖

### Linux

```bash
sudo apt-get update
sudo apt-get install -y cmake build-essential libcurl4-openssl-dev nlohmann-json3-dev
```

### Windows

需要手动安装（或使用vcpkg）：

```bash
# 使用vcpkg（推荐）
git clone https://github.com/microsoft/vcpkg
.\vcpkg\bootstrap-vcpkg.bat
.\vcpkg\vcpkg install curl boost-filesystem nlohmann-json

# 或者手动下载
# 1. CMake: https://cmake.org/download/
# 2. Visual Studio 2019+ with C++桌面开发
# 3. libcurl: https://curl.se/windows/
```

## 构建

```bash
mkdir build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
cmake --build . --config Release
```

## 使用

```bash
./mc_launcher --username <你的用户名> [--version <版本ID>]
```

### 选项

| 选项 | 说明 |
|------|------|
| `--username` | Minecraft用户名（必需） |
| `--uuid` | 玩家UUID |
| `--version` | 游戏版本ID |
| `--width` | 窗口宽度 |
| `--height` | 窗口高度 |
| `--debug` | 启用调试日志 |
| `--help` | 显示帮助 |

## GitHub Actions

工作流文件位于 `.github/workflows/build.yml`

- Linux: 自动安装所有依赖
- Windows: 仅提示依赖安装方法

## 许可证

MIT
