package restore

import (
	"time"
)

// RestoreType mirrors backup types for consistency
type RestoreType string

const (
	RestoreTypeFlatpak         RestoreType = "flatpak"
	RestoreTypeRPM             RestoreType = "rpm"
	RestoreTypeRepos           RestoreType = "repos"
	RestoreTypeGnomeExtensions RestoreType = "gnome_extensions"
	RestoreTypeGnomeSettings   RestoreType = "gnome_settings"
	RestoreTypeDotfiles        RestoreType = "dotfiles"
	RestoreTypeFonts           RestoreType = "fonts"
)

// RestoreResult holds the result of a restore operation
type RestoreResult struct {
	Type         RestoreType `json:"type"`
	Success      bool        `json:"success"`
	ItemsTotal   int         `json:"items_total"`
	ItemsSuccess int         `json:"items_success"`
	ItemsFailed  int         `json:"items_failed"`
	Errors       []string    `json:"errors,omitempty"`
	Timestamp    time.Time   `json:"timestamp"`
	DryRun       bool        `json:"dry_run"`
}

// RestoreOptions configures restore behavior
type RestoreOptions struct {
	BackupPath             string   `json:"backup_path"`
	DryRun                 bool     `json:"dry_run"`
	IncludeFlatpak         bool     `json:"include_flatpak"`
	IncludeRPM             bool     `json:"include_rpm"`
	IncludeRepos           bool     `json:"include_repos"`
	IncludeGnomeExtensions bool     `json:"include_gnome_extensions"`
	IncludeGnomeSettings   bool     `json:"include_gnome_settings"`
	IncludeDotfiles        bool     `json:"include_dotfiles"`
	IncludeFonts           bool     `json:"include_fonts"`
	MergeDotfiles          bool     `json:"merge_dotfiles"`               // false = overwrite
	SelectiveSettings      []string `json:"selective_settings,omitempty"` // Specific dconf paths
}

// DefaultRestoreOptions returns sensible defaults
func DefaultRestoreOptions() RestoreOptions {
	return RestoreOptions{
		DryRun:                 true, // Default to dry run for safety
		IncludeFlatpak:         true,
		IncludeRPM:             true,
		IncludeRepos:           true,
		IncludeGnomeExtensions: true,
		IncludeGnomeSettings:   true,
		IncludeDotfiles:        true,
		IncludeFonts:           true,
		MergeDotfiles:          false,
	}
}

// Restorer is the interface for components that can be restored
type Restorer interface {
	// Name returns the display name of this restore component
	Name() string
	// Type returns the restore type
	Type() RestoreType
	// Available checks if this restore is possible on the system
	Available() bool
	// Preview returns what would be restored without making changes
	Preview(backupDir string) ([]string, error)
	// Restore performs the restore from the backup directory
	Restore(backupDir string, dryRun bool) (RestoreResult, error)
}

// AllRestoreTypes returns all restore types
func AllRestoreTypes() []RestoreType {
	return []RestoreType{
		RestoreTypeFlatpak,
		RestoreTypeRPM,
		RestoreTypeRepos,
		RestoreTypeGnomeExtensions,
		RestoreTypeGnomeSettings,
		RestoreTypeDotfiles,
		RestoreTypeFonts,
	}
}

// RestoreTypeName returns a human-readable name for a restore type
func RestoreTypeName(t RestoreType) string {
	names := map[RestoreType]string{
		RestoreTypeFlatpak:         "Flatpak Applications",
		RestoreTypeRPM:             "RPM Packages",
		RestoreTypeRepos:           "DNF Repositories",
		RestoreTypeGnomeExtensions: "GNOME Extensions",
		RestoreTypeGnomeSettings:   "GNOME Settings",
		RestoreTypeDotfiles:        "Dotfiles",
		RestoreTypeFonts:           "User Fonts",
	}
	if name, ok := names[t]; ok {
		return name
	}
	return string(t)
}
