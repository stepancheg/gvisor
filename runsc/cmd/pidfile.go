package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// WritePidFile writes pid file atomically if possible.
func WritePidFile(path string, pid int) error {
	pidStr := []byte(strconv.Itoa(pid))

	if fileExists(path) {
		// If path exists, write in place, because file could be pipe or something.
		if err := os.WriteFile(path, pidStr, 0644); err != nil {
			return fmt.Errorf("failed to write pid file %s: %w", path, err)
		}
	} else {
		// Otherwise write using temp file to make write atomic.
		dir := filepath.Dir(path)
		tempFile, err := os.CreateTemp(dir, "pid-tmp-*")
		if err != nil {
			return fmt.Errorf("failed to create temp pid file in dir %s: %w", dir, err)
		}

		tempFileRenamed := false
		defer func(tempFile *os.File) {
			_ = tempFile.Close()
			if !tempFileRenamed {
				_ = os.Remove(tempFile.Name())
			}
		}(tempFile)

		if err := os.Chmod(tempFile.Name(), 0644); err != nil {
			return fmt.Errorf("failed to chmod pid file %s: %w", tempFile.Name(), err)
		}

		if _, err := tempFile.Write(pidStr); err != nil {
			return fmt.Errorf("failed to write pid file %s: %w", tempFile.Name(), err)
		}

		if err := tempFile.Close(); err != nil {
			return fmt.Errorf("failed to close temp pid file %s: %w", tempFile.Name(), err)
		}

		if err := os.Rename(tempFile.Name(), path); err != nil {
			return fmt.Errorf("failed to rename temp pid file %s -> %s: %w", tempFile.Name(), path, err)
		}
		tempFileRenamed = true
	}

	return nil
}
