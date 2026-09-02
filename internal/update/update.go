package update

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
		return Check{}, Release{}, releaseAPIError(response.StatusCode, response.Status)
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

// releaseAPIError classifies a non-200 answer from the release metadata endpoint.
//
// The endpoint is GitHub, not Lago, so its failures are never ExitServer: that code is
// documented as "Lago server 5xx error" and a script reading it would conclude Lago is
// down when only the update check failed. Every failure to fetch release metadata is a
// network-class error (ExitNetwork), with a suggestion that names the likely cause.
func releaseAPIError(statusCode int, status string) *apperr.Error {
	suggestion := "Retry later, or upgrade with the command that matches your install: `brew upgrade getlago/tap/lago` or `go install github.com/getlago/lago-cli/cmd/lago@latest`."
	switch statusCode {
	case http.StatusNotFound:
		suggestion = "No published release was found. The repository may be private or have no release yet; upgrade with `brew upgrade getlago/tap/lago` or `go install github.com/getlago/lago-cli/cmd/lago@latest`."
	case http.StatusForbidden, http.StatusTooManyRequests:
		suggestion = "GitHub refused or rate-limited the request, often because of a proxy or too many unauthenticated calls. Retry later, or upgrade with `brew upgrade getlago/tap/lago` or `go install github.com/getlago/lago-cli/cmd/lago@latest`."
	}
	return &apperr.Error{ExitCode: apperr.ExitNetwork, Status: statusCode, Message: "GitHub release API returned " + status, Suggestion: suggestion}
}

// IsDevelopment reports whether a version string identifies a build that no release
// channel produced: `dev`, a commit hash, a local `VERSION=` override. Such a binary was
// built from source, so there is no release to compare it against and nothing to fetch.
func IsDevelopment(version string) bool {
	return !semver.IsValid(normalizedVersion(version))
}

// Method is how the running binary was installed, which determines the only correct
// upgrade command to print.
type Method string

const (
	Homebrew  Method = "homebrew"
	GoInstall Method = "go-install"
	Unknown   Method = "unknown"
)

// UpgradeCommand reports how the running binary was installed and the exact command
// that upgrades it.
//
// Lago CLI ships through two channels, Homebrew and `go install`, and neither is
// self-updating: Homebrew owns its Cellar and `go install` rebuilds from source. So
// `lago upgrade` prints a command instead of replacing the binary. The download,
// checksum-verify and atomic-replace path this replaced belonged to the parked script
// installer; see dist-channels/parked/README.md.
func UpgradeCommand() (Method, string, error) {
	executable, err := os.Executable()
	if err != nil {
		return Unknown, "", apperr.Wrap(apperr.ExitGeneral, "locate Lago CLI executable", err)
	}
	if resolved, err := filepath.EvalSymlinks(executable); err == nil {
		executable = resolved
	}
	method := Detect(executable)
	switch method {
	case Homebrew:
		return method, "brew upgrade getlago/tap/lago", nil
	case GoInstall:
		return method, "go install github.com/getlago/lago-cli/cmd/lago@latest", nil
	default:
		return method, "", nil
	}
}

// Detect classifies an executable path into the channel that installed it.
//
// Homebrew is identified by its Cellar or its prefix; `go install` by GOBIN, GOPATH/bin,
// or a path ending in go/bin. Anything else is Unknown, and Unknown prints both commands
// rather than guessing: telling someone to run `brew upgrade` on a binary Homebrew does
// not own produces a confusing Homebrew error instead of an upgrade.
func Detect(executable string) Method {
	path := filepath.ToSlash(executable)
	lower := strings.ToLower(path)
	if strings.Contains(lower, "/cellar/") || strings.Contains(lower, "/homebrew/") {
		return Homebrew
	}
	directory := filepath.ToSlash(filepath.Dir(path))
	for _, candidate := range []string{os.Getenv("GOBIN"), goPathBin()} {
		if candidate == "" {
			continue
		}
		if directory == filepath.ToSlash(filepath.Clean(candidate)) {
			return GoInstall
		}
	}
	if strings.HasSuffix(directory, "/go/bin") {
		return GoInstall
	}
	return Unknown
}

func goPathBin() string {
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		// GOPATH may be a list; only its first element receives `go install` output.
		first := strings.Split(gopath, string(os.PathListSeparator))[0]
		if first != "" {
			return filepath.Join(first, "bin")
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "go", "bin")
}

func normalizedVersion(version string) string {
	version = strings.TrimSpace(version)
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	return version
}
