#!/bin/sh
set -e

# 1. Platform Detection
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

# 2. Fetch Version
if [ -z "$VERSION" ]; then
    echo "🔍 Fetching latest version tag..."
    VERSION=$(curl -s "https://api.github.com/repos/Sudhanshu-Ambastha/jar-cart/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$VERSION" ] || [ "$VERSION" = "null" ]; then
        echo "❌ Failed to resolve latest version."
        exit 1
    fi
fi

FILE_NAME="jar-cart-${TARGET_ARCH}-${PLATFORM}.${EXTENSION}"
URL="https://github.com/Sudhanshu-Ambastha/jar-cart/releases/download/${VERSION}/${FILE_NAME}"
CHECKSUM_URL="https://github.com/Sudhanshu-Ambastha/jar-cart/releases/download/${VERSION}/checksums.txt"

# Use /tmp for staging to keep the live bin directory safe until verification passes
TMP_DIR="/tmp/jar-cart-update"
mkdir -p "$TMP_DIR"
INSTALL_DIR="$HOME/.jar-cart/bin"
mkdir -p "$INSTALL_DIR"

# 3. Download to Temp
echo "⚡ Downloading $FILE_NAME ($VERSION)..."
curl -qLsSf "$URL" -o "$TMP_DIR/$FILE_NAME"
curl -qLsSf "$CHECKSUM_URL" -o "$TMP_DIR/checksums.txt"

# 4. Verify
echo "🛡️ Verifying integrity..."
EXPECTED_HASH=$(grep "$FILE_NAME" "$TMP_DIR/checksums.txt" | awk '{print $1}')

if command -v sha256sum >/dev/null 2>&1; then
    ACTUAL_HASH=$(sha256sum "$TMP_DIR/$FILE_NAME" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
    ACTUAL_HASH=$(shasum -a 256 "$TMP_DIR/$FILE_NAME" | awk '{print $1}')
else
    echo "⚠️ Warning: No checksum tool found. Skipping verification."
    ACTUAL_HASH="$EXPECTED_HASH"
fi

if [ "$ACTUAL_HASH" != "$EXPECTED_HASH" ]; then
    echo "❌ Hash mismatch! Update aborted. Your existing version remains untouched."
    rm -rf "$TMP_DIR"
    exit 1
fi

# 5. Atomic Update (Only happens after verification)
echo "📦 Unpacking..."
# Clean old files just before swapping
rm -f "$INSTALL_DIR/jar-cart" "$INSTALL_DIR/checksums.txt"

if [ "$EXTENSION" = "tar.gz" ]; then
    tar -xzf "$TMP_DIR/$FILE_NAME" -C "$INSTALL_DIR"
else
    unzip -qo "$TMP_DIR/$FILE_NAME" -d "$INSTALL_DIR"
fi

# Cleanup temp folder
rm -rf "$TMP_DIR"

echo "---"
echo "✨ jar-cart $VERSION successfully installed!"
echo "👉 Add this to your shell config (~/.bashrc or ~/.zshrc):"
echo '    export PATH="$HOME/.jar-cart/bin:$PATH"'