// main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"

	"backup-tool/backup"
	"backup-tool/config"

	"github.com/joho/godotenv"
)

// envVarRegex находит шаблоны вида $ENV:KEY
var envVarRegex = regexp.MustCompile(`\$ENV:([A-Za-z_][A-Za-z0-9_]*)`)

func main() {
	configPath := flag.String("config", "config.json", "Путь к файлу конфигурации")
	envPath := flag.String("env", ".env", "Путь к .env файлу (опционально)")
	flag.Parse()

	// Загружаем .env (если существует)
	if _, err := os.Stat(*envPath); err == nil {
		if err := godotenv.Load(*envPath); err != nil {
			fmt.Printf("⚠️ Ошибка загрузки %s: %v\n", *envPath, err)
		} else {
			fmt.Printf("✅ Загружен .env файл: %s\n", *envPath)
		}
	}

	// Читаем конфиг
	data, err := os.ReadFile(*configPath)
	if err != nil {
		panic(fmt.Errorf("не удалось прочитать конфиг: %w", err))
	}

	// Подменяем $ENV:... → значения из окружения
	data = replaceEnvVars(data)

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Errorf("ошибка парсинга JSON: %w", err))
	}

	// Создаём директорию для бэкапов
	if err := os.MkdirAll(cfg.LocalBackupPath, 0755); err != nil {
		panic(fmt.Errorf("не удалось создать %s: %w", cfg.LocalBackupPath, err))
	}

	// Бэкапы
	if err := backup.BackupDirs(cfg.LocalBackupPath, cfg.Dirs); err != nil {
		fmt.Printf("⚠️ Ошибка бэкапа директорий: %v\n", err)
	}
	if err := backup.BackupFiles(cfg.LocalBackupPath, cfg.Files); err != nil {
		fmt.Printf("⚠️ Ошибка бэкапа файлов: %v\n", err)
	}
	if err := backup.BackupDatabases(cfg.LocalBackupPath, cfg.Databases, cfg.DatabaseUsers); err != nil {
		fmt.Printf("❌ Ошибка бэкапа БД: %v\n", err)
	}
	if err := backup.UploadToSMB(cfg.LocalBackupPath, cfg.Upload); err != nil {
		fmt.Printf("⚠️ Ошибка загрузки на SMB: %v\n", err)
	}

	fmt.Println("✅ Все задачи завершены.")
}

// replaceEnvVars заменяет все $ENV:KEY на os.Getenv("KEY")
func replaceEnvVars(data []byte) []byte {
	return envVarRegex.ReplaceAllFunc(data, func(match []byte) []byte {
		key := string(match[5:]) // пропускаем "$ENV:"
		value := os.Getenv(key)
		if value == "" {
			fmt.Printf("⚠️ Переменная окружения %s не задана — оставляем как есть\n", key)
			return match // оставляем исходную строку, если нет значения
		}
		return []byte(value)
	})
}
