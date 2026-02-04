#!/bin/bash
#
# IronStack WP - 1-Click Installer
# WordPress VPS Control Panel
#

set -e

VERSION="1.0.0"
GITHUB_REPO="maxaatest/ironstack-go"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/ironstack"
LOG_DIR="/var/log/ironstack"
BACKUP_DIR="/backups"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
PURPLE='\033[0;35m'
NC='\033[0m'

# Banner
show_banner() {
    echo -e "${PURPLE}"
    cat << "EOF"
 ██╗██████╗  ██████╗ ███╗   ██╗███████╗████████╗ █████╗  ██████╗██╗  ██╗
 ██║██╔══██╗██╔═══██╗████╗  ██║██╔════╝╚══██╔══╝██╔══██╗██╔════╝██║ ██╔╝
 ██║██████╔╝██║   ██║██╔██╗ ██║███████╗   ██║   ███████║██║     █████╔╝ 
 ██║██╔══██╗██║   ██║██║╚██╗██║╚════██║   ██║   ██╔══██║██║     ██╔═██╗ 
 ██║██║  ██║╚██████╔╝██║ ╚████║███████║   ██║   ██║  ██║╚██████╗██║  ██╗
 ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚══════╝   ╚═╝   ╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝
EOF
    echo -e "${NC}"
    echo -e "${CYAN}WordPress VPS Control Panel - v${VERSION}${NC}"
    echo ""
}

# Check requirements
check_requirements() {
    echo -e "${YELLOW}[1/5] Checking requirements...${NC}"
    
    # Must be root
    if [ "$EUID" -ne 0 ]; then
        echo -e "${RED}Error: Please run as root${NC}"
        exit 1
    }
    
    # Check OS
    if [ ! -f /etc/debian_version ] && [ ! -f /etc/lsb-release ]; then
        echo -e "${RED}Error: Only Debian/Ubuntu supported${NC}"
        exit 1
    }
    
    # Check architecture
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        *) echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"; exit 1 ;;
    esac
    
    echo -e "${GREEN}  ✓ Running as root${NC}"
    echo -e "${GREEN}  ✓ OS: $(lsb_release -ds 2>/dev/null || cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)${NC}"
    echo -e "${GREEN}  ✓ Architecture: $ARCH${NC}"
}

# Create directories
setup_directories() {
    echo -e "${YELLOW}[2/5] Setting up directories...${NC}"
    
    mkdir -p $CONFIG_DIR
    mkdir -p $LOG_DIR
    mkdir -p $BACKUP_DIR
    mkdir -p /var/www
    
    echo -e "${GREEN}  ✓ Created $CONFIG_DIR${NC}"
    echo -e "${GREEN}  ✓ Created $LOG_DIR${NC}"
    echo -e "${GREEN}  ✓ Created $BACKUP_DIR${NC}"
}

# Download binary
download_binary() {
    echo -e "${YELLOW}[3/5] Downloading IronStack...${NC}"
    
    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/ironstack-linux-${ARCH}"
    
    if command -v curl &> /dev/null; then
        curl -fsSL "$DOWNLOAD_URL" -o "${INSTALL_DIR}/ironstack"
    elif command -v wget &> /dev/null; then
        wget -q "$DOWNLOAD_URL" -O "${INSTALL_DIR}/ironstack"
    else
        echo -e "${RED}Error: curl or wget required${NC}"
        exit 1
    fi
    
    chmod +x "${INSTALL_DIR}/ironstack"
    echo -e "${GREEN}  ✓ Downloaded to ${INSTALL_DIR}/ironstack${NC}"
}

# Install dependencies
install_dependencies() {
    echo -e "${YELLOW}[4/5] Installing dependencies...${NC}"
    
    apt-get update -qq
    apt-get install -y -qq curl wget ca-certificates gnupg lsb-release > /dev/null 2>&1
    
    echo -e "${GREEN}  ✓ Dependencies installed${NC}"
}

# Configure system
configure_system() {
    echo -e "${YELLOW}[5/5] Configuring system...${NC}"
    
    # Create config file
    cat > "${CONFIG_DIR}/config.yaml" << EOF
# IronStack Configuration
version: ${VERSION}
web_root: /var/www
backup_dir: /backups
log_level: info
EOF
    
    # Add to PATH if not already
    if ! grep -q "ironstack" /etc/profile.d/*.sh 2>/dev/null; then
        echo 'export PATH=$PATH:/usr/local/bin' > /etc/profile.d/ironstack.sh
    fi
    
    echo -e "${GREEN}  ✓ Configuration created${NC}"
}

# Main installation
main() {
    show_banner
    
    check_requirements
    setup_directories
    install_dependencies
    download_binary
    configure_system
    
    echo ""
    echo -e "${GREEN}════════════════════════════════════════════${NC}"
    echo -e "${GREEN}  ✓ IronStack installed successfully!${NC}"
    echo -e "${GREEN}════════════════════════════════════════════${NC}"
    echo ""
    echo -e "  Run: ${CYAN}ironstack${NC}"
    echo ""
    echo -e "  Documentation: ${CYAN}https://github.com/${GITHUB_REPO}${NC}"
    echo ""
}

# Uninstall
uninstall() {
    echo -e "${YELLOW}Uninstalling IronStack...${NC}"
    rm -f "${INSTALL_DIR}/ironstack"
    rm -rf "${CONFIG_DIR}"
    echo -e "${GREEN}IronStack uninstalled${NC}"
}

# Handle arguments
case "${1:-}" in
    uninstall|remove)
        uninstall
        ;;
    *)
        main
        ;;
esac
