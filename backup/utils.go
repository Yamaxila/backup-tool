package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func ensureBackupSubdir(root, category, subName string) (string, error) {
	dirPath := filepath.Join(root, category, subName)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}
	return dirPath, nil
}

// runTar creates a tar.gz archive with specified contents.
// targetArchive - path to the archive being created (must have .tar.gz extension)
// baseDir - base directory for tar -C command
// entryName - name of file or directory to archive
func runTar(targetArchive, baseDir, entryName string) error {
	// Check that target archive has correct extension
	if filepath.Ext(targetArchive) != ".gz" {
		return fmt.Errorf("archive must have .tar.gz extension, got: %s", targetArchive)
	}

	cmd := exec.Command("tar", "-czf", targetArchive, "-C", baseDir, entryName)
	
	// Capture command output for more informative errors
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tar execution error: %w, output: %s", err, string(output))
	}
	
	return nil
}
