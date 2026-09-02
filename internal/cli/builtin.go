package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/config"
	"github.com/getlago/lago-cli/internal/diagnostics"
	"github.com/getlago/lago-cli/internal/generated"
	"github.com/getlago/lago-cli/internal/transport"
	cliupdate "github.com/getlago/lago-cli/internal/update"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newVersionCommand(app *App) *cobra.Command {
	var check bool
	var channel string
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Show the Lago CLI version",
		Example: "  lago version\n  lago version --check",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			result := map[string]any{
				"version":      app.Version,
				"go_version":   runtime.Version(),
				"platform":     runtime.GOOS + "/" + runtime.GOARCH,
				"spec_version": generated.SpecVersion,
				"spec_sha256":  generated.SpecSHA256,
			}
			if check {
				updateCheck, _, err := cliupdate.Latest(cmd.Context(), app.Version, channel, "lago-cli/"+app.Version, updateAPIBase())
				if err != nil {
					return err
				}
				result["update"] = updateCheck
			}
			return app.Renderer().Render(result)
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "Check whether a newer release is available")
	cmd.Flags().StringVar(&channel, "channel", "stable", "Release channel: stable or beta")
	return cmd
}

// newUpgradeCommand prints the upgrade command for how this binary was installed.
//
// It does not replace the running binary. Lago CLI ships through Homebrew and
// `go install`, and neither is self-updating: Homebrew owns its Cellar, and `go install`
// rebuilds from source. Replacing a Homebrew-managed binary in place would leave brew
// reporting a version it no longer has. See dist-channels/parked/README.md.
func newUpgradeCommand(app *App) *cobra.Command {
	var channel string
	cmd := &cobra.Command{
		Use:     "upgrade",
		Short:   "Print the command that upgrades this Lago CLI installation",
		Example: "  lago upgrade\n  lago upgrade --channel beta",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if cliupdate.IsDevelopment(app.Version) {
				// A source build has no release to compare against, so asking GitHub
				// can only produce a misleading answer (or a misleading failure when
				// the repository is private or the network is filtered). Say what the
				// binary is and how to rebuild it, without any network call.
				fmt.Fprintf(app.Out, "Lago CLI %s is a development build, not a release. Rebuild from source to update it:\n", app.Version)
				fmt.Fprintln(app.Out, "\n    go install github.com/getlago/lago-cli/cmd/lago@latest")
				fmt.Fprintln(app.Out, "\nOr run `make build` in a checkout of github.com/getlago/lago-cli.")
				return nil
			}
			check, _, err := cliupdate.Latest(cmd.Context(), app.Version, channel, "lago-cli/"+app.Version, updateAPIBase())
			if err != nil {
				return err
			}
			method, command, err := cliupdate.UpgradeCommand()
			if err != nil {
				return err
			}
			if !check.UpdateAvailable {
				fmt.Fprintf(app.Out, "Lago CLI %s is already current on the %s channel.\n", app.Version, channel)
				return nil
			}
			fmt.Fprintf(app.Out, "Lago CLI %s is available (installed: %s).\n", check.Latest, app.Version)
			switch method {
			case cliupdate.Homebrew, cliupdate.GoInstall:
				fmt.Fprintf(app.Out, "\n    %s\n", command)
			default:
				// Neither channel owns this binary, so both commands are printed rather
				// than guessing one. Sending someone to `brew upgrade` for a binary
				// Homebrew does not manage produces a brew error, not an upgrade.
				fmt.Fprintln(app.Out, "\nLago CLI does not self-update. Run whichever command matches how you installed it:")
				fmt.Fprintln(app.Out, "\n    brew upgrade getlago/tap/lago")
				fmt.Fprintln(app.Out, "    go install github.com/getlago/lago-cli/cmd/lago@latest")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&channel, "channel", "stable", "Release channel to check: stable or beta")
	return cmd
}

func newInitCommand(app *App) *cobra.Command {
	var region string
	var updateCheck bool
	var updateCheckSet bool
	var use bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Configure a Lago profile and validate its credentials",
		Example: "  lago init\n" +
			"  lago init --api-key lago_test_FAKE000000000000000000000000 --region eu --mode test\n" +
			"  lago init --api-key \"$LAGO_API_KEY\" --region self-hosted --api-url https://billing.example.test",
		Args: cobra.NoArgs,
		PreRun: func(cmd *cobra.Command, _ []string) {
			updateCheckSet = cmd.Flags().Changed("update-check")
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := config.DefaultPath()
			if err != nil {
				return apperr.Wrap(apperr.ExitGeneral, "resolve configuration path", err)
			}
			cfg, err := config.Load(path)
			if err != nil {
				return apperr.Wrap(apperr.ExitGeneral, "load configuration", err)
			}
			profileName := firstNonBlank(app.profile, os.Getenv("LAGO_PROFILE"), cfg.CurrentProfile, "default")
			existing := cfg.Profiles[profileName]
			// QA C-8, S-5: --insecure used to be written to the profile whether or not it
			// was passed, so a re-init without the flag silently cleared it and an init
			// with it silently disabled TLS verification for every later command. The
			// stored value now changes only when the flag is passed, and the result is
			// announced below whenever it ends up true.
			insecure := existing.Insecure
			if app.flagChanged("insecure") {
				insecure = app.insecure
			}
			apiKey := firstNonBlank(app.apiKey, os.Getenv("LAGO_API_KEY"), existing.APIKey)
			selectedRegion := firstNonBlank(region, existing.Region)
			mode := firstNonBlank(app.mode, os.Getenv("LAGO_MODE"), existing.Mode)

			interactive := isInteractive(app)
			reader := bufio.NewReader(app.In)
			if interactive && !updateCheckSet && !cfg.UpdateConsent {
				answer, promptErr := prompt(reader, app.Out, "Allow an anonymous release check at most once per day? (y/N)", "N")
				if promptErr != nil {
					return promptErr
				}
				updateCheck = strings.EqualFold(answer, "y") || strings.EqualFold(answer, "yes")
				updateCheckSet = true
			}
			if apiKey == "" && interactive {
				apiKey, err = prompt(reader, app.Out, "API key", "")
				if err != nil {
					return err
				}
			}
			if selectedRegion == "" && interactive {
				selectedRegion, err = prompt(reader, app.Out, "Region (us/eu/self-hosted)", config.RegionUS)
				if err != nil {
					return err
				}
			}
			selectedRegion = firstNonBlank(selectedRegion, config.RegionUS)
			apiURL := firstNonBlank(app.apiURL, os.Getenv("LAGO_API_URL"), existing.APIURL)
			switch selectedRegion {
			case config.RegionUS, config.RegionEU:
				cloudURL := transport.USAPI
				if selectedRegion == config.RegionEU {
					cloudURL = transport.EUAPI
				}
				// A region and an explicit URL that disagree is an ambiguity, not a
				// precedence question: silently preferring one is how somebody ends up
				// writing to the wrong continent. They are only accepted together when
				// they normalize to the same host.
				if app.flagChanged("api-url") && apiURL != "" {
					if conflict := conflictingTarget(apiURL, cloudURL, insecure); conflict != nil {
						return conflict
					}
				}
				if apiURL == "" || region != "" && !app.flagChanged("api-url") {
					apiURL = cloudURL
				}
			case config.RegionSelf:
				if apiURL == "" && interactive {
					apiURL, err = prompt(reader, app.Out, "Self-hosted API URL", existing.APIURL)
					if err != nil {
						return err
					}
				}
			default:
				return apperr.New(apperr.ExitUsage, "region must be us, eu, or self-hosted", "Pass --region us, --region eu, or --region self-hosted with --api-url.")
			}
			if apiKey == "" {
				return apperr.New(apperr.ExitUsage, "API key is required in a non-interactive terminal", "Pass --api-key or set LAGO_API_KEY.")
			}
			if apiURL == "" {
				return apperr.New(apperr.ExitUsage, "API URL is required for a self-hosted profile", "Pass --api-url https://your-lago.example.")
			}
			// Normalize before saving. QA pasted the full API path and the raw value went
			// into the config file, so `whoami` reported a URL that was not the one being
			// called. The profile now records the resolved base URL, and every later read
			// is the same string the client uses.
			normalized, err := transport.NormalizeBaseURL(apiURL, insecure)
			if err != nil {
				return err
			}
			apiURL = normalized.String()
			mode = firstNonBlank(mode, config.ModeLive)
			if mode != config.ModeLive && mode != config.ModeTest {
				return apperr.New(apperr.ExitUsage, "mode must be live or test", "Pass --mode live or --mode test.")
			}
			timeout := app.timeout
			if !app.flagChanged("timeout") && existing.Timeout != "" {
				if parsed, parseErr := time.ParseDuration(existing.Timeout); parseErr == nil {
					timeout = parsed
				}
			}
			client, err := transport.New(transport.Config{BaseURL: apiURL, APIKey: apiKey, Timeout: timeout, Insecure: insecure, NoRetry: app.noRetry, Verbose: app.verbose, UserAgent: "lago-cli/" + app.Version, Err: app.Err, DialContext: app.dialContext})
			if err != nil {
				return err
			}
			response, err := client.Do(cmd.Context(), transport.Request{Method: http.MethodGet, Path: "/organizations", Idempotent: true})
			if err != nil {
				return err
			}
			organizationID, organizationName := organizationIdentity(response.Body)
			if cfg.Profiles == nil {
				cfg.Profiles = map[string]config.Profile{}
			}
			cfg.Version = config.CurrentVersion
			// QA F-13: `init --profile staging` switched current_profile, so every later
			// command silently targeted the profile that was just configured. The first
			// profile ever written becomes current, because there is nothing else to
			// point at; after that, switching is opt-in with --use.
			switched := use || cfg.CurrentProfile == ""
			if switched {
				cfg.CurrentProfile = profileName
			}
			if cfg.Channel == "" {
				cfg.Channel = "stable"
			}
			if updateCheckSet {
				cfg.UpdateCheck = updateCheck
				cfg.UpdateConsent = true
			}
			cfg.Profiles[profileName] = config.Profile{Region: selectedRegion, APIURL: apiURL, APIKey: apiKey, Mode: mode, Timeout: timeout.String(), Insecure: insecure, OrganizationID: organizationID, Organization: organizationName}
			if err := config.Save(path, cfg); err != nil {
				return apperr.Wrap(apperr.ExitGeneral, "save configuration", err)
			}
			fmt.Fprintf(app.Out, "Connected to Lago as %s.\n", firstNonBlank(organizationName, organizationID, "your organization"))
			fmt.Fprintf(app.Out, "Saved %s profile %q to %s (mode: %s).\n", selectedRegion, profileName, path, mode)
			if !switched && cfg.CurrentProfile != profileName {
				fmt.Fprintf(app.Err, "Profile %q saved; %q remains the current profile. Pass --use to switch, or --profile %q per command.\n", profileName, cfg.CurrentProfile, profileName)
			}
			if insecure {
				origin := "is persisted"
				if !app.flagChanged("insecure") {
					origin = "was kept from the existing profile and stays persisted"
				}
				fmt.Fprintf(app.Err, "WARNING: insecure = true %s in profile %q; TLS verification is disabled for every command that uses it. Run `lago init --profile %q --insecure=false` to clear it.\n", origin, profileName, profileName)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&region, "region", "", "Lago region: us, eu, or self-hosted")
	cmd.Flags().BoolVar(&updateCheck, "update-check", false, "Allow a once-daily anonymous release check")
	cmd.Flags().BoolVar(&use, "use", false, "Make this profile the current profile (the first profile is current by default)")
	return cmd
}

func newWhoamiCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "whoami",
		Short:   "Show the active profile and organization",
		Example: "  lago whoami\n  lago whoami --profile staging --output json",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			value, response, err := app.Request(cmd.Context(), transport.Request{Method: http.MethodGet, Path: "/organizations", Idempotent: true})
			if err != nil {
				return err
			}
			// resolved_api_url is the host the request actually went to. api_url is what
			// the profile holds. They differ whenever a base URL was configured without
			// the /api/v1 prefix, and an operator debugging "which environment am I on"
			// needs the one the client used, not the one they typed.
			result := map[string]any{
				"profile":          app.resolved.Name,
				"region":           app.resolved.Profile.Region,
				"mode":             app.resolved.Profile.Mode,
				"api_url":          app.resolved.Profile.APIURL,
				"resolved_api_url": app.ResolvedAPIURL(),
				"organization":     value,
			}
			return app.Render(result, response)
		},
	}
}

