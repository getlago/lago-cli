package cli

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// QA X-3: delete, terminate and apply dumped the full resource, contradicting the terse
// contract every create and update already followed. A state transition now prints the
// identifiers plus the new status.
func TestQA_X3_DeleteAndTerminatePrintIdentifiers(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, request *http.Request) {
		switch {
		case strings.Contains(request.URL.Path, "/subscriptions/"):
			_, _ = response.Write([]byte(`{"subscription":{"lago_id":"sub_1","external_id":"quickstart_subscription","status":"terminated","plan_code":"quickstart","current_billing_period_started_at":"2026-09-01T00:00:00Z","terminated_at":"2026-09-02T00:00:00Z"}}`))
		case strings.Contains(request.URL.Path, "/applied_coupons"):
			_, _ = response.Write([]byte(`{"applied_coupon":{"lago_id":"ac_1","external_customer_id":"cus_1","coupon_code":"WELCOME","status":"active","amount_cents":1000}}`))
		default:
			_, _ = response.Write([]byte(`{"customer":{"lago_id":"cus_1","external_id":"acme","name":"Acme","currency":"USD","created_at":"2026-01-01T00:00:00Z"}}`))
		}
	})
	profileAt(t, server.URL)

	stdout, _, err := execute(t, "", "subscriptions", "terminate", "quickstart_subscription", "--confirm", "quickstart_subscription")
	if err != nil {
		t.Fatalf("terminate failed: %v", err)
	}
	for _, want := range []string{"LAGO_ID      sub_1", "EXTERNAL_ID  quickstart_subscription", "STATUS       terminated"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("terminate output missing %q:\n%s", want, stdout)
		}
	}
	for _, absent := range []string{"CURRENT_BILLING_PERIOD_STARTED_AT", "TERMINATED_AT", "PLAN_CODE"} {
		if strings.Contains(stdout, absent) {
			t.Errorf("terminate output still dumps %s:\n%s", absent, stdout)
		}
	}

	stdout, _, err = execute(t, "", "customers", "delete", "acme", "--confirm", "acme")
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if !strings.Contains(stdout, "LAGO_ID") || strings.Contains(stdout, "CREATED_AT") {
		t.Errorf("delete output is not terse:\n%s", stdout)
	}

	stdout, _, err = execute(t, "", "coupons", "apply", "--external-customer-id", "cus_1", "--coupon-code", "WELCOME")
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if !strings.Contains(strings.Join(strings.Fields(stdout), " "), "STATUS active") || strings.Contains(stdout, "AMOUNT_CENTS") {
		t.Errorf("apply output is not terse:\n%s", stdout)
	}
}

// QA M-empty: `wallets create-wallet-transaction` answers with an array and rendered as
// a two-line LAGO_ID block; it is one row per transaction with the status.
func TestQA_MEmpty_WalletTransactionsRenderOneRowPerItem(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`{"wallet_transactions":[{"lago_id":"wt_1","status":"settled","transaction_type":"inbound","amount":"10.0"},{"lago_id":"wt_2","status":"pending","transaction_type":"inbound","amount":"5.0"}]}`))
	})
	profileAt(t, server.URL)
	stdout, _, err := execute(t, "", "wallets", "create-wallet-transaction", "--wallet-id", "w_1", "--paid-credits", "10.0", "--granted-credits", "5.0")
	if err != nil {
		t.Fatalf("create-wallet-transaction failed: %v", err)
	}
	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("want header + 2 rows:\n%s", stdout)
	}
	if got := strings.Join(strings.Fields(lines[0]), " "); got != "LAGO_ID STATUS" {
		t.Errorf("header = %q", got)
	}
	if !strings.Contains(lines[1], "wt_1") || !strings.Contains(lines[2], "pending") {
		t.Errorf("rows:\n%s", stdout)
	}
}

// QA C-3, S-9: whoami printed the whole organization as a JSON blob in one cell. It is
// a short identity block in reading order, ending with the host requests go to.
func TestQA_C3_WhoamiIdentityBlock(t *testing.T) {
	server := jsonAPI(t, func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`{"organization":{"lago_id":"org_1","name":"Example Organization","default_currency":"EUR","timezone":"Europe/Paris","email":"billing@acme.test","billing_configuration":{"invoice_footer":"x"}}}`))
	})
	profileAt(t, server.URL)

	stdout, _, err := execute(t, "", "whoami")
	if err != nil {
		t.Fatalf("whoami failed: %v", err)
	}
	var keys []string
	for _, line := range strings.Split(strings.TrimRight(stdout, "\n"), "\n") {
		keys = append(keys, strings.Fields(line)[0])
	}
	if got := strings.Join(keys, " "); got != "NAME LAGO_ID DEFAULT_CURRENCY TIMEZONE PROFILE MODE RESOLVED_API_URL" {
		t.Errorf("whoami rows = %q:\n%s", got, stdout)
	}
	for _, want := range []string{"Example Organization", "org_1", "EUR", "Europe/Paris", "default", "test", server.URL + "/api/v1"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("whoami missing %q:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "{") || strings.Contains(stdout, "ORGANIZATION") || strings.Contains(stdout, "billing@") {
		t.Errorf("whoami still dumps the organization:\n%s", stdout)
	}

	stdout, _, err = execute(t, "", "whoami", "--output", "json")
	if err != nil {
		t.Fatalf("whoami --output json failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatal(err)
	}
	organization, _ := payload["organization"].(map[string]any)
	if organization["lago_id"] != "org_1" || organization["email"] != "billing@acme.test" {
		t.Errorf("JSON organization is not the object itself: %v", payload["organization"])
	}
	for _, key := range []string{"profile", "mode", "region", "api_url", "resolved_api_url"} {
		if _, ok := payload[key]; !ok {
			t.Errorf("JSON output lacks %s", key)
		}
	}
}
