#!/bin/bash
# MC Launcher WebUI - Linux 一键安装脚本
# 作者: NekoMymaomao
# 版本: 1.0.0

set -e

REPO_URL="https://github.com/NekoMymaomao/mc-launcher-webui.git"
TARGET_DIR="${HOME}/mc-launcher"
ARCH=$(uname -m)

echo "========================================"
echo "  MC Launcher WebUI - Linux 安装程序"
echo "========================================"
echo ""

# 检查 Git
if ! command -v git &> /dev/null; then
    echo "❌ 错误: 未找到 git，请先安装 Git"
    echo "   Ubuntu/Debian: sudo apt install git"
    echo "   CentOS/RHEL: sudo yum install git"
    exit 1
fi
echo "✅ Git 已安装"

# 克隆仓库
if [ -d "$TARGET_DIR" ]; then
    echo "📁 检测到已有安装目录: $TARGET_DIR"
    read -p "是否覆盖现有安装? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$TARGET_DIR"
    else
        echo "取消安装"
        exit 0
    fi
fi

echo ""
echo "📥 正在克隆仓库..."
git clone --depth 1 "$REPO_URL" "$TARGET_DIR"

if [ ! -d "$TARGET_DIR" ]; then
    echo "❌ 克隆失败"
    exit 1
fi

echo "✅ 仓库已克隆到: $TARGET_DIR"

# 检查架构并下载对应二进制
cd "$TARGET_DIR"

echo ""
echo "🔍 检测系统架构: $ARCH"

case $ARCH in
    x86_64|amd64)
        BIN="mc-launcher-linux-amd64"
        ;;
    aarch64|arm64)
        BIN="mc-launcher-linux-arm64"
        ;;
    *)
        echo "❌ 不支持的架构: $ARCH"
        echo "   支持的架构: x86_64, amd64, aarch64, arm64"
        exit 1
        ;;
esac

echo "✅ 检测到架构: $BIN"

# 检查核心文件是否存在
if [ ! -f "Core/mc-launcher/$BIN" ]; then
    echo "⚠️  未找到预编译二进制: Core/mc-launcher/$BIN"
    echo "   尝试从 GitHub Actions 下载..."
    
    # 获取最新的 release 或 artifact
    LATEST_RUN=$(curl -s -H "Accept: application/vnd.github.v3+json" \
        "https://api.github.com/repos/NekoMymaomao/mc-launcher-webui/actions/runs?branch=master&per_page=1" | \
        python3 -c "import sys,json; r=json.load(sys.stdin); print(r['workflow_runs'][0]['id'])" 2>/dev/null || echo "")
    
    if [ -n "$LATEST_RUN" ]; then
        echo "   下载中..."
        curl -L -o /tmp/core-dist.zip \
            "https://github.com/NekoMymaomao/mc-launcher-webui/actions/runs/$LATEST_RUN/artifacts/core-dist/zip" \
            -H "Accept: application/vnd.github.v3+json" 2>/dev/null || true
        
        if [ -f /tmp/core-dist.zip ]; then
            unzip -o /tmp/core-dist.zip -d Core/mc-launcher/ 2>/dev/null || true
            rm /tmp/core-dist.zip
        fi
    fi
fi

# 设置权限
chmod +x "Core/mc-launcher/$BIN" 2>/dev/null || true

# 创建启动脚本
cat > launcher.sh << 'LAUNCHER'
#!/bin/bash
# MC Launcher WebUI - 启动脚本

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN="$SCRIPT_DIR/Core/mc-launcher/mc-launcher-linux-$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')"

echo "========================================"
echo "  MC Launcher WebUI"
echo "========================================"
echo ""

# 启动核心服务器
echo "🚀 启动 MC Launcher 服务器..."
echo "   地址: http://localhost:9876"
echo "   按 Ctrl+C 停止"
echo ""

exec "$BIN" -addr 0.0.0.0:9876
LAUNCHER

chmod +x launcher.sh

# 创建 README
cat > README-INSTALL.md << 'README'
# MC Launcher WebUI - Linux 安装说明

## 安装完成

✅ 已成功安装到: $TARGET_DIR

## 使用方法

### 启动服务器
```bash
cd $TARGET_DIR
./launcher.sh
```

### 访问 Web UI
启动服务器后，在浏览器打开:
- http://localhost:9876
- http://127.0.0.1:9876

### 更新版本
```bash
cd $TARGET_DIR
git pull
```

## 文件说明
- `launcher.sh` - 一键启动脚本
- `Core/mc-launcher/` - 核心二进制文件
- `frontend/` - Web UI 前端文件
- `android/` - Android 项目
- `.github/workflows/` - CI/CD 配置

## 卸载
```bash
rm -rf $TARGET_DIR
```
README

echo ""
echo "========================================"
echo "  ✅ 安装完成!"
echo "========================================"
echo ""
echo "📁 安装位置: $TARGET_DIR"
echo "🚀 启动命令: cd $TARGET_DIR && ./launcher.sh"
echo "🌐 访问地址: http://localhost:9876"
echo ""
echo "========================================"
