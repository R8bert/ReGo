package views

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
	spinner    *components.Spinner
	progress   *components.AnimatedProgress
	frame      int
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
	// Get package manager for dynamic label
	pkgLabel := "System Packages"
	pkgDesc := "User-installed packages"
	switch backup.DetectPackageManager() {
	case backup.PMDNF:
		pkgLabel = "RPM Packages"
		pkgDesc = "User-installed DNF/RPM packages"
	case backup.PMAPT:
		pkgLabel = "APT Packages"
		pkgDesc = "User-installed apt packages"
	case backup.PMPacman:
		pkgLabel = "Pacman Packages"
		pkgDesc = "User-installed pacman packages"
	}

	items := []components.CheckboxItem{
		{ID: "flatpaks", Title: "Flatpak Apps", Description: "Installed Flatpak applications", Checked: true},
		{ID: "rpm", Title: pkgLabel, Description: pkgDesc, Checked: true},
		{ID: "extensions", Title: "GNOME Extensions", Description: "Shell extensions", Checked: backup.IsGNOME()},
		{ID: "settings", Title: "GNOME Settings", Description: "Desktop customizations (dconf)", Checked: backup.IsGNOME()},
		{ID: "kde", Title: "KDE Plasma", Description: "Plasma widgets list", Checked: backup.IsKDE()},
		{ID: "repos", Title: "Repositories", Description: "Third-party repos", Checked: true},
	}
	return LightBackupView{
		checkboxes: components.NewCheckboxList(items),
		spinner:    components.NewSpinner(),
		progress:   components.NewAnimatedProgress(5),
		path:       backup.GetDefaultLightBackupPath(),
	}
}

func (v LightBackupView) Init() tea.Cmd { return nil }

func (v LightBackupView) Update(msg tea.Msg) (LightBackupView, tea.Cmd, string) {
	switch msg := msg.(type) {
	case components.TickMsg:
		v.frame++
		v.spinner.Tick()
		v.progress.Tick()
		if v.phase == LightPhaseRunning {
			return v, components.Tick(), ""
		}
		return v, nil, ""
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
					v.progress = components.NewAnimatedProgress(len(v.checkboxes.GetSelected()))
					return v, tea.Batch(v.runBackup(), components.Tick()), ""
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
		case "kde":
			opts.KDE = true
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

// Animated spinner frames
var saveSpinner = []string{"◐", "◓", "◑", "◒"}
var successAnim = []string{"✓", "★", "✓", "☆"}

func (v LightBackupView) View() string {
	s := styles.RenderLogo() + "\n\n"
	s += styles.TitleStyle.Render("⚡ Quick Save") + "\n"
	s += styles.DescriptionStyle.Render("Select what to backup:") + "\n\n"

	switch v.phase {
	case LightPhaseSelect:
		s += v.checkboxes.View() + "\n"
		s += styles.DimStyle.Render("Output: "+v.path) + "\n\n"
		// Animated hint
		cursor := []string{"▸", "►", "▸", "▶"}[v.frame/3%4]
		s += styles.SuccessStyle.Render(cursor+" Press ENTER to save") + "\n"
		s += styles.FooterStyle.Render("Space: Toggle • a: All • Esc: Back")

	case LightPhaseRunning:
		// Animated spinner
		spinner := saveSpinner[v.frame/2%len(saveSpinner)]
		s += "\n"
		s += styles.WarningStyle.Render("  "+spinner+" Scanning system "+spinner) + "\n\n"

		// Animated dots
		dots := []string{"", ".", "..", "..."}[v.frame/3%4]
		s += styles.DimStyle.Render("  Collecting package lists"+dots) + "\n"

		// Progress bar animation
		barWidth := 20
		pos := v.frame % (barWidth * 2)
		if pos >= barWidth {
			pos = barWidth*2 - pos
		}
		bar := ""
		for i := 0; i < barWidth; i++ {
			if i >= pos-2 && i <= pos+2 {
				bar += "█"
			} else {
				bar += "░"
			}
		}
		s += "\n  [" + styles.SuccessStyle.Render(bar) + "]\n"

	case LightPhaseDone:
		if v.error != nil {
			s += styles.ErrorStyle.Render("✗ Failed: "+v.error.Error()) + "\n"
		} else {
			// Success animation
			star := successAnim[v.frame/3%len(successAnim)]
			s += "\n"
			s += styles.SuccessStyle.Render("  "+star+" Saved successfully! "+star) + "\n\n"
			s += fmt.Sprintf("  📄 File: %s\n", styles.SelectedStyle.Render(v.path))
			s += fmt.Sprintf("  📊 Size: %s\n\n", formatSize(v.size))
			if v.stats != nil {
				if v.stats["flatpaks"] > 0 {
					s += fmt.Sprintf("     • %d Flatpaks\n", v.stats["flatpaks"])
				}
				if v.stats["rpm"] > 0 {
					s += fmt.Sprintf("     • %d RPM packages\n", v.stats["rpm"])
				}
				if v.stats["extensions"] > 0 {
					s += fmt.Sprintf("     • %d Extensions\n", v.stats["extensions"])
				}
				if v.stats["repos"] > 0 {
					s += fmt.Sprintf("     • %d Repos\n", v.stats["repos"])
				}
			}
			s += "\n  " + styles.DescriptionStyle.Render("Copy to USB, cloud, or email!")
		}
		// Blinking prompt
		if v.frame/5%2 == 0 {
			s += "\n\n  " + styles.DimStyle.Render("▸ Press any key to continue")
		} else {
			s += "\n\n  " + styles.DimStyle.Render("  Press any key to continue")
		}
	}

	return s
}

// LightRestoreView for restoring from a light backup
type LightRestoreView struct {
	phase        int // 0=file select, 1=checking, 2=confirm, 3=done
	frame        int
	path         string
	backup       *backup.LightBackup
	restoreCheck *backup.RestoreCheck
	checkStatus  string
	results      string
	error        error
}

type lightRestoreDoneMsg struct {
	results string
	err     error
}

type checkDoneMsg struct {
	check *backup.RestoreCheck
}

func NewLightRestoreView() LightRestoreView {
	return LightRestoreView{path: backup.GetDefaultLightBackupPath()}
}

func (v LightRestoreView) Init() tea.Cmd { return components.Tick() }

func (v LightRestoreView) Update(msg tea.Msg) (LightRestoreView, tea.Cmd, string) {
	switch msg := msg.(type) {
	case components.TickMsg:
		v.frame++
		return v, components.Tick(), ""
	case checkDoneMsg:
		v.restoreCheck = msg.check
		v.phase = 2
		return v, nil, ""
	case lightRestoreDoneMsg:
		v.phase = 3
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
				v.checkStatus = "Checking installed packages..."
				return v, v.runCheck(), ""
			case "esc":
				return v, nil, "back"
			}
		case 2:
			switch msg.String() {
			case "enter":
				return v, v.runRestore(), ""
			case "esc":
				v.phase = 0
				v.error = nil
			}
		case 3:
			return v, nil, "back"
		}
	}
	return v, nil, ""
}

