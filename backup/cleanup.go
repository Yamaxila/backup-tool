// Package backup
package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"backup-tool/utils"
)

// cleanupOldBackups removes old backup files based on their lifetime.
func cleanupOldBackups(dir, prefix string, lifetime int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("⚠️ Failed to read %s: %v\n", dir, err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".tar.gz") {
			fullPath := filepath.Join(dir, name)
			if utils.IsBackupOlderThan(fullPath, lifetime) {
				if err := os.Remove(fullPath); err != nil {
					fmt.Printf("❌ Failed to delete %s: %v\n", name, err)
				} else {
					fmt.Printf("🗑️ Deleted old backup: %s\n", name)
				}
			}
		}
	}
}
