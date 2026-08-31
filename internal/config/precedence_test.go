package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func clearEnvironment(t *testing.T) {
	t.Helper()
	for _, name := range []string{"LAGO_API_KEY", "LAGO_API_URL", "LAGO_MODE", "LAGO_PROFILE", "LAGO_TIMEOUT", "LAGO_CONFIG_FILE"} {
		t.Setenv(name, "")
	}
}

func twoProfileFile() File {
	file := NewFile()
	file.CurrentProfile = "staging"
	file.Profiles["staging"] = Profile{Region: "self-hosted", APIURL: "https://staging.example.com/api/v1", APIKey: "key-staging", Mode: ModeTest}
	file.Profiles["production"] = Profile{Region: RegionUS, APIURL: "https://api.example.com/api/v1", APIKey: "key-production", Mode: ModeLive, Timeout: "45s"}
	return file
}

// Precedence is flags > env > profile > defaults. Each layer is asserted separately
// so a regression names the layer that broke.
func TestResolvePrecedenceLayerByLayer(t *testing.T) {
	file := twoProfileFile()

	t.Run("profile beats defaults", func(t *testing.T) {
		clearEnvironment(t)
		resolved, err := Resolve(file, Overrides{})
		if err != nil {
			t.Fatal(err)
		}
		if resolved.Name != "staging" || resolved.Profile.APIKey != "key-staging" || resolved.Profile.Mode != ModeTest {
			t.Fatalf("current_profile ignored: %+v", resolved)
		}
	})

	t.Run("env beats profile", func(t *testing.T) {
		clearEnvironment(t)
		t.Setenv("LAGO_PROFILE", "production")
		t.Setenv("LAGO_MODE", ModeTest)
		resolved, err := Resolve(file, Overrides{})
		if err != nil {
			t.Fatal(err)
		}
		if resolved.Name != "production" || resolved.Profile.Mode != ModeTest {
			t.Fatalf("env did not override the profile: %+v", resolved)
		}
	})

	t.Run("flags beat env", func(t *testing.T) {
		clearEnvironment(t)
		t.Setenv("LAGO_API_URL", "https://env.example.com/api/v1")
		t.Setenv("LAGO_MODE", ModeLive)
		resolved, err := Resolve(file, Overrides{
			APIURL: "https://flag.example.com/api/v1", APIURLSet: true,
			Mode: ModeTest, ModeSet: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		if resolved.Profile.APIURL != "https://flag.example.com/api/v1" || resolved.Profile.Mode != ModeTest {
			t.Fatalf("flags did not override the environment: %+v", resolved)
		}
	})

	t.Run("unknown profile still resolves but is marked absent", func(t *testing.T) {
		clearEnvironment(t)
		resolved, err := Resolve(file, Overrides{Profile: "missing"})
		if err != nil {
			t.Fatal(err)
		}
		if resolved.FromConfig {
			t.Error("an unknown profile was reported as configured")
		}
	})
}

// Credentials supplied without an explicit mode must resolve to live. Failing toward
// caution is the recorded decision: the API exposes no authoritative mode field.
func TestCredentialOverrideWithoutModeFailsTowardLive(t *testing.T) {
	file := twoProfileFile()

	for _, testCase := range []struct {
		name      string
		env       map[string]string
		overrides Overrides
		wantMode  string
	}{
		{"env key alone", map[string]string{"LAGO_API_KEY": "key-from-env"}, Overrides{}, ModeLive},
		{"env url alone", map[string]string{"LAGO_API_URL": "https://env.example.com"}, Overrides{}, ModeLive},
		{"flag key alone", nil, Overrides{APIKey: "key-from-flag", APIKeySet: true}, ModeLive},
		{"flag url alone", nil, Overrides{APIURL: "https://flag.example.com", APIURLSet: true}, ModeLive},
		{"env key with explicit test mode", map[string]string{"LAGO_API_KEY": "k", "LAGO_MODE": ModeTest}, Overrides{}, ModeTest},
		{"flag key with explicit test mode", nil, Overrides{APIKey: "k", APIKeySet: true, Mode: ModeTest, ModeSet: true}, ModeTest},
		{"no override keeps the test profile", nil, Overrides{}, ModeTest},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			clearEnvironment(t)
			for name, value := range testCase.env {
				t.Setenv(name, value)
			}
			resolved, err := Resolve(file, testCase.overrides)
			if err != nil {
				t.Fatal(err)
			}
			if resolved.Profile.Mode != testCase.wantMode {
				t.Fatalf("mode = %q, want %q", resolved.Profile.Mode, testCase.wantMode)
			}
		})
	}
}

