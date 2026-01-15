// utils/time.go
package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var timestampRegex = regexp.MustCompile(`_(\d{8}_\d{6})\.tar\.gz$`)

func GetBackupTimeFromName(filename string) (time.Time, bool) {
	matches := timestampRegex.FindStringSubmatch(filename)
	if len(matches) < 2 {
		return time.Time{}, false
	}
	t, err := time.Parse("20060102_150405", matches[1])
	return t, err == nil
}

func IsBackupOlderThan(fullPath string, days int) bool {
	filename := filepath.Base(fullPath)
	backupTime, ok := GetBackupTimeFromName(filename)
	if !ok {
		if info, err := os.Stat(fullPath); err == nil {
			return info.ModTime().Before(time.Now().AddDate(0, 0, -days))
		}
		return false
	}
	return backupTime.Before(time.Now().AddDate(0, 0, -days))
}
