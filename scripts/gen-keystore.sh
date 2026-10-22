#!/bin/bash
# 生成 APK 签名密钥
# 用法: bash gen-keystore.sh [keystore_path]

KEYSSTORE_PATH="${1:-release.keystore}"
KEYSTORE_PASSWORD="android"
KEY_ALIAS="mc-launcher"
KEY_PASSWORD="android"

# 生成密钥
keytool -genkeypair -v \
  -keystore "$KEYSSTORE_PATH" \
  -keyalg RSA -keysize 2048 -validity 10000 \
  -alias "$KEY_ALIAS" \
  -storepass "$KEYSTORE_PASSWORD" \
  -keypass "$KEY_PASSWORD" \
  -dname "CN=MC Launcher, OU=Dev, O=Org, L=City, S=State, C=CN" 2>&1

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ 密钥已生成: $KEYSSTORE_PATH"
    echo "  Store password: $KEYSTORE_PASSWORD"
    echo "  Key password:   $KEY_PASSWORD"
    echo "  Alias:          $KEY_ALIAS"
else
    echo "✗ 密钥生成失败"
    exit 1
fi