func newDoctorCommand(app *App) *cobra.Command {
	var bundle bool
	var bundlePath string
	cmd := &cobra.Command{
		Use:     "doctor",
		Short:   "Diagnose configuration, permissions, network, and authentication",
		Example: "  lago doctor\n  lago doctor --output json",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			bundle = bundle || bundlePath != ""
			checks := []map[string]any{}
			bundleChecks := map[string]bool{}
			path, pathErr := config.DefaultPath()
			checks = append(checks, diagnostic("config_path", pathErr == nil, firstNonBlank(errorText(pathErr), path)))
			bundleChecks["config_path"] = pathErr == nil
			if pathErr == nil {
				if mode, err := config.FileMode(path); err == nil {
					secure := runtime.GOOS == "windows" || mode == 0o600
					checks = append(checks, diagnostic("config_permissions", secure, mode.String()))
					bundleChecks["config_permissions"] = secure
				} else if !errors.Is(err, os.ErrNotExist) {
					checks = append(checks, diagnostic("config_permissions", false, err.Error()))
					bundleChecks["config_permissions"] = false
				}
			}
			client, clientErr := app.Client(true)
			if clientErr != nil {
				checks = append(checks, diagnostic("configuration", false, clientErr.Error()))
				bundleChecks["configuration"] = false
				if bundle {
					if err := writeDoctorBundle(app, bundlePath, bundleChecks); err != nil {
						return err
					}
				}
				_ = app.Renderer().Render(map[string]any{"ok": false, "checks": checks})
				return clientErr
			}
			checks = append(checks, diagnostic("configuration", true, "profile resolved"))
			bundleChecks["configuration"] = true
			// The resolved URL is the single most useful line in a support ticket: it
			// distinguishes "wrong credentials" from "right credentials, wrong host".
			checks = append(checks, diagnostic("api_url", true, app.ResolvedAPIURL()))
			bundleChecks["api_url"] = true
			response, requestErr := client.Do(cmd.Context(), transport.Request{Method: http.MethodGet, Path: "/organizations", Idempotent: true})
			if requestErr != nil {
				checks = append(checks, diagnostic("api", false, requestErr.Error()))
				bundleChecks["api"] = false
				if bundle {
					if err := writeDoctorBundle(app, bundlePath, bundleChecks); err != nil {
						return err
					}
				}
				_ = app.Renderer().Render(map[string]any{"ok": false, "checks": checks})
				return requestErr
			}
			checks = append(checks, diagnostic("api", true, fmt.Sprintf("HTTP %d; request_id=%s", response.Status, response.RequestID)))
			bundleChecks["api"] = true
			if bundle {
				if err := writeDoctorBundle(app, bundlePath, bundleChecks); err != nil {
					return err
				}
			}
			return app.Renderer().Render(map[string]any{"ok": true, "checks": checks})
		},
	}
	cmd.Flags().BoolVar(&bundle, "bundle", false, "Write a sanitized diagnostic archive")
	cmd.Flags().StringVar(&bundlePath, "bundle-path", "", "Diagnostic archive path (implies --bundle)")
	return cmd
}

