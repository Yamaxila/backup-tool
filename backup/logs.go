// Package backup
package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"backup-tool/config"
)

// BackupLogs archives log files and then truncates the original log files.
// Creates structure: <localBackupPath>/logs/<basename>/log_YYYYMMDD_HHMMSS.tar.gz
func BackupLogs(localPath string, items []config.Item) error {
	for _, item := range items {
		srcPath := item.Path
		info, err := os.Stat(srcPath)
		if os.IsNotExist(err) {
			fmt.Printf("⚠️ Log file %s does not exist — skipping\n", srcPath)
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking log file %s: %w", srcPath, err)
		}
		if info.IsDir() {
			return fmt.Errorf("%s is a directory, expected log file", srcPath)
		}

		baseName := filepath.Base(srcPath)
		subDir, err := ensureBackupSubdir(localPath, "logs", baseName)
		if err != nil {
			return fmt.Errorf("failed to create subdirectory for log %s: %w", baseName, err)
		}

		archiveName := fmt.Sprintf("log_%s.tar.gz", time.Now().Format("20060102_150405"))
		archivePath := filepath.Join(subDir, archiveName)

		parentDir := filepath.Dir(srcPath)
		if err := runTar(archivePath, parentDir, baseName); err != nil {
			return fmt.Errorf("error archiving log file %s: %w", srcPath, err)
		}

		// Verify that archive was actually created
		if _, err := os.Stat(archivePath); os.IsNotExist(err) {
			return fmt.Errorf("archive was not created: %s", archivePath)
		}

		// Truncate original log file after successful backup
		if err := os.Truncate(srcPath, 0); err != nil {
			return fmt.Errorf("failed to truncate log file %s after backup: %w", srcPath, err)
		}

		fmt.Printf("✅ Log file %s → %s (source truncated)\n", srcPath, archivePath)
		cleanupOldBackups(subDir, "log_", item.Lifetime)
	}
	return nil
}

