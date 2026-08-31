package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateWritesRootReferenceAndManPage(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	markdown := filepath.Join(directory, "markdown")
	man := filepath.Join(directory, "man")
	completions := filepath.Join(directory, "completions")
	if err := generate(markdown, man, completions); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{filepath.Join(markdown, "lago.md"), filepath.Join(man, "lago.1"), filepath.Join(completions, "lago.bash"), filepath.Join(completions, "_lago"), filepath.Join(completions, "lago.fish"), filepath.Join(completions, "lago.ps1")} {
		if info, err := os.Stat(path); err != nil || info.Size() == 0 {
			t.Fatalf("generated file %s: info=%v error=%v", path, info, err)
		}
	}
	markdownFiles, err := filepath.Glob(filepath.Join(markdown, "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range markdownFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.HasSuffix(content, []byte("\n")) || bytes.HasSuffix(content, []byte("\n\n")) {
			t.Errorf("%s must end with exactly one newline", path)
		}
		for lineNumber, line := range strings.Split(string(content), "\n") {
			if strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t") || strings.HasSuffix(line, "\r") {
				t.Errorf("%s:%d has trailing whitespace", path, lineNumber+1)
			}
		}
	}
}