func TestModeMustBeLiveOrTest(t *testing.T) {
	clearEnvironment(t)
	file := NewFile()
	file.Profiles["default"] = Profile{Mode: "sandbox"}
	if _, err := Resolve(file, Overrides{}); err == nil {
		t.Fatal("an unknown mode was accepted")
	} else if !strings.Contains(err.Error(), "live or test") {
		t.Errorf("mode error does not name the valid choices: %v", err)
	}
}

// Timeout resolution follows the same precedence and rejects unusable values.
func TestTimeoutPrecedenceAndValidation(t *testing.T) {
	file := twoProfileFile()

	t.Run("profile timeout is used", func(t *testing.T) {
		clearEnvironment(t)
		resolved, err := Resolve(file, Overrides{Profile: "production"})
		if err != nil {
			t.Fatal(err)
		}
		if resolved.Timeout != 45*time.Second {
			t.Fatalf("timeout = %s, want 45s", resolved.Timeout)
		}
	})

	t.Run("default is thirty seconds", func(t *testing.T) {
		clearEnvironment(t)
		resolved, err := Resolve(file, Overrides{Profile: "staging"})
		if err != nil {
			t.Fatal(err)
		}
		if resolved.Timeout != 30*time.Second {
			t.Fatalf("timeout = %s, want 30s", resolved.Timeout)
		}
	})

	t.Run("env overrides the profile", func(t *testing.T) {
		clearEnvironment(t)
		t.Setenv("LAGO_TIMEOUT", "5s")
		resolved, err := Resolve(file, Overrides{Profile: "production"})
		if err != nil {
			t.Fatal(err)
		}
		if resolved.Timeout != 5*time.Second {
			t.Fatalf("timeout = %s, want 5s", resolved.Timeout)
		}
	})

	t.Run("flag overrides the environment", func(t *testing.T) {
		clearEnvironment(t)
		t.Setenv("LAGO_TIMEOUT", "5s")
		resolved, err := Resolve(file, Overrides{Timeout: 2 * time.Second, TimeoutSet: true})
		if err != nil {
			t.Fatal(err)
		}
		if resolved.Timeout != 2*time.Second {
			t.Fatalf("timeout = %s, want 2s", resolved.Timeout)
		}
	})

	t.Run("invalid values are rejected", func(t *testing.T) {
		clearEnvironment(t)
		t.Setenv("LAGO_TIMEOUT", "soon")
		if _, err := Resolve(file, Overrides{}); err == nil {
			t.Error("an unparseable LAGO_TIMEOUT was accepted")
		}

		clearEnvironment(t)
		broken := NewFile()
		broken.Profiles["default"] = Profile{Mode: ModeTest, Timeout: "later"}
		if _, err := Resolve(broken, Overrides{}); err == nil {
			t.Error("an unparseable profile timeout was accepted")
		}

		clearEnvironment(t)
		if _, err := Resolve(file, Overrides{Timeout: 0, TimeoutSet: true}); err == nil {
			t.Error("a zero timeout was accepted")
		}
		if _, err := Resolve(file, Overrides{Timeout: -time.Second, TimeoutSet: true}); err == nil {
			t.Error("a negative timeout was accepted")
		}
	})
}

