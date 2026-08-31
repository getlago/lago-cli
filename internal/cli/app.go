package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/config"
	"github.com/getlago/lago-cli/internal/output"
	"github.com/getlago/lago-cli/internal/transport"
	"github.com/spf13/cobra"
)

type App struct {
	In      io.Reader
	Out     io.Writer
	Err     io.Writer
	Start   time.Time
	Version string

	root       *cobra.Command
	configPath string
	config     config.File
	resolved   config.Resolved
	client     *transport.Client
	loaded     bool

	profile  string
	apiURL   string
	apiKey   string
	mode     string
	timeout  time.Duration
	insecure bool
	noRetry  bool
	verbose  bool
	timing   bool
	dryRun   bool
	output   string
	query    string
	confirm  string
}

func NewApp(in io.Reader, out, errOut io.Writer, version string) *App {
	return &App{In: in, Out: out, Err: errOut, Start: time.Now(), Version: version, output: output.Table, timeout: 30 * time.Second}
}

func (a *App) Renderer() output.Renderer {
	return output.Renderer{Mode: a.output, Query: a.query, Out: a.Out}
}

func (a *App) Load(requireAuth bool) error {
	if a.loaded {
		if requireAuth && a.resolved.Profile.APIKey == "" {
			return apperr.New(apperr.ExitAuth, "Lago API key is not configured", "Run `lago init` or set LAGO_API_KEY.")
		}
		return nil
	}
	path, err := config.DefaultPath()
	if err != nil {
		return apperr.Wrap(apperr.ExitGeneral, "resolve configuration path", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		return apperr.Wrap(apperr.ExitGeneral, "load configuration", err)
	}
	overrides := config.Overrides{
		Profile:     a.profile,
		APIURL:      a.apiURL,
		APIKey:      a.apiKey,
		Mode:        a.mode,
		Timeout:     a.timeout,
		Insecure:    a.insecure,
		APIURLSet:   a.flagChanged("api-url"),
		APIKeySet:   a.flagChanged("api-key"),
		ModeSet:     a.flagChanged("mode"),
		TimeoutSet:  a.flagChanged("timeout"),
		InsecureSet: a.flagChanged("insecure"),
	}
	resolved, err := config.Resolve(cfg, overrides)
	if err != nil {
		return err
	}
	a.configPath, a.config, a.resolved, a.loaded = path, cfg, resolved, true
	if requireAuth && resolved.Profile.APIKey == "" {
		return apperr.New(apperr.ExitAuth, "Lago API key is not configured", "Run `lago init` or set LAGO_API_KEY.")
	}
	return nil
}

func (a *App) Client(requireAuth bool) (*transport.Client, error) {
	if err := a.Load(requireAuth); err != nil {
		return nil, err
	}
	if a.client != nil {
		return a.client, nil
	}
	client, err := transport.New(transport.Config{
		BaseURL:   a.resolved.Profile.APIURL,
		APIKey:    a.resolved.Profile.APIKey,
		Timeout:   a.resolved.Timeout,
		Insecure:  a.resolved.Profile.Insecure,
		NoRetry:   a.noRetry,
		Verbose:   a.verbose || os.Getenv("LAGO_DEBUG") == "1",
		UserAgent: "lago-cli/" + a.Version,
		Err:       a.Err,
	})
	if err != nil {
		return nil, err
	}
	a.client = client
	return client, nil
}

func (a *App) Request(ctx context.Context, request transport.Request) (any, *transport.Response, error) {
	client, err := a.Client(true)
	if err != nil {
		return nil, nil, err
	}
	if a.resolved.Profile.Mode == config.ModeLive {
		fmt.Fprintf(a.Err, "[LIVE] profile=%s\n", a.resolved.Name)
	}
	if a.resolved.Profile.Insecure {
		fmt.Fprintln(a.Err, "WARNING: --insecure disables transport security checks for this profile.")
	}
	request.DryRun = request.DryRun || a.dryRun
	response, err := client.Do(ctx, request)
	if err != nil {
		return nil, response, err
	}
	if response.DryRunData != nil {
		return response.DryRunData, response, nil
	}
	value, decodeErr := transport.DecodeJSON(response.Body)
	if decodeErr != nil {
		if strings.HasPrefix(response.Headers.Get("Content-Type"), "application/json") {
			return nil, response, apperr.Wrap(apperr.ExitGeneral, "decode Lago API response", decodeErr)
		}
		return string(response.Body), response, nil
	}
	return value, response, nil
}

func (a *App) Render(value any, response *transport.Response) error {
	if err := a.Renderer().Render(value); err != nil {
		return err
	}
	if a.timing && response != nil {
		if elapsed := time.Since(a.Start) - response.Timing.Total; elapsed > 0 {
			response.Timing.CLIOverhead = elapsed
		}
		encoded, _ := json.Marshal(response.Timing)
		fmt.Fprintf(a.Err, "timing: %s\n", encoded)
	}
	return nil
}

func (a *App) Confirm(identifier string) error {
	if a.dryRun {
		return nil
	}
	if err := a.Load(true); err != nil {
		return err
	}
	if a.confirm == identifier {
		return nil
	}
	if !isInteractive(a) {
		return apperr.New(apperr.ExitUsage, fmt.Sprintf("confirmation required for %s", identifier), fmt.Sprintf("Pass --confirm %q or rerun in an interactive terminal.", identifier))
	}
	return promptConfirmation(a.In, a.Err, identifier, a.resolved.Profile.Mode == config.ModeLive)
}

// promptConfirmation holds the confirmation policy, separated from the terminal check
// in Confirm so it can be exercised without a pty. Live mode demands the identifier
// typed back exactly; test mode accepts y/N. Both default to refusing.
func promptConfirmation(in io.Reader, errOut io.Writer, identifier string, live bool) error {
	reader := bufio.NewReader(in)
	if live {
		fmt.Fprintf(errOut, "[LIVE] Type %q to confirm this destructive operation: ", identifier)
		value, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return apperr.Wrap(apperr.ExitGeneral, "read confirmation", err)
		}
		if strings.TrimSpace(value) != identifier {
			return apperr.New(apperr.ExitUsage, "confirmation did not match the resource identifier", fmt.Sprintf("Retry and type %q exactly.", identifier))
		}
		return nil
	}
	fmt.Fprintf(errOut, "Delete or terminate %q in test mode? [y/N]: ", identifier)
	value, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return apperr.Wrap(apperr.ExitGeneral, "read confirmation", err)
	}
	if strings.ToLower(strings.TrimSpace(value)) != "y" {
		return apperr.New(apperr.ExitUsage, "operation cancelled", "No changes were made.")
	}
	return nil
}

func (a *App) flagChanged(name string) bool {
	if a.root == nil {
		return false
	}
	flag := a.root.PersistentFlags().Lookup(name)
	return flag != nil && flag.Changed
}

func parsePathQuery(raw string) (string, url.Values, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", nil, apperr.New(apperr.ExitUsage, "invalid API path", "Use a relative path such as /customers?page=2.")
	}
	if parsed.IsAbs() || parsed.Host != "" {
		return "", nil, apperr.New(apperr.ExitUsage, "absolute API paths are not allowed", "Use --api-url to select the server and pass a relative API path.")
	}
	return parsed.Path, parsed.Query(), nil
}

// isIdempotentMethod reports whether the transport may replay a request on its own,
// without an operator-supplied idempotency key. Only side-effect-free reads qualify.
//
// PUT and DELETE are idempotent in the RFC 9110 sense and still move money in Lago:
// PUT /invoices/{id}/finalize issues an invoice and can trigger a payment attempt.
// Mutations become replayable only by carrying an Idempotency-Key, never by verb.
func isIdempotentMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}
