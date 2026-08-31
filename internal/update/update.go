package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/getlago/lago-cli/internal/apperr"
	"golang.org/x/mod/semver"
)

const DefaultAPIBase = "https://api.github.com/repos/getlago/lago-cli"

type Release struct {
	TagName    string  `json:"tag_name"`
	Prerelease bool    `json:"prerelease"`
	Draft      bool    `json:"draft"`
	HTMLURL    string  `json:"html_url"`
	Assets     []Asset `json:"assets"`
}

type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

type Check struct {
	Current         string `json:"current"`
	Latest          string `json:"latest"`
	Channel         string `json:"channel"`
	UpdateAvailable bool   `json:"update_available"`
	ReleaseURL      string `json:"release_url,omitempty"`
	Development     bool   `json:"development_build,omitempty"`
}

func Latest(ctx context.Context, current, channel, userAgent, apiBase string) (Check, Release, error) {
	if channel != "stable" && channel != "beta" {
		return Check{}, Release{}, apperr.New(apperr.ExitUsage, "update channel must be stable or beta", "Pass --channel stable or --channel beta.")
	}
	if apiBase == "" {
		apiBase = DefaultAPIBase
	}
	client := &http.Client{Timeout: 5 * time.Second}
	endpoint := strings.TrimRight(apiBase, "/") + "/releases/latest"
	if channel == "beta" {
		endpoint = strings.TrimRight(apiBase, "/") + "/releases?per_page=30"
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Check{}, Release{}, apperr.Wrap(apperr.ExitGeneral, "build update request", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Set("User-Agent", userAgent)
	response, err := client.Do(request)
	if err != nil {
		return Check{}, Release{}, &apperr.Error{ExitCode: apperr.ExitNetwork, Message: "check for Lago CLI updates: " + err.Error(), Suggestion: "Check network access or disable passive checks with LAGO_NO_UPDATE_CHECK=1.", Cause: err}
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Check{}, Release{}, &apperr.Error{ExitCode: apperr.ExitServer, Status: response.StatusCode, Message: "GitHub release API returned " + response.Status, Suggestion: "Retry later or upgrade with your package manager."}
	}
	limited := io.LimitReader(response.Body, 4<<20)
	var release Release
	if channel == "stable" {
		if err := json.NewDecoder(limited).Decode(&release); err != nil {
			return Check{}, Release{}, apperr.Wrap(apperr.ExitGeneral, "decode latest Lago CLI release", err)
		}
	} else {
		var releases []Release
		if err := json.NewDecoder(limited).Decode(&releases); err != nil {
			return Check{}, Release{}, apperr.Wrap(apperr.ExitGeneral, "decode Lago CLI beta releases", err)
		}
		for _, candidate := range releases {
			if !candidate.Draft && (candidate.Prerelease || release.TagName == "") {
				if release.TagName == "" || semver.Compare(normalizedVersion(candidate.TagName), normalizedVersion(release.TagName)) > 0 {
					release = candidate
				}
			}
		}
	}
	if release.TagName == "" {
		return Check{}, Release{}, apperr.New(apperr.ExitNotFound, "no release exists in the selected channel", "Try --channel stable or check again later.")
	}
	currentVersion := normalizedVersion(current)
	latestVersion := normalizedVersion(release.TagName)
	development := !semver.IsValid(currentVersion)
	available := !development && semver.Compare(latestVersion, currentVersion) > 0
	return Check{Current: current, Latest: strings.TrimPrefix(release.TagName, "v"), Channel: channel, UpdateAvailable: available, ReleaseURL: release.HTMLURL, Development: development}, release, nil
}

func Install(ctx context.Context, release Release, userAgent string) (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", apperr.Wrap(apperr.ExitGeneral, "locate Lago CLI executable", err)
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return "", apperr.Wrap(apperr.ExitGeneral, "resolve Lago CLI executable", err)
	}
	lowerPath := strings.ToLower(filepath.ToSlash(executable))
	switch {
	case strings.Contains(lowerPath, "/cellar/") || strings.Contains(lowerPath, "/homebrew/"):
		return "", apperr.New(apperr.ExitUsage, "this Lago CLI is managed by Homebrew", "Run `brew upgrade getlago/tap/lago`.")
	case strings.Contains(lowerPath, "/scoop/"):
		return "", apperr.New(apperr.ExitUsage, "this Lago CLI is managed by Scoop", "Run `scoop update lago`.")
	case strings.Contains(lowerPath, "/winget/") || runtime.GOOS == "windows":
		return "", apperr.New(apperr.ExitUsage, "use the Windows package manager to upgrade Lago CLI", "Run `winget upgrade Lago.LagoCLI` or `scoop update lago`.")
	case strings.HasSuffix(filepath.ToSlash(filepath.Dir(executable)), "/go/bin"):
		return "", apperr.New(apperr.ExitUsage, "this Lago CLI appears to be managed by go install", "Run `go install github.com/getlago/lago-cli/cmd/lago@latest`.")
	}
	version := strings.TrimPrefix(release.TagName, "v")
	extension := "tar.gz"
	if runtime.GOOS == "windows" {
		extension = "zip"
	}
	archiveName := fmt.Sprintf("lago_%s_%s_%s.%s", version, runtime.GOOS, runtime.GOARCH, extension)
	archiveURL := assetURL(release.Assets, archiveName)
	checksumsURL := assetURL(release.Assets, "checksums.txt")
	if archiveURL == "" || checksumsURL == "" {
		return "", apperr.New(apperr.ExitNotFound, "release assets for this platform are incomplete", "Use the package-manager install command from the release page.")
	}
	archive, err := download(ctx, archiveURL, userAgent, 128<<20)
	if err != nil {
		return "", err
	}
	checksums, err := download(ctx, checksumsURL, userAgent, 4<<20)
	if err != nil {
		return "", err
	}
	expected := checksumFor(checksums, archiveName)
	digest := sha256.Sum256(archive)
	if expected == "" || !strings.EqualFold(expected, hex.EncodeToString(digest[:])) {
		return "", apperr.New(apperr.ExitValidation, "release checksum verification failed", "Do not install this artifact; retry and report the release through SECURITY.md if it persists.")
	}
	binary, err := extractBinary(archive, extension)
	if err != nil {
		return "", err
	}
	temporary, err := os.CreateTemp(filepath.Dir(executable), ".lago-upgrade-*")
	if err != nil {
		return "", apperr.Wrap(apperr.ExitGeneral, "create upgrade file", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o755); err != nil {
		_ = temporary.Close()
		return "", apperr.Wrap(apperr.ExitGeneral, "make upgraded binary executable", err)
	}
	if _, err := temporary.Write(binary); err != nil {
		_ = temporary.Close()
		return "", apperr.Wrap(apperr.ExitGeneral, "write upgraded binary", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return "", apperr.Wrap(apperr.ExitGeneral, "sync upgraded binary", err)
	}
	if err := temporary.Close(); err != nil {
		return "", apperr.Wrap(apperr.ExitGeneral, "close upgraded binary", err)
	}
	if err := os.Rename(temporaryPath, executable); err != nil {
		return "", apperr.Wrap(apperr.ExitGeneral, "replace Lago CLI executable", err)
	}
	return version, nil
}

func download(ctx context.Context, target, userAgent string, limit int64) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, apperr.Wrap(apperr.ExitGeneral, "build release download", err)
	}
	request.Header.Set("User-Agent", userAgent)
	client := &http.Client{Timeout: 30 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, &apperr.Error{ExitCode: apperr.ExitNetwork, Message: "download Lago CLI release: " + err.Error(), Cause: err}
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, &apperr.Error{ExitCode: apperr.ExitServer, Status: response.StatusCode, Message: "release download returned " + response.Status}
	}
	reader := &io.LimitedReader{R: response.Body, N: limit + 1}
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, apperr.Wrap(apperr.ExitNetwork, "read release download", err)
	}
	if reader.N == 0 {
		return nil, apperr.New(apperr.ExitValidation, "release asset exceeds its safety limit", "Install from the signed release page manually.")
	}
	return data, nil
}

