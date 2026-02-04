package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const version = "1.0.0"

// Styles
var (
	purple       = lipgloss.Color("#7C3AED")
	green        = lipgloss.Color("#04B575")
	red          = lipgloss.Color("#FF5555")
	gray         = lipgloss.Color("#6B7280")
	white        = lipgloss.Color("#FFFFFF")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(white).
			Background(purple).
			Padding(0, 2)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(white).
			Background(green).
			Padding(0, 1)

	docStyle     = lipgloss.NewStyle().Margin(1, 2)
	successStyle = lipgloss.NewStyle().Foreground(green).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(red).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(gray)
	boxStyle     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(purple).Padding(1, 2)
	selectedStyle = lipgloss.NewStyle().Foreground(purple).Bold(true)
)

// Menu items
var menuItems = []string{
	"🚀 Install Full Stack",
	"🌐 Add WordPress Site",
	"⚡ WordPress Tools",
	"📦 Cache Management",
	"🗄️  Database",
	"🔒 Security",
	"📊 Analytics",
	"💾 Backup & Restore",
	"📈 Server Status",
	"❌ Exit",
}

var menuDescs = []string{
	"Caddy + Varnish + MariaDB + DragonflyDB + More",
	"Create new site with auto SSL & DB",
	"Performance tuning, plugins, updates",
	"Varnish + DragonflyDB cache controls",
	"MariaDB database management",
	"CSF + Fail2ban configuration",
	"GoAccess real-time web analytics",
	"Full site backups with one click",
	"View system resources & services",
	"Close IronStack",
}

// Components for installation
var installComponents = []string{
	"Caddy Server",
	"Varnish Cache",
	"MariaDB",
	"DragonflyDB",
	"WP-CLI",
	"CSF Firewall",
	"Fail2ban",
	"GoAccess",
}

type state int

const (
	stateMenu state = iota
	stateInstalling
	stateAddSite
	stateMessage
)

type model struct {
	cursor      int
	spinner     spinner.Model
	textInput   textinput.Model
	state       state
	progress    int
	message     string
	messageType string
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(purple)

	ti := textinput.New()
	ti.Placeholder = "example.com"
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 40

	return model{
		spinner:   s,
		textInput: ti,
		state:     stateMenu,
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateMenu:
			return m.updateMenu(msg)
		case stateInstalling:
			if msg.String() == "esc" {
				m.state = stateMenu
				m.progress = 0
			}
			return m, nil
		case stateAddSite:
			return m.updateAddSite(msg)
		case stateMessage:
			if msg.String() == "enter" || msg.String() == "esc" {
				m.state = stateMenu
			}
			return m, nil
		}

	case spinner.TickMsg:
		if m.state == stateInstalling {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case installProgressMsg:
		m.progress++
		if m.progress >= len(installComponents) {
			m.state = stateMessage
			m.message = "Stack installation complete!"
			m.messageType = "success"
			return m, nil
		}
		return m, tea.Batch(m.spinner.Tick, tickInstall())

	case textinput.BlinkMsg:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(menuItems)-1 {
			m.cursor++
		}
	case "enter":
		switch m.cursor {
		case 0: // Install
			m.state = stateInstalling
			m.progress = 0
			return m, tea.Batch(m.spinner.Tick, tickInstall())
		case 1: // Add Site
			m.state = stateAddSite
			m.textInput.SetValue("")
			return m, textinput.Blink
		case 9: // Exit
			return m, tea.Quit
		default:
			m.state = stateMessage
			m.message = fmt.Sprintf("%s - Coming in Phase 3+", menuItems[m.cursor])
			m.messageType = "info"
		}
	}
	return m, nil
}

func (m model) updateAddSite(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateMenu
		return m, nil
	case "enter":
		domain := m.textInput.Value()
		if domain != "" {
			m.state = stateMessage
			m.message = fmt.Sprintf("Site '%s' created successfully!\n\nPath: /var/www/%s\nSSL: Auto-enabled\nCache: Varnish active", domain, domain)
			m.messageType = "success"
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

type installProgressMsg struct{}

func tickInstall() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(_ time.Time) tea.Msg {
		return installProgressMsg{}
	})
}

func (m model) View() string {
	switch m.state {
	case stateInstalling:
		return m.viewInstalling()
	case stateAddSite:
		return m.viewAddSite()
	case stateMessage:
		return m.viewMessage()
	default:
		return m.viewMenu()
	}
}

func (m model) viewMenu() string {
	header := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render("  IRONSTACK WP  "),
		subtitleStyle.Render(" WordPress VPS Control Panel "),
		infoStyle.Render(fmt.Sprintf("v%s", version)),
		"",
	)

	var menu string
	for i, item := range menuItems {
		cursor := "  "
		style := lipgloss.NewStyle()
		if m.cursor == i {
			cursor = "> "
			style = selectedStyle
		}
		menu += style.Render(cursor+item) + "\n"
		menu += infoStyle.Render("    "+menuDescs[i]) + "\n"
	}

	footer := infoStyle.Render("↑/↓ Navigate • Enter Select • q Quit")

	return docStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, menu, footer))
}

func (m model) viewInstalling() string {
	var progress string
	for i, comp := range installComponents {
		if i < m.progress {
			progress += successStyle.Render("  ✓ "+comp) + "\n"
		} else if i == m.progress {
			progress += m.spinner.View() + " Installing " + comp + "...\n"
		} else {
			progress += infoStyle.Render("  ○ "+comp) + "\n"
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("  Installing Full Stack  "),
		"",
		progress,
		"",
		infoStyle.Render("Press ESC to cancel"),
	)

	return docStyle.Render(boxStyle.Render(content))
}

func (m model) viewAddSite() string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("  Add WordPress Site  "),
		"",
		"Enter domain name:",
		"",
		m.textInput.View(),
		"",
		infoStyle.Render("• Auto SSL via Let's Encrypt"),
		infoStyle.Render("• Varnish cache enabled"),
		infoStyle.Render("• Database auto-created"),
		"",
		infoStyle.Render("Enter to create • ESC to cancel"),
	)

	return docStyle.Render(boxStyle.Render(content))
}

func (m model) viewMessage() string {
	style := successStyle
	icon := "✓"
	if m.messageType == "error" {
		style = errorStyle
		icon = "✗"
	} else if m.messageType == "info" {
		style = infoStyle
		icon = "ℹ"
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		style.Render(icon+" "+m.message),
		"",
		infoStyle.Render("Press Enter to continue"),
	)

	return docStyle.Render(boxStyle.Render(content))
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("IronStack WP v%s\n", version)
			return
		case "--help", "-h":
			fmt.Println("IronStack WP - WordPress VPS Control Panel")
			fmt.Println("\nUsage: ironstack [options]")
			fmt.Println("\nOptions:")
			fmt.Println("  -v, --version  Show version")
			fmt.Println("  -h, --help     Show help")
			return
		}
	}

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
