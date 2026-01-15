// Package config
package config

type Config struct {
	LocalBackupPath string            `json:"localBackupPath"`
	Dirs            []Item            `json:"dirs"`
	Files           []Item            `json:"files"`
	DatabaseUsers   map[string]DBUser `json:"databaseUsers,omitempty"` // ← НОВОЕ
	Databases       []Database        `json:"databases"`
	Upload          Upload            `json:"upload"`
}

type Item struct {
	Path     string `json:"path"`
	Lifetime int    `json:"lifetime"`
}

// DBUser — общие параметры подключения
type DBUser struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

// Database — теперь ссылается на userRef
type Database struct {
	Name     string `json:"name"`
	Type     string `json:"type"`    // postgres, mysql, mongo
	UserRef  string `json:"userRef"` // ← ссылка на ключ в DatabaseUsers
	Lifetime int    `json:"lifetime"`
}

type Upload struct {
	Active      bool   `json:"active"`
	SMBUser     string `json:"smbuser"`
	SMBPassword string `json:"smbpassword"`
	SMBHost     string `json:"smbhost"`
	SMBShare    string `json:"smbshare"`
}
