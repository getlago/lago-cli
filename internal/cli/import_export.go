package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/generated"
	"github.com/getlago/lago-cli/internal/transport"
	"github.com/spf13/cobra"
)

func addImportExport(root *cobra.Command, app *App) {
	customers := findDirectChild(root, "customers")
	if customers != nil {
		for _, operation := range generated.Operations {
			if operation.OperationID == "findAllCustomers" {
				export := operation
				export.Action = "export"
				export.Summary = "Export customers using the stable list response schema"
				customers.AddCommand(newGeneratedCommand(app, export))
				break
			}
		}
	}
	plans := findDirectChild(root, "plans")
	if plans != nil {
		plans.AddCommand(newPlansImportCommand(app))
	}
}

func findDirectChild(root *cobra.Command, name string) *cobra.Command {
	for _, command := range root.Commands() {
		if command.Name() == name {
			return command
		}
	}
	return nil
}

func newPlansImportCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "import FILE",
		Short:   "Create or update plans from JSON with a dry-run diff",
		Example: "  lago plans import plans.json --dry-run\n  lago plans import plans.json --output json",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(filepath.Clean(args[0]))
			if err != nil {
				return apperr.Wrap(apperr.ExitGeneral, "read plans import", err)
			}
			if len(data) > 64<<20 {
				return apperr.New(apperr.ExitUsage, "plans import exceeds 64 MiB", "Split the import into smaller files.")
			}
			plans, err := decodePlanImport(data)
			if err != nil {
				return err
			}
			client, err := app.Client(true)
			if err != nil {
				return err
			}
			report := make([]any, 0, len(plans))
			for index, plan := range plans {
				code, _ := plan["code"].(string)
				if code == "" {
					return apperr.New(apperr.ExitUsage, fmt.Sprintf("plan %d has no string code", index+1), "Every imported plan requires the OpenAPI plan.code field.")
				}
				action := "create"
				path := "/plans"
				method := http.MethodPost
				existingResponse, findErr := client.Do(cmd.Context(), transport.Request{Method: http.MethodGet, Path: "/plans/" + url.PathEscape(code), Idempotent: true})
				if findErr == nil {
					existing, decodeErr := transport.DecodeJSON(existingResponse.Body)
					if decodeErr != nil {
						return apperr.Wrap(apperr.ExitGeneral, "decode existing plan", decodeErr)
					}
					existingPlan := unwrapNamedObject(existing, "plan")
					if subsetEqual(existingPlan, plan) {
						report = append(report, map[string]any{"code": code, "action": "unchanged"})
						continue
					}
					action, method, path = "update", http.MethodPut, "/plans/"+url.PathEscape(code)
				} else if apperr.ExitCode(findErr) != apperr.ExitNotFound {
					return findErr
				}
				body, err := json.Marshal(map[string]any{"plan": plan})
				if err != nil {
					return apperr.Wrap(apperr.ExitGeneral, "encode imported plan", err)
				}
				// A plan write is never replayed on its own: lago-api does not read an
				// Idempotency-Key, and a PUT that races a concurrent edit is not idempotent
				// in the billing sense. A transient failure is reported and rerun by hand.
				value, _, err := app.Request(cmd.Context(), transport.Request{Method: method, Path: path, Body: body, Idempotent: false})
				if err != nil {
					return err
				}
				report = append(report, map[string]any{"code": code, "action": action, "request": value})
			}
			return app.Renderer().Render(map[string]any{"plans": report, "dry_run": app.dryRun})
		},
	}
}

func decodePlanImport(data []byte) ([]map[string]any, error) {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, apperr.New(apperr.ExitUsage, fmt.Sprintf("invalid plans JSON: %v", err), "Use a plan object, an array of plan objects, or {\"plans\":[...]}.")
	}
	var rawPlans []any
	switch typed := value.(type) {
	case []any:
		rawPlans = typed
	case map[string]any:
		if nested, ok := typed["plans"].([]any); ok {
			rawPlans = nested
		} else if nested, ok := typed["plan"].(map[string]any); ok {
			rawPlans = []any{nested}
		} else {
			rawPlans = []any{typed}
		}
	default:
		return nil, apperr.New(apperr.ExitUsage, "plans import must contain JSON objects", "Use a plan object or an array of plan objects.")
	}
	plans := make([]map[string]any, 0, len(rawPlans))
	for index, raw := range rawPlans {
		plan, ok := raw.(map[string]any)
		if !ok {
			return nil, apperr.New(apperr.ExitUsage, fmt.Sprintf("plan %d is not an object", index+1), "Use one JSON object per plan.")
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

func unwrapNamedObject(value any, name string) map[string]any {
	object, _ := value.(map[string]any)
	nested, _ := object[name].(map[string]any)
	return nested
}

func subsetEqual(existing, desired map[string]any) bool {
	for key, desiredValue := range desired {
		existingValue, exists := existing[key]
		if !exists {
			return false
		}
		desiredObject, desiredIsObject := desiredValue.(map[string]any)
		existingObject, existingIsObject := existingValue.(map[string]any)
		if desiredIsObject && existingIsObject {
			if !subsetEqual(existingObject, desiredObject) {
				return false
			}
			continue
		}
		if !reflect.DeepEqual(existingValue, desiredValue) {
			return false
		}
	}
	return true
}
