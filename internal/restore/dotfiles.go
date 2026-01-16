package restore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/r8bert/rego/internal/utils"
)

// DotfilesRestore handles dotfiles restoration
type DotfilesRestore struct {
	merge bool // If true, don't overwrite existing files
}

// NewDotfilesRestore creates a new DotfilesRestore instance
func NewDotfilesRestore() *DotfilesRestore {
	return &DotfilesRestore{merge: false}
}

// NewDotfilesRestoreMerge creates a restore that won't overwrite existing files
func NewDotfilesRestoreMerge() *DotfilesRestore {
	return &DotfilesRestore{merge: true}
}

// Name returns the display name
func (d *DotfilesRestore) Name() string {
	return "Dotfiles"
}

// Type returns the restore type
func (d *DotfilesRestore) Type() RestoreType {
	return RestoreTypeDotfiles
}

// Available returns true as dotfiles can always be restored
func (d *DotfilesRestore) Available() bool {
	return true
}

// DotfilesData matches the backup structure
type DotfilesData struct {
	Files      []DotfileInfo `json:"files"`
	BackupDir  string        `json:"backup_dir"`
	SourceHome string        `json:"source_home"`
}

// DotfileInfo contains dotfile information
type DotfileInfo struct {
	Path         string `json:"path"`
	RelativePath string `json:"relative_path"`
	Size         int64  `json:"size"`
	IsDir        bool   `json:"is_dir"`
	Exists       bool   `json:"exists"`
}

// Preview returns what would be restored
func (d *DotfilesRestore) Preview(backupDir string) ([]string, error) {
	data, err := d.loadBackupData(backupDir)
	if err != nil {
		return nil, err
	}

	var items []string
	for _, file := range data.Files {
		items = append(items, file.RelativePath)
	}

	return items, nil
}

// loadBackupData loads the dotfiles backup data
func (d *DotfilesRestore) loadBackupData(backupDir string) (*DotfilesData, error) {
	dataPath := filepath.Join(backupDir, "dotfiles.json")
	content, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read dotfiles backup: %w", err)
	}

	var data DotfilesData
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to parse dotfiles backup: %w", err)
	}

	return &data, nil
}

// Restore performs the dotfiles restoration
func (d *DotfilesRestore) Restore(backupDir string, dryRun bool) (RestoreResult, error) {
	result := RestoreResult{
		Type:      RestoreTypeDotfiles,
		Timestamp: time.Now(),
		DryRun:    dryRun,
	}

	data, err := d.loadBackupData(backupDir)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	result.ItemsTotal = len(data.Files)

	if dryRun {
		result.Success = true
		result.ItemsSuccess = result.ItemsTotal
		return result, nil
	}

	home, err := utils.GetHomeDir()
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	dotfilesDir := filepath.Join(backupDir, "dotfiles")

	for _, file := range data.Files {
		srcPath := filepath.Join(dotfilesDir, file.RelativePath)
		dstPath := filepath.Join(home, file.RelativePath)

		// Check if source exists in backup
		if !utils.FileExists(srcPath) && !utils.DirExists(srcPath) {
			result.ItemsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("Backup not found: %s", file.RelativePath))
			continue
		}

		// In merge mode, skip existing files
		if d.merge && utils.FileExists(dstPath) {
			result.ItemsSuccess++ // Count as success (preserved)
			continue
		}

		// Create backup of existing file
		if utils.FileExists(dstPath) {
			backupPath := dstPath + ".rego-backup"
			utils.CopyFile(dstPath, backupPath)
		}

		var copyErr error
		if file.IsDir {
			copyErr = utils.CopyDir(srcPath, dstPath)
		} else {
			copyErr = utils.CopyFile(srcPath, dstPath)
		}

		if copyErr != nil {
			result.ItemsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to restore %s: %v", file.RelativePath, copyErr))
		} else {
			result.ItemsSuccess++
		}
	}

	result.Success = result.ItemsFailed == 0
	return result, nil
}

// SetMerge sets whether to merge or overwrite
func (d *DotfilesRestore) SetMerge(merge bool) {
	d.merge = merge
}