// A config the CLI cannot understand must fail loudly rather than silently losing
// profiles: an unreadable profile means credentials go missing at request time.
func TestLoadRejectsUnreadableConfigurations(t *testing.T) {
	directory := t.TempDir()

	t.Run("missing file yields defaults", func(t *testing.T) {
		loaded, err := Load(filepath.Join(directory, "absent.toml"))
		if err != nil {
			t.Fatalf("a missing config must not be an error: %v", err)
		}
		if loaded.Version != CurrentVersion || loaded.Profiles == nil || loaded.Aliases == nil {
			t.Fatalf("defaults are not initialised: %+v", loaded)
		}
	})

	t.Run("malformed TOML is an error", func(t *testing.T) {
		path := filepath.Join(directory, "broken.toml")
		if err := os.WriteFile(path, []byte("this is not = = toml"), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(path); err == nil {
			t.Fatal("malformed TOML was accepted")
		}
	})

	t.Run("a newer config version is refused", func(t *testing.T) {
		path := filepath.Join(directory, "future.toml")
		if err := os.WriteFile(path, []byte("version = 999\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := Load(path)
		if err == nil {
			t.Fatal("a config from a newer CLI was accepted")
		}
		if !strings.Contains(err.Error(), "newer Lago CLI") {
			t.Errorf("version error does not explain the cause: %v", err)
		}
	})

	t.Run("a version-less config is upgraded in memory", func(t *testing.T) {
		path := filepath.Join(directory, "legacy.toml")
		if err := os.WriteFile(path, []byte("current_profile = \"default\"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		loaded, err := Load(path)
		if err != nil {
			t.Fatal(err)
		}
		if loaded.Version != CurrentVersion {
			t.Fatalf("version = %d, want %d", loaded.Version, CurrentVersion)
		}
	})
}

// The config holds API keys, so it is created 0600 inside a 0700 directory and a
// round trip must not widen those permissions or lose a profile.
func TestSaveRoundTripsAndKeepsPermissionsTight(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.toml")
	original := twoProfileFile()
	original.Aliases = map[string][]string{"ls": {"customers", "list"}}

	if err := Save(path, original); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Profiles) != 2 || loaded.Profiles["production"].APIKey != "key-production" {
		t.Fatalf("profiles lost in the round trip: %+v", loaded.Profiles)
	}
	if loaded.CurrentProfile != "staging" || len(loaded.Aliases["ls"]) != 2 {
		t.Fatalf("metadata lost in the round trip: %+v", loaded)
	}

	if runtime.GOOS != "windows" {
		mode, err := FileMode(path)
		if err != nil || mode != 0o600 {
			t.Fatalf("config mode = %o (err %v), want 0600", mode, err)
		}
		info, err := os.Stat(filepath.Dir(path))
		if err != nil || info.Mode().Perm() != 0o700 {
			t.Fatalf("config directory mode = %o, want 0700", info.Mode().Perm())
		}
	}

	// Overwriting must preserve the tight mode rather than inheriting the umask.
	if err := Save(path, original); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		if mode, _ := FileMode(path); mode != 0o600 {
			t.Fatalf("mode after overwrite = %o, want 0600", mode)
		}
	}
}

func TestFileModeReportsMissingFiles(t *testing.T) {
	t.Parallel()
	if _, err := FileMode(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Fatal("FileMode did not report a missing file")
	}
}

// LAGO_CONFIG_FILE must win over the platform default so tests and CI can point at
// a throwaway config without touching the operator's real credentials.
func TestDefaultPathHonoursExplicitOverride(t *testing.T) {
	clearEnvironment(t)
	t.Setenv("LAGO_CONFIG_FILE", "/tmp/explicit-lago.toml")
	path, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	if path != "/tmp/explicit-lago.toml" {
		t.Fatalf("DefaultPath = %q, want the explicit override", path)
	}

	t.Setenv("LAGO_CONFIG_FILE", "")
	fallback, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(filepath.ToSlash(fallback), "lago/config.toml") {
		t.Fatalf("default path = %q, want it to end in lago/config.toml", fallback)
	}
	if !filepath.IsAbs(fallback) {
		t.Fatalf("default path %q is not absolute", fallback)
	}
}

func TestFirstNonEmptySkipsBlankValues(t *testing.T) {
	t.Parallel()
	if got := firstNonEmpty("", "   ", "chosen", "later"); got != "chosen" {
		t.Fatalf("firstNonEmpty = %q", got)
	}
	if got := firstNonEmpty("", "  "); got != "" {
		t.Fatalf("firstNonEmpty = %q, want empty", got)
	}
}

// Save must fail rather than leave a partial or world-readable config behind. Each
// case blocks a different step of the atomic write.
func TestSaveFailsClosedOnUnwritableTargets(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("permission-based failures need a non-root POSIX environment")
	}

	t.Run("parent path is a file, not a directory", func(t *testing.T) {
		blocker := filepath.Join(t.TempDir(), "blocker")
		if err := os.WriteFile(blocker, []byte("not a directory"), 0o600); err != nil {
			t.Fatal(err)
		}
		err := Save(filepath.Join(blocker, "config.toml"), NewFile())
		if err == nil {
			t.Fatal("Save succeeded with a file in place of the config directory")
		}
		if !strings.Contains(err.Error(), "configuration directory") {
			t.Errorf("error does not name the failing step: %v", err)
		}
	})

	t.Run("temporary file cannot be created", func(t *testing.T) {
		// SecureDirectory repairs the mode before the write, so removing write
		// permission is not enough; make the path itself unusable instead.
		directory := filepath.Join(t.TempDir(), "sub")
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
		nested := filepath.Join(directory, "config.toml", "deeper", "config.toml")
		if err := os.WriteFile(filepath.Join(directory, "config.toml"), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := Save(nested, NewFile()); err == nil {
			t.Fatal("Save succeeded through a regular file")
		}
	})
}

// A config directory left group- or world-readable is repaired on the next Save
// rather than being written into as found. API keys live here.
func TestSaveRepairsLoosePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX modes only")
	}
	directory := filepath.Join(t.TempDir(), "loose")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "config.toml")
	if err := os.WriteFile(path, []byte("version = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Save(path, twoProfileFile()); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(directory)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Errorf("directory mode = %o after Save, want 0700", info.Mode().Perm())
	}
	mode, err := FileMode(path)
	if err != nil {
		t.Fatal(err)
	}
	if mode != 0o600 {
		t.Errorf("config mode = %o after Save, want 0600", mode)
	}
}

// A symlinked config is a redirection attack: following it would write the operator's
// API keys wherever the link points.
func TestSaveRefusesToFollowASymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation is restricted on Windows")
	}
	directory := t.TempDir()
	target := filepath.Join(directory, "elsewhere.toml")
	link := filepath.Join(directory, "config.toml")
	if err := os.WriteFile(target, []byte("version = 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	err := Save(link, twoProfileFile())
	if err == nil {
		t.Fatal("Save followed a symlink")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Errorf("error does not name the cause: %v", err)
	}
	contents, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if strings.Contains(string(contents), "key-production") {
		t.Fatal("Save wrote credentials through the symlink")
	}
}

// The temporary file used for the atomic write must never be left behind, and must
// never be readable by anyone but the owner while it exists.
func TestSaveLeavesNoTemporaryFiles(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "config.toml")
	if err := Save(path, twoProfileFile()); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".config-") {
			t.Errorf("Save left a temporary file behind: %s", entry.Name())
		}
	}
	if len(entries) != 1 {
		t.Errorf("expected only config.toml, found %d entries", len(entries))
	}
}

