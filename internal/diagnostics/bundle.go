package diagnostics

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/getlago/lago-cli/internal/permissions"
)

type Data struct {
	Version     string
	SpecVersion string
	SpecSHA256  string
	Region      string
	Mode        string
	APIURL      string
	Checks      map[string]bool
}

func WriteBundle(path string, data Data) error {
	files := map[string]any{
		"version.json": map[string]any{
			"cli_version":  data.Version,
			"spec_version": data.SpecVersion,
			"spec_sha256":  data.SpecSHA256,
		},
		"system.json": map[string]any{
			"os":           runtime.GOOS,
			"architecture": runtime.GOARCH,
			"go_version":   runtime.Version(),
		},
		"profile.json": map[string]any{
			"region":  data.Region,
			"mode":    data.Mode,
			"api_url": sanitizedURL(data.APIURL),
		},
		"checks.json": data.Checks,
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create diagnostics directory: %w", err)
	}
	if err := permissions.SecureDirectory(filepath.Dir(path)); err != nil {
		return fmt.Errorf("secure diagnostics directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".lago-diagnostics-*")
	if err != nil {
		return fmt.Errorf("create diagnostics archive: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := permissions.SecureFile(temporaryPath); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("secure diagnostics archive: %w", err)
	}
	gzipWriter := gzip.NewWriter(temporary)
	tarWriter := tar.NewWriter(gzipWriter)
	for _, name := range []string{"version.json", "system.json", "profile.json", "checks.json"} {
		content, err := json.MarshalIndent(files[name], "", "  ")
		if err != nil {
			return closeWithError(temporary, gzipWriter, tarWriter, err)
		}
		content = append(content, '\n')
		header := &tar.Header{Name: name, Mode: 0o600, Size: int64(len(content)), ModTime: time.Unix(0, 0).UTC()}
		if err := tarWriter.WriteHeader(header); err != nil {
			return closeWithError(temporary, gzipWriter, tarWriter, err)
		}
		if _, err := tarWriter.Write(content); err != nil {
			return closeWithError(temporary, gzipWriter, tarWriter, err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		_ = gzipWriter.Close()
		_ = temporary.Close()
		return err
	}
	if err := gzipWriter.Close(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace diagnostics archive: %w", err)
	}
	return permissions.SecureFile(path)
}

func sanitizedURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" {
		return "<redacted>"
	}
	return parsed.Scheme + "://<redacted-host>" + parsed.EscapedPath()
}

func closeWithError(file *os.File, gzipWriter *gzip.Writer, tarWriter *tar.Writer, original error) error {
	_ = tarWriter.Close()
	_ = gzipWriter.Close()
	_ = file.Close()
	return original
}
