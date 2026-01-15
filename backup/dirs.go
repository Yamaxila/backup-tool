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

func BackupDirs(localPath string, items []config.Item) error {
	for _, item := range items {
		srcPath := item.Path

		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			fmt.Printf("⚠️ Директория %s не существует — пропускаем\n", srcPath)
			continue
		}

		baseName := filepath.Base(srcPath)
		parentDir := filepath.Dir(srcPath)

		archiveName := fmt.Sprintf("dir_%s_%s.tar.gz", baseName, time.Now().Format("20060102_150405"))
		archivePath := filepath.Join(localPath, archiveName)

		// Команда: tar -czf archive.tar.gz -C parentDir baseName
		cmd := exec.Command("tar", "-czf", archivePath, "-C", parentDir, baseName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("ошибка архивации директории %s: %w", srcPath, err)
		}

		fmt.Printf("✅ Директория %s → %s\n", srcPath, archivePath)
		cleanupOldBackups(localPath, "dir_"+baseName, item.Lifetime)
	}
	return nil
}
