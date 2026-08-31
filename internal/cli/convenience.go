package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/config"
	"github.com/getlago/lago-cli/internal/generated"
	"github.com/spf13/cobra"
)

func newDocsCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:       "docs RESOURCE",
		Short:     "Open Lago API documentation for a resource",
		Example:   "  lago docs customers\n  lago docs billable-metrics",
		Args:      cobra.ExactArgs(1),
		ValidArgs: generatedResources(),
		RunE: func(_ *cobra.Command, args []string) error {
			resource := normalizeCommandName(args[0])
			url := ""
			for _, operation := range generated.Operations {
				if operation.Resource == resource && operation.DocsURL != "" {
					url = operation.DocsURL
					break
				}
			}
			if url == "" {
				return apperr.New(apperr.ExitNotFound, fmt.Sprintf("no documentation URL is declared for %s", resource), "Run `lago --help` to list generated resources.")
			}
			if !isInteractive(app) {
				fmt.Fprintln(app.Out, url)
				return nil
			}
			if err := openBrowser(url); err != nil {
				fmt.Fprintln(app.Out, url)
				return apperr.Wrap(apperr.ExitGeneral, "open browser", err)
			}
			fmt.Fprintf(app.Out, "Opened %s\n", url)
			return nil
		},
	}
}

func newAliasCommand(app *App) *cobra.Command {
	alias := &cobra.Command{Use: "alias", Short: "Manage user-defined command aliases"}
	alias.AddCommand(&cobra.Command{
		Use:     "list",
		Short:   "List command aliases",
		Example: "  lago alias list",
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			path, err := config.DefaultPath()
			if err != nil {
				return err
			}
			cfg, err := config.Load(path)
			if err != nil {
				return err
			}
			rows := make([]any, 0, len(cfg.Aliases))
			for _, name := range sortedStringKeys(cfg.Aliases) {
				rows = append(rows, map[string]any{"name": name, "expansion": strings.Join(cfg.Aliases[name], " ")})
			}
			return app.Renderer().Render(rows)
		},
	})
	alias.AddCommand(&cobra.Command{
		Use:     "set NAME EXPANSION",
		Short:   "Create or replace a command alias",
		Example: `  lago alias set cust "customers"`,
		Args:    cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			expansion := strings.Fields(args[1])
			if name == "" || strings.HasPrefix(name, "-") || strings.ContainsAny(name, " \t\r\n") || len(expansion) == 0 {
				return apperr.New(apperr.ExitUsage, "invalid alias", "Use a single command word and a non-empty expansion.")
			}
			for _, reserved := range append(generatedResources(), "alias", "api", "completion", "docs", "doctor", "fixtures", "help", "init", "logs", "seed", "upgrade", "version", "whoami") {
				if name == reserved {
					return apperr.New(apperr.ExitUsage, fmt.Sprintf("%q is a built-in command", name), "Choose an alias name that does not shadow the command tree.")
				}
			}
			path, err := config.DefaultPath()
			if err != nil {
				return err
			}
			cfg, err := config.Load(path)
			if err != nil {
				return err
			}
			if cfg.Aliases == nil {
				cfg.Aliases = map[string][]string{}
			}
			cfg.Aliases[name] = expansion
			if err := config.Save(path, cfg); err != nil {
				return err
			}
			fmt.Fprintf(app.Out, "Alias %q expands to %q.\n", name, strings.Join(expansion, " "))
			return nil
		},
	})
	alias.AddCommand(&cobra.Command{
		Use:     "delete NAME",
		Aliases: []string{"remove"},
		Short:   "Delete a command alias",
		Example: "  lago alias delete cust",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			path, err := config.DefaultPath()
			if err != nil {
				return err
			}
			cfg, err := config.Load(path)
			if err != nil {
				return err
			}
			if _, exists := cfg.Aliases[args[0]]; !exists {
				return apperr.New(apperr.ExitNotFound, fmt.Sprintf("alias %q does not exist", args[0]), "Run `lago alias list` to inspect configured aliases.")
			}
			delete(cfg.Aliases, args[0])
			return config.Save(path, cfg)
		},
	})
	return alias
}

func expandAliasArguments(arguments []string) ([]string, error) {
	if len(arguments) == 0 || strings.HasPrefix(arguments[0], "-") {
		return arguments, nil
	}
	path, err := config.DefaultPath()
	if err != nil {
		return nil, apperr.Wrap(apperr.ExitGeneral, "resolve configuration path", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, apperr.Wrap(apperr.ExitGeneral, "load configuration", err)
	}
	seen := map[string]bool{}
	for range 10 {
		expansion, exists := cfg.Aliases[arguments[0]]
		if !exists {
			return arguments, nil
		}
		if seen[arguments[0]] {
			return nil, apperr.New(apperr.ExitUsage, "alias expansion contains a cycle", "Update the aliases with `lago alias set`.")
		}
		seen[arguments[0]] = true
		arguments = append(append([]string{}, expansion...), arguments[1:]...)
	}
	return nil, apperr.New(apperr.ExitUsage, "alias expansion is too deep", "Use fewer than ten nested aliases.")
}

func generatedResources() []string {
	resources := map[string]bool{}
	for _, operation := range generated.Operations {
		resources[operation.Resource] = true
	}
	return sortedStringKeys(resources)
}

func openBrowser(url string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("open", url)
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		command = exec.Command("xdg-open", url)
	}
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Start()
}
