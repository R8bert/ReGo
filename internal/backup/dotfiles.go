package backup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/r8bert/rego/internal/utils"
)

// DotfilesBackup handles dotfiles backup
type DotfilesBackup struct {
	dotfiles []string
}

// NewDotfilesBackup creates a new DotfilesBackup instance
func NewDotfilesBackup() *DotfilesBackup {
	return &DotfilesBackup{
		dotfiles: DefaultDotfiles(),
	}
}

// NewDotfilesBackupWithList creates a DotfilesBackup with a custom list
func NewDotfilesBackupWithList(files []string) *DotfilesBackup {
	return &DotfilesBackup{
		dotfiles: files,
	}
}

// Name returns the display name
func (d *DotfilesBackup) Name() string {
	return "Dotfiles"
}

// Type returns the backup type
func (d *DotfilesBackup) Type() BackupType {
	return BackupTypeDotfiles
}

// Available returns true as dotfiles are always available to backup
func (d *DotfilesBackup) Available() bool {
	return true
}

// DotfileInfo contains information about a dotfile
type DotfileInfo struct {
	Path         string `json:"path"`
	RelativePath string `json:"relative_path"`
	Size         int64  `json:"size"`
	IsDir        bool   `json:"is_dir"`
	Exists       bool   `json:"exists"`
}

// List returns dotfiles that exist and can be backed up
func (d *DotfilesBackup) List() ([]BackupItem, error) {
	home, err := utils.GetHomeDir()
	if err != nil {
		return nil, err
	}

	var items []BackupItem
	for _, dotfile := range d.dotfiles {
		fullPath := filepath.Join(home, dotfile)

		info, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			continue // Skip non-existent files
		}

		item := BackupItem{
			Name:        dotfile,
			Type:        BackupTypeDotfiles,
			Description: fullPath,
			Metadata:    make(map[string]string),
		}

		if info != nil {
			if info.IsDir() {
				item.Metadata["type"] = "directory"
			} else {
				item.Metadata["type"] = "file"
				item.Metadata["size"] = formatSize(info.Size())
			}
		}

		items = append(items, item)
	}

	return items, nil
}

// formatSize returns a human-readable size string
func formatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
	)

	switch {
	case size >= MB:
		return string(rune(size/MB)) + " MB"
	case size >= KB:
		return string(rune(size/KB)) + " KB"
	default:
		return string(rune(size)) + " B"
	}
}

// DotfilesData represents the backup data structure
type DotfilesData struct {
	Files      []DotfileInfo `json:"files"`
	BackupDir  string        `json:"backup_dir"`
	SourceHome string        `json:"source_home"`
}

// Backup performs the dotfiles backup
func (d *DotfilesBackup) Backup(backupDir string) (BackupResult, error) {
	result := BackupResult{
		Type:      BackupTypeDotfiles,
		Timestamp: time.Now(),
	}

	home, err := utils.GetHomeDir()
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	dotfilesDir := filepath.Join(backupDir, "dotfiles")
	if err := utils.EnsureDir(dotfilesDir); err != nil {
		result.Error = err.Error()
		return result, err
	}

	var files []DotfileInfo
	var items []BackupItem

	for _, dotfile := range d.dotfiles {
		srcPath := filepath.Join(home, dotfile)
		dstPath := filepath.Join(dotfilesDir, dotfile)

		info, err := os.Stat(srcPath)
		if os.IsNotExist(err) {
			continue // Skip non-existent files
		}

		fileInfo := DotfileInfo{
			Path:         srcPath,
			RelativePath: dotfile,
			Exists:       true,
		}

		if info.IsDir() {
			fileInfo.IsDir = true
			err = utils.CopyDir(srcPath, dstPath)
		} else {
			fileInfo.Size = info.Size()
			err = utils.CopyFile(srcPath, dstPath)
		}

		if err != nil {
			utils.Warn("Failed to backup %s: %v", dotfile, err)
			continue
		}

		files = append(files, fileInfo)
		items = append(items, BackupItem{
			Name: dotfile,
			Type: BackupTypeDotfiles,
		})
	}

	data := DotfilesData{
		Files:      files,
		BackupDir:  dotfilesDir,
		SourceHome: home,
	}

	// Write metadata
	filePath := filepath.Join(backupDir, "dotfiles.json")
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
	result.Items = items
	result.ItemCount = len(items)
	result.FilePath = filePath

	return result, nil
}

// SetDotfiles updates the list of dotfiles to backup
func (d *DotfilesBackup) SetDotfiles(files []string) {
	d.dotfiles = files
}

// GetDotfiles returns the current list of dotfiles
func (d *DotfilesBackup) GetDotfiles() []string {
	return d.dotfiles
}
