package views

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/r8bert/rego/internal/backup"
	"github.com/r8bert/rego/ui/components"
	"github.com/r8bert/rego/ui/styles"
)

type LightPhase int

const (
	LightPhaseSelect LightPhase = iota
	LightPhaseRunning
	LightPhaseDone
)

type LightBackupView struct {
	phase      LightPhase
	checkboxes *components.CheckboxList
	path       string
	size       int64
	stats      map[string]int
	error      error
}

type lightBackupDoneMsg struct {
	path  string
	size  int64
	stats map[string]int
	err   error
}

func NewLightBackupView() LightBackupView {
	items := []components.CheckboxItem{
		{ID: "flatpaks", Title: "Flatpak Apps", Description: "Installed Flatpak applications", Checked: true},
		{ID: "rpm", Title: "RPM Packages", Description: "User-installed RPM packages", Checked: true},
		{ID: "extensions", Title: "GNOME Extensions", Description: "Shell extensions", Checked: true},
		{ID: "settings", Title: "GNOME Settings", Description: "Desktop customizations (dconf)", Checked: true},
		{ID: "repos", Title: "Repositories", Description: "Third-party DNF repos", Checked: true},
	}
	return LightBackupView{
		checkboxes: components.NewCheckboxList(items),
		path:       backup.GetDefaultLightBackupPath(),
	}
}

func (v LightBackupView) Init() tea.Cmd { return nil }

func (v LightBackupView) Update(msg tea.Msg) (LightBackupView, tea.Cmd, string) {
	switch msg := msg.(type) {
	case lightBackupDoneMsg:
		v.phase = LightPhaseDone
		v.path = msg.path
		v.size = msg.size
		v.stats = msg.stats
		v.error = msg.err
		return v, nil, ""
	case tea.KeyMsg:
		switch v.phase {
		case LightPhaseSelect:
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
					v.phase = LightPhaseRunning
					return v, v.runBackup(), ""
				}
			case "esc", "q":
				return v, nil, "back"
			}
		case LightPhaseDone:
			return v, nil, "back"
		}
	}
	return v, nil, ""
}

func (v LightBackupView) getOptions() backup.LightBackupOptions {
	selected := v.checkboxes.GetSelected()
	opts := backup.LightBackupOptions{}
	for _, item := range selected {
		switch item.ID {
		case "flatpaks":
			opts.Flatpaks = true
		case "rpm":
			opts.RPM = true
		case "extensions":
			opts.Extensions = true
		case "settings":
			opts.Settings = true
		case "repos":
			opts.Repos = true
		}
	}
	return opts
}

func (v LightBackupView) runBackup() tea.Cmd {
	return func() tea.Msg {
		opts := v.getOptions()
		b, err := backup.CreateLightBackupWithOptions(opts)
		if err != nil {
			return lightBackupDoneMsg{err: err}
		}

		path := backup.GetDefaultLightBackupPath()
		if err := b.SaveToFile(path); err != nil {
			return lightBackupDoneMsg{err: err}
		}

		var size int64
		if info, err := os.Stat(path); err == nil {
			size = info.Size()
		}

		return lightBackupDoneMsg{path: path, size: size, stats: b.Stats()}
	}
}

func formatSize(b int64) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	if b < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
}

