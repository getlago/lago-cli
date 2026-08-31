package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/getlago/lago-cli/internal/config"
	cliupdate "github.com/getlago/lago-cli/internal/update"
)

type passiveUpdateResult struct {
	check cliupdate.Check
	cfg   config.File
	path  string
	err   error
}

func startPassiveUpdate(arguments []string, version string) <-chan passiveUpdateResult {
	if os.Getenv("LAGO_NO_UPDATE_CHECK") != "" || excludedPassiveCommand(arguments) {
		return nil
	}
	path, err := config.DefaultPath()
	if err != nil {
		return nil
	}
	cfg, err := config.Load(path)
	if err != nil || !cfg.UpdateConsent || !cfg.UpdateCheck {
		return nil
	}
	if checked, parseErr := time.Parse(time.RFC3339, cfg.LastUpdateCheck); parseErr == nil && time.Since(checked) < 24*time.Hour {
		return nil
	}
	channel := firstNonBlank(cfg.Channel, "stable")
	result := make(chan passiveUpdateResult, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 750*time.Millisecond)
		defer cancel()
		check, _, checkErr := cliupdate.Latest(ctx, version, channel, "lago-cli/"+version, cliupdate.DefaultAPIBase)
		result <- passiveUpdateResult{check: check, cfg: cfg, path: path, err: checkErr}
	}()
	return result
}

func finishPassiveUpdate(result <-chan passiveUpdateResult, errOut interface{ Write([]byte) (int, error) }) {
	if result == nil {
		return
	}
	select {
	case update := <-result:
		if update.err != nil {
			return
		}
		update.cfg.LastUpdateCheck = time.Now().UTC().Format(time.RFC3339)
		update.cfg.LatestVersion = update.check.Latest
		_ = config.Save(update.path, update.cfg)
		if update.check.UpdateAvailable {
			_, _ = fmt.Fprintf(errOut, "A newer Lago CLI is available: %s (current %s). Run `lago upgrade`.\n", update.check.Latest, update.check.Current)
		}
	default:
	}
}

// booleanRootFlags take no value, so the token after them is the command name.
// Every other persistent flag consumes the following token unless it used `=`.
var booleanRootFlags = map[string]bool{
	"--dry-run": true, "--verbose": true, "--timing": true,
	"--insecure": true, "--no-retry": true, "--help": true, "-h": true,
}

// excludedPassiveCommand reports whether the invoked command should skip the
// once-daily update check: commands that must stay fast, run offline, or are
// themselves about updating.
//
// A flag's value is not a command name. Treating it as one meant
// `lago --output json version` started a check that `lago version` skips.
func excludedPassiveCommand(arguments []string) bool {
	skipNext := false
	for _, argument := range arguments {
		if skipNext {
			skipNext = false
			continue
		}
		if strings.HasPrefix(argument, "-") {
			skipNext = !strings.Contains(argument, "=") && !booleanRootFlags[argument]
			continue
		}
		switch argument {
		case "completion", "help", "init", "upgrade", "version":
			return true
		default:
			return false
		}
	}
	return true
}
