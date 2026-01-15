// Package config
package config

type Config struct {
	LocalBackupPath string            `json:"localBackupPath"`
	Dirs            []Item            `json:"dirs"`
	Files           []Item            `json:"files"`
	Logs            []Item            `json:"logs,omitempty"`
	DatabaseUsers   map[string]DBUser `json:"databaseUsers,omitempty"`
	Databases       []Database        `json:"databases"`
	Upload          Upload            `json:"upload"`
}

type Item struct {
	Path     string `json:"path"`
	Lifetime int    `json:"lifetime"`
}

// DBUser contains common database connection parameters
type DBUser struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

// Database now references userRef
type Database struct {
	Name     string `json:"name"`
	Type     string `json:"type"`    // postgres, mysql, mongo
	UserRef  string `json:"userRef"` // reference to key in DatabaseUsers
	Lifetime int    `json:"lifetime"`
}

type Upload struct {
	Active      bool   `json:"active"`
	SMBUser     string `json:"smbuser"`
	SMBPassword string `json:"smbpassword"`
	SMBHost     string `json:"smbhost"`
	SMBShare    string `json:"smbshare"`
	Domain    string `json:"domain"`
}
