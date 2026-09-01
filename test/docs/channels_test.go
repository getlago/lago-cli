// Package docs_test holds the guardrails on what the repository tells users to run.
//
// Documentation is a contract like any other. When a channel is parked, an install
// command left behind in the README is not a stale sentence: it is an instruction to
// run something that does not work.
package docs_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// parkedChannels are the distribution channels removed for 1.0. Each pattern is what a
// reader would actually copy, not the bare product name: the word "docker" is fine in
// prose about somebody's own deployment, `docker run ghcr.io/getlago/lago-cli` is not.
//
// Un-parking a channel means deleting its entry here in the same commit that publishes
// and smoke-tests it. See dist-channels/parked/README.md for the re-enable criteria.
var parkedChannels = []struct {
	channel string
	pattern *regexp.Regexp
}{
	{"shell installer", regexp.MustCompile(`install\.sh`)},
	{"PowerShell installer", regexp.MustCompile(`install\.ps1`)},
	{"Docker image", regexp.MustCompile(`ghcr\.io/getlago|docker\s+run.*lago-cli|dockers_v2|Dockerfile\.release`)},
	{"Scoop", regexp.MustCompile(`(?i)scoop\s+(install|update|bucket)|scoops:|scoop-bucket`)},
	{"Winget", regexp.MustCompile(`(?i)winget\s+(install|upgrade)|winget:|winget-pkgs|Lago\.LagoCLI`)},
}

// documentedSurfaces are the files that tell a user or a release what to run. A parked
// channel appearing in any of them is a live instruction to use something unsupported.
var documentedSurfaces = []string{
	"README.md",
	"CONTRIBUTING.md",
	"ARCHITECTURE.md",
	"SECURITY.md",
	"CHANGELOG.md",
	".goreleaser.yml",
	".github/workflows",
	"docs",
	"scripts",
	"Makefile",
}

func TestNoParkedChannelIsDocumentedOrPublished(t *testing.T) {
	root := repositoryRoot(t)
	for _, surface := range documentedSurfaces {
		for _, path := range filesUnder(t, filepath.Join(root, surface)) {
			// #nosec G304 -- paths are walked from the repository root by this test.
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			relative, _ := filepath.Rel(root, path)
			for _, parked := range parkedChannels {
				if match := parked.pattern.FindString(string(content)); match != "" {
					t.Errorf("%s references the parked %s channel (%q). Remove it, or un-park the channel per dist-channels/parked/README.md.",
						relative, parked.channel, match)
				}
			}
		}
	}
}

// The two supported channels must both be documented. A channel that is supported but
// undocumented fails the same way a documented-but-parked one does: the user cannot
// install the tool the way the project intends.
func TestSupportedChannelsAreDocumented(t *testing.T) {
	// #nosec G304 -- the repository's own README.
	readme, err := os.ReadFile(filepath.Join(repositoryRoot(t), "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"brew install getlago/tap/lago",
		"go install github.com/getlago/lago-cli/cmd/lago@latest",
	} {
		if !strings.Contains(string(readme), required) {
			t.Errorf("README does not document the supported install command %q", required)
		}
	}
}

// The parked files themselves must stay out of the build and off the release. They live
// under dist-channels/parked precisely so nothing references them.
func TestParkedFilesAreNotReferenced(t *testing.T) {
	root := repositoryRoot(t)
	parked := filepath.Join(root, "dist-channels", "parked")
	for _, name := range []string{"install.sh", "install.ps1", "Dockerfile.release", "README.md"} {
		if _, err := os.Stat(filepath.Join(parked, name)); err != nil {
			t.Errorf("dist-channels/parked/%s is missing; parked channel code must stay recoverable: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "scripts", "install.sh")); err == nil {
		t.Error("scripts/install.sh is back in the build tree")
	}
}

// filesUnder returns every regular file at or under target, skipping the parked
// directory, which documents the parked channels on purpose.
func filesUnder(t *testing.T, target string) []string {
	t.Helper()
	info, err := os.Stat(target)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("stat %s: %v", target, err)
	}
	if !info.IsDir() {
		return []string{target}
	}
	var paths []string
	err = filepath.WalkDir(target, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == "parked" {
				return filepath.SkipDir
			}
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", target, err)
	}
	return paths
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}
