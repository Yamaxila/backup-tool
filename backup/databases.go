// Package backup
package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"backup-tool/config"
)

// BackupDatabases creates backups of databases and archives them into tar.gz files.
func BackupDatabases(localPath string, dbs []config.Database, users map[string]config.DBUser) error {
	for _, db := range dbs {
		user, exists := users[db.UserRef]
		if !exists {
			return fmt.Errorf("databaseUsers.%s not found for database %s", db.UserRef, db.Name)
		}

		subDir, err := ensureBackupSubdir(localPath, "databases", db.Name)
		if err != nil {
			return fmt.Errorf("failed to create subdirectory for database %s: %w", db.Name, err)
		}

		archiveName := fmt.Sprintf("db_%s.tar.gz", time.Now().Format("20060102_150405"))
		archivePath := filepath.Join(subDir, archiveName)

		tempDir, err := os.MkdirTemp("", "dbbackup-*")
		if err != nil {
			return fmt.Errorf("failed to create temporary directory: %w", err)
		}
		defer os.RemoveAll(tempDir)

		switch strings.ToLower(db.Type) {
		case "postgres":
			tarFile := filepath.Join(tempDir, "dump.tar")
			cmd := exec.Command("pg_dump", "-h", user.Host, "-p", fmt.Sprint(user.Port), "-U", user.User, "-F", "t", "-f", tarFile, db.Name)
			cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", user.Password))
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("pg_dump error for %s: %w", db.Name, err)
			}
			if err := runTar(archivePath, tempDir, "dump.tar"); err != nil {
				return fmt.Errorf("error archiving PostgreSQL backup: %w", err)
			}

		case "mysql":
			sqlFile := filepath.Join(tempDir, "dump.sql")
			cmd := exec.Command("mysqldump",
				"-h", user.Host,
				"-P", fmt.Sprint(user.Port),
				"-u", user.User,
				"--password="+user.Password,
				db.Name,
				"--result-file", sqlFile)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("mysqldump error for %s: %w", db.Name, err)
			}
			if err := runTar(archivePath, tempDir, "dump.sql"); err != nil {
				return fmt.Errorf("error archiving MySQL backup: %w", err)
			}

		case "mongo":
			dumpDir := filepath.Join(tempDir, "dump")
			cmd := exec.Command("mongodump",
				"--host", fmt.Sprintf("%s:%d", user.Host, user.Port),
				"--db", db.Name,
				"--out", dumpDir)
			if user.User != "" {
				cmd.Args = append(cmd.Args, "--username", user.User)
				if user.Password != "" {
					cmd.Args = append(cmd.Args, "--password", user.Password)
				}
			}
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("mongodump error for %s: %w", db.Name, err)
			}
			if err := runTar(archivePath, tempDir, "dump"); err != nil {
				return fmt.Errorf("error archiving MongoDB backup: %w", err)
			}

		default:
			return fmt.Errorf("unsupported database type: %s", db.Type)
		}

		// Verify that archive was actually created
		if _, err := os.Stat(archivePath); os.IsNotExist(err) {
			return fmt.Errorf("archive was not created: %s", archivePath)
		}

		fmt.Printf("✅ Database backup %s → %s\n", db.Name, archivePath)
		cleanupOldBackups(subDir, "db_", db.Lifetime)
	}
	return nil
}