func writeDoctorBundle(app *App, path string, checks map[string]bool) error {
	if path == "" {
		path = fmt.Sprintf("lago-diagnostics-%s.tar.gz", time.Now().UTC().Format("20060102T150405Z"))
	}
	if err := diagnostics.WriteBundle(path, diagnostics.Data{Version: app.Version, SpecVersion: generated.SpecVersion, SpecSHA256: generated.SpecSHA256, Region: app.resolved.Profile.Region, Mode: app.resolved.Profile.Mode, APIURL: app.resolved.Profile.APIURL, Checks: checks}); err != nil {
		return apperr.Wrap(apperr.ExitGeneral, "write diagnostic bundle", err)
	}
	fmt.Fprintf(app.Err, "Sanitized diagnostic bundle: %s\n", path)
	return nil
}

func newAPICommand(app *App) *cobra.Command {
	var data string
	var headers []string
	var idempotencyKey string
	cmd := &cobra.Command{
		Use:   "api METHOD PATH",
		Short: "Make an authenticated request to any Lago API endpoint",
		Example: "  lago api GET /customers?page=2\n" +
			"  lago api POST /events --data @event.json --idempotency-key event-42\n" +
			"  printf '{\"event\":{}}' | lago api POST /events --data -",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			method := strings.ToUpper(args[0])
			path, query, err := parsePathQuery(args[1])
			if err != nil {
				return err
			}
			if err := app.Load(true); err != nil {
				return err
			}
			// A raw request inherits the spec's danger classification. Live mode goes
			// through the same gate as a generated delete; test mode stays the ungated
			// escape hatch the command exists for. See DECISIONS.md.
			if dangerous, _ := classifyRequest(method, path); dangerous && app.resolved.Profile.Mode == config.ModeLive {
				if err := app.Confirm(path); err != nil {
					return err
				}
			}
			body, err := readData(app.In, data)
			if err != nil {
				return err
			}
			httpHeaders := make(http.Header)
			for _, header := range headers {
				name, value, found := strings.Cut(header, ":")
				if !found || strings.TrimSpace(name) == "" {
					return apperr.New(apperr.ExitUsage, fmt.Sprintf("invalid header %q", header), "Use --header 'Name: value'.")
				}
				if strings.EqualFold(strings.TrimSpace(name), "Authorization") {
					return apperr.New(apperr.ExitUsage, "Authorization is managed by Lago CLI", "Select credentials with --profile or LAGO_API_KEY.")
				}
				httpHeaders.Add(strings.TrimSpace(name), strings.TrimSpace(value))
			}
			if idempotencyKey != "" {
				httpHeaders.Set("Idempotency-Key", idempotencyKey)
			}
			idempotent := isIdempotentMethod(method) || idempotencyKey != ""
			value, response, err := app.Request(cmd.Context(), transport.Request{Method: method, Path: path, Query: query, Headers: httpHeaders, Body: body, Idempotent: idempotent})
			if err != nil {
				return err
			}
			return app.Render(value, response)
		},
	}
	cmd.Flags().StringVarP(&data, "data", "d", "", "Request body JSON, @file, or - for stdin")
	cmd.Flags().StringSliceVarP(&headers, "header", "H", nil, "Additional request header (repeatable)")
	cmd.Flags().StringVar(&idempotencyKey, "idempotency-key", "", "Idempotency key for safe mutation retries")
	return cmd
}

