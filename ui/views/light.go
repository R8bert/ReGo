package views

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/r8bert/rego/internal/backup"
	"github.com/r8bert/rego/internal/utils"
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
	phase       int // 0=file select, 1=confirm, 2=running, 3=done
	frame       int
	path        string
	backup      *backup.LightBackup
	dryRun      bool
	results     string
	error       error
	currentStep string
	progress    int
	total       int
	logs        []string
}

type lightRestoreDoneMsg struct {
	results string
	err     error
}

type restoreProgressMsg struct {
	step    string
	current int
	total   int
	log     string
}

func NewLightRestoreView() LightRestoreView {
	return LightRestoreView{path: backup.GetDefaultLightBackupPath(), dryRun: true}
}

func (v LightRestoreView) Init() tea.Cmd { return components.Tick() }

func (v LightRestoreView) Update(msg tea.Msg) (LightRestoreView, tea.Cmd, string) {
	switch msg := msg.(type) {
	case components.TickMsg:
		v.frame++
		if v.phase == 2 {
			return v, components.Tick(), ""
		}
		return v, components.Tick(), ""
	case restoreProgressMsg:
		v.currentStep = msg.step
		v.progress = msg.current
		v.total = msg.total
		if msg.log != "" {
			v.logs = append(v.logs, msg.log)
		}
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
			case "esc":
				return v, nil, "back"
			}
		case 1:
			switch msg.String() {
			case "d":
				v.dryRun = !v.dryRun
			case "enter":
				v.phase = 2
				v.logs = []string{}
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

func (v LightRestoreView) runRestore() tea.Cmd {
	return func() tea.Msg {
		var results []string

		if v.dryRun {
			// Just show what would be done
			if len(v.backup.Flatpaks) > 0 {
				results = append(results, fmt.Sprintf("[DRY RUN] Would install %d Flatpaks", len(v.backup.Flatpaks)))
			}
			if len(v.backup.RPMPackages) > 0 {
				results = append(results, fmt.Sprintf("[DRY RUN] Would install %d RPM packages", len(v.backup.RPMPackages)))
			}
			if len(v.backup.APTPackages) > 0 {
				results = append(results, fmt.Sprintf("[DRY RUN] Would install %d APT packages", len(v.backup.APTPackages)))
			}
			if len(v.backup.GnomeExtensions) > 0 {
				results = append(results, fmt.Sprintf("[DRY RUN] Would restore %d GNOME extensions", len(v.backup.GnomeExtensions)))
			}
			if v.backup.DconfSettings != "" {
				results = append(results, "[DRY RUN] Would restore dconf settings")
			}
			return lightRestoreDoneMsg{results: "Dry run complete:\n" + joinLines(results)}
		}

		// Actually perform restore

		// 1. Flatpaks
		if len(v.backup.Flatpaks) > 0 {
			for _, app := range v.backup.Flatpaks {
				utils.RunCommand("flatpak", "install", "-y", "flathub", app)
			}
			results = append(results, fmt.Sprintf("Installed %d Flatpaks", len(v.backup.Flatpaks)))
		}

		// 2. System packages (RPM/APT)
		pm := backup.DetectPackageManager()
		switch pm {
		case backup.PMDNF:
			if len(v.backup.RPMPackages) > 0 {
				args := append([]string{"install", "-y"}, v.backup.RPMPackages...)
				utils.RunCommand("sudo", append([]string{"dnf"}, args...)...)
				results = append(results, fmt.Sprintf("Installed %d RPM packages", len(v.backup.RPMPackages)))
			}
		case backup.PMAPT:
			if len(v.backup.APTPackages) > 0 {
				args := append([]string{"install", "-y"}, v.backup.APTPackages...)
				utils.RunCommand("sudo", append([]string{"apt"}, args...)...)
				results = append(results, fmt.Sprintf("Installed %d APT packages", len(v.backup.APTPackages)))
			}
		}

		// 3. GNOME Extensions
		if len(v.backup.GnomeExtensions) > 0 {
			for _, ext := range v.backup.GnomeExtensions {
				utils.RunCommand("gnome-extensions", "enable", ext)
			}
			results = append(results, fmt.Sprintf("Enabled %d GNOME extensions", len(v.backup.GnomeExtensions)))
		}

		// 4. Dconf settings
		if v.backup.DconfSettings != "" {
			// Write to temp file and load
			tmpFile := "/tmp/rego-dconf-restore.txt"
			os.WriteFile(tmpFile, []byte(v.backup.DconfSettings), 0644)
			utils.RunCommand("dconf", "load", "/")
			os.Remove(tmpFile)
			results = append(results, "Restored dconf settings")
		}

		if len(results) == 0 {
			return lightRestoreDoneMsg{results: "Nothing to restore"}
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
	s := styles.TitleStyle.Render("📥 Load Backup") + "\n\n"

	switch v.phase {
	case 0:
		// Animated search
		search := []string{"🔍", "🔎", "🔍", "🔎"}[v.frame/3%4]
		s += search + " Looking for: " + styles.SelectedStyle.Render(v.path) + "\n\n"
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
		s += fmt.Sprintf("Backup from: %s (%s)\n", v.backup.Hostname, v.backup.CreatedAt.Format("2006-01-02"))
		if v.backup.Distro != "" {
			s += fmt.Sprintf("Distro: %s\n\n", v.backup.Distro)
		} else {
			s += "\n"
		}
		stats := v.backup.Stats()
		if stats["flatpaks"] > 0 {
			s += fmt.Sprintf("   • %d Flatpaks\n", stats["flatpaks"])
		}
		if stats["rpm"] > 0 {
			s += fmt.Sprintf("   • %d RPM packages\n", stats["rpm"])
		}
		if stats["apt"] > 0 {
			s += fmt.Sprintf("   • %d APT packages\n", stats["apt"])
		}
		if stats["extensions"] > 0 {
			s += fmt.Sprintf("   • %d GNOME extensions\n", stats["extensions"])
		}
		s += "\n" + styles.DimStyle.Render("d: Toggle mode • Enter: Restore • Esc: Back")
	case 2:
		// Running phase - show progress
		spinner := []string{"◐", "◓", "◑", "◒"}[v.frame/2%4]
		s += styles.WarningStyle.Render(spinner+" Installing...") + "\n\n"

		if v.currentStep != "" {
			s += styles.NormalStyle.Render("Current: ") + styles.SelectedStyle.Render(v.currentStep) + "\n"
		}

		// Progress bar
		if v.total > 0 {
			pct := float64(v.progress) / float64(v.total)
			barWidth := 30
			filled := int(pct * float64(barWidth))
			bar := ""
			for i := 0; i < barWidth; i++ {
				if i < filled {
					bar += "█"
				} else {
					bar += "░"
				}
			}
			s += fmt.Sprintf("\n[%s] %d/%d\n", styles.SuccessStyle.Render(bar), v.progress, v.total)
		}

		// Recent logs
		if len(v.logs) > 0 {
			s += "\n" + styles.DimStyle.Render("Log:") + "\n"
			start := 0
			if len(v.logs) > 5 {
				start = len(v.logs) - 5
			}
			for _, log := range v.logs[start:] {
				s += styles.DimStyle.Render("  "+log) + "\n"
			}
		}
	case 3:
		s += v.results + "\n\n" + styles.DimStyle.Render("Press any key to continue")
	}
	return s
}
