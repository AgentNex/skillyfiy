#!/bin/sh
set -e

REPO="AgentNex/skillyfiy"

# 1. Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  linux*)  OS="linux" ;;
  darwin*) OS="darwin" ;;
  *) echo "Error: Unsupported operating system: $OS"; exit 1 ;;
esac

# 2. Detect Architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Error: Unsupported CPU architecture: $ARCH"; exit 1 ;;
esac

# 3. Detect Installation Directory (Termux vs Standard Linux/Mac)
if [ -n "$PREFIX" ] && [ -d "$PREFIX/bin" ]; then
  INSTALL_DIR="$PREFIX/bin"
elif [ -w "/usr/local/bin" ]; then
  INSTALL_DIR="/usr/local/bin"
else
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
fi

# 4. Fetch Latest Release Version from GitHub API
TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$TAG" ]; then
  TAG="v0.1.0"
fi

TARBALL="skillyfiy_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$TAG/$TARBALL"

echo "➜ Downloading Skillyfiy ${TAG} for ${OS}/${ARCH}..."
TMP_DIR="$(mktemp -d)"

if curl -sL "$URL" | tar -xz -C "$TMP_DIR" 2>/dev/null; then
  mv "$TMP_DIR/skillyfiy" "$INSTALL_DIR/skillyfiy"
  chmod +x "$INSTALL_DIR/skillyfiy"
  rm -rf "$TMP_DIR"
  echo "✓ Successfully installed Skillyfiy to $INSTALL_DIR/skillyfiy"
  echo "➜ Run 'skillyfiy --version' or 'skillyfiy' to start."
else
  echo "Falling back to local go install..."
  go install "github.com/$REPO@latest"
fi
