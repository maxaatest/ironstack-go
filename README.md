# IronStack WP (Go Edition)

**WordPress VPS Control Panel - 100x Speed Edition**

Built with Go + [Bubble Tea](https://github.com/charmbracelet/bubbletea) for beautiful terminal UI.

## 🚀 Install

```bash
curl -sSL https://raw.githubusercontent.com/maxaatest/ironstack-go/main/install.sh | bash
```

## Features

- 🎨 **Beautiful TUI** - Bubble Tea powered interface
- ⚡ **100x Speed** - Varnish + DragonflyDB caching
- 🔒 **Security** - CSF + Fail2ban
- 📊 **Analytics** - GoAccess real-time stats

## Stack

| Component | Purpose |
|-----------|---------|
| Caddy | Auto SSL |
| Varnish | Page cache |
| FrankenPHP | Fast PHP |
| DragonflyDB | Object cache |
| MariaDB | Database |
| CSF + Fail2ban | Security |
| GoAccess | Analytics |

## Build

```bash
go build -o ironstack
./ironstack
```

## Screenshot

```
╭──────────────────────────────────────────────────────╮
│  IRONSTACK WP                                        │
│  WordPress VPS Control Panel - 100x Speed            │
│                                                      │
│  > 🚀 Install Full Stack                             │
│    🌐 Add WordPress Site                             │
│    ⚡ WordPress Tools                                │
│    📦 Cache Management                               │
│    🗄️  Database                                      │
│    🔒 Security                                       │
│    📊 Analytics                                      │
│    💾 Backup & Restore                              │
│    📈 Server Status                                  │
│    ❌ Exit                                           │
│                                                      │
│  ↑/↓ Navigate • Enter Select • q Quit              │
╰──────────────────────────────────────────────────────╯
```

## License

MIT
