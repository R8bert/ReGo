package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/r8bert/rego/internal/utils"
)

// Exporter handles creating portable backup archives
type Exporter struct{}

func NewExporter() *Exporter { return &Exporter{} }

// ExportToFile creates a single portable .tar.gz file from a backup
func (e *Exporter) ExportToFile(backupDir, outputPath string) error {
	// If no extension, add .tar.gz
	if filepath.Ext(outputPath) == "" {
		outputPath += ".tar.gz"
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Walk the backup directory and add all files
	return filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create header
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		// Update name to be relative to backup dir
		relPath, err := filepath.Rel(backupDir, path)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// If it's a file, write contents
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

// ExportQuick does a backup and immediately exports to a single file
func (e *Exporter) ExportQuick(opts BackupOptions, outputPath string) error {
	// Create temp backup
	tmpDir, err := os.MkdirTemp("", "rego-backup-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	opts.BackupPath = tmpDir
	mgr := NewManager()
	_, err = mgr.RunBackup(opts, nil)
	if err != nil {
		return err
	}

	return e.ExportToFile(tmpDir, outputPath)
}

// ImportFromFile extracts a .tar.gz backup to a directory
func (e *Exporter) ImportFromFile(archivePath, outputDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath := filepath.Join(outputDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := utils.EnsureDir(filepath.Dir(targetPath)); err != nil {
				return err
			}
			outFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}

// GetDefaultExportPath returns a suggested export filename
func GetDefaultExportPath() string {
	home, _ := utils.GetHomeDir()
	timestamp := time.Now().Format("2006-01-02")
	hostname, _ := os.Hostname()
	return filepath.Join(home, fmt.Sprintf("rego-backup-%s-%s.tar.gz", hostname, timestamp))
}
