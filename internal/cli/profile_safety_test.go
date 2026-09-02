package cli

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/config"
)

// organizationServer answers /organizations so `init` can validate credentials.
func organizationServer(t *testing.T) string {
	t.Helper()
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`{"organization":{"lago_id":"org_1","name":"Example Organization"}}`))
	})
	return server.URL
}

func initProfile(t *testing.T, serverURL, profile string, extra ...string) (string, string, error) {
	t.Helper()
	argv := append([]string{"--profile", profile, "--api-url", serverURL, "--api-key", "lago_test_FAKE000000000000000000000000", "--mode", "test", "--insecure", "init", "--region", "self-hosted"}, extra...)
	return execute(t, "", argv...)
}

func loadConfig(t *testing.T, path string) config.File {
	t.Helper()
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

// QA F-13: `init --profile staging` switched current_profile, so every later command
// silently targeted the profile just configured. The first profile becomes current;
// after that, switching is opt-in with --use.
func TestQA_F13_SecondProfileDoesNotBecomeCurrentWithoutUse(t *testing.T) {
	setCleanEnvironment(t)
	path := filepath.Join(t.TempDir(), "config.toml")
	t.Setenv("LAGO_CONFIG_FILE", path)
	url := organizationServer(t)

	if _, _, err := initProfile(t, url, "first"); err != nil {
		t.Fatalf("first init failed: %v", err)
	}
	if got := loadConfig(t, path).CurrentProfile; got != "first" {
		t.Fatalf("the first profile did not become current: %q", got)
	}

	_, stderr, err := initProfile(t, url, "second")
	if err != nil {
		t.Fatalf("second init failed: %v", err)
	}
	cfg := loadConfig(t, path)
	if cfg.CurrentProfile != "first" {
		t.Errorf("current_profile switched to %q without --use", cfg.CurrentProfile)
	}
	if _, ok := cfg.Profiles["second"]; !ok {
		t.Error("second profile was not saved")
	}
	if !strings.Contains(stderr, `Profile "second" saved; "first" remains the current profile`) || !strings.Contains(stderr, "--use") {
		t.Errorf("stderr does not explain the current profile:\n%s", stderr)
	}

	if _, _, err := initProfile(t, url, "second", "--use"); err != nil {
		t.Fatalf("init --use failed: %v", err)
	}
	if got := loadConfig(t, path).CurrentProfile; got != "second" {
		t.Errorf("--use did not switch current_profile: %q", got)
	}
}

// QA C-8, S-5: --insecure was persisted silently, and a re-init without the flag
// silently cleared it. It is announced whenever it ends up true, and kept unless the
// flag is passed again.
func TestQA_C8_InsecureIsAnnouncedAndKept(t *testing.T) {
	setCleanEnvironment(t)
	path := filepath.Join(t.TempDir(), "config.toml")
	t.Setenv("LAGO_CONFIG_FILE", path)
	url := organizationServer(t)

	_, stderr, err := initProfile(t, url, "dev")
	if err != nil {
		t.Fatalf("init --insecure failed: %v", err)
	}
	if !strings.Contains(stderr, `insecure = true is persisted in profile "dev"`) || !strings.Contains(stderr, "--insecure=false") {
		t.Errorf("insecure persistence was not announced:\n%s", stderr)
	}
	if !loadConfig(t, path).Profiles["dev"].Insecure {
		t.Fatal("insecure was not persisted")
	}

	// Re-init without the flag: the stored value is kept, and said so.
	_, stderr, err = execute(t, "", "--profile", "dev", "--api-url", url, "--api-key", "lago_test_FAKE000000000000000000000000", "--mode", "test", "init", "--region", "self-hosted")
	if err != nil {
		t.Fatalf("re-init without --insecure failed: %v", err)
	}
	if !loadConfig(t, path).Profiles["dev"].Insecure {
		t.Error("re-init without the flag cleared insecure")
	}
	if !strings.Contains(stderr, "was kept from the existing profile") {
		t.Errorf("kept insecure value was not announced:\n%s", stderr)
	}
}

// QA S-3: loose config permissions were reported by `lago doctor` only.
func TestQA_S3_LooseConfigPermissionsWarnOnStderr(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no POSIX mode bits on Windows")
	}
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`{"organization":{"lago_id":"org_1","name":"Example"}}`))
	})
	path := profileAt(t, server.URL)
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	_, stderr, err := execute(t, "", "--output", "json", "whoami")
	if err != nil {
		t.Fatalf("whoami refused to run: %v", err)
	}
	if !strings.Contains(stderr, "WARNING: "+path+" has permissions 0644") || !strings.Contains(stderr, "chmod 600 "+path) {
		t.Errorf("loose permissions were not reported:\n%s", stderr)
	}
	if strings.Count(stderr, "WARNING: "+path) != 1 {
		t.Errorf("permission warning printed more than once:\n%s", stderr)
	}
}

func TestQA_S3_SecureConfigDoesNotWarn(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`{"organization":{"lago_id":"org_1"}}`))
	})
	path := profileAt(t, server.URL)
	_, stderr, err := execute(t, "", "--output", "json", "whoami")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stderr, "permissions") {
		t.Errorf("a 0600 config produced a warning:\n%s", stderr)
	}
	_ = path
}

// QA F-2, S-20: an alias may not bake in credentials or disable TLS verification.
func TestQA_F2_AliasRejectsCredentialAndTLSFlags(t *testing.T) {
	server := jsonAPI(t, func(http.ResponseWriter, *http.Request) {})
	path := profileAt(t, server.URL)
	for _, expansion := range []string{
		"customers list --api-key lago_live_FAKE000000000000000000000000",
		"customers list --api-key=lago_live_FAKE000000000000000000000000",
		"customers list --api-url https://api.eu.getlago.com",
		"customers list --api-url=https://api.eu.getlago.com",
		"customers list --insecure",
		"customers list --insecure=true",
	} {
		_, _, err := execute(t, "", "alias", "set", "bad", expansion)
		if err == nil {
			t.Errorf("alias %q was accepted", expansion)
			continue
		}
		if apperr.ExitCode(err) != apperr.ExitUsage || !strings.Contains(err.Error(), "alias expansion may not set --") {
			t.Errorf("alias %q: unexpected refusal %v", expansion, err)
		}
	}
	if aliases := loadConfig(t, path).Aliases; len(aliases) != 0 {
		t.Errorf("a rejected alias was saved: %v", aliases)
	}
	if _, _, err := execute(t, "", "alias", "set", "staging", "customers list --profile staging --mode test"); err != nil {
		t.Fatalf("an alias naming a profile was rejected: %v", err)
	}
	if got := loadConfig(t, path).Aliases["staging"]; strings.Join(got, " ") != "customers list --profile staging --mode test" {
		t.Errorf("alias saved as %v", got)
	}
}
