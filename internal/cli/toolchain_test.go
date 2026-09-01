package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/mod/semver"
)

// minimumGoVersion is the toolchain floor this repository builds on. It is asserted
// here, not just declared in go.mod, so that a bump is a reviewed change in two places
// rather than an edit nobody notices, and so a build on an older toolchain fails with
// a sentence instead of an obscure compile error.
//
// Bumping it means bumping go.mod, the Dockerfile, and the README together, and
// re-running the security gates: gosec in particular pins to a release that can read
// the new toolchain's export data.
const minimumGoVersion = "1.27.0"

func TestGoModDeclaresTheSupportedToolchain(t *testing.T) {
	t.Parallel()
	// #nosec G304 -- the repository's own go.mod, at a path fixed by this test.
	data, err := os.ReadFile(filepath.Join("..", "..", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	declared := regexp.MustCompile(`(?m)^go (\S+)$`).FindStringSubmatch(string(data))
	if declared == nil {
		t.Fatal("go.mod declares no go directive")
	}
	if declared[1] != minimumGoVersion {
		t.Fatalf("go.mod declares go %s but this repository supports %s; bump both together, with the Dockerfile and README",
			declared[1], minimumGoVersion)
	}
}

// The toolchain actually compiling this test must satisfy the floor. go.mod enforces
// it for the build, and this catches a CI runner pinned to something older by hand.
func TestBuildingToolchainMeetsTheFloor(t *testing.T) {
	t.Parallel()
	current := "v" + strings.TrimPrefix(runtime.Version(), "go")
	// A development toolchain ("devel ...") carries no comparable version.
	if !semver.IsValid(semver.Canonical(current)) {
		t.Skipf("non-release toolchain %s", runtime.Version())
	}
	if semver.Compare(semver.Canonical(current), "v"+minimumGoVersion) < 0 {
		t.Fatalf("built with %s, which is below the %s floor declared in go.mod", runtime.Version(), minimumGoVersion)
	}
}

// The container build must not drift from go.mod. It is the one place a Go version is
// written as a literal string rather than read from the module file.
func TestDockerfileTracksTheSameToolchain(t *testing.T) {
	t.Parallel()
	// #nosec G304 -- the repository's own Dockerfile, at a path fixed by this test.
	data, err := os.ReadFile(filepath.Join("..", "..", "Dockerfile"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "golang:"+minimumGoVersion) {
		t.Errorf("Dockerfile does not build on golang:%s; it drifted from go.mod", minimumGoVersion)
	}
}
