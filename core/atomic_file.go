package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// EnsurePrivateDir creates a directory and enforces owner-only permissions.
func EnsurePrivateDir(path string) error {
	if err := os.MkdirAll(path, 0700); err != nil {
		return err
	}
	if err := os.Chmod(path, 0700); err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("private storage path %q is not a directory", path)
	}
	if info.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("private directory %q has unsafe permissions %o", path, info.Mode().Perm())
	}
	return nil
}

// WriteFileAtomic writes data through a synced temporary file and rename.
func WriteFileAtomic(path string, data []byte, mode fs.FileMode) error {
	if path == "" {
		return fmt.Errorf("atomic file path is empty")
	}
	parent := filepath.Dir(path)
	if err := EnsurePrivateDir(parent); err != nil {
		return fmt.Errorf("prepare atomic file directory: %w", err)
	}
	if mode == 0 {
		mode = 0600
	}
	tmp, err := os.CreateTemp(parent, "."+filepath.Base(path)+".tmp-")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tmpName)
		}
	}()
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	removeTemp = false
	if dir, err := os.Open(parent); err == nil {
		_ = dir.Sync()
		_ = dir.Close()
	}
	return nil
}

func preserveCorruptFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return WriteFileAtomic(path+".bak", data, 0600)
}
