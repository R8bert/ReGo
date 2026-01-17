package backup

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/r8bert/rego/internal/utils"
)

// KDEBackup handles KDE Plasma configuration backup
type KDEBackup struct {
	home string
}

func NewKDEBackup() *KDEBackup {
	home, _ := utils.GetHomeDir()
	return &KDEBackup{home: home}
}

// IsKDE checks if KDE Plasma is installed/running
func IsKDE() bool {
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")
	if strings.Contains(strings.ToLower(desktop), "kde") || strings.Contains(strings.ToLower(desktop), "plasma") {
		return true
	}
	home, _ := utils.GetHomeDir()
	return utils.FileExists(filepath.Join(home, ".config", "plasma-org.kde.plasma.desktop-appletsrc"))
}

// IsGNOME checks if GNOME is installed/running
func IsGNOME() bool {
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")
	if strings.Contains(strings.ToLower(desktop), "gnome") {
		return true
	}
	return utils.CommandExists("gnome-extensions") || utils.CommandExists("dconf")
}

// KDEConfigFiles returns the list of important KDE config files
func (k *KDEBackup) KDEConfigFiles() []string {
	return []string{
		// Plasma Desktop
		".config/plasma-org.kde.plasma.desktop-appletsrc",
		".config/plasmashellrc",
		".config/plasmarc",
		".config/plasmadashboardrc",

		// KWin (window manager)
		".config/kwinrc",
		".config/kwinrulesrc",
		".config/kglobalshortcutsrc",

		// System Settings
		".config/kdeglobals",
		".config/kcminputrc",
		".config/kscreenlockerrc",
		".config/powermanagementprofilesrc",

		// Look and Feel
		".config/Trolltech.conf",
		".config/gtkrc",
		".config/gtkrc-2.0",
		".config/gtk-3.0/settings.ini",
		".config/gtk-4.0/settings.ini",

		// Konsole
		".config/konsolerc",
		".local/share/konsole/*.profile",
		".local/share/konsole/*.colorscheme",

		// Dolphin
		".config/dolphinrc",

		// Kate/KWrite
		".config/katerc",
		".config/kaboretrc",

		// Application settings
		".config/kiorc",
		".config/kiaboretrc",
		".config/kaboretrc",
	}
}

// KDEDataDirs returns directories with KDE data
func (k *KDEBackup) KDEDataDirs() []string {
	return []string{
		".local/share/plasma",
		".local/share/kwin",
		".local/share/color-schemes",
		".local/share/aurorae", // Window decorations
		".local/share/plasma/look-and-feel",
		".local/share/plasma/plasmoids", // Widgets
		".local/share/plasma/desktoptheme",
		".local/share/wallpapers",
	}
}

// BackupConfigs copies KDE config files to destination
func (k *KDEBackup) BackupConfigs(destDir string) (int, error) {
	count := 0
	configDest := filepath.Join(destDir, "kde-config")
	utils.EnsureDir(configDest)

	for _, relPath := range k.KDEConfigFiles() {
		// Handle wildcards
		if strings.Contains(relPath, "*") {
			basePath := filepath.Join(k.home, filepath.Dir(relPath))
			pattern := filepath.Base(relPath)
			if utils.DirExists(basePath) {
				matches, _ := filepath.Glob(filepath.Join(basePath, pattern))
				for _, match := range matches {
					dest := filepath.Join(configDest, filepath.Base(match))
					if utils.CopyFile(match, dest) == nil {
						count++
					}
				}
			}
		} else {
			src := filepath.Join(k.home, relPath)
			if utils.FileExists(src) {
				// Preserve directory structure for nested files
				dest := filepath.Join(configDest, filepath.Base(relPath))
				if utils.CopyFile(src, dest) == nil {
					count++
				}
			}
		}
	}
	return count, nil
}

// BackupData copies KDE data directories to destination
func (k *KDEBackup) BackupData(destDir string) (int, error) {
	count := 0
	dataDest := filepath.Join(destDir, "kde-data")
	utils.EnsureDir(dataDest)

	for _, relPath := range k.KDEDataDirs() {
		src := filepath.Join(k.home, relPath)
		if utils.DirExists(src) {
			dirName := filepath.Base(relPath)
			dest := filepath.Join(dataDest, dirName)
			if utils.CopyDir(src, dest) == nil {
				files, _ := utils.ListFilesRecursive(dest)
				count += len(files)
			}
		}
	}
	return count, nil
}

// GetInstalledWidgets returns a list of installed Plasma widgets
func (k *KDEBackup) GetInstalledWidgets() []string {
	var widgets []string
	widgetDir := filepath.Join(k.home, ".local", "share", "plasma", "plasmoids")
	if utils.DirExists(widgetDir) {
		entries, _ := os.ReadDir(widgetDir)
		for _, e := range entries {
			if e.IsDir() {
				widgets = append(widgets, e.Name())
			}
		}
	}
	return widgets
}
