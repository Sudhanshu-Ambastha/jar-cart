set -e

OS="$(uname -s)"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64|amd64)   TARGET_ARCH="x86_64" ;;
    aarch64|arm64)  TARGET_ARCH="aarch64" ;;
    *)              TARGET_ARCH="x86_64" ;;
esac

case "$OS" in
    Linux)  PLATFORM="linux"; EXTENSION="tar.gz" ;;
    Darwin) PLATFORM="macos"; EXTENSION="tar.gz" ;;
    *)     
        if echo "$OS" | grep -qE "MINGW|MSYS|CYGWIN"; then
            PLATFORM="windows"; EXTENSION="zip"
        else
            echo "❌ Unsupported OS: $OS"
            exit 1
        fi
        ;;
esac

if [ -z "$VERSION" ]; then
    echo "🔍 Fetching latest version tag..."
    VERSION=$(curl -fsSLH "Accept: application/vnd.github+json" "https://api.github.com/repos/Sudhanshu-Ambastha/jar-cart/releases/latest" | grep '"tag_name":' | head -n 1 | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$VERSION" ] || [ "$VERSION" = "null" ]; then
        echo "❌ Failed to resolve latest version."
        exit 1
    fi
fi

FILE_NAME="jar-cart-${TARGET_ARCH}-${PLATFORM}.${EXTENSION}"
URL="https://github.com/Sudhanshu-Ambastha/jar-cart/releases/download/${VERSION}/${FILE_NAME}"
CHECKSUM_URL="${URL}.sha256"

TMP_DIR="/tmp/jar-cart-install-$$"
INSTALL_DIR="$HOME/.jar-cart/bin"
mkdir -p "$TMP_DIR"
mkdir -p "$INSTALL_DIR"

echo "⚡ Downloading $FILE_NAME ($VERSION)..."
if ! curl -fL "$URL" -o "$TMP_DIR/$FILE_NAME"; then
    echo "❌ Failed to download binary. Check your internet connection."
    exit 1
fi

if ! curl -fL "$CHECKSUM_URL" -o "$TMP_DIR/checksum.sha256"; then
    echo "❌ Failed to download checksum. Check your internet connection."
    exit 1
fi

echo "🛡️ Verifying integrity..."
EXPECTED_HASH=$(cat "$TMP_DIR/checksum.sha256" | tr -d '[:space:]')

if command -v sha256sum >/dev/null 2>&1; then
    ACTUAL_HASH=$(sha256sum "$TMP_DIR/$FILE_NAME" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
    ACTUAL_HASH=$(shasum -a 256 "$TMP_DIR/$FILE_NAME" | awk '{print $1}')
else
    echo "❌ Error: No checksum tool (sha256sum/shasum) found. Installation aborted for security."
    rm -rf "$TMP_DIR"
    exit 1
fi

if [ "$ACTUAL_HASH" != "$EXPECTED_HASH" ]; then
    echo "❌ Hash mismatch! Expected $EXPECTED_HASH, got $ACTUAL_HASH."
    rm -rf "$TMP_DIR"
    exit 1
fi

echo "📦 Unpacking into $INSTALL_DIR..."
if [ "$EXTENSION" = "tar.gz" ]; then
    tar -xzf "$TMP_DIR/$FILE_NAME" -C "$INSTALL_DIR"
else
    unzip -qo "$TMP_DIR/$FILE_NAME" -d "$INSTALL_DIR"
fi

BINARY_NAME="jar-cart"
if [ -f "$INSTALL_DIR/jar-cart" ]; then
    chmod +x "$INSTALL_DIR/jar-cart"
    ln -sf "$INSTALL_DIR/jar-cart" "$INSTALL_DIR/jc"
elif [ -f "$INSTALL_DIR/jar-cart.exe" ]; then
    ln -sf "$INSTALL_DIR/jar-cart.exe" "$INSTALL_DIR/jc.exe"
fi

rm -rf "$TMP_DIR"

echo "---"
echo "✨ jar-cart $VERSION successfully installed!"
echo "👉 Add this to your shell config (~/.bashrc or ~/.zshrc):"
echo '    export PATH="$HOME/.jar-cart/bin:$PATH"'
echo "👉 Then run: source ~/.bashrc (or your config file)"
echo "🚀 You can now use both 'jar-cart' and the short 'jc' alias!"