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

	// Package lists (just names, easy to reinstall)
	Flatpaks    []string `json:"flatpaks,omitempty"`
	RPMPackages []string `json:"rpm_packages,omitempty"`

	// GNOME
	GnomeExtensions []string `json:"gnome_extensions,omitempty"`
	DconfSettings   string   `json:"dconf_settings,omitempty"` // Raw dconf dump

	// Repos (just the names)
	Repos []string `json:"repos,omitempty"`
}

// CreateLightBackup creates a minimal single-file backup
func CreateLightBackup() (*LightBackup, error) {
	hostname, _ := os.Hostname()

	backup := &LightBackup{
		Version:   "1.0",
		CreatedAt: time.Now(),
		Hostname:  hostname,
		User:      os.Getenv("USER"),
	}

	// Flatpaks - just the app IDs
	if utils.CommandExists("flatpak") {
		lines, _ := utils.RunCommandLines("flatpak", "list", "--app", "--columns=application")
		for _, line := range lines {
			if line != "" {
				backup.Flatpaks = append(backup.Flatpaks, line)
			}
		}
	}

	// RPM packages - user installed
	if utils.CommandExists("dnf") {
		result := utils.RunCommand("dnf", "repoquery", "--userinstalled", "--qf", "%{name}")
		if result.Error == nil {
			for _, line := range splitLines(result.Stdout) {
				if line != "" && !isBasePackage(line) {
					backup.RPMPackages = append(backup.RPMPackages, line)
				}
			}
		}
	}

	// GNOME extensions - just UUIDs
	if utils.CommandExists("gnome-extensions") {
		lines, _ := utils.RunCommandLines("gnome-extensions", "list")
		backup.GnomeExtensions = lines
	}

	// Dconf settings - full dump (usually only a few KB)
	if utils.CommandExists("dconf") {
		result := utils.RunCommand("dconf", "dump", "/")
		if result.Error == nil {
			backup.DconfSettings = result.Stdout
		}
	}

	// Repos - just names of third-party repos
	reposBackup := NewReposBackup()
	if items, err := reposBackup.List(); err == nil {
		for _, item := range items {
			backup.Repos = append(backup.Repos, item.Name)
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
		"extensions": len(b.GnomeExtensions),
		"repos":      len(b.Repos),
	}
}
