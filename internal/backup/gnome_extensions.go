package backup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/r8bert/rego/internal/utils"
)

// GnomeExtensionsBackup handles GNOME extensions backup
type GnomeExtensionsBackup struct{}

// NewGnomeExtensionsBackup creates a new GnomeExtensionsBackup instance
func NewGnomeExtensionsBackup() *GnomeExtensionsBackup {
	return &GnomeExtensionsBackup{}
}

// Name returns the display name
func (g *GnomeExtensionsBackup) Name() string {
	return "GNOME Extensions"
}

// Type returns the backup type
func (g *GnomeExtensionsBackup) Type() BackupType {
	return BackupTypeGnomeExtensions
}

// Available checks if GNOME extensions are available
func (g *GnomeExtensionsBackup) Available() bool {
	// Check if gnome-extensions command exists or extensions directory exists
	if utils.CommandExists("gnome-extensions") {
		return true
	}

	home, err := utils.GetHomeDir()
	if err != nil {
		return false
	}

	extDir := filepath.Join(home, ".local", "share", "gnome-shell", "extensions")
	return utils.DirExists(extDir)
}

// ExtensionInfo contains information about a GNOME extension
type ExtensionInfo struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
	URL         string `json:"url,omitempty"`
	Enabled     bool   `json:"enabled"`
	HasSettings bool   `json:"has_settings"`
}

// List returns installed GNOME extensions
func (g *GnomeExtensionsBackup) List() ([]BackupItem, error) {
	extensions, err := g.listExtensions()
	if err != nil {
		return nil, err
	}

	var items []BackupItem
	for _, ext := range extensions {
		items = append(items, BackupItem{
			Name:        ext.UUID,
			Type:        BackupTypeGnomeExtensions,
			Description: ext.Name,
			Metadata: map[string]string{
				"enabled": boolToString(ext.Enabled),
				"version": ext.Version,
			},
		})
	}

	return items, nil
}

// listExtensions gets details about all installed extensions
func (g *GnomeExtensionsBackup) listExtensions() ([]ExtensionInfo, error) {
	var extensions []ExtensionInfo

	// Try using gnome-extensions command first
	if utils.CommandExists("gnome-extensions") {
		result := utils.RunCommand("gnome-extensions", "list", "--details")
		if result.Error == nil {
			extensions = g.parseExtensionsList(result.Stdout)
		}
	}

	// If command failed or no extensions found, scan directory
	if len(extensions) == 0 {
		dirExtensions, err := g.scanExtensionsDirectory()
		if err != nil {
			return nil, err
		}
		extensions = dirExtensions
	}

	// Get enabled status
	enabledExtensions := g.getEnabledExtensions()
	for i := range extensions {
		extensions[i].Enabled = enabledExtensions[extensions[i].UUID]
	}

	return extensions, nil
}

// parseExtensionsList parses the output of gnome-extensions list
func (g *GnomeExtensionsBackup) parseExtensionsList(output string) []ExtensionInfo {
	var extensions []ExtensionInfo

	// Simple parsing - each extension is on its own line
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// The UUID is the first word
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		uuid := parts[0]
		if strings.Contains(uuid, "@") {
			extensions = append(extensions, ExtensionInfo{
				UUID: uuid,
				Name: uuid,
			})
		}
	}

	return extensions
}

