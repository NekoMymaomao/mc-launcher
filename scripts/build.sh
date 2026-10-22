#!/bin/bash
set -e

echo "========================================"
echo "     MC Launcher WebUI Build Script"
echo "========================================"
echo ""

# 切换到 Core 目录
cd Core

# 参数
PLATFORM=${1:-"all"}
GOOS=""
GOARCH=""
OUTPUT_NAME="mc-launcher"

case "$PLATFORM" in
    linux-amd64)
        GOOS="linux"
        GOARCH="amd64"
        OUTPUT_NAME="mc-launcher-linux-amd64"
        ;;
    linux-arm64)
        GOOS="linux"
        GOARCH="arm64"
        OUTPUT_NAME="mc-launcher-linux-arm64"
        ;;
    windows-amd64)
        GOOS="windows"
        GOARCH="amd64"
        OUTPUT_NAME="mc-launcher-windows-amd64.exe"
        ;;
    all)
        # 构建所有平台
        for target in linux/amd64 linux/arm64 windows/amd64; do
            IFS='/' read -r GOOS GOARCH <<< "$target"
            if [ "$GOOS" = "windows" ]; then
                OUTPUT_NAME="mc-launcher-windows-amd64.exe"
            else
                OUTPUT_NAME="mc-launcher-${GOOS}-${GOARCH}"
            fi
            echo ""
            echo "Building for ${GOOS}/${GOARCH}..."
            CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags="-s -w" -o "${OUTPUT_NAME}" .
        done
        exit 0
        ;;
    *)
        echo "用法: $0 [linux-amd64|linux-arm64|windows-amd64|all]"
        exit 1
        ;;
esac

# 构建
echo "Building for ${GOOS}/${GOARCH}..."
CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags="-s -w" -o "${OUTPUT_NAME}" .

echo ""
echo "Build complete: ${OUTPUT_NAME}"
echo "Output location: Core/${OUTPUT_NAME}"
