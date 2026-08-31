package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestSaveLoadAndPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.toml")
	cfg := NewFile()
	cfg.CurrentProfile = "test"
	cfg.Profiles["test"] = Profile{Region: RegionUS, APIURL: "https://api.getlago.com/api/v1", APIKey: "lago_test_FAKE000000000000000000000000", Mode: ModeTest}
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Profiles["test"].APIKey != cfg.Profiles["test"].APIKey {
		t.Fatal("saved profile did not round trip")
	}
	if runtime.GOOS != "windows" {
		mode, err := FileMode(path)
		if err != nil {
			t.Fatal(err)
		}
		if mode != 0o600 {
			t.Fatalf("config mode = %o, want 600", mode)
		}
	}
}

func TestSaveRefusesSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation may require elevated Windows privileges")
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	if err := os.WriteFile(target, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "config.toml")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := Save(link, NewFile()); err == nil {
		t.Fatal("Save unexpectedly replaced a symlink")
	}
}

func TestResolvePrecedenceAndUnknownCredentialsFailLive(t *testing.T) {
	t.Setenv("LAGO_PROFILE", "")
	t.Setenv("LAGO_API_URL", "https://env.example.test")
	t.Setenv("LAGO_API_KEY", "env-key")
	t.Setenv("LAGO_MODE", "")
	t.Setenv("LAGO_TIMEOUT", "")
	cfg := NewFile()
	cfg.CurrentProfile = "staging"
	cfg.Profiles["staging"] = Profile{APIURL: "https://config.example.test", APIKey: "config-key", Mode: ModeTest, Timeout: "12s"}
	resolved, err := Resolve(cfg, Overrides{APIURL: "https://flag.example.test", APIKey: "flag-key", APIURLSet: true, APIKeySet: true})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Profile.APIURL != "https://flag.example.test" || resolved.Profile.APIKey != "flag-key" {
		t.Fatalf("flags did not win: %#v", resolved.Profile)
	}
	if resolved.Profile.Mode != ModeLive {
		t.Fatalf("unknown credential override mode = %q, want live", resolved.Profile.Mode)
	}
	if resolved.Timeout != 12*time.Second {
		t.Fatalf("timeout = %s", resolved.Timeout)
	}
}
