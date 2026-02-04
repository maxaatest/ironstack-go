package ui

import "github.com/charmbracelet/lipgloss"

// Colors
var (
	Purple    = lipgloss.Color("#7D56F4")
	Green     = lipgloss.Color("#04B575")
	Red       = lipgloss.Color("#FF5555")
	Yellow    = lipgloss.Color("#FFCC00")
	Blue      = lipgloss.Color("#00AAFF")
	Gray      = lipgloss.Color("#3C3C3C")
	White     = lipgloss.Color("#FAFAFA")
	LightGray = lipgloss.Color("#DDDDDD")
)

// Styles
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(White).
			Background(Purple).
			Padding(0, 2).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(White).
			Background(Green).
			Padding(0, 1)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Purple).
			Padding(1, 2)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Yellow).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(Gray).
			Italic(true)

	HighlightStyle = lipgloss.NewStyle().
			Foreground(Purple).
			Bold(true)

	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(White).
				Background(Purple).
				Padding(0, 1)

	TableRowStyle = lipgloss.NewStyle().
			Padding(0, 1)

	StatusRunning = lipgloss.NewStyle().
			Foreground(Green).
			SetString("● running")

	StatusStopped = lipgloss.NewStyle().
			Foreground(Red).
			SetString("○ stopped")
)

// Banner returns the IronStack ASCII banner
func Banner() string {
	banner := `
██╗██████╗  ██████╗ ███╗   ██╗███████╗████████╗ █████╗  ██████╗██╗  ██╗
██║██╔══██╗██╔═══██╗████╗  ██║██╔════╝╚══██╔══╝██╔══██╗██╔════╝██║ ██╔╝
██║██████╔╝██║   ██║██╔██╗ ██║███████╗   ██║   ███████║██║     █████╔╝ 
██║██╔══██╗██║   ██║██║╚██╗██║╚════██║   ██║   ██╔══██║██║     ██╔═██╗ 
██║██║  ██║╚██████╔╝██║ ╚████║███████║   ██║   ██║  ██║╚██████╗██║  ██╗
╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚══════╝   ╚═╝   ╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝`

	return lipgloss.NewStyle().Foreground(Purple).Bold(true).Render(banner)
}
