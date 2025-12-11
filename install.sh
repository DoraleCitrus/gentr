#!/bin/sh
set -e

# GitHub 用户名和项目名
OWNER="DoraleCitrus"
REPO="gentr"
BINARY="gentr"

# 1. 检测操作系统
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux)
    OS="linux"
    ;;
  darwin)
    OS="darwin"
    ;;
  mingw*|msys*)
    OS="windows"
    echo "This script does not support Windows directly. Please download the zip from GitHub Releases."
    exit 1
    ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

# 2. 检测架构
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)
    ARCH="amd64"
    ;;
  arm64|aarch64)
    ARCH="arm64"
    ;;
  *)
    echo "Unsupported Architecture: $ARCH"
    exit 1
    ;;
esac

# 3. 获取最新版本号
LATEST_URL="https://api.github.com/repos/$OWNER/$REPO/releases/latest"
VERSION=$(curl -s $LATEST_URL | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
  echo "Error: Could not determine latest version."
  exit 1
fi

echo "Detected System: $OS / $ARCH"
echo "Latest Version: $VERSION"

# 4. 构建下载链接
FILE_NAME="${REPO}_${VERSION#v}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$OWNER/$REPO/releases/download/$VERSION/$FILE_NAME"

echo "Downloading from: $DOWNLOAD_URL"

# 5. 下载并安装
TEMP_DIR=$(mktemp -d)
curl -sL "$DOWNLOAD_URL" -o "$TEMP_DIR/$FILE_NAME"

echo "Extracting..."
tar -xzf "$TEMP_DIR/$FILE_NAME" -C "$TEMP_DIR"

echo "Installing to /usr/local/bin..."
# 需要 sudo 权限
if [ -w "/usr/local/bin" ]; then
    mv "$TEMP_DIR/$BINARY" "/usr/local/bin/$BINARY"
else
    sudo mv "$TEMP_DIR/$BINARY" "/usr/local/bin/$BINARY"
fi

# 清理
rm -rf "$TEMP_DIR"

echo "Successfully installed gentr $VERSION!"
echo "Run 'gentr' to start."