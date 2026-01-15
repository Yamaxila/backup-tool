// Package backup
package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"backup-tool/config"
)

func BackupFiles(localPath string, items []config.Item) error {
	for _, item := range items {
		srcPath := item.Path

		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			fmt.Printf("⚠️ Файл %s не существует — пропускаем\n", srcPath)
			continue
		}

		baseName := filepath.Base(srcPath)
		parentDir := filepath.Dir(srcPath)

		archiveName := fmt.Sprintf("file_%s_%s.tar.gz", baseName, time.Now().Format("20060102_150405"))
		archivePath := filepath.Join(localPath, archiveName)

		// Команда: tar -czf archive.tar.gz -C parentDir baseName
		cmd := exec.Command("tar", "-czf", archivePath, "-C", parentDir, baseName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("ошибка архивации файла %s: %w", srcPath, err)
		}

		fmt.Printf("✅ Файл %s → %s\n", srcPath, archivePath)
		cleanupOldBackups(localPath, "file_"+baseName, item.Lifetime)
	}
	return nil
}
