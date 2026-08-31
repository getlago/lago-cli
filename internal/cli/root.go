package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/config"
	"github.com/getlago/lago-cli/internal/generated"
	"github.com/spf13/cobra"
)

var version = "dev"

func Version() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

func NewRoot(app *App) *cobra.Command {
	root := &cobra.Command{
		Use:           "lago",
		Short:         "The official CLI for Lago billing",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := configPath()
			if err == nil {
				if _, statErr := os.Stat(path); errors.Is(statErr, os.ErrNotExist) && os.Getenv("LAGO_API_KEY") == "" {
					fmt.Fprintln(app.Out, "Welcome to Lago CLI.")
					fmt.Fprintln(app.Out, "Run `lago init` to configure an API key and profile.")
					fmt.Fprintln(app.Out, "Then run `lago seed demo` or `lago --help` to get started.")
					return nil
				}
			}
			return cmd.Help()
		},
	}
	app.root = root
	flags := root.PersistentFlags()
	flags.StringVar(&app.profile, "profile", "", "Named profile to use")
	flags.StringVar(&app.apiURL, "api-url", "", "Override the Lago API URL")
	flags.StringVar(&app.apiKey, "api-key", "", "Override the Lago API key")
	flags.StringVar(&app.mode, "mode", "", "Environment mode: live or test")
	flags.DurationVar(&app.timeout, "timeout", app.timeout, "Total request timeout")
	flags.BoolVar(&app.insecure, "insecure", false, "Allow insecure HTTP or TLS for self-hosted Lago")
	flags.BoolVar(&app.noRetry, "no-retry", false, "Disable automatic retries")
	flags.BoolVar(&app.verbose, "verbose", false, "Print redacted request and response details")
	flags.BoolVar(&app.timing, "timing", false, "Print request latency breakdown")
	flags.BoolVar(&app.dryRun, "dry-run", false, "Print mutating requests without sending them")
	flags.StringVarP(&app.output, "output", "o", "table", "Output format: table, json, or yaml")
	flags.StringVar(&app.query, "query", "", "JMESPath expression applied to the response")
	flags.StringVar(&app.confirm, "confirm", "", "Confirm a dangerous operation with its resource identifier")

	root.AddCommand(newVersionCommand(app))
	root.AddCommand(newUpgradeCommand(app))
	root.AddCommand(newInitCommand(app))
	root.AddCommand(newWhoamiCommand(app))
	root.AddCommand(newDoctorCommand(app))
	root.AddCommand(newAPICommand(app))
	root.AddCommand(newDocsCommand(app))
	root.AddCommand(newAliasCommand(app))
	root.AddCommand(newFixturesCommand(app))
	root.AddCommand(newSeedCommand(app))
	root.AddCommand(newLogsCommand(app))
	root.AddCommand(newCompletionCommand(root))
	addGenerated(root, app, generated.Operations)
	addImportExport(root, app)
	return root
}

func Execute(in io.Reader, out, errOut io.Writer) int {
	return ExecuteArgs(os.Args[1:], in, out, errOut)
}

func ExecuteArgs(args []string, in io.Reader, out, errOut io.Writer) int {
	passiveUpdate := startPassiveUpdate(args, Version())
	app := NewApp(in, out, errOut, Version())
	root := NewRoot(app)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	arguments, err := expandAliasArguments(args)
	if err == nil {
		root.SetArgs(arguments)
		err = root.ExecuteContext(ctx)
	}
	if err != nil {
		var appError *apperr.Error
		if !errors.As(err, &appError) {
			err = apperr.Wrap(apperr.ExitUsage, err.Error(), err)
		}
	}
	if err == nil {
		finishPassiveUpdate(passiveUpdate, errOut)
		return apperr.ExitSuccess
	}
	if app.output == "json" {
		fmt.Fprintln(errOut, string(apperr.Encode(err)))
	} else {
		var appErr *apperr.Error
		if errors.As(err, &appErr) {
			fmt.Fprintf(errOut, "Error: %s\n", appErr.Message)
			if appErr.Code != "" {
				fmt.Fprintf(errOut, "Lago code: %s\n", appErr.Code)
			}
			if appErr.RequestID != "" {
				fmt.Fprintf(errOut, "Request ID: %s\n", appErr.RequestID)
			}
			if appErr.Suggestion != "" {
				fmt.Fprintf(errOut, "Suggestion: %s\n", appErr.Suggestion)
			}
		} else {
			fmt.Fprintf(errOut, "Error: %s\n", err)
		}
	}
	return apperr.ExitCode(err)
}

func configPath() (string, error) {
	return config.DefaultPath()
}

func normalizeCommandName(value string) string { return strings.ReplaceAll(value, "_", "-") }
