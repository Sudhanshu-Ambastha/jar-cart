set -e

TAG_REF="$1"
PAT_TOKEN="$2"

LINUX_SHA256=$(cat ./dist/linux/jar-cart-x86_64-linux.tar.gz.sha256)
MACOS_SHA256=$(cat ./dist/macos/jar-cart-x86_64-macos.tar.gz.sha256)

echo "Triggering Homebrew tap update dispatch for version $TAG_REF..."

curl -X POST \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer $PAT_TOKEN" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  https://api.github.com/repos/Sudhanshu-Ambastha/homebrew-tap/dispatches \
  -d "{
    \"event_type\": \"update-jar-cart\",
    \"client_payload\": {
      \"ref\": \"$TAG_REF\",
      \"linux_sha256\": \"$LINUX_SHA256\",
      \"macos_sha256\": \"$MACOS_SHA256\"
    }
  }"

echo -e "\nDispatch event sent successfully!"