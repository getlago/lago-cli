//go:build !windows

package permissions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSecureModes(t *testing.T) {
	t.Parallel()
	directory := filepath.Join(t.TempDir(), "private")
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(directory, "config")
	if err := os.WriteFile(file, []byte("fake"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SecureDirectory(directory); err != nil {
		t.Fatal(err)
	}
	if err := SecureFile(file); err != nil {
		t.Fatal(err)
	}
	directoryInfo, _ := os.Stat(directory)
	fileInfo, _ := os.Stat(file)
	if directoryInfo.Mode().Perm() != 0o700 || fileInfo.Mode().Perm() != 0o600 {
		t.Fatalf("directory=%o file=%o", directoryInfo.Mode().Perm(), fileInfo.Mode().Perm())
	}
}
