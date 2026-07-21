set -e

TAG_REF="$1"
LINUX_SHA256="$2"
MACOS_SHA256="$3"

VERSION="${TAG_REF#v}"
MAJOR_MINOR=$(echo "$VERSION" | cut -d. -f1,2)
CLASS_SUFFIX=$(echo "$MAJOR_MINOR" | tr -d '.')

cat > /tmp/formula_body.txt << EOF
  homepage "https://github.com/Sudhanshu-Ambastha/jar-cart"
  license "MIT"
  on_macos do
    url "https://github.com/Sudhanshu-Ambastha/jar-cart/releases/download/${TAG_REF}/jar-cart-x86_64-macos.tar.gz"
    sha256 "${MACOS_SHA256}"
  end
  on_linux do
    url "https://github.com/Sudhanshu-Ambastha/jar-cart/releases/download/${TAG_REF}/jar-cart-x86_64-linux.tar.gz"
    sha256 "${LINUX_SHA256}"
  end
  def install
    bin.install "jar-cart"
    bin.install_symlink bin/"jar-cart" => "jc"
  end
  test do
    assert_match "jar-cart v${VERSION}", shell_output("#{bin}/jar-cart --version")
  end
EOF

{
  echo 'class JarCart < Formula'
  echo '  desc "Lightning-fast, no-build Java build orchestrator"'
  cat /tmp/formula_body.txt
  echo 'end'
} > tap/Formula/jar-cart.rb

{
  echo "class JarCartAT${CLASS_SUFFIX} < Formula"
  echo "  desc \"Lightning-fast, no-build Java build orchestrator (${MAJOR_MINOR}.x line)\""
  echo '  keg_only "this is a pinned version; install jar-cart for the latest release"'
  cat /tmp/formula_body.txt
  echo 'end'
} > "tap/Formula/jar-cart@${MAJOR_MINOR}.rb"