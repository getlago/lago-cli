// Package billing_test is the offline billing-accuracy harness. It runs on every PR
// against an httptest server, unlike test/e2e/golden_test.go, which is build-tagged
// and skips without trusted staging credentials — meaning the only billing harness
// this repository had could never fail in CI.
//
// Every assertion here is about money: exact totals, zero-decimal currencies,
// idempotent replay, and the JSON number precision that carries amounts to the
// terminal. These are the failures that cost customers real money.
package billing_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/getlago/lago-cli/internal/cli"
	"github.com/getlago/lago-cli/internal/config"
)

const testAPIKey = "lago_test_FAKE000000000000000000000000"

// newCLI returns a runner bound to a test-mode profile pointing at server.
func newCLI(t *testing.T, serverURL string) func(stdin string, argv ...string) (string, string, error) {
	t.Helper()
	for _, name := range []string{"LAGO_API_KEY", "LAGO_API_URL", "LAGO_MODE", "LAGO_PROFILE", "LAGO_TIMEOUT", "LAGO_DEBUG"} {
		t.Setenv(name, "")
	}
	path := filepath.Join(t.TempDir(), "config.toml")
	file := config.File{
		Version:        1,
		CurrentProfile: "default",
		Profiles: map[string]config.Profile{
			"default": {Region: "self-hosted", APIKey: testAPIKey, APIURL: serverURL, Mode: config.ModeTest, Insecure: true},
		},
	}
	if err := config.Save(path, file); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LAGO_CONFIG_FILE", path)

	return func(stdin string, argv ...string) (string, string, error) {
		var stdout, stderr bytes.Buffer
		app := cli.NewApp(strings.NewReader(stdin), &stdout, &stderr, "billing-test")
		root := cli.NewRoot(app)
		root.SetArgs(argv)
		root.SetOut(&stdout)
		root.SetErr(&stderr)
		err := root.Execute()
		return stdout.String(), stderr.String(), err
	}
}

func jsonServer(t *testing.T, handler func(http.ResponseWriter, *http.Request)) *httptest.Server {
	t.Helper()
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		handler(response, request)
	}))
	t.Cleanup(server.Close)
	return server
}

// The golden invoice must render byte-for-byte. Mutating a single fee cent must
// break it: a golden file that tolerates a changed amount tests nothing.
func TestGoldenInvoiceTotalsRenderExactly(t *testing.T) {
	golden, err := os.ReadFile(filepath.Join("..", "..", "testdata", "golden-invoice.json"))
	if err != nil {
		t.Fatal(err)
	}
	var expected map[string]any
	decoder := json.NewDecoder(bytes.NewReader(golden))
	decoder.UseNumber()
	if err := decoder.Decode(&expected); err != nil {
		t.Fatal(err)
	}

	for _, testCase := range []struct {
		name      string
		feeCents  int64
		wantMatch bool
	}{
		{"golden totals match", 100, true},
		{"one cent more breaks the golden", 101, false},
		{"one cent less breaks the golden", 99, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			server := jsonServer(t, func(response http.ResponseWriter, _ *http.Request) {
				fmt.Fprintf(response, `{"invoice":{"currency":"USD","fees_amount_cents":%d,`+
					`"sub_total_excluding_taxes_amount_cents":%d,"taxes_amount_cents":0,`+
					`"total_amount_cents":%d,"fees":[{"amount_cents":%d}]}}`,
					testCase.feeCents, testCase.feeCents, testCase.feeCents, testCase.feeCents)
			})
			run := newCLI(t, server.URL)
			stdout, _, err := run("", "--output", "json", "api", "GET", "/invoices/inv_1")
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			var payload map[string]any
			d := json.NewDecoder(strings.NewReader(stdout))
			d.UseNumber()
			if err := d.Decode(&payload); err != nil {
				t.Fatalf("decode %q: %v", stdout, err)
			}
			invoice, _ := payload["invoice"].(map[string]any)
			if invoice == nil {
				t.Fatalf("no invoice in %q", stdout)
			}

			matched := true
			for _, field := range []string{"currency", "fees_amount_cents", "sub_total_excluding_taxes_amount_cents", "taxes_amount_cents", "total_amount_cents"} {
				if fmt.Sprint(invoice[field]) != fmt.Sprint(expected[field]) {
					matched = false
				}
			}
			if matched != testCase.wantMatch {
				t.Fatalf("golden match = %v, want %v (fee %d cents)", matched, testCase.wantMatch, testCase.feeCents)
			}
		})
	}
}

