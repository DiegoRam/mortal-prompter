#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}"
echo "═══════════════════════════════════════════════════════════"
echo "  MORTAL PROMPTER - Installer"
echo "═══════════════════════════════════════════════════════════"
echo -e "${NC}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo -e "${RED}Unsupported architecture: $ARCH${NC}"; exit 1 ;;
esac

case $OS in
    darwin|linux) ;;
    *) echo -e "${RED}Unsupported OS: $OS${NC}"; exit 1 ;;
esac

# Get latest version
REPO="minimalart/mortal-prompter"
LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    echo -e "${RED}Could not fetch latest version${NC}"
    exit 1
fi

echo -e "${YELLOW}Installing mortal-prompter $LATEST_VERSION for $OS/$ARCH...${NC}"

# Build download URL
FILENAME="mortal-prompter_${LATEST_VERSION#v}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_VERSION/$FILENAME"

# Temporary directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download and extract
echo "Downloading from $DOWNLOAD_URL..."
curl -sL "$DOWNLOAD_URL" -o "$TMP_DIR/mortal-prompter.tar.gz"

if [ ! -s "$TMP_DIR/mortal-prompter.tar.gz" ]; then
    echo -e "${RED}Download failed or file is empty${NC}"
    exit 1
fi

tar -xzf "$TMP_DIR/mortal-prompter.tar.gz" -C "$TMP_DIR"

# Install
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    echo -e "${YELLOW}Requires administrator permissions to install in $INSTALL_DIR${NC}"
    sudo mv "$TMP_DIR/mortal-prompter" "$INSTALL_DIR/"
else
    mv "$TMP_DIR/mortal-prompter" "$INSTALL_DIR/"
fi

chmod +x "$INSTALL_DIR/mortal-prompter"

echo -e "${GREEN}"
echo "═══════════════════════════════════════════════════════════"
echo "  INSTALLATION COMPLETE!"
echo "═══════════════════════════════════════════════════════════"
echo -e "${NC}"
echo "Run 'mortal-prompter --help' to get started"
echo ""
echo -e "${YELLOW}FIGHT!${NC}"
