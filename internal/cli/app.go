package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
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

	// dialContext lets this package's tests point the production hostnames at a local
	// server, so URL handling is proven against api.getlago.com and api.eu.getlago.com
	// rather than only against 127.0.0.1. Nil everywhere outside tests.
	dialContext func(ctx context.Context, network, address string) (net.Conn, error)

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
	return output.Renderer{Mode: a.outputMode(), Query: a.query, Out: a.Out, Err: a.Err}
}

// outputMode resolves --output, switching an unspecified format to JSON when --query is
// used.
//
// A JMESPath result is structured data: a projection, a filtered list, a scalar. The
// table renderer has nothing useful to do with most of those, and QA's `--query` returned
// an empty table that read as "no results" rather than "wrong expression". Choosing JSON
// makes the default correct and greppable. An explicit --output is always honoured,
// including --output table, so nothing is taken away.
//
// The switch is announced on stderr rather than performed silently: a changed output
// format is exactly the kind of thing a script author needs to know happened.
func (a *App) outputMode() string {
	if a.query == "" || a.flagChanged("output") {
		return a.output
	}
	if a.output == output.Table {
		return output.JSON
	}
	return a.output
}

// noteQueryOutputSwitch prints the one-line stderr notice for the --query JSON switch.
// It is separate from outputMode so the resolution stays free of side effects and can be
// called from anywhere a renderer is built.
func (a *App) noteQueryOutputSwitch() {
	if a.query != "" && !a.flagChanged("output") && a.output == output.Table && a.Err != nil {
		fmt.Fprintln(a.Err, "--query implies --output json; pass --output table explicitly to render the result as a table.")
	}
}

// IdentifierRenderer is the renderer for a create or update: terse identifiers in the
// default table output, the complete resource under --output json|yaml.
//
// An explicit --query is honoured as written. The operator has already said which
// fields they want; silently reducing the response first would drop the very keys
// their expression addresses.
func (a *App) IdentifierRenderer() output.Renderer {
	renderer := a.Renderer()
	renderer.Identifiers = a.query == ""
	return renderer
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
	warnLooseConfigPermissions(a.Err, path)
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

		DialContext: a.dialContext,
	})
	if err != nil {
		return nil, err
	}
	a.client = client
	return client, nil
}

// ResolvedAPIURL is the base URL requests are actually sent to, after normalization.
//
// It is what `whoami` and `doctor` report, because the value in the profile is what the
// operator typed and the resolved value is what the client calls. QA lost time to a
// profile that displayed one host while requests went to another.
func (a *App) ResolvedAPIURL() string {
	normalized, err := transport.NormalizeBaseURL(a.resolved.Profile.APIURL, a.resolved.Profile.Insecure)
	if err != nil {
		return a.resolved.Profile.APIURL
	}
	return normalized.String()
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
		// The success path reports timing after rendering; a failure never renders, so
		// report here. F-10/F-11: retry behaviour is only observable through this line.
		a.reportTiming(response)
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
	return a.render(a.Renderer(), value, response)
}

// RenderMutation renders a create or update response through the identifier renderer.
//
// A --dry-run response is the request envelope, not a resource, so it never goes
// through the identifier reduction; see render for how the envelope is printed.
func (a *App) RenderMutation(value any, response *transport.Response) error {
	if response != nil && response.DryRunData != nil {
		return a.render(a.Renderer(), value, response)
	}
	return a.render(a.IdentifierRenderer(), value, response)
}

func (a *App) render(renderer output.Renderer, value any, response *transport.Response) error {
	// A --dry-run envelope is `{method, url, headers, body}` with the payload nested
	// under body. Table cells summarise nested values, which would reduce the payload
	// the flag exists to show to `{2 fields}`, so the envelope prints as JSON instead:
	// it is a request, not a resource, and JSON is the form it will be sent in.
	if response != nil && response.DryRunData != nil && (renderer.Mode == "" || renderer.Mode == output.Table) {
		renderer.Mode = output.JSON
	}
	if err := renderer.Render(value); err != nil {
		return err
	}
	a.reportTiming(response)
	return nil
}

// reportTiming prints the `timing:` line for one request when --timing is set. It is
// called once per request, from render on success and from Request on failure.
func (a *App) reportTiming(response *transport.Response) {
	if !a.timing || response == nil {
		return
	}
	if elapsed := time.Since(a.Start) - response.Timing.Total; elapsed > 0 {
		response.Timing.CLIOverhead = elapsed
	}
	encoded, _ := json.Marshal(response.Timing)
	fmt.Fprintf(a.Err, "timing: %s\n", encoded)
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

// isIdempotentMethod reports whether the transport may replay a request on its own.
// Only side-effect-free reads qualify.
//
// PUT and DELETE are idempotent in the RFC 9110 sense and still move money in Lago:
// PUT /invoices/{id}/finalize issues an invoice and can trigger a payment attempt.
// Mutations are never replayed by verb, and lago-api does not read an Idempotency-Key
// header, so the CLI offers none; the one exception is a usage event that carries a
// timestamp, whose transaction_id plus timestamp the server deduplicates.
func isIdempotentMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}
