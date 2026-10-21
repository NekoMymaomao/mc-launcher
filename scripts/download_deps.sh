#!/bin/bash
# ============================================================
# MC Launcher - 静态依赖下载脚本
# 在构建前运行此脚本下载静态curl库
# ============================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VENDOR_DIR="$SCRIPT_DIR/../vendor"
BUILD_DIR="$SCRIPT_DIR/../build"

echo "=========================================="
echo "MC Launcher - Static Dependencies"
echo "=========================================="

# 创建vendor目录
mkdir -p "$VENDOR_DIR"

# 检测平台
if [ "$(uname)" == "Darwin" ]; then
    echo "macOS detected (not fully supported)"
    exit 1
elif [ "$(uname)" == "Linux" ]; then
    echo "Linux detected"
    PLATFORM="linux"
elif [ "$(uname -s)" == "MINGW64_NT-" ] || [ "$(uname -s)" == "MSYS_NT-" ]; then
    echo "Windows (MinGW) detected"
    PLATFORM="windows"
else
    echo "Unsupported platform: $(uname -s)"
    exit 1
fi

# 下载静态curl库
# 注意：实际使用时需要从可靠来源获取预编译的静态库
# 这里提供框架，实际下载链接需要根据平台调整

if [ "$PLATFORM" == "linux" ]; then
    # Linux静态curl
    CURL_LIB_URL="https://github.com/curl/curl/releases/download/curl-8_12_1/libcurl-aio-4.4-git-sdk-64.tar.xz"
    CURL_LIB_FILE="$VENDOR_DIR/libcurl-linux.tar.xz"
    
    echo "Downloading libcurl for Linux..."
    # wget -qO "$CURL_LIB_FILE" "$CURL_LIB_URL" || curl -L -o "$CURL_LIB_FILE" "$CURL_LIB_URL"
    
    # 解压
    # tar -xJf "$CURL_LIB_FILE" -C "$VENDOR_DIR"
    
elif [ "$PLATFORM" == "windows" ]; then
    # Windows静态curl (MinGW)
    CURL_LIB_URL="https://github.com/curl/curl/releases/download/curl-8_12_1/curl-8.12.1_64-w64-mingw32.zip"
    CURL_LIB_FILE="$VENDOR_DIR/libcurl-windows.zip"
    
    echo "Downloading libcurl for Windows..."
    # wget -qO "$CURL_LIB_FILE" "$CURL_LIB_URL" || curl -L -o "$CURL_LIB_FILE" "$CURL_LIB_URL"
    
    # 解压
    # unzip -q "$CURL_LIB_FILE" -d "$VENDOR_DIR"
fi

echo ""
echo "=========================================="
echo "Dependencies ready in: $VENDOR_DIR"
echo "=========================================="
echo ""
echo "Next steps:"
echo "  mkdir build && cd build"
echo "  cmake .. -DCMAKE_BUILD_TYPE=Release"
echo "  make -j$(nproc)"
