package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	docStyle = lipgloss.NewStyle().Margin(1, 2)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3C3C3C")).
			Italic(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2)
)

// Menu item
type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// App states
type state int

const (
	stateMenu state = iota
	stateInstalling
	stateSubMenu
)

// Model
type model struct {
	list       list.Model
	spinner    spinner.Model
	state      state
	selected   string
	installing bool
	progress   int
	err        error
}

func initialModel() model {
	items := []list.Item{
		item{title: "🚀 Install Full Stack", desc: "Caddy + Varnish + FrankenPHP + MariaDB + More"},
		item{title: "🌐 Add WordPress Site", desc: "Create new site with auto SSL"},
		item{title: "⚡ WordPress Tools", desc: "Tuning, plugins, updates, security"},
		item{title: "📦 Cache Management", desc: "Varnish + DragonflyDB controls"},
		item{title: "🗄️  Database", desc: "MariaDB management"},
		item{title: "🔒 Security", desc: "CSF + Fail2ban configuration"},
		item{title: "📊 Analytics", desc: "GoAccess real-time stats"},
		item{title: "💾 Backup & Restore", desc: "Full site backups"},
		item{title: "📈 Server Status", desc: "View system resources"},
		item{title: "❌ Exit", desc: "Close IronStack"},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#DDDDDD")).
		Background(lipgloss.Color("#5A3FC0"))

	l := list.New(items, delegate, 60, 20)
	l.Title = "IronStack WP v1.0"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	return model{
		list:    l,
		spinner: s,
		state:   stateMenu,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.state == stateMenu {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.selected = i.title
					if i.title == "❌ Exit" {
						return m, tea.Quit
					}
					if i.title == "🚀 Install Full Stack" {
						m.state = stateInstalling
						m.installing = true
						return m, tea.Batch(m.spinner.Tick, tickInstall())
					}
				}
			}
		case "esc":
			m.state = stateMenu
		}

	case spinner.TickMsg:
		if m.installing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case installProgressMsg:
		m.progress = int(msg)
		if m.progress >= 100 {
			m.installing = false
			m.state = stateMenu
		}
		return m, tickInstall()

	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

type installProgressMsg int

func tickInstall() tea.Cmd {
	return tea.Tick(100*1000000, func(t interface{}) tea.Msg {
		return installProgressMsg(10)
	})
}

func (m model) View() string {
	switch m.state {
	case stateInstalling:
		return m.viewInstalling()
	default:
		return m.viewMenu()
	}
}

func (m model) viewMenu() string {
	header := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render("  IRONSTACK WP  "),
		subtitleStyle.Render(" WordPress VPS Control Panel - 100x Speed "),
		"",
	)

	footer := infoStyle.Render("↑/↓ Navigate • Enter Select • q Quit")

	return docStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			header,
			m.list.View(),
			footer,
		),
	)
}

func (m model) viewInstalling() string {
	components := []string{
		"Caddy Server",
		"FrankenPHP",
		"Varnish Cache",
		"MariaDB",
		"DragonflyDB",
		"WP-CLI",
		"CSF Firewall",
		"Fail2ban",
		"GoAccess",
	}

	progress := ""
	for i, comp := range components {
		if i < m.progress/10 {
			progress += successStyle.Render("  ✓ "+comp) + "\n"
		} else if i == m.progress/10 {
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

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
