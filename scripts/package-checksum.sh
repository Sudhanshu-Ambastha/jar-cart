ASSET_NAME="$1"
OS_TYPE="$2"

if [ "$OS_TYPE" == "windows" ]; then
  certutil -hashfile "$ASSET_NAME" SHA256 | grep -v "SHA256" | grep -v "CertUtil" | tr -d '\r\n ' > "${ASSET_NAME}.sha256"
else
  shasum -a 256 "$ASSET_NAME" | awk '{print $1}' > "${ASSET_NAME}.sha256"
fi