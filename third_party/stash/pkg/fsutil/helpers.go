package fsutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func FileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return !info.IsDir(), nil
	}
	return false, err
}

func FindInPaths(paths []string, baseName string) string {
	for _, p := range paths {
		if p == "" {
			continue
		}
		candidate := filepath.Join(p, baseName)
		if exists, _ := FileExists(candidate); exists {
			return candidate
		}
	}
	return ""
}

func SafeMove(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	} else {
		if copyErr := copyFile(src, dst); copyErr != nil {
			return fmt.Errorf("copying file during SafeMove failed: %w; rename also failed: %v", copyErr, err)
		}
		if removeErr := os.Remove(src); removeErr != nil {
			_ = os.Remove(dst)
			return fmt.Errorf("removing old file during SafeMove failed: %w; rename also failed: %v", removeErr, err)
		}
	}
	return nil
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := out.Close()
		if err == nil {
			err = closeErr
		}
	}()

	_, err = io.Copy(out, in)
	return err
}
