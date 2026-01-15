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

func BackupDatabases(localPath string, dbs []config.Database, users map[string]config.DBUser) error {
	for _, db := range dbs {
		user, exists := users[db.UserRef]
		if !exists {
			return fmt.Errorf("не найден databaseUsers.%s для БД %s", db.UserRef, db.Name)
		}

		archiveName := fmt.Sprintf("db_%s_%s.tar.gz", db.Name, time.Now().Format("20060102_150405"))
		archivePath := filepath.Join(localPath, archiveName)

		// Временная директория для дампа
		tempDir, err := os.MkdirTemp("", "dbbackup-*")
		if err != nil {
			return fmt.Errorf("не удалось создать временную директорию: %w", err)
		}
		defer os.RemoveAll(tempDir)

		switch strings.ToLower(db.Type) {
		case "postgres":
			tarFile := filepath.Join(tempDir, "dump.tar")
			cmd := exec.Command("pg_dump", "-h", user.Host, "-p", fmt.Sprint(user.Port), "-U", user.User, "-F", "t", "-f", tarFile, db.Name)
			cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", user.Password))
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("ошибка pg_dump для %s: %w", db.Name, err)
			}
			// Упаковываем dump.tar в архив
			if err := runTar(archivePath, tempDir, "dump.tar"); err != nil {
				return fmt.Errorf("ошибка архивации PostgreSQL: %w", err)
			}

		case "mysql":
			fullSqlPath := filepath.Join(tempDir, "dump.sql")
			cmd := exec.Command("mysqldump",
				"-h", user.Host,
				"-P", fmt.Sprint(user.Port),
				"-u", user.User,
				"--password="+user.Password,
				db.Name,
				"--result-file", fullSqlPath)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("ошибка mysqldump для %s: %w", db.Name, err)
			}
			if err := runTar(archivePath, tempDir, "dump.sql"); err != nil {
				return fmt.Errorf("ошибка архивации MySQL: %w", err)
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
				return fmt.Errorf("ошибка mongodump для %s: %w", db.Name, err)
			}
			// mongodump создаёт dump/<dbname>, но мы архивируем всю папку dump
			if err := runTar(archivePath, tempDir, "dump"); err != nil {
				return fmt.Errorf("ошибка архивации MongoDB: %w", err)
			}

		default:
			return fmt.Errorf("неподдерживаемый тип БД: %s", db.Type)
		}

		fmt.Printf("✅ Бэкап БД %s → %s\n", db.Name, archivePath)
		cleanupOldBackups(localPath, "db_"+db.Name, db.Lifetime)
	}
	return nil
}

// runTar выполняет: tar -czf target.tar.gz -C baseDir entry
func runTar(targetArchive, baseDir, entryName string) error {
	cmd := exec.Command("tar", "-czf", targetArchive, "-C", baseDir, entryName)
	return cmd.Run()
}