// Zero-decimal currencies carry no minor unit: 3000 JPY is 3000, not 30.00.
// The amount must survive the round trip with its digits unchanged.
func TestZeroDecimalAndHighPrecisionAmountsSurviveUnchanged(t *testing.T) {
	for _, testCase := range []struct{ currency, amount string }{
		{"JPY", "3000"},             // zero-decimal
		{"KRW", "150000"},           // zero-decimal
		{"USD", "100"},              // two-decimal minor units
		{"BHD", "1234"},             // three-decimal minor units
		{"EUR", "9007199254740993"}, // beyond exact float64
		{"USD", "0"},                // zero
	} {
		t.Run(testCase.currency+"_"+testCase.amount, func(t *testing.T) {
			server := jsonServer(t, func(response http.ResponseWriter, _ *http.Request) {
				fmt.Fprintf(response, `{"invoice":{"currency":%q,"total_amount_cents":%s}}`, testCase.currency, testCase.amount)
			})
			run := newCLI(t, server.URL)

			stdout, _, err := run("", "--output", "json", "api", "GET", "/invoices/inv_1")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(stdout, `"total_amount_cents": `+testCase.amount) {
				t.Errorf("JSON output altered the amount:\n%s", stdout)
			}

			yamlOut, _, err := run("", "--output", "yaml", "api", "GET", "/invoices/inv_1")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(yamlOut, "total_amount_cents: "+testCase.amount) {
				t.Errorf("YAML output altered or quoted the amount:\n%s", yamlOut)
			}
		})
	}
}

// A retried event must carry the same Idempotency-Key, so the server can collapse
// the duplicate. A replay that mints a new key produces a double charge.
func TestEventReplayReusesTheSameIdempotencyKey(t *testing.T) {
	var mutex sync.Mutex
	keys := map[string]int{}
	attempts := 0

	server := jsonServer(t, func(response http.ResponseWriter, request *http.Request) {
		mutex.Lock()
		attempts++
		current := attempts
		keys[request.Header.Get("Idempotency-Key")]++
		mutex.Unlock()
		if current == 1 {
			response.WriteHeader(http.StatusInternalServerError)
			_, _ = response.Write([]byte(`{"status":500,"error":"Internal Server Error"}`))
			return
		}
		_, _ = response.Write([]byte(`{"event":{"lago_id":"evt_1"}}`))
	})

	path := filepath.Join(t.TempDir(), "events.ndjson")
	if err := os.WriteFile(path, []byte(`{"transaction_id":"txn_1","external_subscription_id":"sub_1","code":"requests"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	run := newCLI(t, server.URL)
	if _, stderr, err := run("", "--output", "json", "events", "send", "--file", path); err != nil {
		t.Fatalf("event send failed: %v\n%s", err, stderr)
	}

	mutex.Lock()
	defer mutex.Unlock()
	if attempts < 2 {
		t.Fatalf("server saw %d attempts; the retry never happened", attempts)
	}
	if len(keys) != 1 {
		t.Fatalf("retry used %d distinct idempotency keys, want 1: %v", len(keys), keys)
	}
	for key, count := range keys {
		if key == "" {
			t.Fatal("event was sent with no Idempotency-Key")
		}
		if count != attempts {
			t.Fatalf("key %s seen %d times across %d attempts", key, count, attempts)
		}
	}
}

// A mutation the CLI cannot prove is safe to replay must be sent exactly once.
// Retrying an invoice finalize or a payment retry can charge a customer twice.
func TestMoneyMovingMutationsAreSentExactlyOnce(t *testing.T) {
	for _, testCase := range []struct{ name, method, path string }{
		{"finalize", "PUT", "/invoices/inv_1/finalize"},
		{"void", "POST", "/invoices/inv_1/void"},
		{"retry payment", "POST", "/invoices/inv_1/retry_payment"},
		{"void credit note", "PUT", "/credit_notes/cn_1/void"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var mutex sync.Mutex
			attempts := 0
			server := jsonServer(t, func(response http.ResponseWriter, _ *http.Request) {
				mutex.Lock()
				attempts++
				mutex.Unlock()
				response.WriteHeader(http.StatusInternalServerError)
				_, _ = response.Write([]byte(`{"status":500,"error":"Internal Server Error"}`))
			})
			run := newCLI(t, server.URL)

			if _, _, err := run("", "api", testCase.method, testCase.path); err == nil {
				t.Fatal("expected the 500 to surface as an error")
			}

			mutex.Lock()
			defer mutex.Unlock()
			if attempts != 1 {
				t.Fatalf("%s %s was sent %d times; a money-moving mutation must not be auto-retried",
					testCase.method, testCase.path, attempts)
			}
		})
	}
}

// Reads may be retried freely: they move no money and a transient 500 should not
// surface to the operator.
func TestReadsAreStillRetried(t *testing.T) {
	var mutex sync.Mutex
	attempts := 0
	server := jsonServer(t, func(response http.ResponseWriter, _ *http.Request) {
		mutex.Lock()
		attempts++
		current := attempts
		mutex.Unlock()
		if current == 1 {
			response.WriteHeader(http.StatusInternalServerError)
			_, _ = response.Write([]byte(`{"status":500}`))
			return
		}
		_, _ = response.Write([]byte(`{"invoice":{"lago_id":"inv_1"}}`))
	})
	run := newCLI(t, server.URL)
	if _, _, err := run("", "api", "GET", "/invoices/inv_1"); err != nil {
		t.Fatalf("read was not retried past a transient 500: %v", err)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if attempts < 2 {
		t.Fatalf("read was attempted %d times, want at least 2", attempts)
	}
}
