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
				updateCheck, _, err := cliupdate.Latest(cmd.Context(), app.Version, channel, "lago-cli/"+app.Version, cliupdate.DefaultAPIBase)
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

func newUpgradeCommand(app *App) *cobra.Command {
	var channel string
	return &cobra.Command{
		Use:     "upgrade",
		Short:   "Upgrade a script-installed Lago CLI",
		Example: "  lago upgrade\n  lago upgrade --channel beta",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			check, release, err := cliupdate.Latest(cmd.Context(), app.Version, channel, "lago-cli/"+app.Version, cliupdate.DefaultAPIBase)
			if err != nil {
				return err
			}
			if !check.Development && !check.UpdateAvailable {
				fmt.Fprintf(app.Out, "Lago CLI %s is already current on the %s channel.\n", app.Version, channel)
				return nil
			}
			installed, err := cliupdate.Install(cmd.Context(), release, "lago-cli/"+app.Version)
			if err != nil {
				return err
			}
			fmt.Fprintf(app.Out, "Upgraded Lago CLI to %s (%s channel).\n", installed, channel)
			return nil
		},
	}
}

func newInitCommand(app *App) *cobra.Command {
	var region string
	var updateCheck bool
	var updateCheckSet bool
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
			case config.RegionUS:
				if apiURL == "" || region != "" {
					apiURL = transport.USAPI
				}
			case config.RegionEU:
				if apiURL == "" || region != "" {
					apiURL = transport.EUAPI
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
				return apperr.New(apperr.ExitUsage, "API URL is required for a self-hosted profile", "Pass --api-url https://your-lago.example/api/v1.")
			}
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
			client, err := transport.New(transport.Config{BaseURL: apiURL, APIKey: apiKey, Timeout: timeout, Insecure: app.insecure, NoRetry: app.noRetry, Verbose: app.verbose, UserAgent: "lago-cli/" + app.Version, Err: app.Err})
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
			cfg.CurrentProfile = profileName
			if cfg.Channel == "" {
				cfg.Channel = "stable"
			}
			if updateCheckSet {
				cfg.UpdateCheck = updateCheck
				cfg.UpdateConsent = true
			}
			cfg.Profiles[profileName] = config.Profile{Region: selectedRegion, APIURL: apiURL, APIKey: apiKey, Mode: mode, Timeout: timeout.String(), Insecure: app.insecure, OrganizationID: organizationID, Organization: organizationName}
			if err := config.Save(path, cfg); err != nil {
				return apperr.Wrap(apperr.ExitGeneral, "save configuration", err)
			}
			fmt.Fprintf(app.Out, "Connected to Lago as %s.\n", firstNonBlank(organizationName, organizationID, "your organization"))
			fmt.Fprintf(app.Out, "Saved %s profile %q to %s (mode: %s).\n", selectedRegion, profileName, path, mode)
			return nil
		},
	}
	cmd.Flags().StringVar(&region, "region", "", "Lago region: us, eu, or self-hosted")
	cmd.Flags().BoolVar(&updateCheck, "update-check", false, "Allow a once-daily anonymous release check")
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
			result := map[string]any{"profile": app.resolved.Name, "region": app.resolved.Profile.Region, "mode": app.resolved.Profile.Mode, "api_url": app.resolved.Profile.APIURL, "organization": value}
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
