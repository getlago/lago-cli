//go:build e2e

package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestGoldenInvoice(t *testing.T) {
	apiKey := os.Getenv("LAGO_API_KEY")
	if apiKey == "" {
		t.Skip("trusted staging credentials are not configured")
	}
	binary := filepath.Join("..", "..", "bin", "lago")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	configPath := filepath.Join(t.TempDir(), "config.toml")
	environment := append(os.Environ(), "LAGO_CONFIG_FILE="+configPath, "LAGO_MODE=test")
	target := os.Getenv("LAGO_E2E_TARGET")
	initArgs := []string{"init", "--region", target, "--mode", "test", "--profile", "e2e"}
	if target == "self-hosted" {
		initArgs = append(initArgs, "--api-url", os.Getenv("LAGO_API_URL"))
	}
	run(t, binary, environment, initArgs...)
	prefix := fmt.Sprintf("cli-e2e-%d", time.Now().UTC().UnixNano())
	output := run(t, binary, environment, "--profile", "e2e", "--output", "json", "seed", "demo", "--prefix", prefix)

	var result map[string]any
	decoder := json.NewDecoder(bytes.NewReader(output))
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		t.Fatal(err)
	}
	invoice := fixtureInvoice(t, result)
	expectedData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "golden-invoice.json"))
	if err != nil {
		t.Fatal(err)
	}
	var expected map[string]any
	expectedDecoder := json.NewDecoder(bytes.NewReader(expectedData))
	expectedDecoder.UseNumber()
	if err := expectedDecoder.Decode(&expected); err != nil {
		t.Fatal(err)
	}
	actual := map[string]any{}
	for _, field := range []string{"currency", "fees_amount_cents", "sub_total_excluding_taxes_amount_cents", "taxes_amount_cents", "total_amount_cents"} {
		actual[field] = invoice[field]
	}
	fees, _ := invoice["fees"].([]any)
	actual["line_item_count"] = json.Number(fmt.Sprint(len(fees)))
	if fmt.Sprint(actual) != fmt.Sprint(expected) {
		actualJSON, _ := json.MarshalIndent(actual, "", "  ")
		t.Fatalf("golden invoice mismatch (sanitized totals only):\n%s", actualJSON)
	}

	variables, _ := result["variables"].(map[string]any)
	cleanup(binary, environment, variables)
}

func fixtureInvoice(t *testing.T, result map[string]any) map[string]any {
	t.Helper()
	steps, _ := result["steps"].([]any)
	for _, raw := range steps {
		step, _ := raw.(map[string]any)
		if step["id"] != "invoice-preview" {
			continue
		}
		response, _ := step["response"].(map[string]any)
		invoice, _ := response["invoice"].(map[string]any)
		if invoice != nil {
			return invoice
		}
	}
	t.Fatal("fixture returned no invoice preview")
	return nil
}

func cleanup(binary string, environment []string, variables map[string]any) {
	commands := [][]string{
		{"--profile", "e2e", "subscriptions", "terminate", fmt.Sprint(variables["subscription_id"]), "--confirm", fmt.Sprint(variables["subscription_id"])},
		{"--profile", "e2e", "customers", "delete", fmt.Sprint(variables["customer_id"]), "--confirm", fmt.Sprint(variables["customer_id"])},
		{"--profile", "e2e", "plans", "delete", fmt.Sprint(variables["plan_code"]), "--confirm", fmt.Sprint(variables["plan_code"])},
	}
	for _, arguments := range commands {
		command := exec.Command(binary, arguments...)
		command.Env = environment
		_ = command.Run()
	}
}

func run(t *testing.T, binary string, environment []string, arguments ...string) []byte {
	t.Helper()
	command := exec.Command(binary, arguments...)
	command.Env = environment
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("lago command failed: %v\n%s", err, stderr.String())
	}
	return stdout.Bytes()
}
