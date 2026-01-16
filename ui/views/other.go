package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/r8bert/rego/ui/styles"
)

type AboutView struct{}

func NewAboutView() AboutView     { return AboutView{} }
func (v AboutView) Init() tea.Cmd { return nil }

func (v AboutView) Update(msg tea.Msg) (AboutView, tea.Cmd, string) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "esc" || msg.String() == "enter" || msg.String() == "q" {
			return v, nil, "back"
		}
	}
	return v, nil, ""
}

func (v AboutView) View() string {
	s := styles.RenderLogo() + "\n\n"
	s += styles.TitleStyle.Render("About ReGo") + "\n\n"
	s += "Version: 1.0.0\n\n"
	s += styles.DescriptionStyle.Render("ReGo helps you seamlessly backup and restore your Linux system\n")
	s += styles.DescriptionStyle.Render("configuration when reinstalling your operating system.\n\n")
	s += styles.SubtitleStyle.Render("Features:") + "\n"
	s += "  • Backup/restore Flatpak applications\n"
	s += "  • Backup/restore RPM packages\n"
	s += "  • Backup/restore DNF repositories\n"
	s += "  • Backup/restore GNOME extensions\n"
	s += "  • Backup/restore GNOME settings (dconf)\n"
	s += "  • Backup/restore dotfiles\n"
	s += "  • Backup/restore user fonts\n\n"
	s += styles.FooterStyle.Render("Press Esc or Enter to go back")
	return s
}

type SettingsView struct {
	items  []string
	cursor int
}

func NewSettingsView() SettingsView {
	return SettingsView{items: []string{"Backup Location", "Default Dotfiles", "Theme"}}
}
func (v SettingsView) Init() tea.Cmd { return nil }

func (v SettingsView) Update(msg tea.Msg) (SettingsView, tea.Cmd, string) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "k":
			if v.cursor > 0 {
				v.cursor--
			}
		case "down", "j":
			if v.cursor < len(v.items)-1 {
				v.cursor++
			}
		case "esc":
			return v, nil, "back"
		}
	}
	return v, nil, ""
}

func (v SettingsView) View() string {
	s := styles.TitleStyle.Render("⚙️ Settings") + "\n\n"
	s += styles.DescriptionStyle.Render("Settings configuration coming soon...") + "\n\n"
	for i, item := range v.items {
		cursor := "  "
		if i == v.cursor {
			cursor = styles.SelectedStyle.Render("▸ ")
		}
		s += cursor + item + "\n"
	}
	s += "\n" + styles.FooterStyle.Render("Esc: Back")
	return s
}

type ProfilesView struct {
	message string
}

func NewProfilesView() ProfilesView {
	return ProfilesView{message: "Profile management coming soon..."}
}
func (v ProfilesView) Init() tea.Cmd { return nil }

func (v ProfilesView) Update(msg tea.Msg) (ProfilesView, tea.Cmd, string) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "esc" {
			return v, nil, "back"
		}
	}
	return v, nil, ""
}

func (v ProfilesView) View() string {
	s := styles.TitleStyle.Render("📁 Manage Profiles") + "\n\n"
	s += styles.DescriptionStyle.Render(v.message) + "\n\n"
	s += styles.FooterStyle.Render("Esc: Back")
	return s
}
