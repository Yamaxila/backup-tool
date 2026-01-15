// main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"backup-tool/backup"
	"backup-tool/config"
	"github.com/joho/godotenv"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to configuration file")
	envPath := flag.String("env", ".env", "Path to .env file (optional)")
	flag.Parse()

	// Load .env file if it exists
	if _, err := os.Stat(*envPath); err == nil {
		if err := godotenv.Load(*envPath); err != nil {
			fmt.Printf("⚠️ Error loading %s: %v\n", *envPath, err)
		} else {
			fmt.Printf("✅ Loaded .env file: %s\n", *envPath)
		}
	}

	// Load configuration with environment variable substitution
	cfg, err := loadConfig(*configPath)
	if err != nil {
		panic(fmt.Errorf("error loading configuration: %w", err))
	}

	// Create root backup directory
	if err := os.MkdirAll(cfg.LocalBackupPath, 0755); err != nil {
		panic(fmt.Errorf("failed to create %s: %w", cfg.LocalBackupPath, err))
	}

	// === 1. Backups ===
	if err := backup.BackupDirs(cfg.LocalBackupPath, cfg.Dirs); err != nil {
		fmt.Printf("⚠️ Error backing up directories: %v\n", err)
	}
	if err := backup.BackupFiles(cfg.LocalBackupPath, cfg.Files); err != nil {
		fmt.Printf("⚠️ Error backing up files: %v\n", err)
	}
	if err := backup.BackupLogs(cfg.LocalBackupPath, cfg.Logs); err != nil {
		fmt.Printf("⚠️ Error backing up logs: %v\n", err)
	}
	if err := backup.BackupDatabases(cfg.LocalBackupPath, cfg.Databases, cfg.DatabaseUsers); err != nil {
		fmt.Printf("❌ Error backing up databases: %v\n", err)
	}

	// === 2. Upload to SMB + Cleanup on SMB ===
	if cfg.Upload.Active {
		// Upload ALL contents of LocalBackupPath to SMB
		if err := backup.UploadToSMB(cfg.LocalBackupPath, cfg.Upload); err != nil {
			fmt.Printf("⚠️ Error uploading to SMB: %v\n", err)
		}

		// Prepare list of items for cleanup on SMB
		var smbItems []backup.SMBItem

		for _, dir := range cfg.Dirs {
			smbItems = append(smbItems, backup.SMBItem{
				Prefix:   "dir_" + filepath.Base(dir.Path),
				Lifetime: dir.Lifetime,
			})
		}
		for _, file := range cfg.Files {
			smbItems = append(smbItems, backup.SMBItem{
				Prefix:   "file_" + filepath.Base(file.Path),
				Lifetime: file.Lifetime,
			})
		}
		for _, logItem := range cfg.Logs {
			smbItems = append(smbItems, backup.SMBItem{
				Prefix:   "log_" + filepath.Base(logItem.Path),
				Lifetime: logItem.Lifetime,
			})
		}
		for _, db := range cfg.Databases {
			smbItems = append(smbItems, backup.SMBItem{
				Prefix:   "db_" + db.Name,
				Lifetime: db.Lifetime,
			})
		}

		// Clean up old backups on SMB
		if err := backup.CleanupSMB(cfg.Upload, smbItems); err != nil {
			fmt.Printf("⚠️ Error cleaning up SMB: %v\n", err)
		}
	}

	fmt.Println("✅ All tasks completed.")
}

// loadConfig reads JSON configuration from disk and populates config.Config structure.
// Environment variable substitution can be added here if needed.
func loadConfig(path string) (*config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file %s: %w", path, err)
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing JSON in %s: %w", path, err)
	}

	return &cfg, nil
}
