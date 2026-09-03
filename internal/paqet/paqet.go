package paqet

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

// FileExists checks if a file exists at the given path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return false
}

// EnsureDir creates the parent directories for the given file path if they don't exist.
func EnsureDir(filePath string) error {
	dir := filepath.Dir(filePath)
	return os.MkdirAll(dir, 0755)
}

// GetBinaryName returns the binary name corresponding to the current OS and architecture.
func GetBinaryName() string {
	return fmt.Sprintf("paqet_linux_%s", runtime.GOARCH)
}

// InstallPaqet extracts the embedded binary from srcFS to the specified destination path.
func InstallPaqet(dstPath string, srcFS fs.FS) error {
	if FileExists(dstPath) {
		return nil
	}

	if err := EnsureDir(dstPath); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	binName := GetBinaryName()
	return copyFromFS(srcFS, binName, dstPath)
}

func copyFromFS(srcFS fs.FS, srcName, dstPath string) error {
	srcFile, err := srcFS.Open(srcName)
	if err != nil {
		return fmt.Errorf("embedded binary %q not found: %w", srcName, err)
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dstPath, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("error copying content: %w", err)
	}

	if err := os.Chmod(dstPath, 0755); err != nil {
		return fmt.Errorf("failed to set 0755 permissions: %w", err)
	}

	return dstFile.Sync()
}
