package views

import (
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/r8bert/rego/internal/backup"
	"github.com/r8bert/rego/internal/restore"
	"github.com/r8bert/rego/ui/styles"
	"github.com/r8bert/rego/ui/components"
)

type RestorePhase int

const (
	RestorePhaseSelectBackup RestorePhase = iota
	RestorePhaseSelectComponents
	RestorePhaseConfirm
	RestorePhaseRunning
	RestorePhaseComplete
)

type RestoreView struct {
	phase        RestorePhase
	backups      []backup.BackupManifest
	backupMenu   *components.Menu
	checkboxes   *components.CheckboxList
	confirm      *components.Confirm
	progress     *components.Progress
	dryRun       bool
	selectedPath string
	results      []restore.RestoreResult
	error        error
}

type restoreCompleteMsg struct {
	results []restore.RestoreResult
	err     error
}

func NewRestoreView() RestoreView {
	mgr := backup.NewManager()
	backups, _ := mgr.ListBackups()

	var items []components.MenuItem
	for _, b := range backups {
		items = append(items, components.MenuItem{
			ID: b.BackupPath, Title: filepath.Base(b.BackupPath),
			Description: fmt.Sprintf("%s - %d components", b.CreatedAt.Format("2006-01-02 15:04"), len(b.Components)),
		})
	}
	if len(items) == 0 {
		items = append(items, components.MenuItem{ID: "", Title: "No backups found", Description: "Create a backup first"})
	}

	return RestoreView{backups: backups, backupMenu: components.NewMenu(items), dryRun: true}
}

func (v RestoreView) Init() tea.Cmd { return nil }

func (v RestoreView) Update(msg tea.Msg) (RestoreView, tea.Cmd, string) {
	switch msg := msg.(type) {
	case restoreCompleteMsg:
		v.phase = RestorePhaseComplete
		v.results = msg.results
		v.error = msg.err
		return v, nil, ""
	case tea.KeyMsg:
		switch v.phase {
		case RestorePhaseSelectBackup:
			switch msg.String() {
			case "up", "k":
				v.backupMenu.Up()
			case "down", "j":
				v.backupMenu.Down()
			case "enter":
				sel := v.backupMenu.Selected()
				if sel.ID != "" {
					v.selectedPath = sel.ID
					v.setupComponentSelection()
					v.phase = RestorePhaseSelectComponents
				}
			case "esc":
				return v, nil, "back"
			}
		case RestorePhaseSelectComponents:
			switch msg.String() {
			case "up", "k":
				v.checkboxes.Up()
			case "down", "j":
				v.checkboxes.Down()
			case " ":
				v.checkboxes.Toggle()
			case "a":
				v.checkboxes.ToggleAll()
			case "d":
				v.dryRun = !v.dryRun
			case "enter":
				if len(v.checkboxes.GetSelected()) > 0 {
					mode := "DRY RUN"
					if !v.dryRun {
						mode = "LIVE"
					}
					v.confirm = components.NewConfirm("Start Restore? ("+mode+")",
						"This will restore the selected components from backup.")
					v.phase = RestorePhaseConfirm
				}
			case "esc":
				v.phase = RestorePhaseSelectBackup
			}
		case RestorePhaseConfirm:
			switch msg.String() {
			case "left", "h", "right", "l":
				v.confirm.Toggle()
			case "enter":
				if v.confirm.Confirmed() {
					v.phase = RestorePhaseRunning
					return v, v.runRestore(), ""
				}
				v.phase = RestorePhaseSelectComponents
			case "esc":
				v.phase = RestorePhaseSelectComponents
			}
		case RestorePhaseComplete:
			if msg.String() == "enter" || msg.String() == "esc" {
				return v, nil, "back"
			}
		}
	}
	return v, nil, ""
}

func (v *RestoreView) setupComponentSelection() {
	mgr := restore.NewManager()
	available := mgr.GetAvailableRestorers()
	var items []components.CheckboxItem
	for _, r := range available {
		items = append(items, components.CheckboxItem{
			ID: string(r.Type()), Title: r.Name(), Checked: true,
		})
	}
	v.checkboxes = components.NewCheckboxList(items)
	v.progress = components.NewProgress(len(items))
}

func (v RestoreView) runRestore() tea.Cmd {
	return func() tea.Msg {
		selected := v.checkboxes.GetSelected()
		opts := restore.RestoreOptions{
			BackupPath: v.selectedPath, DryRun: v.dryRun,
			IncludeFlatpak:         hasID(selected, "flatpak"),
			IncludeRPM:             hasID(selected, "rpm"),
			IncludeRepos:           hasID(selected, "repos"),
			IncludeGnomeExtensions: hasID(selected, "gnome_extensions"),
			IncludeGnomeSettings:   hasID(selected, "gnome_settings"),
			IncludeDotfiles:        hasID(selected, "dotfiles"),
			IncludeFonts:           hasID(selected, "fonts"),
		}
		mgr := restore.NewManager()
		results, err := mgr.RunRestore(opts, nil)
		return restoreCompleteMsg{results, err}
	}
}

func (v RestoreView) View() string {
	s := styles.TitleStyle.Render("🔄 Restore System") + "\n\n"

	switch v.phase {
	case RestorePhaseSelectBackup:
		s += styles.DescriptionStyle.Render("Select a backup to restore:") + "\n\n"
		s += v.backupMenu.View() + "\n"
		s += styles.FooterStyle.Render("↑/↓: Navigate • Enter: Select • Esc: Back")
	case RestorePhaseSelectComponents:
		mode := styles.SuccessStyle.Render("[DRY RUN]")
		if !v.dryRun {
			mode = styles.WarningStyle.Render("[LIVE MODE]")
		}
		s += "Mode: " + mode + "\n\n"
		s += styles.DescriptionStyle.Render("Select components to restore:") + "\n\n"
		s += v.checkboxes.View() + "\n"
		s += styles.FooterStyle.Render("Space: Toggle • d: Toggle Dry Run • Enter: Continue")
	case RestorePhaseConfirm:
		s += v.confirm.View()
	case RestorePhaseRunning:
		s += "Restore in progress...\n\n" + v.progress.View()
	case RestorePhaseComplete:
		if v.dryRun {
			s += styles.WarningStyle.Render("DRY RUN - No changes were made") + "\n\n"
		}
		if v.error != nil {
			s += styles.ErrorStyle.Render("Restore failed: "+v.error.Error()) + "\n"
		}
		for _, r := range v.results {
			status := styles.SuccessStyle.Render("✓")
			if !r.Success {
				status = styles.ErrorStyle.Render("✗")
			}
			s += fmt.Sprintf("  %s %s: %d/%d items\n", status, r.Type, r.ItemsSuccess, r.ItemsTotal)
		}
		s += "\n" + styles.FooterStyle.Render("Press Enter to continue")
	}
	return s
}
