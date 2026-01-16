package backup

import (
	"time"
)

// BackupType represents the type of backup component
type BackupType string

const (
	BackupTypeFlatpak         BackupType = "flatpak"
	BackupTypeRPM             BackupType = "rpm"
	BackupTypeRepos           BackupType = "repos"
	BackupTypeGnomeExtensions BackupType = "gnome_extensions"
	BackupTypeGnomeSettings   BackupType = "gnome_settings"
	BackupTypeDotfiles        BackupType = "dotfiles"
	BackupTypeFonts           BackupType = "fonts"
)

// BackupItem represents a single item that can be backed up
type BackupItem struct {
	Name        string            `json:"name"`
	Type        BackupType        `json:"type"`
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// BackupResult holds the result of a backup operation
type BackupResult struct {
	Type      BackupType   `json:"type"`
	Success   bool         `json:"success"`
	Items     []BackupItem `json:"items"`
	ItemCount int          `json:"item_count"`
	Error     string       `json:"error,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
	FilePath  string       `json:"file_path,omitempty"`
}

// BackupManifest contains metadata about a complete backup
type BackupManifest struct {
	Version     string                      `json:"version"`
	CreatedAt   time.Time                   `json:"created_at"`
	Hostname    string                      `json:"hostname"`
	User        string                      `json:"user"`
	Components  []BackupType                `json:"components"`
	Results     map[BackupType]BackupResult `json:"results"`
	BackupPath  string                      `json:"backup_path"`
	Description string                      `json:"description,omitempty"`
}

// BackupOptions configures backup behavior
type BackupOptions struct {
	IncludeFlatpak         bool     `json:"include_flatpak"`
	IncludeRPM             bool     `json:"include_rpm"`
	IncludeRepos           bool     `json:"include_repos"`
	IncludeGnomeExtensions bool     `json:"include_gnome_extensions"`
	IncludeGnomeSettings   bool     `json:"include_gnome_settings"`
	IncludeDotfiles        bool     `json:"include_dotfiles"`
	IncludeFonts           bool     `json:"include_fonts"`
	DotfilesList           []string `json:"dotfiles_list,omitempty"`
	BackupPath             string   `json:"backup_path"`
	Description            string   `json:"description,omitempty"`
}

// DefaultBackupOptions returns sensible defaults
func DefaultBackupOptions() BackupOptions {
	return BackupOptions{
		IncludeFlatpak:         true,
		IncludeRPM:             true,
		IncludeRepos:           true,
		IncludeGnomeExtensions: true,
		IncludeGnomeSettings:   true,
		IncludeDotfiles:        true,
		IncludeFonts:           true,
		DotfilesList:           DefaultDotfiles(),
	}
}

// DefaultDotfiles returns the default list of dotfiles to backup
func DefaultDotfiles() []string {
	return []string{
		".bashrc",
		".bash_profile",
		".bash_aliases",
		".zshrc",
		".zprofile",
		".profile",
		".gitconfig",
		".gitignore_global",
		".vimrc",
		".tmux.conf",
		".ssh/config",
		".config/fish/config.fish",
		".config/starship.toml",
		".config/alacritty/alacritty.toml",
		".config/kitty/kitty.conf",
	}
}

// Backer is the interface for components that can be backed up
type Backer interface {
	// Name returns the display name of this backup component
	Name() string
	// Type returns the backup type
	Type() BackupType
	// Available checks if this backup type is available on the system
	Available() bool
	// List returns items that would be backed up
	List() ([]BackupItem, error)
	// Backup performs the backup to the specified directory
	Backup(backupDir string) (BackupResult, error)
}

// AllBackupTypes returns all available backup types
func AllBackupTypes() []BackupType {
	return []BackupType{
		BackupTypeFlatpak,
		BackupTypeRPM,
		BackupTypeRepos,
		BackupTypeGnomeExtensions,
		BackupTypeGnomeSettings,
		BackupTypeDotfiles,
		BackupTypeFonts,
	}
}

// BackupTypeName returns a human-readable name for a backup type
func BackupTypeName(t BackupType) string {
	names := map[BackupType]string{
		BackupTypeFlatpak:         "Flatpak Applications",
		BackupTypeRPM:             "RPM Packages",
		BackupTypeRepos:           "DNF Repositories",
		BackupTypeGnomeExtensions: "GNOME Extensions",
		BackupTypeGnomeSettings:   "GNOME Settings",
		BackupTypeDotfiles:        "Dotfiles",
		BackupTypeFonts:           "User Fonts",
	}
	if name, ok := names[t]; ok {
		return name
	}
	return string(t)
}

// BackupTypeDescription returns a description for a backup type
func BackupTypeDescription(t BackupType) string {
	descriptions := map[BackupType]string{
		BackupTypeFlatpak:         "Flatpak applications installed from Flathub and other remotes",
		BackupTypeRPM:             "User-installed RPM packages from DNF/YUM",
		BackupTypeRepos:           "Third-party repository configurations and GPG keys",
		BackupTypeGnomeExtensions: "GNOME Shell extensions and their settings",
		BackupTypeGnomeSettings:   "GNOME desktop customizations (dconf database)",
		BackupTypeDotfiles:        "Shell configurations, git settings, and other dotfiles",
		BackupTypeFonts:           "User-installed fonts from ~/.local/share/fonts",
	}
	if desc, ok := descriptions[t]; ok {
		return desc
	}
	return ""
}
