package styles

import "github.com/charmbracelet/lipgloss"

var (
	Primary   = lipgloss.Color("#7C3AED")
	Secondary = lipgloss.Color("#06B6D4")
	Success   = lipgloss.Color("#10B981")
	Warning   = lipgloss.Color("#F59E0B")
	Danger    = lipgloss.Color("#EF4444")
	Muted     = lipgloss.Color("#6B7280")
	Text      = lipgloss.Color("#F9FAFB")
	TextDim   = lipgloss.Color("#9CA3AF")

	TitleStyle       = lipgloss.NewStyle().Bold(true).Foreground(Primary).MarginBottom(1)
	SubtitleStyle    = lipgloss.NewStyle().Foreground(Secondary).MarginBottom(1)
	DescriptionStyle = lipgloss.NewStyle().Foreground(TextDim)
	SelectedStyle    = lipgloss.NewStyle().Foreground(Primary).Bold(true)
	NormalStyle      = lipgloss.NewStyle().Foreground(Text)
	DimStyle         = lipgloss.NewStyle().Foreground(Muted)
	SuccessStyle     = lipgloss.NewStyle().Foreground(Success)
	WarningStyle     = lipgloss.NewStyle().Foreground(Warning)
	ErrorStyle       = lipgloss.NewStyle().Foreground(Danger)
	BoxStyle         = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Primary).Padding(1, 2)
	FooterStyle      = lipgloss.NewStyle().Foreground(TextDim).MarginTop(1)
	ProgressBarStyle = lipgloss.NewStyle().Foreground(Secondary)
	StatusStyle      = lipgloss.NewStyle().Foreground(TextDim).Italic(true)
)

func RenderLogo() string {
	logo := `
 ██████╗ ███████╗ ██████╗  ██████╗ 
 ██╔══██╗██╔════╝██╔════╝ ██╔═══██╗
 ██████╔╝█████╗  ██║  ███╗██║   ██║
 ██╔══██╗██╔══╝  ██║   ██║██║   ██║
 ██║  ██║███████╗╚██████╔╝╚██████╔╝
 ╚═╝  ╚═╝╚══════╝ ╚═════╝  ╚═════╝ `
	return lipgloss.NewStyle().Foreground(Primary).Render(logo)
}

func RenderTagline() string { return SubtitleStyle.Render("Linux Reinstall Helper") }
