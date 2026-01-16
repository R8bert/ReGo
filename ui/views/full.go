package views

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/r8bert/rego/internal/backup"
	"github.com/r8bert/rego/ui/components"
	"github.com/r8bert/rego/ui/styles"
)

type FullSavePhase int

const (
	FullSavePhaseSelect FullSavePhase = iota
	FullSavePhaseRunning
	FullSavePhaseDone
)

type FullSaveView struct {
	phase      FullSavePhase
	checkboxes *components.CheckboxList
	path       string
	size       int64
	stats      map[string]int
	error      error
}

type fullSaveDoneMsg struct {
	path  string
	size  int64
	stats map[string]int
	err   error
}

func NewFullSaveView() FullSaveView {
	items := []components.CheckboxItem{
		{ID: "flatpaks", Title: "Flatpak Apps", Description: "All installed Flatpak applications", Checked: true},
		{ID: "rpm", Title: "RPM Packages", Description: "User-installed system packages", Checked: true},
		{ID: "repos", Title: "Repositories", Description: "Third-party DNF repos", Checked: true},
		{ID: "extensions", Title: "GNOME Extensions", Description: "Shell extensions and their settings", Checked: true},
		{ID: "settings", Title: "GNOME Settings", Description: "Desktop customizations (dconf)", Checked: true},
		{ID: "dotfiles", Title: "Dotfiles", Description: ".bashrc, .zshrc, .gitconfig, etc.", Checked: true},
		{ID: "fonts", Title: "User Fonts", Description: "~/.local/share/fonts", Checked: true},
		{ID: "ssh", Title: "SSH Config", Description: "~/.ssh/config (no keys)", Checked: true},
		{ID: "autostart", Title: "Autostart Apps", Description: "~/.config/autostart", Checked: true},
		{ID: "backgrounds", Title: "Wallpapers", Description: "Custom wallpapers", Checked: true},
		{ID: "themes", Title: "GTK Themes", Description: "~/.themes and icons", Checked: true},
	}
	return FullSaveView{
		checkboxes: components.NewCheckboxList(items),
		path:       backup.GetDefaultFullBackupPath(),
	}
}

func (v FullSaveView) Init() tea.Cmd { return nil }

func (v FullSaveView) Update(msg tea.Msg) (FullSaveView, tea.Cmd, string) {
	switch msg := msg.(type) {
	case fullSaveDoneMsg:
		v.phase = FullSavePhaseDone
		v.path = msg.path
		v.size = msg.size
		v.stats = msg.stats
		v.error = msg.err
		return v, nil, ""
	case tea.KeyMsg:
		switch v.phase {
		case FullSavePhaseSelect:
			switch msg.String() {
			case "up", "k":
				v.checkboxes.Up()
			case "down", "j":
				v.checkboxes.Down()
			case " ":
				v.checkboxes.Toggle()
			case "a":
				v.checkboxes.ToggleAll()
			case "enter":
				if len(v.checkboxes.GetSelected()) > 0 {
					v.phase = FullSavePhaseRunning
					return v, v.runBackup(), ""
				}
			case "esc", "q":
				return v, nil, "back"
			}
		case FullSavePhaseDone:
			return v, nil, "back"
		}
	}
	return v, nil, ""
}

func (v FullSaveView) getOptions() backup.FullBackupOptions {
	selected := v.checkboxes.GetSelected()
	opts := backup.FullBackupOptions{}
	for _, item := range selected {
		switch item.ID {
		case "flatpaks":
			opts.Flatpaks = true
		case "rpm":
			opts.RPM = true
		case "repos":
			opts.Repos = true
		case "extensions":
			opts.Extensions = true
		case "settings":
			opts.Settings = true
		case "dotfiles":
			opts.Dotfiles = true
		case "fonts":
			opts.Fonts = true
		case "ssh":
			opts.SSHConfig = true
		case "autostart":
			opts.Autostart = true
		case "backgrounds":
			opts.Backgrounds = true
		case "themes":
			opts.Themes = true
		}
	}
	return opts
}

func (v FullSaveView) runBackup() tea.Cmd {
	return func() tea.Msg {
		opts := v.getOptions()
		path := backup.GetDefaultFullBackupPath()
		stats, err := backup.CreateFullBackup(opts, path)

		var size int64
		if err == nil {
			if info, statErr := os.Stat(path); statErr == nil {
				size = info.Size()
			}
		}

		return fullSaveDoneMsg{path: path, size: size, stats: stats, err: err}
	}
}

func (v FullSaveView) View() string {
	s := styles.RenderLogo() + "\n\n"
	s += styles.TitleStyle.Render("💾 Full Save") + "\n"
	s += styles.DescriptionStyle.Render("Create a complete backup with all files") + "\n\n"

	switch v.phase {
	case FullSavePhaseSelect:
		s += v.checkboxes.View() + "\n"
		s += styles.DimStyle.Render("Output: "+v.path) + "\n\n"
		s += styles.FooterStyle.Render("Space: Toggle • a: All • Enter: Save • Esc: Back")

	case FullSavePhaseRunning:
		s += styles.WarningStyle.Render("⏳ Creating backup...") + "\n"
		s += styles.DimStyle.Render("This may take a moment for large files")

	case FullSavePhaseDone:
		if v.error != nil {
			s += styles.ErrorStyle.Render("✗ Failed: "+v.error.Error()) + "\n"
		} else {
			s += styles.SuccessStyle.Render("✓ Backup complete!") + "\n\n"
			s += fmt.Sprintf("📦 File: %s\n", styles.SelectedStyle.Render(v.path))
			s += fmt.Sprintf("📊 Size: %s\n\n", formatSizeFull(v.size))
			if v.stats != nil {
				for k, v := range v.stats {
					if v > 0 {
						s += fmt.Sprintf("   • %d %s\n", v, k)
					}
				}
			}
			s += "\n" + styles.DescriptionStyle.Render("Copy this file to USB, cloud, or external drive!")
		}
		s += "\n\n" + styles.DimStyle.Render("Press any key to continue")
	}

	return s
}

func formatSizeFull(b int64) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	if b < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	}
	if b < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(b)/(1024*1024*1024))
}
