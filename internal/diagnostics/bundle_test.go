package diagnostics

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBundleContainsOnlyAllowlistedSanitizedData(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "bundle.tar.gz")
	data := Data{Version: "1.0.0", SpecVersion: "1.52.1", SpecSHA256: "fake-sha", Region: "self-hosted", Mode: "test", APIURL: "https://private.example.test/api/v1", Checks: map[string]bool{"api": false}}
	if err := WriteBundle(path, data); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Fatalf("mode=%v", info.Mode().Perm())
		}
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	archive := tar.NewReader(gzipReader)
	var contents strings.Builder
	for {
		_, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(&contents, archive)
	}
	if strings.Contains(contents.String(), "private.example.test") || !strings.Contains(contents.String(), "redacted-host") {
		t.Fatalf("bundle was not sanitized: %s", contents.String())
	}
}
