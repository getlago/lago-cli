package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/getlago/lago-cli/internal/cli"
	"github.com/spf13/cobra/doc"
)

func main() {
	markdown := flag.String("markdown", "docs/reference", "Markdown output directory")
	man := flag.String("man", "man", "man-page output directory")
	completions := flag.String("completions", "completions", "shell completion output directory")
	flag.Parse()
	fatalIf(generate(*markdown, *man, *completions))
	fmt.Println("generated CLI reference and man pages")
}

// manPageDate pins the date rendered into every man page's .TH line.
//
// cobra/doc defaults it to time.Now(), which makes the 266 checked-in man pages change
// on the first of every month and fails `make generate-check` on a PR that touched
// nothing. Pinning it also keeps the generated documentation reproducible, matching the
// reproducible-build gate on the binary. Bump it deliberately at a release.
var manPageDate = time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)

func generate(markdown, man, completions string) error {
	// #nosec G301 -- generated public documentation directories intentionally use 0755.
	if err := os.MkdirAll(markdown, 0o755); err != nil {
		return err
	}
	// #nosec G301 -- generated public man-page directories intentionally use 0755.
	if err := os.MkdirAll(man, 0o755); err != nil {
		return err
	}
	// #nosec G301 -- generated public shell completion directories intentionally use 0755.
	if err := os.MkdirAll(completions, 0o755); err != nil {
		return err
	}
	root := cli.NewRoot(cli.NewApp(strings.NewReader(""), io.Discard, io.Discard, cli.Version()))
	root.DisableAutoGenTag = true
	if err := doc.GenMarkdownTreeCustom(root, markdown, func(string) string { return "" }, func(name string) string {
		return strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	}); err != nil {
		return err
	}
	if err := normalizeMarkdownTree(markdown); err != nil {
		return err
	}
	header := &doc.GenManHeader{Title: "LAGO", Section: "1", Source: "Lago CLI", Manual: "Lago CLI Manual", Date: &manPageDate}
	if err := doc.GenManTree(root, header, man); err != nil {
		return err
	}
	if err := root.GenBashCompletionFile(filepath.Join(completions, "lago.bash")); err != nil {
		return err
	}
	if err := root.GenZshCompletionFile(filepath.Join(completions, "_lago")); err != nil {
		return err
	}
	if err := root.GenFishCompletionFile(filepath.Join(completions, "lago.fish"), true); err != nil {
		return err
	}
	return root.GenPowerShellCompletionFile(filepath.Join(completions, "lago.ps1"))
}

func normalizeMarkdownTree(root string) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		// path is yielded by WalkDir under the output root this same program just created
		// and wrote, on a maintainer's checkout. There is no adversary between the walk
		// and the read to win a symlink race with, and no untrusted input reaches path.
		// #nosec G304,G703,G122 -- maintainer-run doc generator rewriting its own output tree.
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		lines := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")
		for i := range lines {
			lines[i] = strings.TrimRight(lines[i], " \t\r")
		}
		normalized := strings.TrimRight(strings.Join(lines, "\n"), "\n") + "\n"
		// #nosec G306,G703,G122 -- generated CLI reference is intentionally public, written back to the path just read.
		return os.WriteFile(path, []byte(normalized), 0o644)
	})
}

func fatalIf(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "docs:", err)
		os.Exit(1)
	}
}