func newCompletionCommand(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "completion SHELL",
		Short:     "Generate shell completion scripts",
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.ExactArgs(1),
		Example:   "  lago completion zsh > ~/.zfunc/_lago\n  lago completion powershell | Out-String | Invoke-Expression",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return root.GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return root.GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return root.GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return root.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			default:
				return apperr.New(apperr.ExitUsage, "unsupported shell", "Choose bash, zsh, fish, or powershell.")
			}
		},
	}
	return cmd
}

func prompt(reader *bufio.Reader, out io.Writer, label, defaultValue string) (string, error) {
	if defaultValue == "" {
		fmt.Fprintf(out, "%s: ", label)
	} else {
		fmt.Fprintf(out, "%s [%s]: ", label, defaultValue)
	}
	value, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", apperr.Wrap(apperr.ExitGeneral, "read interactive input", err)
	}
	return firstNonBlank(strings.TrimSpace(value), defaultValue), nil
}

func isInteractive(app *App) bool {
	file, ok := app.In.(*os.File)
	return ok && term.IsTerminal(int(file.Fd()))
}

func organizationIdentity(body []byte) (string, string) {
	var payload map[string]any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if decoder.Decode(&payload) != nil {
		return "", ""
	}
	value := payload
	if organization, ok := payload["organization"].(map[string]any); ok {
		value = organization
	}
	id, _ := value["lago_id"].(string)
	if id == "" {
		id, _ = value["id"].(string)
	}
	name, _ := value["name"].(string)
	return id, name
}

