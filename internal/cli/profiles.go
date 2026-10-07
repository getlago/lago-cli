package cli

import (
	"fmt"
	"os"

	"github.com/getlago/lago-cli/internal/apperr"
	"github.com/getlago/lago-cli/internal/config"
	"github.com/spf13/cobra"
)

func newProfileCommand(app *App) *cobra.Command {
	profile := &cobra.Command{Use: "profile", Short: "Manage configured Lago profiles"}
	profile.AddCommand(newProfileAddCommand(app))
	profile.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List configured profiles without their API keys",
		Example: "  lago profile list\n  lago profile list --output json",
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			path, err := config.DefaultPath()
			if err != nil {
				return apperr.Wrap(apperr.ExitGeneral, "resolve configuration path", err)
			}
			cfg, err := config.Load(path)
			if err != nil {
				return apperr.Wrap(apperr.ExitGeneral, "load configuration", err)
			}
			// The active profile follows the same precedence as config.Resolve, so the
			// marker names the profile the next command would actually use.
			active := firstNonBlank(app.profile, os.Getenv("LAGO_PROFILE"), cfg.CurrentProfile, "default")
			rows := make([]any, 0, len(cfg.Profiles))
			for _, name := range sortedStringKeys(cfg.Profiles) {
				p := cfg.Profiles[name]
				// The API key is never part of the row: listing profiles is something
				// people paste into tickets and screen shares.
				row := map[string]any{
					"name":         name,
					"active":       name == active,
					"region":       p.Region,
					"mode":         p.Mode,
					"api_url":      p.APIURL,
					"organization": firstNonBlank(p.Organization, p.OrganizationID),
				}
				if p.Insecure {
					row["insecure"] = true
				}
				rows = append(rows, row)
			}
			if len(rows) == 0 && app.outputMode() == "table" {
				fmt.Fprintln(app.Out, "No profiles configured. Run `lago profile add <name>` to create one.")
				return nil
			}
			return app.Renderer().Render(rows)
		},
	})
	return profile
}

// newProfileAddCommand is `lago init` with the profile name as an argument: the same
// prompts, credential check, and --use semantics, so the two never drift apart.
func newProfileAddCommand(app *App) *cobra.Command {
	cmd := newInitCommand(app)
	initPreRun := cmd.PreRun
	cmd.Use = "add NAME"
	cmd.Short = "Add or update a named profile and validate its credentials"
	cmd.Example = "  lago profile add staging\n" +
		"  lago profile add staging --api-key \"$LAGO_API_KEY\" --region eu --mode test --use"
	cmd.Args = cobra.ExactArgs(1)
	cmd.PreRun = nil
	cmd.PreRunE = func(c *cobra.Command, args []string) error {
		name := args[0]
		if err := validateProfileName(name); err != nil {
			return err
		}
		if app.profile != "" && app.profile != name {
			return apperr.New(apperr.ExitUsage, fmt.Sprintf("--profile %q conflicts with profile name %q", app.profile, name), "Pass the profile name once, as the argument.")
		}
		app.profile = name
		if initPreRun != nil {
			initPreRun(c, args)
		}
		return nil
	}
	return cmd
}
