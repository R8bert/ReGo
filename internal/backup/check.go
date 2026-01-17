package backup

import (
	"os/exec"
	"strings"
)

// GetInstalledFlatpaks returns a list of currently installed Flatpak app IDs
func GetInstalledFlatpaks() []string {
	cmd := exec.Command("flatpak", "list", "--app", "--columns=application")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var installed []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			installed = append(installed, line)
		}
	}
	return installed
}

// GetInstalledRPM returns a list of currently installed RPM package names
func GetInstalledRPM() []string {
	cmd := exec.Command("rpm", "-qa", "--qf", "%{NAME}\n")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var installed []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			installed = append(installed, line)
		}
	}
	return installed
}

// GetInstalledAPT returns a list of currently installed APT package names
func GetInstalledAPT() []string {
	cmd := exec.Command("dpkg-query", "-W", "-f=${Package}\n")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var installed []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			installed = append(installed, line)
		}
	}
	return installed
}

// GetEnabledGnomeExtensions returns a list of enabled GNOME extension UUIDs
func GetEnabledGnomeExtensions() []string {
	cmd := exec.Command("gnome-extensions", "list", "--enabled")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var enabled []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			enabled = append(enabled, line)
		}
	}
	return enabled
}

// FilterMissing returns items from 'wanted' that are not in 'installed'
func FilterMissing(wanted, installed []string) []string {
	installedSet := make(map[string]bool)
	for _, item := range installed {
		installedSet[item] = true
	}

	var missing []string
	for _, item := range wanted {
		if !installedSet[item] {
			missing = append(missing, item)
		}
	}
	return missing
}

// RestoreCheck holds the results of checking what needs to be installed
type RestoreCheck struct {
	FlatpaksToInstall  []string
	FlatpaksSkipped    int
	RPMToInstall       []string
	RPMSkipped         int
	APTToInstall       []string
	APTSkipped         int
	ExtensionsToEnable []string
	ExtensionsSkipped  int
	HasDconfSettings   bool
}

// CheckRestore analyzes a backup and returns what actually needs to be installed
func CheckRestore(b *LightBackup) *RestoreCheck {
	check := &RestoreCheck{}

	// Check Flatpaks
	if len(b.Flatpaks) > 0 {
		installed := GetInstalledFlatpaks()
		check.FlatpaksToInstall = FilterMissing(b.Flatpaks, installed)
		check.FlatpaksSkipped = len(b.Flatpaks) - len(check.FlatpaksToInstall)
	}

	// Check RPM packages
	if len(b.RPMPackages) > 0 && DetectPackageManager() == PMDNF {
		installed := GetInstalledRPM()
		check.RPMToInstall = FilterMissing(b.RPMPackages, installed)
		check.RPMSkipped = len(b.RPMPackages) - len(check.RPMToInstall)
	}

	// Check APT packages
	if len(b.APTPackages) > 0 && DetectPackageManager() == PMAPT {
		installed := GetInstalledAPT()
		check.APTToInstall = FilterMissing(b.APTPackages, installed)
		check.APTSkipped = len(b.APTPackages) - len(check.APTToInstall)
	}

	// Check GNOME extensions
	if len(b.GnomeExtensions) > 0 {
		enabled := GetEnabledGnomeExtensions()
		check.ExtensionsToEnable = FilterMissing(b.GnomeExtensions, enabled)
		check.ExtensionsSkipped = len(b.GnomeExtensions) - len(check.ExtensionsToEnable)
	}

	check.HasDconfSettings = b.DconfSettings != ""

	return check
}
