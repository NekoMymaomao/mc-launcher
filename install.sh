#!/bin/bash
# MC Launcher WebUI - 跨平台安装脚本

OS=$(uname -s | tr '[:upper:]' '[:lower:]')

case "$OS" in
    linux*)
        echo "检测到 Linux 系统"
        exec bash "$(dirname "$0")/install-linux.sh"
        ;;
    darwin*)
        echo "检测到 macOS 系统"
        echo "⚠️  macOS 支持尚未完善，请参考文档手动安装"
        echo "   https://github.com/NekoMymaomao/mc-launcher-webui"
        exit 1
        ;;
    msys*|cygwin*|mingw*)
        echo "检测到 Windows 系统 (Git Bash)"
        echo "💡 建议使用原生 Windows 安装脚本:"
        echo "   下载 install-windows.bat 并在 cmd 中运行"
        echo ""
        echo "正在切换至 Linux 安装模式..."
        exec bash "$(dirname "$0")/install-linux.sh"
        ;;
    *)
        echo "❌ 不支持的操作系统: $OS"
        echo "   支持的系统: Linux, Windows"
        exit 1
        ;;
esac
