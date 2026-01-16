package restore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/r8bert/rego/internal/utils"
)

// ReposRestore handles repository restoration
type ReposRestore struct{}

// NewReposRestore creates a new ReposRestore instance
func NewReposRestore() *ReposRestore {
	return &ReposRestore{}
}

// Name returns the display name
func (r *ReposRestore) Name() string {
	return "DNF Repositories"
}

// Type returns the restore type
func (r *ReposRestore) Type() RestoreType {
	return RestoreTypeRepos
}

// Available checks if DNF repos directory is accessible
func (r *ReposRestore) Available() bool {
	return utils.DirExists("/etc/yum.repos.d")
}

// ReposData matches the backup structure
type ReposData struct {
	Repos     []RepoInfo `json:"repos"`
	RepoFiles []string   `json:"repo_files"`
}

// RepoInfo contains repository information
type RepoInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	FileName string `json:"filename"`
}

// Preview returns what would be restored
func (r *ReposRestore) Preview(backupDir string) ([]string, error) {
	data, err := r.loadBackupData(backupDir)
	if err != nil {
		return nil, err
	}

	var items []string
	for _, repo := range data.Repos {
		items = append(items, fmt.Sprintf("%s (%s)", repo.ID, repo.FileName))
	}

	return items, nil
}

// loadBackupData loads the repos backup data
func (r *ReposRestore) loadBackupData(backupDir string) (*ReposData, error) {
	dataPath := filepath.Join(backupDir, "repos.json")
	content, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read repos backup: %w", err)
	}

	var data ReposData
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to parse repos backup: %w", err)
	}

	return &data, nil
}

// Restore performs the repository restoration
func (r *ReposRestore) Restore(backupDir string, dryRun bool) (RestoreResult, error) {
	result := RestoreResult{
		Type:      RestoreTypeRepos,
		Timestamp: time.Now(),
		DryRun:    dryRun,
	}

	data, err := r.loadBackupData(backupDir)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	result.ItemsTotal = len(data.RepoFiles)

	if dryRun {
		result.Success = true
		result.ItemsSuccess = result.ItemsTotal
		return result, nil
	}

	// Copy repo files back
	reposBackupDir := filepath.Join(backupDir, "repos.d")
	for _, fileName := range data.RepoFiles {
		srcPath := filepath.Join(reposBackupDir, fileName)
		dstPath := filepath.Join("/etc/yum.repos.d", fileName)

		// Check if file exists in backup
		if !utils.FileExists(srcPath) {
			result.ItemsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("Backup file not found: %s", fileName))
			continue
		}

		// Need sudo for /etc/yum.repos.d
		cmdResult := utils.RunCommand("sudo", "cp", srcPath, dstPath)
		if cmdResult.Error != nil {
			result.ItemsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to copy %s: %s", fileName, cmdResult.Stderr))
		} else {
			result.ItemsSuccess++
		}
	}

	result.Success = result.ItemsFailed == 0
	return result, nil
}
