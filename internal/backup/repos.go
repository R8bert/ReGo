package backup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/r8bert/rego/internal/utils"
)

// ReposBackup handles repository backup
type ReposBackup struct{}

// NewReposBackup creates a new ReposBackup instance
func NewReposBackup() *ReposBackup {
	return &ReposBackup{}
}

// Name returns the display name
func (r *ReposBackup) Name() string {
	return "DNF Repositories"
}

// Type returns the backup type
func (r *ReposBackup) Type() BackupType {
	return BackupTypeRepos
}

// Available checks if DNF repos directory exists
func (r *ReposBackup) Available() bool {
	return utils.DirExists("/etc/yum.repos.d")
}

// RepoInfo contains information about a repository
type RepoInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	BaseURL  string `json:"baseurl,omitempty"`
	Metalink string `json:"metalink,omitempty"`
	Enabled  bool   `json:"enabled"`
	GPGCheck bool   `json:"gpgcheck"`
	GPGKey   string `json:"gpgkey,omitempty"`
	FileName string `json:"filename"`
}

// List returns enabled third-party repositories
func (r *ReposBackup) List() ([]BackupItem, error) {
	repos, err := r.listRepos()
	if err != nil {
		return nil, err
	}

	var items []BackupItem
	for _, repo := range repos {
		// Skip Fedora base repos
		if r.isFedoraBaseRepo(repo.ID) {
			continue
		}

		items = append(items, BackupItem{
			Name:        repo.ID,
			Type:        BackupTypeRepos,
			Description: repo.Name,
			Metadata: map[string]string{
				"filename": repo.FileName,
				"enabled":  boolToString(repo.Enabled),
			},
		})
	}

	return items, nil
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// isFedoraBaseRepo checks if a repo is a standard Fedora repo
func (r *ReposBackup) isFedoraBaseRepo(repoID string) bool {
	baseRepos := []string{
		"fedora",
		"fedora-modular",
		"fedora-cisco-openh264",
		"updates",
		"updates-modular",
		"updates-testing",
		"updates-testing-modular",
		"fedora-debuginfo",
		"fedora-source",
		"updates-debuginfo",
		"updates-source",
	}

	for _, base := range baseRepos {
		if repoID == base {
			return true
		}
	}

	return false
}

// listRepos parses all .repo files
func (r *ReposBackup) listRepos() ([]RepoInfo, error) {
	reposDir := "/etc/yum.repos.d"

	entries, err := os.ReadDir(reposDir)
	if err != nil {
		return nil, err
	}

	var repos []RepoInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".repo") {
			continue
		}

		filePath := filepath.Join(reposDir, entry.Name())
		fileRepos, err := r.parseRepoFile(filePath)
		if err != nil {
			utils.Warn("Failed to parse %s: %v", filePath, err)
			continue
		}

		repos = append(repos, fileRepos...)
	}

	return repos, nil
}

// parseRepoFile parses a single .repo file
func (r *ReposBackup) parseRepoFile(filePath string) ([]RepoInfo, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	fileName := filepath.Base(filePath)
	var repos []RepoInfo
	var currentRepo *RepoInfo

	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// New section
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if currentRepo != nil {
				repos = append(repos, *currentRepo)
			}
			currentRepo = &RepoInfo{
				ID:       strings.Trim(line, "[]"),
				FileName: fileName,
				Enabled:  true, // Default
			}
			continue
		}

		if currentRepo == nil {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "name":
			currentRepo.Name = value
		case "baseurl":
			currentRepo.BaseURL = value
		case "metalink":
			currentRepo.Metalink = value
		case "enabled":
			currentRepo.Enabled = value == "1" || value == "true"
		case "gpgcheck":
			currentRepo.GPGCheck = value == "1" || value == "true"
		case "gpgkey":
			currentRepo.GPGKey = value
		}
	}

	if currentRepo != nil {
		repos = append(repos, *currentRepo)
	}

	return repos, nil
}

// ReposData represents the backup data structure
type ReposData struct {
	Repos     []RepoInfo `json:"repos"`
	RepoFiles []string   `json:"repo_files"`
}

// Backup performs the repos backup
func (r *ReposBackup) Backup(backupDir string) (BackupResult, error) {
	result := BackupResult{
		Type:      BackupTypeRepos,
		Timestamp: time.Now(),
	}

	items, err := r.List()
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Get full repo info
	repos, _ := r.listRepos()

	// Filter to non-base repos
	var thirdPartyRepos []RepoInfo
	for _, repo := range repos {
		if !r.isFedoraBaseRepo(repo.ID) {
			thirdPartyRepos = append(thirdPartyRepos, repo)
		}
	}

	// Copy actual repo files
	reposBackupDir := filepath.Join(backupDir, "repos.d")
	if err := utils.EnsureDir(reposBackupDir); err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Get unique filenames
	fileSet := make(map[string]bool)
	for _, repo := range thirdPartyRepos {
		fileSet[repo.FileName] = true
	}

	var copiedFiles []string
	for fileName := range fileSet {
		srcPath := filepath.Join("/etc/yum.repos.d", fileName)
		dstPath := filepath.Join(reposBackupDir, fileName)

		if err := utils.CopyFile(srcPath, dstPath); err != nil {
			utils.Warn("Failed to copy %s: %v", fileName, err)
			continue
		}
		copiedFiles = append(copiedFiles, fileName)
	}

	// Also backup GPG keys if accessible
	r.backupGPGKeys(backupDir)

	data := ReposData{
		Repos:     thirdPartyRepos,
		RepoFiles: copiedFiles,
	}

	// Write metadata
	filePath := filepath.Join(backupDir, "repos.json")
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

// backupGPGKeys attempts to backup imported GPG keys
func (r *ReposBackup) backupGPGKeys(backupDir string) {
	keysDir := filepath.Join(backupDir, "gpg-keys")
	if err := utils.EnsureDir(keysDir); err != nil {
		return
	}

	// Export RPM GPG keys
	result := utils.RunCommand("rpm", "-qa", "gpg-pubkey*")
	if result.Error != nil {
		return
	}

	for _, key := range strings.Split(result.Stdout, "\n") {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		exportResult := utils.RunCommand("rpm", "-qi", key)
		if exportResult.Error == nil {
			keyFile := filepath.Join(keysDir, key+".txt")
			utils.WriteFile(keyFile, []byte(exportResult.Stdout))
		}
	}
}
