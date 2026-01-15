## Backup Tool

Backup Tool is a small, focused utility for **file system and database backups** with:

- **Config‑driven** setup (single JSON file)
- **Directory and file backups** into compressed `tar.gz` archives
- **PostgreSQL / MySQL / MongoDB** backups
- Optional **upload to SMB share** and **retention policy** for old backups (locally and on SMB)
- Integration with **systemd service + timer** for scheduled runs (e.g. daily at 02:00)
- Optional **`.env` file support** for secrets and connection parameters

---

### Requirements

- **Go**: `>= 1.25` (may work on lower versions, but not tested)
- Unix‑like OS with:
  - `tar`
  - For databases (optional, depending on what you use):
    - PostgreSQL: `pg_dump`
    - MySQL/MariaDB: `mysqldump`
    - MongoDB: `mongodump`

---

### Build and Run

From the project root:

```bash
go mod tidy
go build -o backup-tool
chmod +x ./backup-tool
```

Run manually with a config:

```bash
./backup-tool -config ./config.json
```

Environment variables from `.env` (if present in the project root) are loaded automatically at startup.

---

### Configuration

The application is configured via a single JSON file.  
See `config.json.example` for a full example. Basic structure:

```json
{
  "localBackupPath": "/backups",
  "databaseUsers": {
    "main_pg": {
      "user": "postgres",
      "password": "postgres",
      "host": "127.0.0.1",
      "port": 5432
    }
  },
  "dirs": [
    { "path": "/var/www", "lifetime": 7 }
  ],
  "files": [
    { "path": "/etc/nginx/nginx.conf", "lifetime": 14 }
  ],
  "logs": [
    { "path": "/var/log/nginx/access.log", "lifetime": 7 }
  ],
  "databases": [
    {
      "name": "appdb",
      "type": "postgres",
      "userRef": "main_pg",
      "lifetime": 30
    }
  ],
  "upload": {
    "active": true,
    "smbuser": "DOMAIN\\backup",
    "smbpassword": "password",
    "smbhost": "192.168.1.234",
    "smbshare": "\\\\server\\backups"
  }
}
```

- **`localBackupPath`**: root directory where backups are written locally.
- **`databaseUsers`**: reusable DB connection profiles, referenced by `userRef`.
- **`dirs`**:
  - `path`: directory to back up
  - `lifetime` (days): how long to keep archives for this directory
- **`files`**:
  - `path`: single file to back up
  - `lifetime` (days): retention for this file’s backups
- **`logs`**:
  - `path`: log file to back up
  - `lifetime` (days): retention for this log’s backups
  - after successful backup the **source log file is truncated** (log rotation behavior)
- **`databases`**:
  - `name`: database name
  - `type`: `postgres`, `mysql`, or `mongo`
  - `userRef`: key in `databaseUsers` used to connect
  - `lifetime` (days): retention for DB backups
- **`upload`**:
  - `active`: enable/disable SMB upload and cleanup
  - `smbuser`, `smbpassword`, `smbhost`, `smbshare`: SMB connection parameters

---

### What the Tool Does

- **Directories**  
  For each entry in `dirs`, the tool creates:

  - Local path:  
    `<localBackupPath>/dirs/<basename>/dir_YYYYMMDD_HHMMSS.tar.gz`

- **Files**  
  For each entry in `files`:

  - Local path:  
    `<localBackupPath>/files/<basename>/file_YYYYMMDD_HHMMSS.tar.gz`

- **Logs**  
  For each entry in `logs`:

  - Local path:  
    `<localBackupPath>/logs/<basename>/log_YYYYMMDD_HHMMSS.tar.gz`  
  - After archiving the log file, the original log is truncated to size 0.

- **Databases**  
  For each entry in `databases`:

  - Local path:  
    `<localBackupPath>/databases/<dbName>/db_YYYYMMDD_HHMMSS.tar.gz`

In each of these subdirectories, old backups are automatically removed according to the `lifetime` setting.

If `upload.active` is `true`, the entire `localBackupPath` tree is mirrored to the SMB share, and old archives are also cleaned up on SMB.

---

### .env Support

If `.env` exists in the project root, it is loaded on startup (via `github.com/joho/godotenv`).  
You can safely move secrets (passwords, hosts, etc.) into environment variables and reference them in your config or systemd environment.

Example `.env`:

```bash
PG_MAIN_PASSWORD=postgres
MONGO_LOCAL_PASSWORD=secret
SMB_PASSWORD=VerySecret
```

---

### systemd Integration (Scheduled Backups)

In the `systemd` directory you will find ready‑to‑use units:

- `systemd/backup-tool.service`
- `systemd/backup-tool.timer`

The timer is configured to run the backup **daily at 02:00**.

#### 1. Install units

```bash
sudo cp systemd/backup-tool.service /etc/systemd/system/
sudo cp systemd/backup-tool.timer /etc/systemd/system/
sudo systemctl daemon-reload
```

Make sure the binary is built and paths inside the `.service` match your installation:

- `WorkingDirectory=/home/yamaxila/projects/backupsProject`
- `ExecStart=/home/yamaxila/projects/backupsProject/backup-tool -config /home/yamaxila/projects/backupsProject/config.json`
- `EnvironmentFile=-/home/yamaxila/projects/backupsProject/.env`

Adjust these if you deploy the binary elsewhere.

#### 2. Enable and start timer

```bash
sudo systemctl enable --now backup-tool.timer
```

Check status:

```bash
systemctl status backup-tool.timer
```

View last run logs:

```bash
journalctl -u backup-tool.service
```

---

### Development Notes

- Project module name: `backup-tool` (see `go.mod`).
- Main entry point: `main.go`.
- Core logic:
  - `backup/dirs.go`, `backup/files.go`, `backup/databases.go`
  - `backup/upload.go`, `backup/smb.go`, `backup/cleanup.go`
  - `backup/utils.go`, `utils/time.go`, `config/config.go`

---

### License

This project is intended for internal use.  
If you plan to redistribute it, consider adding an explicit license file. 