func extractBinary(archive []byte, format string) ([]byte, error) {
	if format == "zip" {
		reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
		if err != nil {
			return nil, apperr.Wrap(apperr.ExitValidation, "open release ZIP", err)
		}
		for _, file := range reader.File {
			if filepath.Base(file.Name) != "lago.exe" {
				continue
			}
			opened, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer opened.Close()
			return io.ReadAll(io.LimitReader(opened, 128<<20))
		}
	} else {
		gzipReader, err := gzip.NewReader(bytes.NewReader(archive))
		if err != nil {
			return nil, apperr.Wrap(apperr.ExitValidation, "open release archive", err)
		}
		defer gzipReader.Close()
		tarReader := tar.NewReader(gzipReader)
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, apperr.Wrap(apperr.ExitValidation, "read release archive", err)
			}
			if filepath.Base(header.Name) == "lago" && header.Typeflag == tar.TypeReg {
				return io.ReadAll(io.LimitReader(tarReader, 128<<20))
			}
		}
	}
	return nil, apperr.New(apperr.ExitValidation, "release archive contains no Lago binary", "Install from the signed release page manually.")
}

func assetURL(assets []Asset, name string) string {
	for _, asset := range assets {
		if asset.Name == name {
			return asset.URL
		}
	}
	return ""
}

func checksumFor(data []byte, name string) string {
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == name {
			return fields[0]
		}
	}
	return ""
}

func normalizedVersion(version string) string {
	version = strings.TrimSpace(version)
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	return version
}
