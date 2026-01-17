package styles

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Premium color palette
	Primary     = lipgloss.Color("#8B5CF6") // Vibrant purple
	PrimaryDark = lipgloss.Color("#6D28D9")
	Secondary   = lipgloss.Color("#22D3EE") // Cyan
	Accent      = lipgloss.Color("#F472B6") // Pink
	Success     = lipgloss.Color("#34D399") // Emerald
	Warning     = lipgloss.Color("#FBBF24") // Amber
	Danger      = lipgloss.Color("#F87171") // Red
	Muted       = lipgloss.Color("#64748B") // Slate
	Text        = lipgloss.Color("#F1F5F9") // Light slate
	TextDim     = lipgloss.Color("#94A3B8")
	Surface     = lipgloss.Color("#1E293B") // Dark surface
	Border      = lipgloss.Color("#334155")

	// Base styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Text).
			Background(Primary).
			Padding(0, 2)

	SubtitleStyle    = lipgloss.NewStyle().Foreground(Secondary).Bold(true)
	DescriptionStyle = lipgloss.NewStyle().Foreground(TextDim)
	SelectedStyle    = lipgloss.NewStyle().Foreground(Primary).Bold(true)
	NormalStyle      = lipgloss.NewStyle().Foreground(Text)
	DimStyle         = lipgloss.NewStyle().Foreground(Muted)
	SuccessStyle     = lipgloss.NewStyle().Foreground(Success).Bold(true)
	WarningStyle     = lipgloss.NewStyle().Foreground(Warning)
	ErrorStyle       = lipgloss.NewStyle().Foreground(Danger).Bold(true)
	AccentStyle      = lipgloss.NewStyle().Foreground(Accent)
	FooterStyle      = lipgloss.NewStyle().Foreground(Muted)
	StatusStyle      = lipgloss.NewStyle().Foreground(TextDim).Italic(true)
	ProgressBarStyle = lipgloss.NewStyle().Foreground(Secondary)
	BoxStyle         = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Primary).Padding(1, 2)

	// Card style - main content container
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(1, 2).
			MarginTop(1)

	// Highlighted card
	HighlightCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(1, 2)

	// Menu item styles
	MenuItemActive = lipgloss.NewStyle().
			Foreground(Text).
			Background(Primary).
			Padding(0, 2).
			Bold(true)

	MenuItemNormal = lipgloss.NewStyle().
			Foreground(TextDim).
			Padding(0, 2)

	// Badge styles
	GnomeBadge = lipgloss.NewStyle().
			Background(lipgloss.Color("#4A86CF")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Bold(true)

	KDEBadge = lipgloss.NewStyle().
			Background(lipgloss.Color("#1D99F3")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Bold(true)

	// Checkbox styles
	CheckedStyle   = lipgloss.NewStyle().Foreground(Success).Bold(true)
	UncheckedStyle = lipgloss.NewStyle().Foreground(Muted)
)

const AppWidth = 56

// Logo with gradient effect
func RenderLogo() string {
	lines := []string{
		" ██████╗ ███████╗ ██████╗  ██████╗ ",
		" ██╔══██╗██╔════╝██╔════╝ ██╔═══██╗",
		" ██████╔╝█████╗  ██║  ███╗██║   ██║",
		" ██╔══██╗██╔══╝  ██║   ██║██║   ██║",
		" ██║  ██║███████╗╚██████╔╝╚██████╔╝",
		" ╚═╝  ╚═╝╚══════╝ ╚═════╝  ╚═════╝ ",
	}
	colors := []lipgloss.Color{
		"#A78BFA", "#8B5CF6", "#7C3AED", "#6D28D9", "#5B21B6", "#4C1D95",
	}
	var result string
	for i, line := range lines {
		result += lipgloss.NewStyle().Foreground(colors[i]).Render(line) + "\n"
	}
	return result
}

// Tagline with desktop badge
func RenderTagline() string {
	tagline := SubtitleStyle.Render("✦ Linux Reinstall Helper ✦")
	badge := GetDesktopBadge()
	if badge != "" {
		return tagline + "  " + badge
	}
	return tagline
}

// Desktop detection
func GetDesktopEnv() string {
	desktop := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))
	if strings.Contains(desktop, "gnome") {
		return "gnome"
	}
	if strings.Contains(desktop, "kde") || strings.Contains(desktop, "plasma") {
		return "kde"
	}
	return "unknown"
}

func GetDesktopBadge() string {
	switch GetDesktopEnv() {
	case "gnome":
		return GnomeBadge.Render(" GNOME ")
	case "kde":
		return KDEBadge.Render(" KDE ")
	default:
		return ""
	}
}

func GetDesktopSuggestion() string {
	switch GetDesktopEnv() {
	case "gnome":
		return SuccessStyle.Render("●") + " GNOME detected"
	case "kde":
		return SuccessStyle.Render("●") + " KDE Plasma detected"
	default:
		return DimStyle.Render("○ Desktop not detected")
	}
}

// GetSystemInfo returns a formatted system info string
func GetSystemInfo() string {
	// Get distro from /etc/os-release
	distro := "Linux"
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				distro = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
				break
			}
		}
	}
	return distro
}

// Divider with style
func Divider() string {
	return DimStyle.Render("─────────────────────────────────────────────────────")
}

func DividerShort() string {
	return DimStyle.Render("────────────────────────────")
}

// Render a menu item card
func RenderMenuItem(icon, title, desc string, selected bool, frame int) string {
	if selected {
		cursors := []string{"▸", "►", "▹", "►"}
		cursor := cursors[frame/3%len(cursors)]
		header := lipgloss.NewStyle().Foreground(Primary).Bold(true).Render(cursor + " " + icon + " " + title)
		return header + "\n   " + DescriptionStyle.Render(desc)
	}
	return DimStyle.Render("  "+icon+" ") + lipgloss.NewStyle().Foreground(TextDim).Render(title)
}

// Render a checkbox item
func RenderCheckbox(title string, checked, selected bool) string {
	box := "○"
	style := DimStyle
	if checked {
		box = "●"
		style = CheckedStyle
	}

	cursor := "  "
	if selected {
		cursor = AccentStyle.Render("▸ ")
		if checked {
			return cursor + CheckedStyle.Render(box+" "+title)
		}
		return cursor + NormalStyle.Render(box+" "+title)
	}
	return cursor + style.Render(box+" "+title)
}

// Progress bar
func RenderProgress(current, total int, width int) string {
	if total == 0 {
		return ""
	}
	pct := float64(current) / float64(total)
	filled := int(pct * float64(width))
	empty := width - filled

	bar := SuccessStyle.Render(strings.Repeat("█", filled)) +
		DimStyle.Render(strings.Repeat("░", empty))

	return "  [" + bar + "] " + SuccessStyle.Render(fmt.Sprintf("%d%%", int(pct*100)))
}
