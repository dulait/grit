package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// replaceBinary overwrites the executable at execPath with newBinary.
// On Unix it uses an atomic rename; on Windows it renames the current
// binary out of the way first.
func replaceBinary(execPath string, newBinary []byte) error {
	if runtime.GOOS == "windows" {
		return replaceWindows(execPath, newBinary)
	}
	return replaceUnix(execPath, newBinary)
}

func replaceUnix(execPath string, newBinary []byte) error {
	dir := filepath.Dir(execPath)

	info, err := os.Stat(execPath)
	if err != nil {
		return fmt.Errorf("stat current binary: %w", err)
	}

	tmp, err := os.CreateTemp(dir, "grit-new-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(newBinary); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing new binary: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Chmod(tmpPath, info.Mode()); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("setting permissions: %w", err)
	}

	if err := os.Rename(tmpPath, execPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replacing binary: %w", err)
	}
	return nil
}

func replaceWindows(execPath string, newBinary []byte) error {
	oldPath := execPath + ".old"

	// Remove a leftover .old file from a previous update, if any.
	os.Remove(oldPath)

	if err := os.Rename(execPath, oldPath); err != nil {
		return fmt.Errorf("renaming current binary: %w", err)
	}

	if err := os.WriteFile(execPath, newBinary, 0o755); err != nil {
		// Attempt to restore the original binary.
		os.Rename(oldPath, execPath)
		return fmt.Errorf("writing new binary: %w", err)
	}

	// Best-effort cleanup of the old binary.
	os.Remove(oldPath)
	return nil
}
