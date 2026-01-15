// Package backup
package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"backup-tool/config"
)

// BackupDirs archives directories into tar.gz archives.
// Creates structure: <localBackupPath>/dirs/<basename>/dir_YYYYMMDD_HHMMSS.tar.gz
func BackupDirs(localPath string, items []config.Item) error {
	for _, item := range items {
		srcPath := item.Path
		info, err := os.Stat(srcPath)
		if os.IsNotExist(err) {
			fmt.Printf("⚠️ Directory %s does not exist — skipping\n", srcPath)
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking directory %s: %w", srcPath, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("%s is not a directory", srcPath)
		}

		baseName := filepath.Base(srcPath)
		subDir, err := ensureBackupSubdir(localPath, "dirs", baseName)
		if err != nil {
			return fmt.Errorf("failed to create subdirectory for %s: %w", baseName, err)
		}

		archiveName := fmt.Sprintf("dir_%s.tar.gz", time.Now().Format("20060102_150405"))
		archivePath := filepath.Join(subDir, archiveName)

		parentDir := filepath.Dir(srcPath)
		if err := runTar(archivePath, parentDir, baseName); err != nil {
			return fmt.Errorf("error archiving directory %s: %w", srcPath, err)
		}

		// Verify that archive was actually created
		if _, err := os.Stat(archivePath); os.IsNotExist(err) {
			return fmt.Errorf("archive was not created: %s", archivePath)
		}

		fmt.Printf("✅ Directory %s → %s\n", srcPath, archivePath)
		cleanupOldBackups(subDir, "dir_", item.Lifetime)
	}
	return nil
}
