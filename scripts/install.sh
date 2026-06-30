#!/bin/sh
set -e

OS="$(uname -s)"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64|amd64)   TARGET_ARCH="x86_64" ;;
    aarch64|arm64)  TARGET_ARCH="aarch64" ;;
    *)              TARGET_ARCH="x86_64" ;;
esac

if [ "$OS" = "Linux" ]; then
    PLATFORM="linux"; EXTENSION="tar.gz"
elif [ "$OS" = "Darwin" ]; then
    PLATFORM="macos"; EXTENSION="tar.gz"
elif echo "$OS" | grep -qE "MINGW|MSYS|CYGWIN"; then
    PLATFORM="windows"; EXTENSION="zip"
else
    echo "❌ Unsupported OS: $OS"
    exit 1
fi

if [ -z "$VERSION" ]; then
    echo "🔍 Fetching latest version tag..."
    VERSION=$(curl -s "https://api.github.com/repos/Sudhanshu-Ambastha/jar-cart/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$VERSION" ] || [ "$VERSION" = "null" ]; then
        echo "❌ Failed to resolve latest version. Please set the VERSION environment variable."
        exit 1
    fi
fi

FILE_NAME="jar-cart-${TARGET_ARCH}-${PLATFORM}.${EXTENSION}"
URL="https://github.com/Sudhanshu-Ambastha/jar-cart/releases/download/${VERSION}/${FILE_NAME}"
INSTALL_DIR="$HOME/.jar-cart/bin"

echo "⚡ Downloading $FILE_NAME ($VERSION)..."
mkdir -p "$INSTALL_DIR"

if [ "$EXTENSION" = "tar.gz" ]; then
    curl -qLsSf "$URL" | tar -xz -C "$INSTALL_DIR"
else
    TEMP_ZIP="/tmp/jar-cart.zip"
    curl -qLsSf "$URL" -o "$TEMP_ZIP"
    if command -v unzip >/dev/null 2>&1; then
        unzip -qo "$TEMP_ZIP" -d "$INSTALL_DIR"
    else
        powershell.exe -NoProfile -Command "Expand-Archive -Path '$TEMP_ZIP' -DestinationPath '$INSTALL_DIR' -Force"
    fi
    rm -f "$TEMP_ZIP"
fi

echo "---"
echo "✨ jar-cart $VERSION successfully installed!"
if [ "$PLATFORM" = "windows" ]; then
    echo "🚀 Make sure $INSTALL_DIR is in your PATH."
else
    echo "👉 Add this to your shell config (~/.bashrc or ~/.zshrc):"
    echo "   export PATH=\"\$HOME/.jar-cart/bin:\$PATH\""
fi