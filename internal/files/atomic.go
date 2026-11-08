package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// WriteFileAtomic writes data to a temporary file next to path and renames it
// over path. Readers and a crash or power loss in the middle of the write see
// either the old or the new contents, never a truncated file.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) (err error) {
	tmpPath := filepath.Join(filepath.Dir(path),
		"."+filepath.Base(path)+".tmp"+strconv.FormatInt(time.Now().UnixNano(), 36))

	file, err := Fs.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		if err != nil {
			_ = Fs.Remove(tmpPath)
		}
	}()

	if _, err = file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("failed to write temporary file: %w", err)
	}
	if err = file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("failed to sync temporary file: %w", err)
	}
	if err = file.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}
	if err = Fs.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to replace %s: %w", path, err)
	}
	return nil
}
