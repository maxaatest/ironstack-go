package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Colors
var (
	Purple    = lipgloss.Color("#7D56F4")
	Green     = lipgloss.Color("#04B575")
	Red       = lipgloss.Color("#FF5555")
	Yellow    = lipgloss.Color("#FFCC00")
	White     = lipgloss.Color("#FAFAFA")
	Gray      = lipgloss.Color("#3C3C3C")
	LightGray = lipgloss.Color("#DDDDDD")
)

// Styles
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(White).
			Background(Purple).
			Padding(0, 2)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(White).
			Background(Green).
			Padding(0, 1)

	DocStyle = lipgloss.NewStyle().Margin(1, 2)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(Gray).
			Italic(true)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Purple).
			Padding(1, 2)
)

// Item represents a menu item
type Item struct {
	Title string
	Desc  string
}

func (i Item) FilterValue() string { return i.Title }

// ItemDelegate handles item rendering
type ItemDelegate struct{}

func (d ItemDelegate) Height() int                             { return 2 }
func (d ItemDelegate) Spacing() int                            { return 0 }
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d ItemDelegate) Render(w interface{ WriteString(string) (int, error) }, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s\n%s", i.Title, lipgloss.NewStyle().Foreground(LightGray).Render(i.Desc))

	fn := lipgloss.NewStyle().PaddingLeft(2).Render
	if index == m.Index() {
		fn = func(s string) string {
			return lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(White).
				Background(Purple).
				Bold(true).
				Render("> " + s)
		}
	}

	w.WriteString(fn(str))
}

// NewSpinner creates a styled spinner
func NewSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(Purple)
	return s
}

// Banner returns the ASCII logo
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
