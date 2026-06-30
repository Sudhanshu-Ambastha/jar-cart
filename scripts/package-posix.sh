#!/bin/sh
set -e

BINARY_NAME=$1
ASSET_NAME=$2

if [ -z "$BINARY_NAME" ] || [ -z "$ASSET_NAME" ]; then
    echo "❌ Usage: package-posix.sh <binary_name> <asset_name>"
    exit 1
fi

if [ ! -f "$BINARY_NAME" ]; then
    echo "❌ Binary $BINARY_NAME not found!"
    exit 1
fi

echo "🚀 Setting execution permissions..."
chmod +x "$BINARY_NAME"

echo "📦 Compressing $BINARY_NAME into $ASSET_NAME..."
tar -czf "$ASSET_NAME" "$BINARY_NAME"

echo "✨ Successfully packed $ASSET_NAME"