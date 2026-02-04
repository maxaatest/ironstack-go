# IronStack WP Documentation

## Quick Start

### Installation

```bash
curl -sSL https://raw.githubusercontent.com/maxaatest/ironstack-go/main/install.sh | bash
```

### First Run

```bash
ironstack
```

## Commands

### CLI Flags

```bash
ironstack --help      # Show help
ironstack --version   # Show version
```

## Features

### 1. Full Stack Installation
Installs and configures:
- Caddy (web server with auto SSL)
- Varnish (page cache)
- MariaDB (database)
- DragonflyDB (object cache)
- WP-CLI (WordPress management)
- CSF (firewall)
- Fail2ban (brute-force protection)
- GoAccess (analytics)

### 2. Site Management
- Create WordPress sites with auto SSL
- Clone sites for staging
- Push staging to production
- Domain aliases

### 3. WordPress Tools
- Auto performance tuning
- Plugin/theme management
- Database optimization
- Cache management

### 4. Security
- IP blocking/allowing
- Port management
- WordPress hardening
- Fail2ban jails

### 5. Backup & Restore
- Full site backups
- Database-only backups
- One-click restore

### 6. Monitoring
- Real-time server stats
- Service status
- GoAccess analytics
- Alert system

## Configuration

Config file: `/etc/ironstack/config.yaml`

```yaml
version: 1.0.0
web_root: /var/www
backup_dir: /backups
log_level: info
```

## Directory Structure

```
/var/www/              # Site files
  └── example.com/
      ├── public/      # WordPress files
      ├── logs/        # Access logs
      └── backups/     # Site backups

/etc/ironstack/        # Configuration
/var/log/ironstack/    # Logs
/backups/              # Global backups
```

## API Reference

See [internal/](./internal/) for Go packages:

- `installer/` - Component installation
- `config/` - Configuration generation
- `site/` - Site management
- `wordpress/` - WP-CLI wrapper
- `security/` - CSF & Fail2ban
- `backup/` - Backup system
- `cache/` - Cache management
- `monitoring/` - Server monitoring

## Building from Source

```bash
# Clone
git clone https://github.com/maxaatest/ironstack-go.git
cd ironstack-go

# Build
make build

# Or build for Linux
make linux

# Install
sudo make install
```

## Requirements

- Ubuntu 20.04+ or Debian 11+
- 1GB RAM minimum (2GB recommended)
- Root access

## Support

GitHub Issues: https://github.com/maxaatest/ironstack-go/issues

## License

MIT License
