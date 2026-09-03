package cli

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/config"
	"github.com/getlago/lago-cli/internal/transport"
	"github.com/google/uuid"
	"github.com/jmespath/go-jmespath"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

//go:embed fixtures/demo.yaml
var fixtureAssets embed.FS

type fixture struct {
	Version int            `yaml:"version"`
	Name    string         `yaml:"name"`
	Vars    map[string]any `yaml:"vars"`
	Steps   []fixtureStep  `yaml:"steps"`
}

type fixtureStep struct {
	ID      string            `yaml:"id"`
	Method  string            `yaml:"method"`
	Path    string            `yaml:"path"`
	Query   map[string]string `yaml:"query"`
	Body    any               `yaml:"body"`
	Capture map[string]string `yaml:"capture"`
}

var fixtureVariable = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

func newFixturesCommand(app *App) *cobra.Command {
	fixtures := &cobra.Command{Use: "fixtures", Short: "Run declarative Lago API scenarios"}
	var variables []string
	run := &cobra.Command{
		Use:     "run FILE",
		Short:   "Execute a multi-step fixture with variable interpolation",
		Example: "  lago fixtures run scenario.yaml\n  lago fixtures run scenario.yaml --var customer_code=example --dry-run",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireTestProfile(app, "fixture execution"); err != nil {
				return err
			}
			data, err := os.ReadFile(filepath.Clean(args[0]))
			if err != nil {
				return apperr.Wrap(apperr.ExitGeneral, "read fixture", err)
			}
			overrides, err := parseFixtureVariables(variables)
			if err != nil {
				return err
			}
			return runFixture(cmd, app, data, overrides)
		},
	}
	run.Flags().StringArrayVar(&variables, "var", nil, "Fixture variable as name=value (repeatable)")
	fixtures.AddCommand(run)
	return fixtures
}

func newSeedCommand(app *App) *cobra.Command {
	seed := &cobra.Command{Use: "seed", Short: "Populate a test account with reproducible data"}
	var prefix string
	demo := &cobra.Command{
		Use:     "demo",
		Short:   "Create a metric, plan, customer, subscription, event, and invoice preview",
		Example: "  lago seed demo\n  lago seed demo --prefix local-demo --output json",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := requireTestProfile(app, "demo seeding"); err != nil {
				return err
			}
			data, err := fixtureAssets.ReadFile("fixtures/demo.yaml")
			if err != nil {
				return apperr.Wrap(apperr.ExitGeneral, "load bundled demo fixture", err)
			}
			if prefix == "" {
				prefix = "demo-" + strings.ReplaceAll(uuid.NewString()[:8], "-", "")
			}
			return runFixture(cmd, app, data, map[string]any{"prefix": prefix})
		},
	}
	demo.Flags().StringVar(&prefix, "prefix", "", "Unique external-ID prefix")
	seed.AddCommand(demo)
	return seed
}

func runFixture(cmd *cobra.Command, app *App, data []byte, overrides map[string]any) error {
	if err := rejectRemovedFixtureKeys(data); err != nil {
		return err
	}
	var scenario fixture
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&scenario); err != nil {
		return apperr.New(apperr.ExitUsage, fmt.Sprintf("invalid fixture: %v", err), "Use version 1 with id, method, path, body, and capture fields.")
	}
	if scenario.Version != 1 || len(scenario.Steps) == 0 {
		return apperr.New(apperr.ExitUsage, "fixture must use version 1 and contain at least one step", "See the bundled `lago seed demo` fixture format.")
	}
	variables := map[string]any{}
	for name, value := range scenario.Vars {
		variables[name] = value
	}
	for name, value := range overrides {
		variables[name] = value
	}
	if err := confirmDestructiveSteps(app, scenario); err != nil {
		return err
	}
	seen := map[string]bool{}
	results := make([]any, 0, len(scenario.Steps))
	for index, step := range scenario.Steps {
		if step.ID == "" || seen[step.ID] {
			return apperr.New(apperr.ExitUsage, fmt.Sprintf("fixture step %d has a missing or duplicate id", index+1), "Give every fixture step a unique id.")
		}
		seen[step.ID] = true
		method := strings.ToUpper(firstNonBlank(step.Method, http.MethodPost))
		pathValue, err := interpolateString(step.Path, variables)
		if err != nil {
			return fixtureStepError(step.ID, err)
		}
		query := make(url.Values)
		for name, raw := range step.Query {
			value, err := interpolateString(raw, variables)
			if err != nil {
				return fixtureStepError(step.ID, err)
			}
			query.Set(name, value)
		}
		bodyValue, err := interpolateFixtureValue(step.Body, variables)
		if err != nil {
			return fixtureStepError(step.ID, err)
		}
		var body []byte
		if bodyValue != nil {
			body, err = json.Marshal(bodyValue)
			if err != nil {
				return fixtureStepError(step.ID, err)
			}
		}
		fmt.Fprintf(app.Err, "fixture: %s (%s %s)\n", step.ID, method, pathValue)
		value, response, err := app.Request(cmd.Context(), transport.Request{Method: method, Path: pathValue, Query: query, Body: body, Idempotent: isIdempotentMethod(method)})
		if err != nil {
			if app.timing && response != nil {
				fmt.Fprintf(app.Err, "fixture: %s attempts=%d total=%s (failed)\n", step.ID, response.Attempts, response.Timing.Total)
			}
			return fixtureStepError(step.ID, err)
		}
		results = append(results, map[string]any{"id": step.ID, "response": value})
		for name, expression := range step.Capture {
			if app.dryRun {
				variables[name] = fmt.Sprintf("<dry-run:%s.%s>", step.ID, expression)
				continue
			}
			captured, captureErr := jmespath.Search(expression, value)
			if captureErr != nil || captured == nil {
				return fixtureStepError(step.ID, fmt.Errorf("capture %s with %q failed", name, expression))
			}
			variables[name] = captured
		}
		if app.timing && response != nil {
			fmt.Fprintf(app.Err, "fixture: %s attempts=%d total=%s\n", step.ID, response.Attempts, response.Timing.Total)
		}
	}
	return app.Renderer().Render(map[string]any{"fixture": scenario.Name, "steps": results, "variables": variables, "dry_run": app.dryRun})
}