func (v LightBackupView) View() string {
	s := styles.RenderLogo() + "\n\n"
	s += styles.TitleStyle.Render("⚡ Quick Save") + "\n"
	s += styles.DescriptionStyle.Render("Select what to backup:") + "\n\n"

	switch v.phase {
	case LightPhaseSelect:
		s += v.checkboxes.View() + "\n"
		s += styles.DimStyle.Render("Output: "+v.path) + "\n\n"
		s += styles.FooterStyle.Render("Space: Toggle • a: All • Enter: Save • Esc: Back")

	case LightPhaseRunning:
		s += styles.WarningStyle.Render("⏳ Scanning system...") + "\n"

	case LightPhaseDone:
		if v.error != nil {
			s += styles.ErrorStyle.Render("✗ Failed: "+v.error.Error()) + "\n"
		} else {
			s += styles.SuccessStyle.Render("✓ Saved!") + "\n\n"
			s += fmt.Sprintf("📄 File: %s\n", styles.SelectedStyle.Render(v.path))
			s += fmt.Sprintf("📊 Size: %s\n\n", formatSize(v.size))
			if v.stats != nil {
				if v.stats["flatpaks"] > 0 {
					s += fmt.Sprintf("   • %d Flatpaks\n", v.stats["flatpaks"])
				}
				if v.stats["rpm"] > 0 {
					s += fmt.Sprintf("   • %d RPM packages\n", v.stats["rpm"])
				}
				if v.stats["extensions"] > 0 {
					s += fmt.Sprintf("   • %d GNOME extensions\n", v.stats["extensions"])
				}
				if v.stats["repos"] > 0 {
					s += fmt.Sprintf("   • %d Repositories\n", v.stats["repos"])
				}
			}
			s += "\n" + styles.DescriptionStyle.Render("Copy this file to USB, cloud, or email it!")
		}
		s += "\n\n" + styles.DimStyle.Render("Press any key to continue")
	}

	return s
}

// LightRestoreView for restoring from a light backup
type LightRestoreView struct {
	phase   int
	path    string
	backup  *backup.LightBackup
	dryRun  bool
	results string
	error   error
}

type lightRestoreDoneMsg struct {
	results string
	err     error
}

func NewLightRestoreView() LightRestoreView {
	return LightRestoreView{path: backup.GetDefaultLightBackupPath(), dryRun: true}
}

func (v LightRestoreView) Init() tea.Cmd { return nil }

func (v LightRestoreView) Update(msg tea.Msg) (LightRestoreView, tea.Cmd, string) {
	switch msg := msg.(type) {
	case lightRestoreDoneMsg:
		v.phase = 2
		v.results = msg.results
		v.error = msg.err
		return v, nil, ""
	case tea.KeyMsg:
		switch v.phase {
		case 0:
			switch msg.String() {
			case "enter":
				b, err := backup.LoadLightBackup(v.path)
				if err != nil {
					v.error = err
					return v, nil, ""
				}
				v.backup = b
				v.phase = 1
			case "esc":
				return v, nil, "back"
			}
		case 1:
			switch msg.String() {
			case "d":
				v.dryRun = !v.dryRun
			case "enter":
				return v, v.runRestore(), ""
			case "esc":
				v.phase = 0
				v.error = nil
			}
		case 2:
			return v, nil, "back"
		}
	}
	return v, nil, ""
}

func (v LightRestoreView) runRestore() tea.Cmd {
	return func() tea.Msg {
		result := fmt.Sprintf("Would install:\n• %d Flatpaks\n• %d RPM packages\n• %d Extensions",
			len(v.backup.Flatpaks), len(v.backup.RPMPackages), len(v.backup.GnomeExtensions))
		return lightRestoreDoneMsg{results: result}
	}
}

func (v LightRestoreView) View() string {
	s := styles.TitleStyle.Render("📥 Load Backup") + "\n\n"

	switch v.phase {
	case 0:
		s += "Looking for: " + styles.SelectedStyle.Render(v.path) + "\n\n"
		if v.error != nil {
			s += styles.ErrorStyle.Render("Not found: "+v.error.Error()) + "\n"
		}
		s += styles.DimStyle.Render("Enter: Load • Esc: Back")
	case 1:
		mode := styles.SuccessStyle.Render("[DRY RUN]")
		if !v.dryRun {
			mode = styles.WarningStyle.Render("[LIVE]")
		}
		s += "Mode: " + mode + "\n\n"
		s += fmt.Sprintf("Backup from: %s (%s)\n\n", v.backup.Hostname, v.backup.CreatedAt.Format("2006-01-02"))
		stats := v.backup.Stats()
		s += fmt.Sprintf("   • %d Flatpaks\n", stats["flatpaks"])
		s += fmt.Sprintf("   • %d RPM packages\n", stats["rpm"])
		s += fmt.Sprintf("   • %d Extensions\n", stats["extensions"])
		s += "\n" + styles.DimStyle.Render("d: Toggle mode • Enter: Restore • Esc: Back")
	case 2:
		s += v.results + "\n\n" + styles.DimStyle.Render("Press any key")
	}
	return s
}
