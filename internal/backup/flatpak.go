package backup

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"time"

	"github.com/r8bert/rego/internal/utils"
)

// FlatpakBackup handles Flatpak application backup
type FlatpakBackup struct{}

// NewFlatpakBackup creates a new FlatpakBackup instance
func NewFlatpakBackup() *FlatpakBackup {
	return &FlatpakBackup{}
}

// Name returns the display name
func (f *FlatpakBackup) Name() string {
	return "Flatpak Applications"
}

// Type returns the backup type
func (f *FlatpakBackup) Type() BackupType {
	return BackupTypeFlatpak
}

// Available checks if Flatpak is installed
func (f *FlatpakBackup) Available() bool {
	return utils.CommandExists("flatpak")
}

// List returns all installed Flatpak applications
func (f *FlatpakBackup) List() ([]BackupItem, error) {
	// Get list of installed applications (not runtimes)
	lines, err := utils.RunCommandLines("flatpak", "list", "--app", "--columns=application,name,branch,origin")
	if err != nil {
		return nil, err
	}

	var items []BackupItem
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 1 {
			continue
		}

		appID := strings.TrimSpace(parts[0])
		if appID == "" {
			continue
		}

		item := BackupItem{
			Name:     appID,
			Type:     BackupTypeFlatpak,
			Metadata: make(map[string]string),
		}

		if len(parts) > 1 {
			item.Description = strings.TrimSpace(parts[1])
		}
		if len(parts) > 2 {
			item.Metadata["branch"] = strings.TrimSpace(parts[2])
		}
		if len(parts) > 3 {
			item.Metadata["origin"] = strings.TrimSpace(parts[3])
		}

		items = append(items, item)
	}

	return items, nil
}

// FlatpakData represents the backup data structure
type FlatpakData struct {
	Applications []BackupItem    `json:"applications"`
	Remotes      []FlatpakRemote `json:"remotes"`
}

// FlatpakRemote represents a Flatpak remote
type FlatpakRemote struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Options string `json:"options,omitempty"`
}

// ListRemotes returns all configured Flatpak remotes
func (f *FlatpakBackup) ListRemotes() ([]FlatpakRemote, error) {
	lines, err := utils.RunCommandLines("flatpak", "remotes", "--columns=name,url,options")
	if err != nil {
		return nil, err
	}

	var remotes []FlatpakRemote
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		remote := FlatpakRemote{
			Name: strings.TrimSpace(parts[0]),
			URL:  strings.TrimSpace(parts[1]),
		}

		if len(parts) > 2 {
			remote.Options = strings.TrimSpace(parts[2])
		}

		remotes = append(remotes, remote)
	}

	return remotes, nil
}

// Backup performs the Flatpak backup
func (f *FlatpakBackup) Backup(backupDir string) (BackupResult, error) {
	result := BackupResult{
		Type:      BackupTypeFlatpak,
		Timestamp: time.Now(),
	}

	// Get applications
	apps, err := f.List()
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Get remotes
	remotes, err := f.ListRemotes()
	if err != nil {
		utils.Warn("Failed to list Flatpak remotes: %v", err)
		// Continue anyway, apps are more important
	}

	data := FlatpakData{
		Applications: apps,
		Remotes:      remotes,
	}

	// Write to file
	filePath := filepath.Join(backupDir, "flatpak.json")
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	if err := utils.WriteFile(filePath, jsonData); err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.Success = true
	result.Items = apps
	result.ItemCount = len(apps)
	result.FilePath = filePath

	return result, nil
}
