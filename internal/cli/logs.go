package cli

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/generated"
	"github.com/spf13/cobra"
)

func newLogsCommand(app *App) *cobra.Command {
	logs := &cobra.Command{Use: "logs", Short: "Inspect Lago API request logs"}
	var statuses []string
	var methods []string
	var resource string
	var interval time.Duration
	tail := &cobra.Command{
		Use:     "tail",
		Short:   "Continuously render API logs when they change",
		Example: "  lago logs tail\n  lago logs tail --status 4xx --resource /customers --method POST",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if interval < 500*time.Millisecond {
				return apperr.New(apperr.ExitUsage, "log polling interval must be at least 500ms", "Pass --interval 2s or a longer duration.")
			}
			query := make(url.Values)
			query.Set("page", "1")
			query.Set("per_page", "100")
			for _, status := range statuses {
				query.Add("http_statuses[]", status)
			}
			for _, method := range methods {
				query.Add("http_methods[]", strings.ToUpper(method))
			}
			if resource != "" {
				query.Set("request_paths", resource)
			}
			operation := generated.Operation{Resource: "api-logs", Action: "list", Method: http.MethodGet, Path: "/api_logs", Idempotent: true}
			return runWatch(cmd, app, operation, operation.Path, query, interval)
		},
	}
	tail.Flags().StringSliceVar(&statuses, "status", nil, "HTTP status or class filter (repeatable)")
	tail.Flags().StringSliceVar(&methods, "method", nil, "HTTP method filter (repeatable)")
	tail.Flags().StringVar(&resource, "resource", "", "Request path filter")
	tail.Flags().DurationVar(&interval, "interval", 2*time.Second, "Polling interval")
	logs.AddCommand(tail)
	return logs
}
