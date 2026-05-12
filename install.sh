#!/bin/sh
set -e

REPO="duclm2609/infra-cli"
BINARY="infra"
INSTALL_DIR="/usr/local/bin"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  darwin) OS="darwin" ;;
  linux)  OS="linux" ;;
  *)      echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)       echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

PLATFORM="${OS}-${ARCH}"

# Get latest release tag
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
if [ -z "$LATEST" ]; then
  echo "Failed to fetch latest release" >&2
  exit 1
fi

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/infra-${PLATFORM}.zip"

echo "Installing infra ${LATEST} (${PLATFORM})..."

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$DOWNLOAD_URL" -o "$TMP/infra.zip"
unzip -q "$TMP/infra.zip" -d "$TMP"
chmod +x "$TMP/${BINARY}"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  sudo mv "$TMP/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi

echo "Installed to ${INSTALL_DIR}/${BINARY}"
infra --version
