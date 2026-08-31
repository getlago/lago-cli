package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/permissions"
	"github.com/pelletier/go-toml/v2"
)

const (
	CurrentVersion = 1
	RegionUS       = "us"
	RegionEU       = "eu"
	RegionSelf     = "self-hosted"
	ModeLive       = "live"
	ModeTest       = "test"
)

type File struct {
	Version         int                 `toml:"version"`
	CurrentProfile  string              `toml:"current_profile,omitempty"`
	UpdateCheck     bool                `toml:"update_check"`
	UpdateConsent   bool                `toml:"update_consent"`
	Channel         string              `toml:"channel,omitempty"`
	LastUpdateCheck string              `toml:"last_update_check,omitempty"`
	LatestVersion   string              `toml:"latest_version,omitempty"`
	Aliases         map[string][]string `toml:"aliases,omitempty"`
	Profiles        map[string]Profile  `toml:"profiles,omitempty"`
}

type Profile struct {
	Region         string `toml:"region"`
	APIURL         string `toml:"api_url"`
	AppURL         string `toml:"app_url,omitempty"`
	APIKey         string `toml:"api_key"`
	Mode           string `toml:"mode"`
	Timeout        string `toml:"timeout,omitempty"`
	Insecure       bool   `toml:"insecure,omitempty"`
	OrganizationID string `toml:"organization_id,omitempty"`
	Organization   string `toml:"organization,omitempty"`
}

type Overrides struct {
	Profile  string
	APIURL   string
	APIKey   string
	Mode     string
	Timeout  time.Duration
	Insecure bool

	APIURLSet   bool
	APIKeySet   bool
	ModeSet     bool
	TimeoutSet  bool
	InsecureSet bool
}

type Resolved struct {
	Name       string
	Profile    Profile
	Timeout    time.Duration
	FromConfig bool
}

func DefaultPath() (string, error) {
	if explicit := os.Getenv("LAGO_CONFIG_FILE"); explicit != "" {
		return explicit, nil
	}
	if runtime.GOOS == "windows" {
		base, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(base, "lago", "config.toml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "lago", "config.toml"), nil
}

func NewFile() File {
	return File{Version: CurrentVersion, Channel: "stable", Aliases: map[string][]string{}, Profiles: map[string]Profile{}}
}

func Load(path string) (File, error) {
	// #nosec G304 -- the config location is explicitly user-selectable through LAGO_CONFIG_FILE.
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return NewFile(), nil
	}
	if err != nil {
		return File{}, fmt.Errorf("read configuration: %w", err)
	}
	var cfg File
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return File{}, fmt.Errorf("parse configuration: %w", err)
	}
	if cfg.Version == 0 {
		cfg.Version = CurrentVersion
	}
	if cfg.Version > CurrentVersion {
		return File{}, fmt.Errorf("configuration version %d requires a newer Lago CLI", cfg.Version)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]Profile{}
	}
	if cfg.Aliases == nil {
		cfg.Aliases = map[string][]string{}
	}
	return cfg, nil
}

func Save(path string, cfg File) error {
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to replace symlinked configuration file %s", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create configuration directory: %w", err)
	}
	if err := permissions.SecureDirectory(filepath.Dir(path)); err != nil {
		return fmt.Errorf("secure configuration directory: %w", err)
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode configuration: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".config-*")
	if err != nil {
		return fmt.Errorf("create temporary configuration: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := permissions.SecureFile(tmpName); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("secure temporary configuration: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write configuration: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("sync configuration: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close configuration: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace configuration: %w", err)
	}
	if err := permissions.SecureFile(path); err != nil {
		return fmt.Errorf("secure configuration: %w", err)
	}
	return nil
}

func Resolve(cfg File, o Overrides) (Resolved, error) {
	name := firstNonEmpty(o.Profile, os.Getenv("LAGO_PROFILE"), cfg.CurrentProfile, "default")
	profile, exists := cfg.Profiles[name]
	resolved := Resolved{Name: name, Profile: profile, FromConfig: exists}

	envURL, envKey, envMode := os.Getenv("LAGO_API_URL"), os.Getenv("LAGO_API_KEY"), os.Getenv("LAGO_MODE")
	if envURL != "" {
		resolved.Profile.APIURL = envURL
	}
	if envKey != "" {
		resolved.Profile.APIKey = envKey
	}
	if envMode != "" {
		resolved.Profile.Mode = envMode
	}
	if o.APIURLSet {
		resolved.Profile.APIURL = o.APIURL
	}
	if o.APIKeySet {
		resolved.Profile.APIKey = o.APIKey
	}
	if o.ModeSet {
		resolved.Profile.Mode = o.Mode
	}
	if o.InsecureSet {
		resolved.Profile.Insecure = o.Insecure
	}

	credentialOverride := envURL != "" || envKey != "" || o.APIURLSet || o.APIKeySet
	explicitMode := envMode != "" || o.ModeSet
	if credentialOverride && !explicitMode {
		resolved.Profile.Mode = ModeLive
	}
	if resolved.Profile.Mode == "" {
		resolved.Profile.Mode = ModeLive
	}
	if resolved.Profile.Mode != ModeLive && resolved.Profile.Mode != ModeTest {
		return Resolved{}, apperr.New(apperr.ExitUsage, "mode must be live or test", "Pass --mode live or --mode test.")
	}

	resolved.Timeout = 30 * time.Second
	if profile.Timeout != "" {
		parsed, err := time.ParseDuration(profile.Timeout)
		if err != nil {
			return Resolved{}, fmt.Errorf("invalid timeout in profile %q: %w", name, err)
		}
		resolved.Timeout = parsed
	}
	if envTimeout := os.Getenv("LAGO_TIMEOUT"); envTimeout != "" {
		parsed, err := time.ParseDuration(envTimeout)
		if err != nil {
			return Resolved{}, apperr.New(apperr.ExitUsage, "invalid LAGO_TIMEOUT", "Use a duration such as 30s or 2m.")
		}
		resolved.Timeout = parsed
	}
	if o.TimeoutSet {
		resolved.Timeout = o.Timeout
	}
	if resolved.Timeout <= 0 {
		return Resolved{}, apperr.New(apperr.ExitUsage, "timeout must be greater than zero", "Use a duration such as --timeout 30s.")
	}
	return resolved, nil
}

func FileMode(path string) (fs.FileMode, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Mode().Perm(), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
