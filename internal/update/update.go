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
	suggestion := "Retry later, or upgrade with the command that matches your install: `brew upgrade getlago/tap/lago`, `go install github.com/getlago/lago-cli/cmd/lago@latest`, or re-run the installer from https://getlago.github.io/lago-cli/install.sh."
	switch statusCode {
	case http.StatusNotFound:
		suggestion = "No published release was found. The repository may be private or have no release yet; upgrade with `brew upgrade getlago/tap/lago`, `go install github.com/getlago/lago-cli/cmd/lago@latest`, or re-run the installer from https://getlago.github.io/lago-cli/install.sh."
	case http.StatusForbidden, http.StatusTooManyRequests:
		suggestion = "GitHub refused or rate-limited the request, often because of a proxy or too many unauthenticated calls. Retry later, or upgrade with `brew upgrade getlago/tap/lago`, `go install github.com/getlago/lago-cli/cmd/lago@latest`, or re-run the installer from https://getlago.github.io/lago-cli/install.sh."
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
	Script    Method = "script"
	Unknown   Method = "unknown"
)

// Commands are the upgrade commands per channel, in the order they are documented.
const (
	HomebrewCommand  = "brew upgrade getlago/tap/lago"
	GoInstallCommand = "go install github.com/getlago/lago-cli/cmd/lago@latest"
	ScriptCommand    = "curl -fsSL https://getlago.github.io/lago-cli/install.sh | sh"
)

// UpgradeCommand reports how the running binary was installed and the exact command
// that upgrades it.
//
// Lago CLI ships through three channels, Homebrew, `go install` and the shell
// installer, and none is self-updating: Homebrew owns its Cellar, `go install` rebuilds
// from source, and the installer is idempotent, so re-running it is the upgrade. So
// `lago upgrade` prints a command instead of replacing the binary; the binary never
// downloads and swaps itself. See DECISIONS.md, "Shell installer, hosted on GitHub Pages".
func UpgradeCommand() (Method, string, error) {
	executable, err := os.Executable()
	if err != nil {
		return Unknown, "", apperr.Wrap(apperr.ExitGeneral, "locate Lago CLI executable", err)
	}
	if resolved, err := filepath.EvalSymlinks(executable); err == nil {
		executable = resolved
	}
	method := Detect(executable)
	return method, CommandFor(method, filepath.Dir(executable)), nil
}

// CommandFor is the upgrade command for a binary installed by method into directory.
//
// A script install outside the installer's default directory gets the same line with
// LAGO_INSTALL_DIR set, so re-running it replaces the binary that is actually on the
// PATH instead of leaving a second copy in /usr/local/bin. Unknown yields no command so
// the caller knows to print them all.
func CommandFor(method Method, directory string) string {
	switch method {
	case Homebrew:
		return HomebrewCommand
	case GoInstall:
		return GoInstallCommand
	case Script:
		if filepath.ToSlash(filepath.Clean(directory)) == defaultScriptInstallDir {
			return ScriptCommand
		}
		return strings.Replace(ScriptCommand, "| sh", "| LAGO_INSTALL_DIR="+shellQuote(directory)+" sh", 1)
	default:
		return ""
	}
}

// shellQuote single-quotes a path when it contains anything a shell would interpret,
// so a directory with a space in it survives a paste into a terminal.
func shellQuote(value string) string {
	if strings.IndexFunc(value, func(r rune) bool {
		return !(r == '/' || r == '.' || r == '_' || r == '-' || r == '~' ||
			(r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'))
	}) < 0 {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

// Detect classifies an executable path into the channel that installed it.
//
// Homebrew is identified by its Cellar or its prefix; `go install` by GOBIN, GOPATH/bin,
// or a path ending in go/bin; the shell installer by its default /usr/local/bin, the
// LAGO_INSTALL_DIR override, or ~/.local/bin. Homebrew is checked first because on
// Intel macOS its prefix is /usr/local, and the executable path is symlink-resolved
// before it gets here, so a brew-linked /usr/local/bin/lago reads as its Cellar path.
// Anything else is Unknown, and Unknown prints every command rather than guessing:
// telling someone to run `brew upgrade` on a binary Homebrew does not own produces a
// confusing Homebrew error instead of an upgrade.
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
	for _, candidate := range scriptInstallDirs() {
		if directory == filepath.ToSlash(filepath.Clean(candidate)) {
			return Script
		}
	}
	return Unknown
}

// scriptInstallDirs are the directories install.sh installs into: its default, the
// override it honours, and the directory its error message tells a user without sudo
// to pick. A binary copied there by hand from a release archive is upgraded the same
// way, by re-running the installer, so the heuristic is right for that case too.
const defaultScriptInstallDir = "/usr/local/bin"

func scriptInstallDirs() []string {
	dirs := []string{defaultScriptInstallDir}
	if override := os.Getenv("LAGO_INSTALL_DIR"); override != "" {
		dirs = append(dirs, override)
	}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".local", "bin"))
	}
	return dirs
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