// A config path that cannot be read for a reason other than absence must surface
// the error, not be mistaken for "no config yet" and silently drop credentials.
func TestLoadDistinguishesUnreadableFromAbsent(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	if _, err := Load(directory); err == nil {
		t.Fatal("Load treated a directory as an empty configuration")
	} else if !strings.Contains(err.Error(), "read configuration") {
		t.Errorf("error does not name the failing step: %v", err)
	}
}

// --insecure is a per-profile safety switch, so an explicit flag must be able to
// both set and clear it regardless of what the profile says.
func TestInsecureOverrideAppliesInBothDirections(t *testing.T) {
	clearEnvironment(t)
	file := NewFile()
	file.Profiles["default"] = Profile{Mode: ModeTest, Insecure: true, APIKey: "k", APIURL: "https://example.test"}

	kept, err := Resolve(file, Overrides{})
	if err != nil {
		t.Fatal(err)
	}
	if !kept.Profile.Insecure {
		t.Error("profile insecure flag was dropped")
	}

	cleared, err := Resolve(file, Overrides{Insecure: false, InsecureSet: true})
	if err != nil {
		t.Fatal(err)
	}
	if cleared.Profile.Insecure {
		t.Error("--insecure=false did not override the profile")
	}

	strict := NewFile()
	strict.Profiles["default"] = Profile{Mode: ModeTest}
	enabled, err := Resolve(strict, Overrides{Insecure: true, InsecureSet: true})
	if err != nil {
		t.Fatal(err)
	}
	if !enabled.Profile.Insecure {
		t.Error("--insecure did not enable the flag")
	}
}

// DefaultPath must report an error rather than returning a relative or empty path
// when the home directory cannot be determined.
func TestDefaultPathFailsWhenHomeIsUnknown(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("HOME is not how Windows resolves the config directory")
	}
	clearEnvironment(t)
	t.Setenv("HOME", "")
	if path, err := DefaultPath(); err == nil {
		t.Fatalf("DefaultPath returned %q with no home directory", path)
	}
}