func readData(in io.Reader, value string) ([]byte, error) {
	if value == "" {
		return nil, nil
	}
	if value == "-" {
		data, err := io.ReadAll(io.LimitReader(in, 32<<20))
		if err != nil {
			return nil, apperr.Wrap(apperr.ExitGeneral, "read request body from stdin", err)
		}
		return validateJSON(data)
	}
	if strings.HasPrefix(value, "@") {
		path := strings.TrimPrefix(value, "@")
		if path == "" {
			return nil, apperr.New(apperr.ExitUsage, "missing file after @", "Use --data @payload.json.")
		}
		data, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return nil, apperr.Wrap(apperr.ExitGeneral, "read request body", err)
		}
		if len(data) > 32<<20 {
			return nil, apperr.New(apperr.ExitUsage, "request body exceeds 32 MiB", "Use a bulk streaming command for large inputs.")
		}
		return validateJSON(data)
	}
	return validateJSON([]byte(value))
}

func validateJSON(data []byte) ([]byte, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, apperr.New(apperr.ExitUsage, fmt.Sprintf("request body is not valid JSON: %v", err), "Pass JSON directly, --data @file.json, or --data -.")
	}
	return data, nil
}

// conflictingTarget reports a usage error when an explicit --api-url and a --region
// shorthand name different deployments. Equal after normalization is not a conflict:
// `--region us --api-url https://api.getlago.com/api/v1` is redundant, not wrong.
func conflictingTarget(apiURL, regionURL string, insecure bool) error {
	explicit, err := transport.NormalizeBaseURL(apiURL, insecure)
	if err != nil {
		return err
	}
	shorthand, err := transport.NormalizeBaseURL(regionURL, insecure)
	if err != nil {
		return err
	}
	if explicit.String() == shorthand.String() {
		return nil
	}
	return apperr.New(apperr.ExitUsage,
		fmt.Sprintf("--api-url %s and the selected region (%s) are different deployments", explicit, shorthand),
		"Pass one or the other: --region us|eu for Lago Cloud, or --region self-hosted with --api-url.")
}

func diagnostic(name string, ok bool, detail string) map[string]any {
	return map[string]any{"name": name, "ok": ok, "detail": detail}
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func sortedStringKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

var _ = url.Values{}
