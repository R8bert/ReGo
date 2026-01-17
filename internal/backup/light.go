package backup

import (
	"encoding/json"
	"os"
	"time"

	"github.com/r8bert/rego/internal/utils"
)

// LightBackup is a single-file backup containing just the essential data
type LightBackup struct {
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	Hostname  string    `json:"hostname"`
	User      string    `json:"user"`
	Distro    string    `json:"distro,omitempty"`
	Desktop   string    `json:"desktop,omitempty"` // GNOME, KDE, etc.

	// Package lists (just names, easy to reinstall)
	Flatpaks    []string `json:"flatpaks,omitempty"`
	RPMPackages []string `json:"rpm_packages,omitempty"`
	APTPackages []string `json:"apt_packages,omitempty"`

	// GNOME
	GnomeExtensions []string `json:"gnome_extensions,omitempty"`
	DconfSettings   string   `json:"dconf_settings,omitempty"`

	// KDE Plasma
	KDEWidgets []string `json:"kde_widgets,omitempty"`

	// Repos (just the names)
	Repos []string `json:"repos,omitempty"`
}

// LightBackupOptions controls what to include in the backup
type LightBackupOptions struct {
	Flatpaks   bool
	RPM        bool
	Extensions bool // GNOME extensions
	Settings   bool // GNOME dconf settings
	KDE        bool // KDE Plasma settings
	Repos      bool
}

// DefaultLightBackupOptions returns all options enabled
func DefaultLightBackupOptions() LightBackupOptions {
	return LightBackupOptions{Flatpaks: true, RPM: true, Extensions: true, Settings: true, Repos: true}
}

// CreateLightBackup creates a minimal single-file backup with all options
func CreateLightBackup() (*LightBackup, error) {
	return CreateLightBackupWithOptions(DefaultLightBackupOptions())
}

// CreateLightBackupWithOptions creates a backup with selected components
func CreateLightBackupWithOptions(opts LightBackupOptions) (*LightBackup, error) {
	hostname, _ := os.Hostname()

	backup := &LightBackup{
		Version:   "1.0",
		CreatedAt: time.Now(),
		Hostname:  hostname,
		User:      os.Getenv("USER"),
		Distro:    GetDistroName(),
	}

	// Flatpaks
	if opts.Flatpaks && utils.CommandExists("flatpak") {
		lines, _ := utils.RunCommandLines("flatpak", "list", "--app", "--columns=application")
		for _, line := range lines {
			if line != "" {
				backup.Flatpaks = append(backup.Flatpaks, line)
			}
		}
	}

	// System packages - auto-detect package manager
	if opts.RPM {
		pm := DetectPackageManager()
		switch pm {
		case PMDNF:
			result := utils.RunCommand("dnf", "repoquery", "--userinstalled", "--qf", "%{name}")
			if result.Error == nil {
				for _, line := range splitLines(result.Stdout) {
					if line != "" && !isBasePackage(line) {
						backup.RPMPackages = append(backup.RPMPackages, line)
					}
				}
			}
		case PMAPT:
			apt := NewAPTBackup()
			pkgs, _ := apt.ListUserInstalled()
			backup.APTPackages = pkgs
		}
	}

	// GNOME extensions
	if opts.Extensions && utils.CommandExists("gnome-extensions") {
		lines, _ := utils.RunCommandLines("gnome-extensions", "list")
		backup.GnomeExtensions = lines
	}

	// Dconf settings
	if opts.Settings && utils.CommandExists("dconf") {
		result := utils.RunCommand("dconf", "dump", "/")
		if result.Error == nil {
			backup.DconfSettings = result.Stdout
		}
	}

	// Repos
	if opts.Repos {
		reposBackup := NewReposBackup()
		if items, err := reposBackup.List(); err == nil {
			for _, item := range items {
				backup.Repos = append(backup.Repos, item.Name)
			}
		}
	}

	return backup, nil
}

func splitLines(s string) []string {
	var lines []string
	for _, line := range []byte(s) {
		if line == '\n' {
			continue
		}
	}
	// Simple split
	result := ""
	for _, c := range s {
		if c == '\n' {
			if result != "" {
				lines = append(lines, result)
			}
			result = ""
		} else {
			result += string(c)
		}
	}
	if result != "" {
		lines = append(lines, result)
	}
	return lines
}

func isBasePackage(pkg string) bool {
	base := map[string]bool{
		"fedora-release": true, "fedora-repos": true, "fedora-gpg-keys": true,
		"basesystem": true, "filesystem": true, "setup": true,
		"glibc": true, "glibc-common": true, "bash": true,
		"coreutils": true, "systemd": true, "kernel": true,
	}
	return base[pkg]
}

// SaveToFile saves the backup to a JSON file
func (b *LightBackup) SaveToFile(path string) error {
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadLightBackup loads a backup from a JSON file
func LoadLightBackup(path string) (*LightBackup, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var backup LightBackup
	if err := json.Unmarshal(data, &backup); err != nil {
		return nil, err
	}
	return &backup, nil
}

// GetDefaultLightBackupPath returns the default path for the light backup
func GetDefaultLightBackupPath() string {
	home, _ := utils.GetHomeDir()
	hostname, _ := os.Hostname()
	return home + "/rego-" + hostname + ".json"
}

// Stats returns a summary of what's in the backup
func (b *LightBackup) Stats() map[string]int {
	return map[string]int{
		"flatpaks":   len(b.Flatpaks),
		"rpm":        len(b.RPMPackages),
		"apt":        len(b.APTPackages),
		"extensions": len(b.GnomeExtensions),
		"repos":      len(b.Repos),
	}
}