// scanExtensionsDirectory scans the extensions directory for installed extensions
func (g *GnomeExtensionsBackup) scanExtensionsDirectory() ([]ExtensionInfo, error) {
	home, err := utils.GetHomeDir()
	if err != nil {
		return nil, err
	}

	extDir := filepath.Join(home, ".local", "share", "gnome-shell", "extensions")
	if !utils.DirExists(extDir) {
		return []ExtensionInfo{}, nil
	}

	entries, err := os.ReadDir(extDir)
	if err != nil {
		return nil, err
	}

	var extensions []ExtensionInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		uuid := entry.Name()
		ext := ExtensionInfo{
			UUID: uuid,
			Name: uuid,
		}

		// Try to read metadata.json for more details
		metadataPath := filepath.Join(extDir, uuid, "metadata.json")
		if data, err := os.ReadFile(metadataPath); err == nil {
			var metadata struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				Version     int    `json:"version"`
				URL         string `json:"url"`
			}
			if json.Unmarshal(data, &metadata) == nil {
				ext.Name = metadata.Name
				ext.Description = metadata.Description
				ext.URL = metadata.URL
				if metadata.Version > 0 {
					ext.Version = string(rune(metadata.Version))
				}
			}
		}

		// Check if extension has settings schema
		schemasDir := filepath.Join(extDir, uuid, "schemas")
		ext.HasSettings = utils.DirExists(schemasDir)

		extensions = append(extensions, ext)
	}

	return extensions, nil
}

// getEnabledExtensions returns a map of enabled extension UUIDs
func (g *GnomeExtensionsBackup) getEnabledExtensions() map[string]bool {
	enabled := make(map[string]bool)

	// Use gsettings to get enabled extensions
	if !utils.CommandExists("gsettings") {
		return enabled
	}

	result := utils.RunCommand("gsettings", "get", "org.gnome.shell", "enabled-extensions")
	if result.Error != nil {
		return enabled
	}

	// Parse the array format: ['ext1@example.com', 'ext2@example.com']
	output := strings.TrimSpace(result.Stdout)
	output = strings.Trim(output, "[]")

	for _, ext := range strings.Split(output, ",") {
		ext = strings.TrimSpace(ext)
		ext = strings.Trim(ext, "'\"")
		if ext != "" {
			enabled[ext] = true
		}
	}

	return enabled
}

// ExtensionsData represents the backup data structure
type ExtensionsData struct {
	Extensions        []ExtensionInfo `json:"extensions"`
	EnabledExtensions []string        `json:"enabled_extensions"`
}

// Backup performs the GNOME extensions backup
func (g *GnomeExtensionsBackup) Backup(backupDir string) (BackupResult, error) {
	result := BackupResult{
		Type:      BackupTypeGnomeExtensions,
		Timestamp: time.Now(),
	}

	items, err := g.List()
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	extensions, _ := g.listExtensions()

	// Get enabled extensions list
	var enabledList []string
	for _, ext := range extensions {
		if ext.Enabled {
			enabledList = append(enabledList, ext.UUID)
		}
	}

	data := ExtensionsData{
		Extensions:        extensions,
		EnabledExtensions: enabledList,
	}

	// Write metadata
	filePath := filepath.Join(backupDir, "gnome_extensions.json")
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	if err := utils.WriteFile(filePath, jsonData); err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Backup extension settings (dconf keys for extensions)
	g.backupExtensionSettings(backupDir, extensions)

	result.Success = true
	result.Items = items
	result.ItemCount = len(items)
	result.FilePath = filePath

	return result, nil
}

// backupExtensionSettings exports dconf settings for extensions
func (g *GnomeExtensionsBackup) backupExtensionSettings(backupDir string, extensions []ExtensionInfo) {
	if !utils.CommandExists("dconf") {
		return
	}

	settingsDir := filepath.Join(backupDir, "extension_settings")
	utils.EnsureDir(settingsDir)

	for _, ext := range extensions {
		if !ext.HasSettings {
			continue
		}

		// Each extension stores settings under /org/gnome/shell/extensions/<uuid>/
		path := "/org/gnome/shell/extensions/" + strings.ReplaceAll(ext.UUID, "@", "-") + "/"
		result := utils.RunCommand("dconf", "dump", path)
		if result.Error == nil && result.Stdout != "" {
			settingsFile := filepath.Join(settingsDir, ext.UUID+".dconf")
			utils.WriteFile(settingsFile, []byte(result.Stdout))
		}
	}
}
