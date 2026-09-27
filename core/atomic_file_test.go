package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestAtomicWritePreservesPreviousFileOnParentFailure(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "state.json")
	if err := os.WriteFile(target, []byte(`{"version":1}`), 0600); err != nil {
		t.Fatal(err)
	}
	badParent := filepath.Join(root, "not-a-directory")
	if err := os.WriteFile(badParent, []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomic(filepath.Join(badParent, "state.json"), []byte(`{"version":2}`), 0600); err == nil {
		t.Fatal("WriteFileAtomic() unexpectedly succeeded below a regular file")
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != `{"version":1}` {
		t.Fatalf("previous file = %q, %v", data, err)
	}
}

func TestAtomicEnsurePrivateDirAndMode(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "private")
	if err := EnsurePrivateDir(dir); err != nil {
		t.Fatal(err)
	}
	if mode := mustFileMode(t, dir); mode.Perm()&0077 != 0 {
		t.Fatalf("directory mode = %o, want private", mode.Perm())
	}
	path := filepath.Join(dir, "state.json")
	if err := WriteFileAtomic(path, []byte("ok"), 0600); err != nil {
		t.Fatal(err)
	}
	if mode := mustFileMode(t, path); mode.Perm() != 0600 {
		t.Fatalf("file mode = %o, want 0600", mode.Perm())
	}
}

func TestAtomicConcurrentWritersLeaveCompleteJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(version int) {
			defer wg.Done()
			data, _ := json.Marshal(map[string]int{"version": version})
			if err := WriteFileAtomic(path, data, 0600); err != nil {
				t.Errorf("WriteFileAtomic() error = %v", err)
			}
		}(i)
	}
	wg.Wait()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]int
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("final file is torn JSON %q: %v", data, err)
	}
}

func TestConfigCorruptionPreservesSourceAndWritesBackup(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XSH_CONFIG_DIR", root)
	paths, err := GetPaths()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.ConfigFile, []byte("not = [toml"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig(); err == nil {
		t.Fatal("LoadConfig() accepted corrupt TOML")
	}
	if _, err := os.Stat(paths.ConfigFile + ".bak"); err != nil {
		t.Fatalf("corrupt source backup missing: %v", err)
	}
	data, err := os.ReadFile(paths.ConfigFile)
	if err != nil || string(data) != "not = [toml" {
		t.Fatalf("corrupt source changed: %q, %v", data, err)
	}
}

func mustFileMode(t *testing.T, path string) os.FileMode {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.Mode()
}
