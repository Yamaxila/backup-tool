// Package backup
package backup

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"backup-tool/config"
	"github.com/hirochachacha/go-smb2"
)

// UploadToSMB recursively uploads contents of localPath to SMB share,
// preserving directory structure.
func UploadToSMB(localPath string, upload config.Upload) error {
	if !upload.Active {
		return nil
	}

	// Normalize local path for correct comparison
	localPath, err := filepath.Abs(localPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	localPath = filepath.Clean(localPath)

	conn, err := net.Dial("tcp", upload.SMBHost+":445")
	if err != nil {
		return fmt.Errorf("failed to connect to SMB: %w", err)
	}
	defer conn.Close()

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     upload.SMBUser,
			Password: upload.SMBPassword,
			Domain:   upload.Domain,
		},
	}

	s, err := d.Dial(conn)
	if err != nil {
		return fmt.Errorf("SMB authentication error: %w", err)
	}
	defer s.Logoff()

	fs, err := s.Mount(upload.SMBShare)
	if err != nil {
		return fmt.Errorf("failed to mount share %s: %w", upload.SMBShare, err)
	}
	defer fs.Umount()

	fmt.Println("📤 Starting upload to SMB...")

	// Recursively walk local directory
	return filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking %s: %w", path, err)
		}

		// Relative path from localPath
		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		// Skip root directory (relPath will be ".")
		if relPath == "." {
			return nil
		}

		// Convert separators to /
		smbPath := strings.ReplaceAll(relPath, string(filepath.Separator), "/")

		if info.IsDir() {
			// Create directory on SMB
			if err := fs.Mkdir(smbPath, 0755); err != nil {
				// Check if error is related to directory existence
				if pathErr, ok := err.(*os.PathError); ok {
					if pathErr.Err == syscall.EEXIST || pathErr.Err == syscall.EISDIR {
						// Directory already exists - this is fine
						return nil
					}
				}
				// If error is not related to existence, check text (for cross-platform compatibility)
				errStr := err.Error()
				if !strings.Contains(errStr, "exists") &&
					!strings.Contains(errStr, "EEXIST") &&
					!strings.Contains(errStr, "file exists") {
					return fmt.Errorf("failed to create directory %s on SMB: %w", smbPath, err)
				}
			} else {
				fmt.Printf("📁 Created directory on SMB: %s\n", smbPath)
			}
		} else {
			// Upload file using streaming for large files
			srcFile, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", path, err)
			}
			defer srcFile.Close()

			dstFile, err := fs.Create(smbPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s on SMB: %w", smbPath, err)
			}

			written, err := io.Copy(dstFile, srcFile)
			if err != nil {
				dstFile.Close()
				return fmt.Errorf("error copying %s to SMB: %w", smbPath, err)
			}

			if err := dstFile.Close(); err != nil {
				return fmt.Errorf("error closing file %s on SMB: %w", smbPath, err)
			}

			fmt.Printf("✅ Uploaded: %s (%d bytes)\n", smbPath, written)
		}
		return nil
	})
}
