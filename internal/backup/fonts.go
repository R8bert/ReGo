package backup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/r8bert/rego/internal/utils"
)

// FontsBackup handles user fonts backup
type FontsBackup struct{}

// NewFontsBackup creates a new FontsBackup instance
func NewFontsBackup() *FontsBackup {
	return &FontsBackup{}
}

// Name returns the display name
func (f *FontsBackup) Name() string {
	return "User Fonts"
}

// Type returns the backup type
func (f *FontsBackup) Type() BackupType {
	return BackupTypeFonts
}

// Available checks if user fonts directory exists
func (f *FontsBackup) Available() bool {
	fontsDir, err := f.getUserFontsDir()
	if err != nil {
		return false
	}
	return utils.DirExists(fontsDir)
}

// getUserFontsDir returns the user fonts directory
func (f *FontsBackup) getUserFontsDir() (string, error) {
	home, err := utils.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "fonts"), nil
}

// FontInfo contains information about a font file
type FontInfo struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	RelativePath string `json:"relative_path"`
	Size         int64  `json:"size"`
}

// List returns font files in the user fonts directory
func (f *FontsBackup) List() ([]BackupItem, error) {
	fontsDir, err := f.getUserFontsDir()
	if err != nil {
		return nil, err
	}

	if !utils.DirExists(fontsDir) {
		return []BackupItem{}, nil
	}

	fonts, err := f.scanFontsDirectory(fontsDir)
	if err != nil {
		return nil, err
	}

	var items []BackupItem
	for _, font := range fonts {
		items = append(items, BackupItem{
			Name:        font.Name,
			Type:        BackupTypeFonts,
			Description: font.RelativePath,
			Metadata: map[string]string{
				"size": formatBytes(font.Size),
			},
		})
	}

	return items, nil
}

// formatBytes formats bytes as human-readable string
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
	)

	switch {
	case bytes >= MB:
		return formatInt(bytes/MB) + " MB"
	case bytes >= KB:
		return formatInt(bytes/KB) + " KB"
	default:
		return formatInt(bytes) + " B"
	}
}

func formatInt(n int64) string {
	return string(rune('0' + n%10))
}

// scanFontsDirectory scans the fonts directory for font files
func (f *FontsBackup) scanFontsDirectory(fontsDir string) ([]FontInfo, error) {
	var fonts []FontInfo

	err := filepath.Walk(fontsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			return nil
		}

		// Check for font file extensions
		ext := filepath.Ext(path)
		if !isFontExtension(ext) {
			return nil
		}

		relPath, _ := filepath.Rel(fontsDir, path)
		fonts = append(fonts, FontInfo{
			Name:         filepath.Base(path),
			Path:         path,
			RelativePath: relPath,
			Size:         info.Size(),
		})

		return nil
	})

	return fonts, err
}

// isFontExtension checks if the extension is a font file
func isFontExtension(ext string) bool {
	fontExtensions := map[string]bool{
		".ttf":   true,
		".otf":   true,
		".woff":  true,
		".woff2": true,
		".eot":   true,
		".ttc":   true,
	}
	return fontExtensions[ext]
}

// FontsData represents the backup data structure
type FontsData struct {
	Fonts     []FontInfo `json:"fonts"`
	TotalSize int64      `json:"total_size"`
	FontsDir  string     `json:"fonts_dir"`
}

// Backup performs the fonts backup
func (f *FontsBackup) Backup(backupDir string) (BackupResult, error) {
	result := BackupResult{
		Type:      BackupTypeFonts,
		Timestamp: time.Now(),
	}

	fontsDir, err := f.getUserFontsDir()
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	if !utils.DirExists(fontsDir) {
		// No fonts directory, this is not an error
		result.Success = true
		result.ItemCount = 0
		return result, nil
	}

	// Scan fonts
	fonts, err := f.scanFontsDirectory(fontsDir)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	if len(fonts) == 0 {
		result.Success = true
		result.ItemCount = 0
		return result, nil
	}

	// Copy fonts directory
	fontsBackupDir := filepath.Join(backupDir, "fonts")
	if err := utils.CopyDir(fontsDir, fontsBackupDir); err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Calculate total size
	var totalSize int64
	for _, font := range fonts {
		totalSize += font.Size
	}

	data := FontsData{
		Fonts:     fonts,
		TotalSize: totalSize,
		FontsDir:  fontsDir,
	}

	// Write metadata
	filePath := filepath.Join(backupDir, "fonts.json")
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	if err := utils.WriteFile(filePath, jsonData); err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Convert to BackupItems
	var items []BackupItem
	for _, font := range fonts {
		items = append(items, BackupItem{
			Name: font.Name,
			Type: BackupTypeFonts,
		})
	}

	result.Success = true
	result.Items = items
	result.ItemCount = len(items)
	result.FilePath = filePath

	return result, nil
}
