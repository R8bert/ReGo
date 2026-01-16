package backup

import (
	"path/filepath"
	"time"

	"github.com/r8bert/rego/internal/utils"
)

// GnomeSettingsBackup handles GNOME dconf settings backup
type GnomeSettingsBackup struct{}

// NewGnomeSettingsBackup creates a new GnomeSettingsBackup instance
func NewGnomeSettingsBackup() *GnomeSettingsBackup {
	return &GnomeSettingsBackup{}
}

// Name returns the display name
func (g *GnomeSettingsBackup) Name() string {
	return "GNOME Settings"
}

// Type returns the backup type
func (g *GnomeSettingsBackup) Type() BackupType {
	return BackupTypeGnomeSettings
}

// Available checks if dconf is available
func (g *GnomeSettingsBackup) Available() bool {
	return utils.CommandExists("dconf")
}

// List returns a summary of dconf paths that will be backed up
func (g *GnomeSettingsBackup) List() ([]BackupItem, error) {
	// Return key areas that will be backed up
	paths := []struct {
		name string
		path string
	}{
		{"Desktop Interface", "/org/gnome/desktop/interface/"},
		{"Desktop Background", "/org/gnome/desktop/background/"},
		{"Desktop Sound", "/org/gnome/desktop/sound/"},
		{"Shell Settings", "/org/gnome/shell/"},
		{"Window Manager", "/org/gnome/desktop/wm/"},
		{"Keyboard Shortcuts", "/org/gnome/desktop/wm/keybindings/"},
		{"Terminal Settings", "/org/gnome/terminal/"},
		{"Nautilus Settings", "/org/gnome/nautilus/"},
		{"GTK Settings", "/org/gtk/"},
	}

	var items []BackupItem
	for _, p := range paths {
		// Check if path has any keys
		result := utils.RunCommand("dconf", "list", p.path)
		if result.Error == nil && result.Stdout != "" {
			items = append(items, BackupItem{
				Name:        p.name,
				Type:        BackupTypeGnomeSettings,
				Description: p.path,
			})
		}
	}

	return items, nil
}

// Backup performs the GNOME settings backup using dconf dump
func (g *GnomeSettingsBackup) Backup(backupDir string) (BackupResult, error) {
	result := BackupResult{
		Type:      BackupTypeGnomeSettings,
		Timestamp: time.Now(),
	}

	items, err := g.List()
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Dump entire dconf database
	dconfResult := utils.RunCommand("dconf", "dump", "/")
	if dconfResult.Error != nil {
		result.Error = dconfResult.Error.Error()
		return result, dconfResult.Error
	}

	// Write full dump
	filePath := filepath.Join(backupDir, "gnome_settings.dconf")
	if err := utils.WriteFile(filePath, []byte(dconfResult.Stdout)); err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Also create selective backups for safer restore
	selectiveDir := filepath.Join(backupDir, "gnome_settings_selective")
	utils.EnsureDir(selectiveDir)

	// Key paths to backup selectively
	paths := map[string]string{
		"desktop_interface":   "/org/gnome/desktop/interface/",
		"desktop_background":  "/org/gnome/desktop/background/",
		"desktop_peripherals": "/org/gnome/desktop/peripherals/",
		"desktop_wm":          "/org/gnome/desktop/wm/",
		"shell":               "/org/gnome/shell/",
		"terminal":            "/org/gnome/terminal/",
		"nautilus":            "/org/gnome/nautilus/",
		"gtk_settings":        "/org/gtk/settings/",
		"mutter":              "/org/gnome/mutter/",
	}

	for name, path := range paths {
		dumpResult := utils.RunCommand("dconf", "dump", path)
		if dumpResult.Error == nil && dumpResult.Stdout != "" {
			pathFile := filepath.Join(selectiveDir, name+".dconf")
			utils.WriteFile(pathFile, []byte(dumpResult.Stdout))
		}
	}

	result.Success = true
	result.Items = items
	result.ItemCount = len(items)
	result.FilePath = filePath

	return result, nil
}
