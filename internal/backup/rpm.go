package backup

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"time"

	"github.com/r8bert/rego/internal/utils"
)

// RPMBackup handles RPM package backup
type RPMBackup struct{}

// NewRPMBackup creates a new RPMBackup instance
func NewRPMBackup() *RPMBackup {
	return &RPMBackup{}
}

// Name returns the display name
func (r *RPMBackup) Name() string {
	return "RPM Packages"
}

// Type returns the backup type
func (r *RPMBackup) Type() BackupType {
	return BackupTypeRPM
}

// Available checks if DNF/RPM is available
func (r *RPMBackup) Available() bool {
	return utils.CommandExists("dnf") || utils.CommandExists("rpm")
}

// List returns user-installed RPM packages
func (r *RPMBackup) List() ([]BackupItem, error) {
	var packages []string
	var err error

	// Try dnf first (preferred for user-installed detection)
	if utils.CommandExists("dnf") {
		packages, err = r.listDNFUserInstalled()
		if err != nil {
			utils.Warn("DNF user-installed listing failed, falling back to rpm: %v", err)
			packages, err = r.listAllRPM()
		}
	} else {
		packages, err = r.listAllRPM()
	}

	if err != nil {
		return nil, err
	}

	var items []BackupItem
	for _, pkg := range packages {
		if pkg == "" {
			continue
		}
		items = append(items, BackupItem{
			Name: pkg,
			Type: BackupTypeRPM,
		})
	}

	return items, nil
}

// listDNFUserInstalled gets packages explicitly installed by user
func (r *RPMBackup) listDNFUserInstalled() ([]string, error) {
	// Get user-installed packages using dnf repoquery
	result := utils.RunCommand("dnf", "repoquery", "--userinstalled", "--qf", "%{name}")
	if result.Error != nil {
		// Fallback to history method
		return r.listDNFHistory()
	}

	var packages []string
	for _, line := range strings.Split(result.Stdout, "\n") {
		pkg := strings.TrimSpace(line)
		if pkg != "" && !r.isBasePackage(pkg) {
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// listDNFHistory uses dnf history to find user-installed packages
func (r *RPMBackup) listDNFHistory() ([]string, error) {
	result := utils.RunCommand("dnf", "history", "userinstalled")
	if result.Error != nil {
		return nil, result.Error
	}

	var packages []string
	for _, line := range strings.Split(result.Stdout, "\n") {
		pkg := strings.TrimSpace(line)
		if pkg != "" && !r.isBasePackage(pkg) {
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// listAllRPM gets all installed packages (less precise)
func (r *RPMBackup) listAllRPM() ([]string, error) {
	result := utils.RunCommand("rpm", "-qa", "--qf", "%{NAME}\n")
	if result.Error != nil {
		return nil, result.Error
	}

	var packages []string
	for _, line := range strings.Split(result.Stdout, "\n") {
		pkg := strings.TrimSpace(line)
		if pkg != "" {
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// isBasePackage filters out common base system packages
func (r *RPMBackup) isBasePackage(pkg string) bool {
	// Filter out obvious base packages that come with the system
	basePackages := map[string]bool{
		"fedora-release":      true,
		"fedora-repos":        true,
		"fedora-gpg-keys":     true,
		"basesystem":          true,
		"filesystem":          true,
		"setup":               true,
		"glibc":               true,
		"glibc-common":        true,
		"bash":                true,
		"coreutils":           true,
		"systemd":             true,
		"kernel":              true,
		"kernel-core":         true,
		"kernel-modules":      true,
		"kernel-modules-core": true,
	}

	return basePackages[pkg]
}

// RPMData represents the backup data structure
type RPMData struct {
	Packages      []BackupItem `json:"packages"`
	PackageCount  int          `json:"package_count"`
	PackageMethod string       `json:"package_method"` // "dnf_userinstalled" or "rpm_all"
}

// Backup performs the RPM backup
func (r *RPMBackup) Backup(backupDir string) (BackupResult, error) {
	result := BackupResult{
		Type:      BackupTypeRPM,
		Timestamp: time.Now(),
	}

	packages, err := r.List()
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	method := "dnf_userinstalled"
	if !utils.CommandExists("dnf") {
		method = "rpm_all"
	}

	data := RPMData{
		Packages:      packages,
		PackageCount:  len(packages),
		PackageMethod: method,
	}

	// Write to file
	filePath := filepath.Join(backupDir, "rpm_packages.json")
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
	result.Items = packages
	result.ItemCount = len(packages)
	result.FilePath = filePath

	return result, nil
}