// requireTestProfile refuses to run scripted writes against anything but a test
// profile. Fixtures and the bundled demo create and delete real billing objects, so
// unlike a single generated command they are not offered a live-mode confirmation:
// the whole scenario is out of bounds. Dry runs are refused too, matching seed demo.
func requireTestProfile(app *App, what string) error {
	if err := app.Load(true); err != nil {
		return err
	}
	if app.resolved.Profile.Mode != config.ModeTest {
		return apperr.New(apperr.ExitUsage, what+" is restricted to test profiles", "Select an explicit test profile with --profile or pass --mode test with test credentials.")
	}
	return nil
}

// confirmDestructiveSteps gates a fixture the way a generated delete is gated, before
// any step runs. The scan happens up front so a refusal never leaves a scenario
// half-applied. Variables are not resolved yet, so a `${var}` path segment is treated
// as an opaque value when matching the step against the operation table.
func confirmDestructiveSteps(app *App, scenario fixture) error {
	var destructive []string
	for _, step := range scenario.Steps {
		method := strings.ToUpper(firstNonBlank(step.Method, http.MethodPost))
		pathTemplate := fixtureVariable.ReplaceAllString(step.Path, "_")
		if dangerous, _ := classifyRequest(method, pathTemplate); dangerous {
			destructive = append(destructive, fmt.Sprintf("%s (%s %s)", step.ID, method, step.Path))
		}
	}
	if len(destructive) == 0 {
		return nil
	}
	name := firstNonBlank(scenario.Name, "fixture")
	fmt.Fprintf(app.Err, "fixture %q contains %d destructive step(s): %s\n", name, len(destructive), strings.Join(destructive, ", "))
	return app.Confirm(name)
}

// rejectRemovedFixtureKeys names the removal of `idempotency_key` instead of letting the
// strict decoder report an anonymous unknown field. The header it set was never read by
// lago-api, so a fixture relying on it was relying on a safety that did not exist.
func rejectRemovedFixtureKeys(data []byte) error {
	var loose struct {
		Steps []map[string]any `yaml:"steps"`
	}
	if err := yaml.Unmarshal(data, &loose); err != nil {
		return nil // the strict decoder reports the shape error with its own message
	}
	for index, step := range loose.Steps {
		if _, present := step["idempotency_key"]; !present {
			continue
		}
		id, _ := step["id"].(string)
		if id == "" {
			id = fmt.Sprintf("#%d", index+1)
		}
		return apperr.New(apperr.ExitUsage, fmt.Sprintf("fixture step %q uses idempotency_key, which Lago CLI no longer supports", id), "Remove the idempotency_key lines. Lago CLI never sends an Idempotency-Key header because the Lago API does not read it; give events a transaction_id and timestamp instead.")
	}
	return nil
}

func interpolateFixtureValue(value any, variables map[string]any) (any, error) {
	switch typed := value.(type) {
	case string:
		matches := fixtureVariable.FindStringSubmatch(typed)
		if len(matches) == 2 && matches[0] == typed {
			value, exists := variables[matches[1]]
			if !exists {
				return nil, fmt.Errorf("undefined variable %q", matches[1])
			}
			return value, nil
		}
		return interpolateString(typed, variables)
	case []any:
		result := make([]any, len(typed))
		for index, item := range typed {
			interpolated, err := interpolateFixtureValue(item, variables)
			if err != nil {
				return nil, err
			}
			result[index] = interpolated
		}
		return result, nil
	case map[string]any:
		result := make(map[string]any, len(typed))
		for name, item := range typed {
			interpolated, err := interpolateFixtureValue(item, variables)
			if err != nil {
				return nil, err
			}
			result[name] = interpolated
		}
		return result, nil
	default:
		return value, nil
	}
}

func interpolateString(value string, variables map[string]any) (string, error) {
	var interpolationErr error
	result := fixtureVariable.ReplaceAllStringFunc(value, func(match string) string {
		name := fixtureVariable.FindStringSubmatch(match)[1]
		variable, exists := variables[name]
		if !exists {
			interpolationErr = fmt.Errorf("undefined variable %q", name)
			return match
		}
		return fmt.Sprint(variable)
	})
	return result, interpolationErr
}

func parseFixtureVariables(values []string) (map[string]any, error) {
	result := map[string]any{}
	for _, value := range values {
		name, raw, found := strings.Cut(value, "=")
		if !found || strings.TrimSpace(name) == "" {
			return nil, apperr.New(apperr.ExitUsage, fmt.Sprintf("invalid fixture variable %q", value), "Use --var name=value.")
		}
		result[strings.TrimSpace(name)] = raw
	}
	return result, nil
}

func fixtureStepError(step string, err error) error {
	return &apperr.Error{ExitCode: apperr.ExitCode(err), Message: fmt.Sprintf("fixture step %q failed: %v", step, err), Cause: err}
}
