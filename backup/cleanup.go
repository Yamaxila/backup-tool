// Package backup
package backup

import (
	"backup-tool/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func cleanupOldBackups(dir, prefix string, lifetime int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("⚠️ Cleanup failed: %v\n", err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			path := filepath.Join(dir, entry.Name())
			old, err := utils.IsOlderThanDays(path, lifetime)
			if err == nil && old {
				os.Remove(path)
				fmt.Printf("🗑️ Deleted old backup: %s\n", path)
			}
		}
	}
}
