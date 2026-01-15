package backup

import (
	"fmt"
	"net"
	"strings"
	"time"

	"backup-tool/config"
	"backup-tool/utils"
	"github.com/hirochachacha/go-smb2"
)

type SMBItem struct {
	Prefix   string
	Lifetime int
}

// CleanupSMB removes old backups on SMB share according to specified lifetime.
func CleanupSMB(upload config.Upload, items []SMBItem) error {
	if !upload.Active {
		return nil
	}

	conn, err := net.Dial("tcp", upload.SMBHost+":445")
	if err != nil {
		return fmt.Errorf("failed to connect to SMB: %w", err)
	}
	defer conn.Close()

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     upload.SMBUser,
			Password: upload.SMBPassword,
		},
	}

	s, err := d.Dial(conn)
	if err != nil {
		return fmt.Errorf("SMB authentication error: %w", err)
	}
	defer s.Logoff()

	fs, err := s.Mount(upload.SMBShare)
	if err != nil {
		return fmt.Errorf("failed to mount share: %w", err)
	}
	defer fs.Umount()

	now := time.Now()
	cutoffTime := now.AddDate(0, 0, -1) // Default if lifetime is not specified

	for _, item := range items {
		var smbDir string
		var prefix string

		// Determine SMB path and file prefix
		switch {
		case strings.HasPrefix(item.Prefix, "db_"):
			dbName := strings.TrimPrefix(item.Prefix, "db_")
			smbDir = fmt.Sprintf("databases/%s", dbName)
			prefix = "db_"
		case strings.HasPrefix(item.Prefix, "dir_"):
			dirName := strings.TrimPrefix(item.Prefix, "dir_")
			smbDir = fmt.Sprintf("dirs/%s", dirName)
			prefix = "dir_"
		case strings.HasPrefix(item.Prefix, "file_"):
			fileName := strings.TrimPrefix(item.Prefix, "file_")
			smbDir = fmt.Sprintf("files/%s", fileName)
			prefix = "file_"
		case strings.HasPrefix(item.Prefix, "log_"):
			logName := strings.TrimPrefix(item.Prefix, "log_")
			smbDir = fmt.Sprintf("logs/%s", logName)
			prefix = "log_"
		default:
			fmt.Printf("⚠️ Unknown prefix for cleanup: %s\n", item.Prefix)
			continue
		}

		// Calculate cutoff time for this item
		if item.Lifetime > 0 {
			cutoffTime = now.AddDate(0, 0, -item.Lifetime)
		}

		// Read subdirectory contents on SMB
		fileInfos, err := fs.ReadDir(smbDir)
		if err != nil {
			// Directory may not exist - this is normal, skip
			fmt.Printf("ℹ️ Directory %s not found on SMB (backups may not exist yet)\n", smbDir)
			continue
		}

		deletedCount := 0
		for _, fi := range fileInfos {
			name := fi.Name()
			if fi.IsDir() {
				continue
			}

			// Check that file matches backup format
			if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".tar.gz") {
				continue
			}

			// Try to extract time from filename
			backupTime, ok := utils.GetBackupTimeFromName(name)
			if !ok {
				// If we can't extract time from name, we could use file modification time
				// But for that we need to get full path and check via Stat
				// For simplicity, skip such files or use a more conservative approach
				fmt.Printf("⚠️ Failed to determine backup time from filename: %s (skipping)\n", name)
				continue
			}

			// Check if file should be deleted
			if backupTime.Before(cutoffTime) {
				// Use correct path formation for SMB (always /)
				fullPath := strings.TrimSuffix(smbDir, "/") + "/" + name
				if err := fs.Remove(fullPath); err != nil {
					fmt.Printf("⚠️ Failed to delete %s on SMB: %v\n", fullPath, err)
				} else {
					fmt.Printf("🗑️ Deleted old backup on SMB: %s (age: %d days)\n", 
						fullPath, int(now.Sub(backupTime).Hours()/24))
					deletedCount++
				}
			}
		}

		if deletedCount > 0 {
			fmt.Printf("✅ Cleaned up %d old backups in %s\n", deletedCount, smbDir)
		}
	}

	return nil
}
