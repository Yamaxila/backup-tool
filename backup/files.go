// Package backup
package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"backup-tool/config"
)

// BackupFiles archives individual files into tar.gz archives.
// Creates structure: <localBackupPath>/files/<basename>/file_YYYYMMDD_HHMMSS.tar.gz
func BackupFiles(localPath string, items []config.Item) error {
	for _, item := range items {
		srcPath := item.Path
		info, err := os.Stat(srcPath)
		if os.IsNotExist(err) {
			fmt.Printf("⚠️ File %s does not exist — skipping\n", srcPath)
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking file %s: %w", srcPath, err)
		}
		if info.IsDir() {
			return fmt.Errorf("%s is a directory, use BackupDirs instead", srcPath)
		}

		baseName := filepath.Base(srcPath)
		subDir, err := ensureBackupSubdir(localPath, "files", baseName)
		if err != nil {
			return fmt.Errorf("failed to create subdirectory for file %s: %w", baseName, err)
		}

		archiveName := fmt.Sprintf("file_%s.tar.gz", time.Now().Format("20060102_150405"))
		archivePath := filepath.Join(subDir, archiveName)

		parentDir := filepath.Dir(srcPath)
		if err := runTar(archivePath, parentDir, baseName); err != nil {
			return fmt.Errorf("error archiving file %s: %w", srcPath, err)
		}

		// Verify that archive was actually created
		if _, err := os.Stat(archivePath); os.IsNotExist(err) {
			return fmt.Errorf("archive was not created: %s", archivePath)
		}

		fmt.Printf("✅ File %s → %s\n", srcPath, archivePath)
		cleanupOldBackups(subDir, "file_", item.Lifetime)
	}
	return nil
}
