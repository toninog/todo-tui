package styles

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/blackraven/todo-tui/internal/themes"
)

var (
	AppStyle          lipgloss.Style
	HeaderStyle       lipgloss.Style
	ListSelectedStyle lipgloss.Style
	ListItemStyle     lipgloss.Style
	InlineInputStyle  lipgloss.Style
	StrikeStyle       lipgloss.Style
	BinaryStyle       lipgloss.Style
	HelpStyle         lipgloss.Style
	DueStyle          lipgloss.Style
	OverdueStyle      lipgloss.Style
	ErrorStyle        lipgloss.Style
	SuccessStyle      lipgloss.Style
	CategoryStyle     lipgloss.Style
	PriorityHighStyle lipgloss.Style
	PriorityMedStyle  lipgloss.Style
	PriorityLowStyle  lipgloss.Style
	TabActiveStyle    lipgloss.Style
	TabInactiveStyle  lipgloss.Style
	InputLabelStyle   lipgloss.Style
	InputFieldStyle   lipgloss.Style
)

// Update updates all styles based on the given theme
func Update(t themes.Theme) {
	AppStyle = lipgloss.NewStyle().Padding(1).Background(t.Bg)

	HeaderStyle = lipgloss.NewStyle().
		Foreground(t.Bg).
		Background(t.Accent).
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)

	ListSelectedStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Accent).
		Padding(0, 1).
		Foreground(t.Accent).
		Bold(true)

	ListItemStyle = lipgloss.NewStyle().
		PaddingLeft(4).
		Foreground(t.Fg)

	InlineInputStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Bold(true)

	StrikeStyle = lipgloss.NewStyle().Foreground(t.Dim).Strikethrough(true)
	BinaryStyle = lipgloss.NewStyle().Foreground(t.Warning).Bold(true)
	HelpStyle = lipgloss.NewStyle().Foreground(t.Dim)

	DueStyle = lipgloss.NewStyle().Foreground(t.Secondary).Italic(true)
	OverdueStyle = lipgloss.NewStyle().Foreground(t.Warning).Bold(true).Blink(true)

	ErrorStyle = lipgloss.NewStyle().Foreground(t.Warning).Bold(true)
	SuccessStyle = lipgloss.NewStyle().Foreground(t.Success).Bold(true)

	CategoryStyle = lipgloss.NewStyle().
		Padding(0, 1).
		MarginRight(1)

	PriorityHighStyle = lipgloss.NewStyle().Foreground(t.Warning).Bold(true)
	PriorityMedStyle = lipgloss.NewStyle().Foreground(t.Secondary)
	PriorityLowStyle = lipgloss.NewStyle().Foreground(t.Dim)

	TabActiveStyle = lipgloss.NewStyle().
		Foreground(t.Bg).
		Background(t.Accent).
		Padding(0, 2).
		Bold(true)

	TabInactiveStyle = lipgloss.NewStyle().
		Foreground(t.Dim).
		Background(t.Bg).
		Padding(0, 2)

	InputLabelStyle = lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	InputFieldStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Accent).
		Padding(0, 1)
}

// Init initializes styles with the first theme
func Init() {
	if len(themes.All) > 0 {
		Update(themes.All[0])
	}
}
