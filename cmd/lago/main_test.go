package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/getlago/lago-cli/internal/cli"
)

func TestVersionEntrypointContract(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	if code := cli.ExecuteArgs([]string{"version", "--output", "json"}, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"spec_version": "1.52.1"`) {
		t.Fatalf("version output lacks embedded spec identity: %s", stdout.String())
	}
}
