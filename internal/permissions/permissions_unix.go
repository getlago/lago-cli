//go:build !windows

package permissions

import "os"

func SecureFile(path string) error { return os.Chmod(path, 0o600) }

// SecureDirectory is called only with directory paths. Owner-only 0700 is the
// least permissive usable mode for a directory because traversal needs execute.
func SecureDirectory(path string) error { return os.Chmod(path, 0o700) } // #nosec G302 -- directory, not a regular file
