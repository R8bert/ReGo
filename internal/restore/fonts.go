package restore

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/r8bert/rego/internal/utils"
)

type FontsRestore struct{}

func NewFontsRestore() *FontsRestore { return &FontsRestore{} }

func (f *FontsRestore) Name() string      { return "User Fonts" }
func (f *FontsRestore) Type() RestoreType { return RestoreTypeFonts }
func (f *FontsRestore) Available() bool   { return true }

type FontsData struct {
	Fonts     []FontInfo `json:"fonts"`
	TotalSize int64      `json:"total_size"`
	FontsDir  string     `json:"fonts_dir"`
}

type FontInfo struct {
	Name         string `json:"name"`
	RelativePath string `json:"relative_path"`
	Size         int64  `json:"size"`
}

func (f *FontsRestore) Preview(backupDir string) ([]string, error) {
	fontsDir := filepath.Join(backupDir, "fonts")
	if utils.DirExists(fontsDir) {
		files, _ := utils.ListFilesRecursive(fontsDir)
		var items []string
		for _, file := range files {
			items = append(items, filepath.Base(file))
		}
		return items, nil
	}
	return nil, fmt.Errorf("fonts backup not found")
}

func (f *FontsRestore) Restore(backupDir string, dryRun bool) (RestoreResult, error) {
	result := RestoreResult{Type: RestoreTypeFonts, Timestamp: time.Now(), DryRun: dryRun}

	fontsBackupDir := filepath.Join(backupDir, "fonts")
	if !utils.DirExists(fontsBackupDir) {
		result.Success = true
		return result, nil
	}

	files, _ := utils.ListFilesRecursive(fontsBackupDir)
	result.ItemsTotal = len(files)

	if dryRun {
		result.Success, result.ItemsSuccess = true, result.ItemsTotal
		return result, nil
	}

	home, _ := utils.GetHomeDir()
	userFontsDir := filepath.Join(home, ".local", "share", "fonts")

	if err := utils.CopyDir(fontsBackupDir, userFontsDir); err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	result.ItemsSuccess = result.ItemsTotal
	if utils.CommandExists("fc-cache") {
		utils.RunCommand("fc-cache", "-f", "-v")
	}
	result.Success = true
	return result, nil
}
