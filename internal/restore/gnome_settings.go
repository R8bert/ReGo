package restore

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/r8bert/rego/internal/utils"
)

// GnomeSettingsRestore handles GNOME settings restoration
type GnomeSettingsRestore struct {
	selectivePaths []string
}

// NewGnomeSettingsRestore creates a new GnomeSettingsRestore instance
func NewGnomeSettingsRestore() *GnomeSettingsRestore {
	return &GnomeSettingsRestore{}
}

// NewGnomeSettingsRestoreSelective creates a restore with specific paths only
func NewGnomeSettingsRestoreSelective(paths []string) *GnomeSettingsRestore {
	return &GnomeSettingsRestore{
		selectivePaths: paths,
	}
}

// Name returns the display name
func (g *GnomeSettingsRestore) Name() string {
	return "GNOME Settings"
}

// Type returns the restore type
func (g *GnomeSettingsRestore) Type() RestoreType {
	return RestoreTypeGnomeSettings
}

// Available checks if dconf is available
func (g *GnomeSettingsRestore) Available() bool {
	return utils.CommandExists("dconf")
}

// Preview returns what would be restored
func (g *GnomeSettingsRestore) Preview(backupDir string) ([]string, error) {
	settingsFile := filepath.Join(backupDir, "gnome_settings.dconf")
	if !utils.FileExists(settingsFile) {
		return nil, fmt.Errorf("settings backup not found")
	}

	// List selective backups if available
	selectiveDir := filepath.Join(backupDir, "gnome_settings_selective")
	if utils.DirExists(selectiveDir) {
		files, _ := utils.ListFiles(selectiveDir)
		var items []string
		for _, f := range files {
			items = append(items, filepath.Base(f))
		}
		return items, nil
	}

	return []string{"Full dconf database restore"}, nil
}

// Restore performs the GNOME settings restoration
func (g *GnomeSettingsRestore) Restore(backupDir string, dryRun bool) (RestoreResult, error) {
	result := RestoreResult{
		Type:      RestoreTypeGnomeSettings,
		Timestamp: time.Now(),
		DryRun:    dryRun,
	}

	if dryRun {
		result.Success = true
		result.ItemsTotal = 1
		result.ItemsSuccess = 1
		return result, nil
	}

	// Prefer selective restore if paths specified
	if len(g.selectivePaths) > 0 {
		return g.restoreSelective(backupDir, result)
	}

	// Try selective restore from individual files (safer)
	selectiveDir := filepath.Join(backupDir, "gnome_settings_selective")
	if utils.DirExists(selectiveDir) {
		return g.restoreFromSelectiveDir(selectiveDir, result)
	}

	// Fall back to full restore
	return g.restoreFull(backupDir, result)
}

// restoreFull restores the full dconf database
func (g *GnomeSettingsRestore) restoreFull(backupDir string, result RestoreResult) (RestoreResult, error) {
	settingsFile := filepath.Join(backupDir, "gnome_settings.dconf")
	if !utils.FileExists(settingsFile) {
		result.Errors = append(result.Errors, "settings backup not found")
		return result, fmt.Errorf("settings backup not found")
	}

	content, err := os.ReadFile(settingsFile)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	// Write to temp file and pipe to dconf
	tmpFile := filepath.Join(os.TempDir(), "rego_dconf_restore.dconf")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}
	defer os.Remove(tmpFile)

	cmdResult := utils.RunCommand("sh", "-c", fmt.Sprintf("cat %s | dconf load /", tmpFile))
	if cmdResult.Error != nil {
		result.Errors = append(result.Errors, cmdResult.Stderr)
		return result, cmdResult.Error
	}

	result.Success = true
	result.ItemsTotal = 1
	result.ItemsSuccess = 1
	return result, nil
}

// restoreFromSelectiveDir restores from individual path files
func (g *GnomeSettingsRestore) restoreFromSelectiveDir(selectiveDir string, result RestoreResult) (RestoreResult, error) {
	files, err := utils.ListFiles(selectiveDir)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	result.ItemsTotal = len(files)

	pathMap := map[string]string{
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

	for _, file := range files {
		baseName := filepath.Base(file)
		name := baseName[:len(baseName)-len(".dconf")]

		path, ok := pathMap[name]
		if !ok {
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			result.ItemsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to read %s: %v", name, err))
			continue
		}

		// Write to temp file and pipe to dconf
		tmpFile := filepath.Join(os.TempDir(), "rego_dconf_"+name+".dconf")
		if err := os.WriteFile(tmpFile, content, 0644); err != nil {
			result.ItemsFailed++
			continue
		}

		cmdResult := utils.RunCommand("sh", "-c", fmt.Sprintf("cat %s | dconf load %s", tmpFile, path))
		os.Remove(tmpFile)

		if cmdResult.Error != nil {
			result.ItemsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to restore %s: %s", name, cmdResult.Stderr))
		} else {
			result.ItemsSuccess++
		}
	}

	result.Success = result.ItemsFailed == 0
	return result, nil
}

// restoreSelective restores only specified paths
func (g *GnomeSettingsRestore) restoreSelective(backupDir string, result RestoreResult) (RestoreResult, error) {
	selectiveDir := filepath.Join(backupDir, "gnome_settings_selective")

	result.ItemsTotal = len(g.selectivePaths)

	for _, pathName := range g.selectivePaths {
		file := filepath.Join(selectiveDir, pathName+".dconf")
		if !utils.FileExists(file) {
			result.ItemsFailed++
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			result.ItemsFailed++
			continue
		}

		tmpFile := filepath.Join(os.TempDir(), "rego_dconf_selective.dconf")
		os.WriteFile(tmpFile, content, 0644)

		// The path needs to be constructed based on the name
		// This is simplified - actual implementation needs proper path mapping
		cmdResult := utils.RunCommand("sh", "-c", fmt.Sprintf("cat %s | dconf load /", tmpFile))
		os.Remove(tmpFile)

		if cmdResult.Error != nil {
			result.ItemsFailed++
		} else {
			result.ItemsSuccess++
		}
	}

	result.Success = result.ItemsFailed == 0
	return result, nil
}
