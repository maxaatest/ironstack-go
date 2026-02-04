#!/bin/bash
#
# IronStack WP (Go Edition) - 1-Click Installer
#

set -e

PURPLE='\033[0;35m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${PURPLE}"
echo "██╗██████╗  ██████╗ ███╗   ██╗███████╗████████╗ █████╗  ██████╗██╗  ██╗"
echo "██║██╔══██╗██╔═══██╗████╗  ██║██╔════╝╚══██╔══╝██╔══██╗██╔════╝██║ ██╔╝"
echo "██║██████╔╝██║   ██║██╔██╗ ██║███████╗   ██║   ███████║██║     █████╔╝ "
echo "██║██╔══██╗██║   ██║██║╚██╗██║╚════██║   ██║   ██╔══██║██║     ██╔═██╗ "
echo "██║██║  ██║╚██████╔╝██║ ╚████║███████║   ██║   ██║  ██║╚██████╗██║  ██╗"
echo "╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚══════╝   ╚═╝   ╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝"
echo -e "${NC}"
echo -e "${CYAN}WordPress VPS Control Panel - 100x Speed Edition${NC}"
echo ""

if [ "$EUID" -ne 0 ]; then
    echo "Please run as root"
    exit 1
fi

echo -e "${GREEN}[1/2] Downloading IronStack...${NC}"
curl -fsSL https://github.com/maxaatest/ironstack-go/releases/latest/download/ironstack-linux-amd64 -o /usr/local/bin/ironstack
chmod +x /usr/local/bin/ironstack

echo -e "${GREEN}[2/2] Setting up...${NC}"
mkdir -p /etc/ironstack
mkdir -p /var/log/ironstack
mkdir -p /backups

echo ""
echo -e "${GREEN}✓ IronStack installed!${NC}"
echo ""
echo "Run: ironstack"
echo ""
