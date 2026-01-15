// Package backup
package backup

import (
	"backup-tool/config"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/hirochachacha/go-smb2"
)

func UploadToSMB(localPath string, u config.Upload) error {
	if !u.Active {
		return nil
	}

	conn, err := net.Dial("tcp", u.SMBHost+":445")
	if err != nil {
		return err
	}
	defer conn.Close()

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     u.SMBUser,
			Password: u.SMBPassword,
		},
	}

	s, err := d.Dial(conn)
	if err != nil {
		return err
	}
	defer s.Logoff()

	fs, err := s.Mount(u.SMBShare)
	if err != nil {
		return err
	}
	defer fs.Umount()

	entries, err := os.ReadDir(localPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		srcPath := filepath.Join(localPath, entry.Name())
		dstPath := entry.Name()

		data, err := os.ReadFile(srcPath)
		if err != nil {
			fmt.Printf("⚠️ Failed to read %s: %v\n", srcPath, err)
			continue
		}

		f, err := fs.Create(dstPath)
		if err != nil {
			fmt.Printf("⚠️ Failed to create %s on SMB: %v\n", dstPath, err)
			continue
		}
		f.Write(data)
		f.Close()
		fmt.Printf("📤 Uploaded %s to SMB\n", dstPath)
	}

	return nil
}
