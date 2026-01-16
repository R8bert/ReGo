package restore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/r8bert/rego/internal/utils"
)

// RPMRestore handles RPM package restoration
type RPMRestore struct{}

// NewRPMRestore creates a new RPMRestore instance
func NewRPMRestore() *RPMRestore {
	return &RPMRestore{}
}

// Name returns the display name
func (r *RPMRestore) Name() string {
	return "RPM Packages"
}

// Type returns the restore type
func (r *RPMRestore) Type() RestoreType {
	return RestoreTypeRPM
}

// Available checks if DNF is available
func (r *RPMRestore) Available() bool {
	return utils.CommandExists("dnf")
}

// RPMData matches the backup structure
type RPMData struct {
	Packages      []RPMPackage `json:"packages"`
	PackageCount  int          `json:"package_count"`
	PackageMethod string       `json:"package_method"`
}

// RPMPackage represents an RPM package
type RPMPackage struct {
	Name string `json:"name"`
}

// Preview returns what would be restored
func (r *RPMRestore) Preview(backupDir string) ([]string, error) {
	data, err := r.loadBackupData(backupDir)
	if err != nil {
		return nil, err
	}

	var items []string
	for _, pkg := range data.Packages {
		items = append(items, pkg.Name)
	}

	return items, nil
}

// loadBackupData loads the RPM backup data
func (r *RPMRestore) loadBackupData(backupDir string) (*RPMData, error) {
	dataPath := filepath.Join(backupDir, "rpm_packages.json")
	content, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rpm backup: %w", err)
	}

	var data RPMData
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to parse rpm backup: %w", err)
	}

	return &data, nil
}

// Restore performs the RPM package restoration
func (r *RPMRestore) Restore(backupDir string, dryRun bool) (RestoreResult, error) {
	result := RestoreResult{
		Type:      RestoreTypeRPM,
		Timestamp: time.Now(),
		DryRun:    dryRun,
	}

	data, err := r.loadBackupData(backupDir)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	result.ItemsTotal = len(data.Packages)

	if dryRun {
		result.Success = true
		result.ItemsSuccess = result.ItemsTotal
		return result, nil
	}

	// Build package list
	var packageNames []string
	for _, pkg := range data.Packages {
		packageNames = append(packageNames, pkg.Name)
	}

	if len(packageNames) == 0 {
		result.Success = true
		return result, nil
	}

	// Install all packages in one go
	args := append([]string{"install", "-y"}, packageNames...)
	cmdResult := utils.RunCommandWithTimeout("dnf", 30*time.Minute, args...)

	if cmdResult.Error != nil {
		// Try to parse what succeeded and what failed
		result.Errors = append(result.Errors, cmdResult.Stderr)

		// Count failures from error message
		for _, pkg := range packageNames {
			if strings.Contains(cmdResult.Stderr, pkg) {
				result.ItemsFailed++
			} else {
				result.ItemsSuccess++
			}
		}
	} else {
		result.ItemsSuccess = len(packageNames)
	}

	result.Success = result.ItemsFailed == 0
	return result, nil
}
