set -e

OS="$(uname -s)"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64|amd64)   TARGET_ARCH="x86_64" ;;
    aarch64|arm64)  TARGET_ARCH="aarch64" ;;
    *)              TARGET_ARCH="x86_64" ;;
esac

if [ "$OS" = "Linux" ]; then
    PLATFORM="linux"
    EXTENSION="tar.gz"
elif [ "$OS" = "Darwin" ]; then
    PLATFORM="macos"
    EXTENSION="tar.gz"
elif echo "$OS" | grep -qE "MINGW|MSYS|CYGWIN"; then
    PLATFORM="windows"
    EXTENSION="zip"
else
    echo "❌ Unsupported OS: $OS"
    exit 1
fi

if [ -z "$VERSION" ]; then
    echo "🔍 Fetching latest version tag from GitHub API..."
    VERSION=$(curl -s "https://api.github.com/repos/Sudhanshu-Ambastha/jar-cart/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$VERSION" ] || [ "$VERSION" = "null" ]; then
        echo "⚠️ Could not resolve latest version automatically. Falling back to default v0.0.1"
        VERSION="v0.0.1"
    fi
fi

FILE_NAME="jar-cart-${TARGET_ARCH}-${PLATFORM}.${EXTENSION}"
URL="https://github.com/Sudhanshu-Ambastha/jar-cart/releases/download/${VERSION}/${FILE_NAME}"
INSTALL_DIR="$HOME/.jar-cart/bin"

echo "⚡ Downloading jar-cart $VERSION for $OS ($ARCH)..."
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
echo "✨ jar-cart successfully installed into $INSTALL_DIR"
if [ "$PLATFORM" = "windows" ]; then
    echo "🚀 Make sure $INSTALL_DIR is in your system's Environment Variables path!"
else
    echo "👉 Add this line to your ~/.bashrc or ~/.zshrc file:"
    echo "   export PATH=\"\$HOME/.jar-cart/bin:\$PATH\""
fi