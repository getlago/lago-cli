package cli

import (
	"fmt"
	"io"
	"runtime"

	"github.com/getlago/lago-cli/internal/config"
)

// warnLooseConfigPermissions prints one stderr line when the config file is readable
// by anyone but its owner. `lago doctor` already reported this; QA S-3 asked for it on
// every command, because the file holds API keys and the operator who never runs doctor
// is the one who needs telling. It warns rather than refuses: an ssh-style refusal would
// break every script on a shared CI volume for a condition the operator can fix in one
// command, and the fix is printed. Windows has no POSIX mode bits, so it is skipped, as is
// a config file that does not exist yet.
func warnLooseConfigPermissions(errOut io.Writer, path string) {
	if errOut == nil || runtime.GOOS == "windows" {
		return
	}
	mode, err := config.FileMode(path)
	if err != nil || mode&0o077 == 0 {
		// Owner-only is fine whatever the owner bits: 0600 and a read-only 0400 both pass.
		return
	}
	fmt.Fprintf(errOut, "WARNING: %s has permissions %04o and holds API keys; run: chmod 600 %s\n", path, mode, path)
}
