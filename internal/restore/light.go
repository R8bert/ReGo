package restore

import (
	"fmt"
	"time"

	"github.com/r8bert/rego/internal/backup"
	"github.com/r8bert/rego/internal/utils"
)

// LightRestore restores from a light backup
type LightRestore struct {
	backup *backup.LightBackup
	dryRun bool
}

func NewLightRestore(b *backup.LightBackup, dryRun bool) *LightRestore {
	return &LightRestore{backup: b, dryRun: dryRun}
}

// RestoreFlatpaks installs all Flatpak apps
func (r *LightRestore) RestoreFlatpaks() (int, int, error) {
	if len(r.backup.Flatpaks) == 0 {
		return 0, 0, nil
	}

	if r.dryRun {
		return len(r.backup.Flatpaks), 0, nil
	}

	// Add flathub if not present
	utils.RunCommand("flatpak", "remote-add", "--if-not-exists", "flathub", "https://flathub.org/repo/flathub.flatpakrepo")

	success, failed := 0, 0
	for _, app := range r.backup.Flatpaks {
		result := utils.RunCommandWithTimeout("flatpak", 5*time.Minute, "install", "-y", "--noninteractive", "flathub", app)
		if result.Error != nil {
			failed++
		} else {
			success++
		}
	}
	return success, failed, nil
}

// RestoreRPM installs all RPM packages
func (r *LightRestore) RestoreRPM() (int, int, error) {
	if len(r.backup.RPMPackages) == 0 {
		return 0, 0, nil
	}

	if r.dryRun {
		return len(r.backup.RPMPackages), 0, nil
	}

	args := append([]string{"install", "-y"}, r.backup.RPMPackages...)
	result := utils.RunCommandWithTimeout("dnf", 30*time.Minute, args...)
	if result.Error != nil {
		return 0, len(r.backup.RPMPackages), fmt.Errorf("dnf install failed: %s", result.Stderr)
	}
	return len(r.backup.RPMPackages), 0, nil
}

// RestoreExtensions installs GNOME extensions
func (r *LightRestore) RestoreExtensions() (int, int, error) {
	if len(r.backup.GnomeExtensions) == 0 {
		return 0, 0, nil
	}

	if r.dryRun {
		return len(r.backup.GnomeExtensions), 0, nil
	}

	success, failed := 0, 0
	for _, ext := range r.backup.GnomeExtensions {
		result := utils.RunCommand("gnome-extensions", "install", ext)
		if result.Error != nil {
			// Try enabling if already installed
			utils.RunCommand("gnome-extensions", "enable", ext)
			failed++
		} else {
			success++
		}
	}
	return success, failed, nil
}

// RestoreDconf loads dconf settings
func (r *LightRestore) RestoreDconf() error {
	if r.backup.DconfSettings == "" {
		return nil
	}

	if r.dryRun {
		return nil
	}

	// Write to temp file and load
	tmpFile := "/tmp/rego-dconf-restore"
	if err := utils.WriteFile(tmpFile, []byte(r.backup.DconfSettings)); err != nil {
		return err
	}

	result := utils.RunCommand("sh", "-c", "cat "+tmpFile+" | dconf load /")
	if result.Error != nil {
		return fmt.Errorf("dconf load failed: %s", result.Stderr)
	}
	return nil
}
