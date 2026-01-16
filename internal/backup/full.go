package backup

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/r8bert/rego/internal/utils"
)

// FullBackupOptions controls what to include in the full backup
type FullBackupOptions struct {
	Flatpaks    bool
	RPM         bool
	Repos       bool
	Extensions  bool
	Settings    bool
	Dotfiles    bool
	Fonts       bool
	SSHConfig   bool
	Autostart   bool
	Backgrounds bool
	Themes      bool
}

// DefaultFullBackupOptions returns all options enabled
func DefaultFullBackupOptions() FullBackupOptions {
	return FullBackupOptions{
		Flatpaks: true, RPM: true, Repos: true, Extensions: true,
		Settings: true, Dotfiles: true, Fonts: true, SSHConfig: true,
		Autostart: true, Backgrounds: true, Themes: true,
	}
}

// FullBackupManifest contains metadata about the full backup
type FullBackupManifest struct {
	Version   string         `json:"version"`
	CreatedAt time.Time      `json:"created_at"`
	Hostname  string         `json:"hostname"`
	Stats     map[string]int `json:"stats"`
	Included  []string       `json:"included"`
}

// GetDefaultFullBackupPath returns the default path for full backup
func GetDefaultFullBackupPath() string {
	home, _ := utils.GetHomeDir()
	hostname, _ := os.Hostname()
	date := time.Now().Format("2006-01-02")
	return filepath.Join(home, fmt.Sprintf("rego-full-%s-%s.tar.gz", hostname, date))
}

// CreateFullBackup creates a comprehensive backup archive
func CreateFullBackup(opts FullBackupOptions, outputPath string) (map[string]int, error) {
	stats := make(map[string]int)
	home, _ := utils.GetHomeDir()

	// Create temp directory for backup
	tmpDir, err := os.MkdirTemp("", "rego-full-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	var included []string

	// Package lists (always as JSON)
	if opts.Flatpaks || opts.RPM || opts.Extensions || opts.Settings || opts.Repos {
		lightOpts := LightBackupOptions{
			Flatpaks: opts.Flatpaks, RPM: opts.RPM, Extensions: opts.Extensions,
			Settings: opts.Settings, Repos: opts.Repos,
		}
		lightBackup, _ := CreateLightBackupWithOptions(lightOpts)
		if lightBackup != nil {
			lightBackup.SaveToFile(filepath.Join(tmpDir, "packages.json"))
			stats["flatpaks"] = len(lightBackup.Flatpaks)
			stats["rpm"] = len(lightBackup.RPMPackages)
			stats["extensions"] = len(lightBackup.GnomeExtensions)
			if lightBackup.DconfSettings != "" {
				stats["settings"] = 1
			}
			stats["repos"] = len(lightBackup.Repos)
		}
	}

	// Dotfiles
	if opts.Dotfiles {
		dotfiles := []string{
			".bashrc", ".bash_profile", ".bash_aliases",
			".zshrc", ".zprofile",
			".profile",
			".gitconfig", ".gitignore_global",
			".vimrc", ".nanorc",
			".tmux.conf",
			".config/fish/config.fish",
			".config/starship.toml",
		}
		count := copyFilesToDir(home, dotfiles, filepath.Join(tmpDir, "dotfiles"))
		stats["dotfiles"] = count
		if count > 0 {
			included = append(included, "dotfiles")
		}
	}

	// SSH Config (NOT keys for security)
	if opts.SSHConfig {
		sshFiles := []string{".ssh/config", ".ssh/known_hosts"}
		count := copyFilesToDir(home, sshFiles, filepath.Join(tmpDir, "ssh"))
		stats["ssh"] = count
		if count > 0 {
			included = append(included, "ssh")
		}
	}

	// Fonts
	if opts.Fonts {
		fontsDir := filepath.Join(home, ".local", "share", "fonts")
		if utils.DirExists(fontsDir) {
			destDir := filepath.Join(tmpDir, "fonts")
			utils.CopyDir(fontsDir, destDir)
			files, _ := utils.ListFilesRecursive(destDir)
			stats["fonts"] = len(files)
			if len(files) > 0 {
				included = append(included, "fonts")
			}
		}
	}

	// Autostart
	if opts.Autostart {
		autostartDir := filepath.Join(home, ".config", "autostart")
		if utils.DirExists(autostartDir) {
			destDir := filepath.Join(tmpDir, "autostart")
			utils.CopyDir(autostartDir, destDir)
			files, _ := utils.ListFilesRecursive(destDir)
			stats["autostart"] = len(files)
			if len(files) > 0 {
				included = append(included, "autostart")
			}
		}
	}

	// Backgrounds/Wallpapers
	if opts.Backgrounds {
		bgDirs := []string{
			filepath.Join(home, ".local", "share", "backgrounds"),
			filepath.Join(home, "Pictures", "Wallpapers"),
		}
		destDir := filepath.Join(tmpDir, "backgrounds")
		count := 0
		for _, bgDir := range bgDirs {
			if utils.DirExists(bgDir) {
				utils.CopyDir(bgDir, destDir)
				files, _ := utils.ListFilesRecursive(destDir)
				count = len(files)
			}
		}
		stats["backgrounds"] = count
		if count > 0 {
			included = append(included, "backgrounds")
		}
	}

	// Themes and Icons
	if opts.Themes {
		themeDirs := []string{
			filepath.Join(home, ".themes"),
			filepath.Join(home, ".icons"),
			filepath.Join(home, ".local", "share", "themes"),
			filepath.Join(home, ".local", "share", "icons"),
		}
		count := 0
		for i, themeDir := range themeDirs {
			if utils.DirExists(themeDir) {
				destDir := filepath.Join(tmpDir, fmt.Sprintf("themes_%d", i))
				utils.CopyDir(themeDir, destDir)
				files, _ := utils.ListFilesRecursive(destDir)
				count += len(files)
			}
		}
		stats["themes"] = count
		if count > 0 {
			included = append(included, "themes")
		}
	}

	// Write manifest
	manifest := FullBackupManifest{
		Version:   "1.0",
		CreatedAt: time.Now(),
		Hostname:  getHostnameSimple(),
		Stats:     stats,
		Included:  included,
	}
	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	os.WriteFile(filepath.Join(tmpDir, "manifest.json"), manifestData, 0644)

	// Create tar.gz
	if err := createTarGz(tmpDir, outputPath); err != nil {
		return stats, err
	}

	return stats, nil
}

func copyFilesToDir(baseDir string, files []string, destDir string) int {
	utils.EnsureDir(destDir)
	count := 0
	for _, file := range files {
		src := filepath.Join(baseDir, file)
		if utils.FileExists(src) {
			dst := filepath.Join(destDir, filepath.Base(file))
			if utils.CopyFile(src, dst) == nil {
				count++
			}
		}
	}
	return count
}

func getHostnameSimple() string {
	hostname, _ := os.Hostname()
	return hostname
}

func createTarGz(sourceDir, destPath string) error {
	outFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(sourceDir, path)
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarWriter, file)
			return err
		}
		return nil
	})
}
