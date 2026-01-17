package backup

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/r8bert/rego/internal/utils"
)

// PackageManager represents the system's package manager
type PackageManager string

const (
	PMUnknown PackageManager = "unknown"
	PMDNF     PackageManager = "dnf"
	PMAPT     PackageManager = "apt"
	PMPacman  PackageManager = "pacman"
	PMZypper  PackageManager = "zypper"
)

// DetectPackageManager returns the detected package manager
func DetectPackageManager() PackageManager {
	// Check for apt (Debian/Ubuntu)
	if utils.CommandExists("apt") && utils.FileExists("/etc/debian_version") {
		return PMAPT
	}
	// Check for dnf (Fedora/RHEL)
	if utils.CommandExists("dnf") {
		return PMDNF
	}
	// Check for pacman (Arch)
	if utils.CommandExists("pacman") {
		return PMPacman
	}
	// Check for zypper (openSUSE)
	if utils.CommandExists("zypper") {
		return PMZypper
	}
	return PMUnknown
}

// GetPackageManagerName returns a friendly name
func GetPackageManagerName() string {
	switch DetectPackageManager() {
	case PMAPT:
		return "APT (Debian/Ubuntu)"
	case PMDNF:
		return "DNF (Fedora/RHEL)"
	case PMPacman:
		return "Pacman (Arch)"
	case PMZypper:
		return "Zypper (openSUSE)"
	default:
		return "Unknown"
	}
}

// APTBackup handles apt package backup for Debian/Ubuntu
type APTBackup struct{}

func NewAPTBackup() *APTBackup { return &APTBackup{} }

// ListUserInstalled returns manually installed packages
func (a *APTBackup) ListUserInstalled() ([]string, error) {
	// apt-mark showmanual lists manually installed packages
	result := utils.RunCommand("apt-mark", "showmanual")
	if result.Error != nil {
		return nil, result.Error
	}

	var packages []string
	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !a.isBasePackage(line) {
			packages = append(packages, line)
		}
	}
	return packages, nil
}

// isBasePackage filters out base system packages
func (a *APTBackup) isBasePackage(pkg string) bool {
	base := map[string]bool{
		"apt": true, "dpkg": true, "base-files": true, "init": true,
		"systemd": true, "libc6": true, "bash": true, "coreutils": true,
		"ubuntu-minimal": true, "ubuntu-standard": true, "ubuntu-desktop": true,
		"debian-archive-keyring": true, "ubuntu-keyring": true,
	}
	return base[pkg]
}

// BackupSources backs up APT source lists
func (a *APTBackup) BackupSources(destDir string) (int, error) {
	utils.EnsureDir(destDir)
	count := 0

	// Copy /etc/apt/sources.list
	if utils.FileExists("/etc/apt/sources.list") {
		if utils.CopyFile("/etc/apt/sources.list", filepath.Join(destDir, "sources.list")) == nil {
			count++
		}
	}

	// Copy /etc/apt/sources.list.d/*.list
	sourcesDir := "/etc/apt/sources.list.d"
	if utils.DirExists(sourcesDir) {
		files, _ := filepath.Glob(filepath.Join(sourcesDir, "*.list"))
		for _, f := range files {
			if utils.CopyFile(f, filepath.Join(destDir, filepath.Base(f))) == nil {
				count++
			}
		}
	}

	// Copy trusted GPG keys
	keysDir := "/etc/apt/trusted.gpg.d"
	if utils.DirExists(keysDir) {
		keysDest := filepath.Join(destDir, "trusted.gpg.d")
		utils.EnsureDir(keysDest)
		files, _ := filepath.Glob(filepath.Join(keysDir, "*"))
		for _, f := range files {
			if utils.CopyFile(f, filepath.Join(keysDest, filepath.Base(f))) == nil {
				count++
			}
		}
	}

	return count, nil
}

// GetDistro returns the Linux distribution name
func GetDistro() string {
	// Try /etc/os-release
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "ID=") {
				return strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
			}
		}
	}
	// Fallback checks
	if utils.FileExists("/etc/fedora-release") {
		return "fedora"
	}
	if utils.FileExists("/etc/debian_version") {
		return "debian"
	}
	if utils.FileExists("/etc/arch-release") {
		return "arch"
	}
	return "unknown"
}

// GetDistroName returns a friendly distro name
func GetDistroName() string {
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
			}
		}
	}
	return GetDistro()
}
