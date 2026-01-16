package views

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/r8bert/rego/internal/backup"
	"github.com/r8bert/rego/ui/components"
	"github.com/r8bert/rego/ui/styles"
)

type ExportPhase int

const (
	ExportPhaseSelect ExportPhase = iota
	ExportPhaseRunning
	ExportPhaseComplete
)

type ExportView struct {
	phase      ExportPhase
	checkboxes *components.CheckboxList
	outputPath string
	fileSize   int64
	error      error
}

type exportCompleteMsg struct {
	path string
	size int64
	err  error
}

func NewExportView() ExportView {
	mgr := backup.NewManager()
	available := mgr.GetAvailableBackers()

	var items []components.CheckboxItem
	for _, b := range available {
		items = append(items, components.CheckboxItem{
			ID: string(b.Type()), Title: b.Name(), Checked: true,
		})
	}

	return ExportView{
		checkboxes: components.NewCheckboxList(items),
		outputPath: backup.GetDefaultExportPath(),
	}
}

func (v ExportView) Init() tea.Cmd { return nil }

func (v ExportView) Update(msg tea.Msg) (ExportView, tea.Cmd, string) {
	switch msg := msg.(type) {
	case exportCompleteMsg:
		v.phase = ExportPhaseComplete
		v.outputPath = msg.path
		v.fileSize = msg.size
		v.error = msg.err
		return v, nil, ""
	case tea.KeyMsg:
		switch v.phase {
		case ExportPhaseSelect:
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
					v.phase = ExportPhaseRunning
					return v, v.runExport(), ""
				}
			case "esc":
				return v, nil, "back"
			}
		case ExportPhaseComplete:
			if msg.String() == "enter" || msg.String() == "esc" {
				return v, nil, "back"
			}
		}
	}
	return v, nil, ""
}

func (v ExportView) runExport() tea.Cmd {
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

		exporter := backup.NewExporter()
		outputPath := backup.GetDefaultExportPath()
		err := exporter.ExportQuick(opts, outputPath)

		var size int64
		if err == nil {
			if info, statErr := os.Stat(outputPath); statErr == nil {
				size = info.Size()
			}
		}

		return exportCompleteMsg{path: outputPath, size: size, err: err}
	}
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func (v ExportView) View() string {
	s := styles.TitleStyle.Render("📦 Quick Export") + "\n"
	s += styles.DescriptionStyle.Render("Create a single portable backup file") + "\n\n"

	switch v.phase {
	case ExportPhaseSelect:
		s += "Select what to include:\n\n"
		s += v.checkboxes.View() + "\n"
		s += styles.DimStyle.Render("Output: "+filepath.Base(v.outputPath)) + "\n\n"
		s += styles.FooterStyle.Render("Space: Toggle • Enter: Export • Esc: Back")

	case ExportPhaseRunning:
		s += styles.WarningStyle.Render("⏳ Creating backup archive...") + "\n\n"
		s += "Please wait, this may take a moment."

	case ExportPhaseComplete:
		if v.error != nil {
			s += styles.ErrorStyle.Render("✗ Export failed: " + v.error.Error())
		} else {
			s += styles.SuccessStyle.Render("✓ Export complete!") + "\n\n"
			s += fmt.Sprintf("📁 File: %s\n", styles.SelectedStyle.Render(v.outputPath))
			s += fmt.Sprintf("📊 Size: %s\n\n", formatBytes(v.fileSize))
			s += styles.DescriptionStyle.Render("Copy this file to a USB drive, cloud storage,\nor email it to yourself for safekeeping!")
		}
		s += "\n\n" + styles.FooterStyle.Render("Press Enter to continue")
	}

	return s
}

// ImportView for importing from a portable backup
type ImportView struct {
	phase     int // 0=input, 1=running, 2=complete
	inputPath string
	error     error
}

func NewImportView() ImportView {
	home, _ := os.UserHomeDir()
	return ImportView{inputPath: filepath.Join(home, "rego-backup-*.tar.gz")}
}

func (v ImportView) Init() tea.Cmd { return nil }

func (v ImportView) Update(msg tea.Msg) (ImportView, tea.Cmd, string) {
	if kmsg, ok := msg.(tea.KeyMsg); ok {
		if kmsg.String() == "esc" {
			return v, nil, "back"
		}
	}
	return v, nil, ""
}

func (v ImportView) View() string {
	s := styles.TitleStyle.Render("📥 Import Backup") + "\n\n"
	s += styles.DescriptionStyle.Render("To import a backup file, run:") + "\n\n"
	s += styles.SelectedStyle.Render("  ./rego --import ~/rego-backup-*.tar.gz") + "\n\n"
	s += styles.FooterStyle.Render("Press Esc to go back")
	return s
}
