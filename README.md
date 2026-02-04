<p align="center">
  <img src="https://img.shields.io/badge/version-1.0.0-7C3AED?style=for-the-badge" alt="Version">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/platform-Linux-04B575?style=for-the-badge&logo=linux" alt="Platform">
  <img src="https://img.shields.io/badge/license-MIT-blue?style=for-the-badge" alt="License">
</p>

<h1 align="center">⚡ IronStack WP</h1>

<p align="center">
  <strong>WordPress VPS Control Panel - 100x Speed Edition</strong><br>
  Built with Go + Bubble Tea for beautiful terminal UI
</p>

---

## � About

IronStack WP is a high-performance WordPress VPS control panel that combines the best open-source technologies into a single, easy-to-use binary. It provides automated installation, configuration, and management of your WordPress hosting stack.

**Why IronStack?**
- �🚀 **Fast** - Single binary, no dependencies, starts in milliseconds
- 🎯 **Simple** - Interactive TUI, no complex commands to memorize
- 🔒 **Secure** - Built-in firewall, fail2ban, WordPress hardening
- 📊 **Smart** - Auto-tuning, caching, monitoring out of the box

---

## 🚀 Quick Start

```bash
curl -sSL https://raw.githubusercontent.com/maxaatest/ironstack-go/main/install.sh | bash
```

Then:
```bash
ironstack
```

---

## ✨ Features

| Category | Features |
|----------|----------|
| **Web Server** | Caddy + FrankenPHP, Auto SSL, Varnish Cache |
| **Database** | MariaDB, DragonflyDB (25x faster Redis) |
| **WordPress** | WP-CLI, Auto-tune, Object Cache, WooCommerce |
| **Security** | CSF Firewall, Fail2ban, WordPress Hardening |
| **Management** | Site Cloning, Staging, Domain Aliases |
| **Monitoring** | GoAccess Analytics, Server Stats, Alerts |
| **Backup** | Full/DB/Files Backups, One-click Restore |

---

## 📊 Comparison

| Feature | IronStack | CentminMod | WordOps | SlickStack |
|---------|:---------:|:----------:|:-------:|:----------:|
| Single Binary | ✅ | ❌ | ❌ | ❌ |
| Interactive TUI | ✅ | ✅ | ❌ | ❌ |
| Auto SSL | ✅ | ✅ | ✅ | ✅ |
| Object Cache | ✅ | ❌ | ✅ | ✅ |
| Built-in Firewall | ✅ | ✅ | ❌ | ❌ |
| Memory Usage | ~10MB | ~200MB | ~150MB | ~100MB |

---

## 🛠️ Build from Source

```bash
git clone https://github.com/maxaatest/ironstack-go.git
cd ironstack-go
make linux
sudo make install
```

---

## 📖 Documentation

See [docs/DOCUMENTATION.md](docs/DOCUMENTATION.md)

---

## 📜 License

MIT © IronStack
