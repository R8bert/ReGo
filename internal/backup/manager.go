package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/r8bert/rego/internal/utils"
)

// Manager orchestrates the backup process
type Manager struct {
	backers   map[BackupType]Backer
	backupDir string
}

// NewManager creates a new backup manager
func NewManager() *Manager {
	m := &Manager{
		backers: make(map[BackupType]Backer),
	}

	// Register all available backers
	m.RegisterBacker(NewFlatpakBackup())
	m.RegisterBacker(NewRPMBackup())
	m.RegisterBacker(NewReposBackup())
	m.RegisterBacker(NewGnomeExtensionsBackup())
	m.RegisterBacker(NewGnomeSettingsBackup())
	m.RegisterBacker(NewDotfilesBackup())
	m.RegisterBacker(NewFontsBackup())

	return m
}

// RegisterBacker registers a backup component
func (m *Manager) RegisterBacker(b Backer) {
	m.backers[b.Type()] = b
}

// GetBacker returns a backer by type
func (m *Manager) GetBacker(t BackupType) (Backer, bool) {
	b, ok := m.backers[t]
	return b, ok
}

// GetAvailableBackers returns all available backers on this system
func (m *Manager) GetAvailableBackers() []Backer {
	var available []Backer
	for _, t := range AllBackupTypes() {
		if b, ok := m.backers[t]; ok && b.Available() {
			available = append(available, b)
		}
	}
	return available
}

// BackupProgress represents the progress of a backup operation
type BackupProgress struct {
	CurrentType BackupType
	CurrentName string
	TotalSteps  int
	CurrentStep int
	Completed   []BackupResult
	InProgress  bool
	Error       error
}

// ProgressCallback is called during backup to report progress
type ProgressCallback func(progress BackupProgress)

// RunBackup performs a backup with the given options
func (m *Manager) RunBackup(opts BackupOptions, callback ProgressCallback) (*BackupManifest, error) {
	// Determine backup directory
	backupDir := opts.BackupPath
	if backupDir == "" {
		defaultDir, err := utils.GetBackupDir()
		if err != nil {
			return nil, err
		}
		backupDir = filepath.Join(defaultDir, time.Now().Format("2006-01-02_150405"))
	}

	// Create backup directory
	if err := utils.EnsureDir(backupDir); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	m.backupDir = backupDir

	// Determine which types to backup
	var typesToBackup []BackupType
	if opts.IncludeFlatpak {
		typesToBackup = append(typesToBackup, BackupTypeFlatpak)
	}
	if opts.IncludeRPM {
		typesToBackup = append(typesToBackup, BackupTypeRPM)
	}
	if opts.IncludeRepos {
		typesToBackup = append(typesToBackup, BackupTypeRepos)
	}
	if opts.IncludeGnomeExtensions {
		typesToBackup = append(typesToBackup, BackupTypeGnomeExtensions)
	}
	if opts.IncludeGnomeSettings {
		typesToBackup = append(typesToBackup, BackupTypeGnomeSettings)
	}
	if opts.IncludeDotfiles {
		typesToBackup = append(typesToBackup, BackupTypeDotfiles)
	}
	if opts.IncludeFonts {
		typesToBackup = append(typesToBackup, BackupTypeFonts)
	}

	// Set custom dotfiles if provided
	if len(opts.DotfilesList) > 0 {
		if b, ok := m.backers[BackupTypeDotfiles].(*DotfilesBackup); ok {
			b.SetDotfiles(opts.DotfilesList)
		}
	}

	// Initialize manifest
	manifest := &BackupManifest{
		Version:     "1.0",
		CreatedAt:   time.Now(),
		Hostname:    getHostname(),
		User:        getUsername(),
		Components:  typesToBackup,
		Results:     make(map[BackupType]BackupResult),
		BackupPath:  backupDir,
		Description: opts.Description,
	}

	// Run backups
	progress := BackupProgress{
		TotalSteps: len(typesToBackup),
		InProgress: true,
	}

	for i, backupType := range typesToBackup {
		backer, ok := m.backers[backupType]
		if !ok || !backer.Available() {
			continue
		}

		progress.CurrentStep = i + 1
		progress.CurrentType = backupType
		progress.CurrentName = backer.Name()

		if callback != nil {
			callback(progress)
		}

		result, err := backer.Backup(backupDir)
		if err != nil {
			utils.Error("Backup failed for %s: %v", backupType, err)
		}

		manifest.Results[backupType] = result
		progress.Completed = append(progress.Completed, result)
	}

	// Save manifest
	manifestPath := filepath.Join(backupDir, "manifest.json")
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return manifest, fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := utils.WriteFile(manifestPath, manifestData); err != nil {
		return manifest, fmt.Errorf("failed to write manifest: %w", err)
	}

	progress.InProgress = false
	if callback != nil {
		callback(progress)
	}

	return manifest, nil
}

// ListBackups returns all available backups
func (m *Manager) ListBackups() ([]BackupManifest, error) {
	backupDir, err := utils.GetBackupDir()
	if err != nil {
		return nil, err
	}

	if !utils.DirExists(backupDir) {
		return []BackupManifest{}, nil
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, err
	}

	var backups []BackupManifest
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		manifestPath := filepath.Join(backupDir, entry.Name(), "manifest.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue // Skip directories without manifest
		}

		var manifest BackupManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			continue
		}

		backups = append(backups, manifest)
	}

	return backups, nil
}

// LoadBackup loads a backup manifest from a directory
func (m *Manager) LoadBackup(backupPath string) (*BackupManifest, error) {
	manifestPath := filepath.Join(backupPath, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest BackupManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// getHostname returns the system hostname
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// getUsername returns the current username
func getUsername() string {
	return os.Getenv("USER")
}
