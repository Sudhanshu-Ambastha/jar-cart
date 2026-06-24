#!/bin/sh
set -e

BINARY_NAME=$1
ASSET_NAME=$2

if [ -z "$BINARY_NAME" ] || [ -z "$ASSET_NAME" ]; then
    echo "❌ Missing required arguments: package-posix.sh <binary_name> <asset_name>"
    exit 1
fi

echo "🚀 Setting execution permissions on binary..."
chmod +x "$BINARY_NAME"

echo "📦 Compressing $BINARY_NAME into $ASSET_NAME..."
tar -czf "$ASSET_NAME" "$BINARY_NAME"

echo "✨ Successfully packed $ASSET_NAME"