func (v LightRestoreView) runCheck() tea.Cmd {
	return func() tea.Msg {
		check := backup.CheckRestore(v.backup)
		return checkDoneMsg{check: check}
	}
}

func (v LightRestoreView) runRestore() tea.Cmd {
	return func() tea.Msg {
		var results []string
		c := v.restoreCheck

		// 1. Flatpaks - only install missing ones
		if len(c.FlatpaksToInstall) > 0 {
			for _, app := range c.FlatpaksToInstall {
				cmd := exec.Command("flatpak", "install", "-y", "flathub", app)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Run()
			}
			results = append(results, fmt.Sprintf("Installed %d Flatpaks", len(c.FlatpaksToInstall)))
		}

		// 2. System packages (RPM/APT) - only install missing ones
		if len(c.RPMToInstall) > 0 {
			args := append([]string{"dnf", "install", "-y"}, c.RPMToInstall...)
			cmd := exec.Command("sudo", args...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
			results = append(results, fmt.Sprintf("Installed %d RPM packages", len(c.RPMToInstall)))
		}

		if len(c.APTToInstall) > 0 {
			args := append([]string{"apt", "install", "-y"}, c.APTToInstall...)
			cmd := exec.Command("sudo", args...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
			results = append(results, fmt.Sprintf("Installed %d APT packages", len(c.APTToInstall)))
		}

		// 3. GNOME Extensions - only enable missing ones
		if len(c.ExtensionsToEnable) > 0 {
			for _, ext := range c.ExtensionsToEnable {
				exec.Command("gnome-extensions", "enable", ext).Run()
			}
			results = append(results, fmt.Sprintf("Enabled %d GNOME extensions", len(c.ExtensionsToEnable)))
		}

		// 4. Dconf settings
		if c.HasDconfSettings && v.backup.DconfSettings != "" {
			cmd := exec.Command("dconf", "load", "/")
			cmd.Stdin = strings.NewReader(v.backup.DconfSettings)
			cmd.Run()
			results = append(results, "Restored dconf settings")
		}

		if len(results) == 0 {
			return lightRestoreDoneMsg{results: "Nothing to restore - everything is already installed!"}
		}
		return lightRestoreDoneMsg{results: "Restore complete:\n" + joinLines(results)}
	}
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		result += "  • " + line
		if i < len(lines)-1 {
			result += "\n"
		}
	}
	return result
}

func (v LightRestoreView) View() string {
	// Header
	header := styles.TitleStyle.Render("📥 Load Backup")

	var content string

	switch v.phase {
	case 0:
		// Animated search
		search := []string{"🔍", "🔎", "🔍", "🔎"}[v.frame/3%4]
		content = search + " Looking for backup file:\n\n"
		content += styles.BoxStyle.Render(v.path) + "\n\n"
		if v.error != nil {
			content += styles.ErrorStyle.Render("✗ "+v.error.Error()) + "\n\n"
		}
		content += styles.DimStyle.Render("[Enter] Load  [Esc] Back")

	case 1:
		// Checking phase - show verification in progress
		spinner := []string{"◐", "◓", "◑", "◒"}[v.frame/2%4]
		content = styles.WarningStyle.Render(spinner+" Verifying installed packages...") + "\n\n"

		checks := []string{
			styles.DimStyle.Render("   ◦ Scanning Flatpaks..."),
			styles.DimStyle.Render("   ◦ Scanning system packages..."),
			styles.DimStyle.Render("   ◦ Scanning GNOME extensions..."),
		}
		content += strings.Join(checks, "\n")

	case 2:
		// Confirm phase
		content = styles.SuccessStyle.Render("✓ Verification complete") + "\n\n"

		// Backup info box
		infoLines := []string{
			fmt.Sprintf("Host: %s", v.backup.Hostname),
			fmt.Sprintf("Date: %s", v.backup.CreatedAt.Format("2006-01-02 15:04")),
		}
		if v.backup.Distro != "" {
			infoLines = append(infoLines, fmt.Sprintf("Distro: %s", v.backup.Distro))
		}
		content += styles.CardStyle.Render(strings.Join(infoLines, "\n")) + "\n\n"

		// What will be installed
		content += styles.NormalStyle.Render("Packages to install:") + "\n\n"
		c := v.restoreCheck

		if len(c.FlatpaksToInstall) > 0 {
			line := fmt.Sprintf("   %s %d Flatpaks", styles.SuccessStyle.Render("●"), len(c.FlatpaksToInstall))
			if c.FlatpaksSkipped > 0 {
				line += styles.DimStyle.Render(fmt.Sprintf(" (%d skipped)", c.FlatpaksSkipped))
			}
			content += line + "\n"
		}
		if len(c.RPMToInstall) > 0 {
			line := fmt.Sprintf("   %s %d RPM packages", styles.WarningStyle.Render("●"), len(c.RPMToInstall))
			if c.RPMSkipped > 0 {
				line += styles.DimStyle.Render(fmt.Sprintf(" (%d skipped)", c.RPMSkipped))
			}
			content += line + " " + styles.DimStyle.Render("[sudo]") + "\n"
		}
		if len(c.APTToInstall) > 0 {
			line := fmt.Sprintf("   %s %d APT packages", styles.WarningStyle.Render("●"), len(c.APTToInstall))
			if c.APTSkipped > 0 {
				line += styles.DimStyle.Render(fmt.Sprintf(" (%d skipped)", c.APTSkipped))
			}
			content += line + " " + styles.DimStyle.Render("[sudo]") + "\n"
		}
		if len(c.ExtensionsToEnable) > 0 {
			line := fmt.Sprintf("   %s %d GNOME extensions", styles.AccentStyle.Render("●"), len(c.ExtensionsToEnable))
			if c.ExtensionsSkipped > 0 {
				line += styles.DimStyle.Render(fmt.Sprintf(" (%d skipped)", c.ExtensionsSkipped))
			}
			content += line + "\n"
		}
		if c.HasDconfSettings {
			content += fmt.Sprintf("   %s dconf settings\n", styles.AccentStyle.Render("●"))
		}

		// Show if nothing to install
		if len(c.FlatpaksToInstall) == 0 && len(c.RPMToInstall) == 0 &&
			len(c.APTToInstall) == 0 && len(c.ExtensionsToEnable) == 0 && !c.HasDconfSettings {
			content += styles.SuccessStyle.Render("   ✓ Everything is already installed!") + "\n"
		}

		content += "\n" + styles.DimStyle.Render("[Enter] Install  [Esc] Cancel")

	case 3:
		// Done phase
		content = styles.CardStyle.Render(v.results) + "\n\n"
		content += styles.DimStyle.Render("[Any key] Continue")
	}

	return header + "\n\n" + content
}
