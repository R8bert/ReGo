package restore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/r8bert/rego/internal/utils"
)

// FlatpakRestore handles Flatpak restoration
type FlatpakRestore struct{}

// NewFlatpakRestore creates a new FlatpakRestore instance
func NewFlatpakRestore() *FlatpakRestore {
	return &FlatpakRestore{}
}

// Name returns the display name
func (f *FlatpakRestore) Name() string {
	return "Flatpak Applications"
}

// Type returns the restore type
func (f *FlatpakRestore) Type() RestoreType {
	return RestoreTypeFlatpak
}

// Available checks if Flatpak is installed
func (f *FlatpakRestore) Available() bool {
	return utils.CommandExists("flatpak")
}

// FlatpakData matches the backup structure
type FlatpakData struct {
	Applications []FlatpakApp    `json:"applications"`
	Remotes      []FlatpakRemote `json:"remotes"`
}

// FlatpakApp represents a Flatpak application
type FlatpakApp struct {
	Name     string            `json:"name"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// FlatpakRemote represents a Flatpak remote
type FlatpakRemote struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Options string `json:"options,omitempty"`
}

// Preview returns what would be restored
func (f *FlatpakRestore) Preview(backupDir string) ([]string, error) {
	data, err := f.loadBackupData(backupDir)
	if err != nil {
		return nil, err
	}

	var items []string
	for _, remote := range data.Remotes {
		items = append(items, fmt.Sprintf("Remote: %s (%s)", remote.Name, remote.URL))
	}
	for _, app := range data.Applications {
		items = append(items, fmt.Sprintf("App: %s", app.Name))
	}

	return items, nil
}

// loadBackupData loads the Flatpak backup data
func (f *FlatpakRestore) loadBackupData(backupDir string) (*FlatpakData, error) {
	dataPath := filepath.Join(backupDir, "flatpak.json")
	content, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read flatpak backup: %w", err)
	}

	var data FlatpakData
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to parse flatpak backup: %w", err)
	}

	return &data, nil
}

// Restore performs the Flatpak restoration
func (f *FlatpakRestore) Restore(backupDir string, dryRun bool) (RestoreResult, error) {
	result := RestoreResult{
		Type:      RestoreTypeFlatpak,
		Timestamp: time.Now(),
		DryRun:    dryRun,
	}

	data, err := f.loadBackupData(backupDir)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	result.ItemsTotal = len(data.Applications)

	if dryRun {
		result.Success = true
		result.ItemsSuccess = result.ItemsTotal
		return result, nil
	}

	// Install remotes first
	for _, remote := range data.Remotes {
		if remote.Name == "flathub" {
			// Flathub is special, use the official command
			cmdResult := utils.RunCommand("flatpak", "remote-add", "--if-not-exists", "flathub", "https://flathub.org/repo/flathub.flatpakrepo")
			if cmdResult.Error != nil {
				utils.Warn("Failed to add flathub remote: %v", cmdResult.Error)
			}
		} else if remote.URL != "" {
			cmdResult := utils.RunCommand("flatpak", "remote-add", "--if-not-exists", remote.Name, remote.URL)
			if cmdResult.Error != nil {
				utils.Warn("Failed to add remote %s: %v", remote.Name, cmdResult.Error)
			}
		}
	}

	// Install applications
	for _, app := range data.Applications {
		origin := "flathub" // Default
		if app.Metadata != nil && app.Metadata["origin"] != "" {
			origin = app.Metadata["origin"]
		}

		cmdResult := utils.RunCommandWithTimeout("flatpak", 5*time.Minute, "install", "-y", "--noninteractive", origin, app.Name)
		if cmdResult.Error != nil {
			result.ItemsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to install %s: %s", app.Name, cmdResult.Stderr))
		} else {
			result.ItemsSuccess++
		}
	}

	result.Success = result.ItemsFailed == 0
	return result, nil
}
