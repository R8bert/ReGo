package views

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/r8bert/rego/internal/backup"
	"github.com/r8bert/rego/ui/styles"
	"github.com/r8bert/rego/ui/components"
)

type BackupPhase int

const (
	BackupPhaseSelect BackupPhase = iota
	BackupPhaseConfirm
	BackupPhaseRunning
	BackupPhaseComplete
)

type BackupView struct {
	phase      BackupPhase
	checkboxes *components.CheckboxList
	confirm    *components.Confirm
	progress   *components.Progress
	status     *components.StatusList
	manager    *backup.Manager
	manifest   *backup.BackupManifest
	error      error
}

func NewBackupView() BackupView {
	mgr := backup.NewManager()
	available := mgr.GetAvailableBackers()

	var items []components.CheckboxItem
	for _, b := range available {
		items = append(items, components.CheckboxItem{
			ID: string(b.Type()), Title: b.Name(),
			Description: backup.BackupTypeDescription(b.Type()), Checked: true,
		})
	}

	return BackupView{
		checkboxes: components.NewCheckboxList(items),
		progress:   components.NewProgress(len(items)),
		status:     components.NewStatusList(),
		manager:    mgr,
	}
}

type backupCompleteMsg struct {
	manifest *backup.BackupManifest
	err      error
}
type backupProgressMsg struct {
	name string
	step int
}

func (v BackupView) Init() tea.Cmd { return nil }

func (v BackupView) Update(msg tea.Msg) (BackupView, tea.Cmd, string) {
	switch msg := msg.(type) {
	case backupProgressMsg:
		v.progress.SetCurrent(msg.step)
		v.progress.SetStatus("Backing up " + msg.name + "...")
		v.status.Add(components.StatusItem{Label: msg.name, Status: "running"})
		return v, nil, ""
	case backupCompleteMsg:
		v.phase = BackupPhaseComplete
		v.manifest = msg.manifest
		v.error = msg.err
		return v, nil, ""
	case tea.KeyMsg:
		switch v.phase {
		case BackupPhaseSelect:
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
					v.confirm = components.NewConfirm("Start Backup?", "This will create a backup of the selected components.")
					v.phase = BackupPhaseConfirm
				}
			case "esc":
				return v, nil, "back"
			}
		case BackupPhaseConfirm:
			switch msg.String() {
			case "left", "h", "right", "l":
				v.confirm.Toggle()
			case "enter":
				if v.confirm.Confirmed() {
					v.phase = BackupPhaseRunning
					return v, v.runBackup(), ""
				}
				v.phase = BackupPhaseSelect
			case "esc":
				v.phase = BackupPhaseSelect
			}
		case BackupPhaseComplete:
			if msg.String() == "enter" || msg.String() == "esc" {
				return v, nil, "back"
			}
		}
	}
	return v, nil, ""
}

func (v BackupView) runBackup() tea.Cmd {
	return func() tea.Msg {
		selected := v.checkboxes.GetSelected()
		opts := backup.BackupOptions{
			IncludeFlatpak:         hasID(selected, "flatpak"),
			IncludeRPM:             hasID(selected, "rpm"),
			IncludeRepos:           hasID(selected, "repos"),
			IncludeGnomeExtensions: hasID(selected, "gnome_extensions"),
			IncludeGnomeSettings:   hasID(selected, "gnome_settings"),
			IncludeDotfiles:        hasID(selected, "dotfiles"),
			IncludeFonts:           hasID(selected, "fonts"),
		}
		manifest, err := v.manager.RunBackup(opts, nil)
		return backupCompleteMsg{manifest, err}
	}
}

func hasID(items []components.CheckboxItem, id string) bool {
	for _, i := range items {
		if i.ID == id {
			return true
		}
	}
	return false
}

func (v BackupView) View() string {
	s := styles.TitleStyle.Render("📦 Create Backup") + "\n\n"

	switch v.phase {
	case BackupPhaseSelect:
		s += styles.DescriptionStyle.Render("Select components to backup:") + "\n\n"
		s += v.checkboxes.View() + "\n"
		s += styles.FooterStyle.Render("Space: Toggle • a: Select All • Enter: Continue • Esc: Back")
	case BackupPhaseConfirm:
		s += v.confirm.View()
	case BackupPhaseRunning:
		s += "Backup in progress...\n\n" + v.progress.View()
	case BackupPhaseComplete:
		if v.error != nil {
			s += styles.ErrorStyle.Render("Backup failed: " + v.error.Error())
		} else if v.manifest != nil {
			s += styles.SuccessStyle.Render("✓ Backup completed successfully!") + "\n\n"
			s += fmt.Sprintf("Location: %s\n", v.manifest.BackupPath)
			s += fmt.Sprintf("Time: %s\n", v.manifest.CreatedAt.Format(time.RFC822))
			count := 0
			for _, r := range v.manifest.Results {
				count += r.ItemCount
			}
			s += fmt.Sprintf("Items backed up: %d\n", count)
		}
		s += "\n" + styles.FooterStyle.Render("Press Enter to continue")
	}
	return s
}
