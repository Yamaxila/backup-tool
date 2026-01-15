// Package utils
package utils

import (
	"os"
	"time"
)

// IsOlderThanDays checks if file is older than N days
func IsOlderThanDays(path string, days int) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	return info.ModTime().Before(cutoff), nil
}
