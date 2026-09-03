package transport

import (
	"net/url"
	"strings"
	"testing"

	"github.com/getlago/lago-cli/internal/apperr"
)

// One normalizer, one table, all three deployment targets.
//
// QA pasted the full API path and got a request to /api/v1/api/v1. Every shape an
// operator can plausibly paste is pinned here, for cloud US, cloud EU, and self-hosted
// alike: nothing in this function may special-case a hostname.
func TestNormalizeBaseURLAcceptsEveryDeploymentShape(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name     string
		raw      string
		insecure bool
		want     string
	}{
		// --- cloud US ------------------------------------------------------------
		{name: "us base", raw: "https://api.getlago.com", want: "https://api.getlago.com/api/v1"},
		{name: "us with prefix", raw: "https://api.getlago.com/api/v1", want: "https://api.getlago.com/api/v1"},
		{name: "us trailing slash", raw: "https://api.getlago.com/", want: "https://api.getlago.com/api/v1"},
		{name: "us prefix and trailing slash", raw: "https://api.getlago.com/api/v1/", want: "https://api.getlago.com/api/v1"},
		{name: "us doubled prefix from an older config", raw: "https://api.getlago.com/api/v1/api/v1", want: "https://api.getlago.com/api/v1"},
		{name: "us surrounded by whitespace", raw: "  https://api.getlago.com  ", want: "https://api.getlago.com/api/v1"},
		{name: "us with a pasted query and fragment", raw: "https://api.getlago.com/api/v1?token=leak#top", want: "https://api.getlago.com/api/v1"},

		// --- cloud EU ------------------------------------------------------------
		{name: "eu base", raw: "https://api.eu.getlago.com", want: "https://api.eu.getlago.com/api/v1"},
		{name: "eu with prefix", raw: "https://api.eu.getlago.com/api/v1", want: "https://api.eu.getlago.com/api/v1"},
		{name: "eu trailing slash", raw: "https://api.eu.getlago.com/", want: "https://api.eu.getlago.com/api/v1"},

		// --- self-hosted ---------------------------------------------------------
		{name: "self-hosted base", raw: "https://lago.acme.test", want: "https://lago.acme.test/api/v1"},
		{name: "self-hosted with prefix", raw: "https://lago.acme.test/api/v1", want: "https://lago.acme.test/api/v1"},
		{name: "self-hosted custom port", raw: "https://lago.acme.test:8443", want: "https://lago.acme.test:8443/api/v1"},
		{name: "self-hosted custom port with prefix", raw: "https://lago.acme.test:8443/api/v1", want: "https://lago.acme.test:8443/api/v1"},
		{name: "self-hosted sub-path behind a proxy", raw: "https://tools.acme.com/lago", want: "https://tools.acme.com/lago/api/v1"},
		{name: "self-hosted sub-path with prefix", raw: "https://tools.acme.com/lago/api/v1", want: "https://tools.acme.com/lago/api/v1"},
		{name: "self-hosted sub-path and trailing slash", raw: "https://tools.acme.com/lago/", want: "https://tools.acme.com/lago/api/v1"},
		{name: "self-hosted nested sub-path", raw: "https://tools.acme.com/billing/lago", want: "https://tools.acme.com/billing/lago/api/v1"},
		{name: "self-hosted uppercase host", raw: "https://LAGO.ACME.INTERNAL", want: "https://LAGO.ACME.INTERNAL/api/v1"},

		// --- local development ---------------------------------------------------
		{name: "localhost with --insecure", raw: "http://localhost:3000", insecure: true, want: "http://localhost:3000/api/v1"},
		{name: "localhost with prefix and --insecure", raw: "http://localhost:3000/api/v1", insecure: true, want: "http://localhost:3000/api/v1"},
		{name: "loopback address with --insecure", raw: "http://127.0.0.1:3000", insecure: true, want: "http://127.0.0.1:3000/api/v1"},
		{name: "https localhost needs no --insecure", raw: "https://localhost:3000", want: "https://localhost:3000/api/v1"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			parsed, err := NormalizeBaseURL(testCase.raw, testCase.insecure)
			if err != nil {
				t.Fatalf("NormalizeBaseURL(%q) failed: %v", testCase.raw, err)
			}
			if parsed.String() != testCase.want {
				t.Errorf("NormalizeBaseURL(%q) = %q, want %q", testCase.raw, parsed.String(), testCase.want)
			}
		})
	}
}

// Normalization must be idempotent. A URL the CLI wrote into a config file and read back
// has to normalize to itself, or the /api/v1 doubles on every round trip.
func TestNormalizationIsIdempotent(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{
		"https://api.getlago.com", "https://api.eu.getlago.com",
		"https://lago.acme.test:8443", "https://tools.acme.com/lago",
	} {
		once, err := NormalizeBaseURL(raw, false)
		if err != nil {
			t.Fatalf("%s: %v", raw, err)
		}
		twice, err := NormalizeBaseURL(once.String(), false)
		if err != nil {
			t.Fatalf("%s (second pass): %v", once, err)
		}
		if once.String() != twice.String() {
			t.Errorf("%s normalized to %q then %q", raw, once, twice)
		}
	}
}

// The region shorthand must resolve to exactly the same URL as passing the base
// explicitly, or `--region us` and `--api-url https://api.getlago.com` would be two
// different deployments that look like one.
func TestRegionShorthandMatchesTheExplicitForm(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct{ shorthand, explicit string }{
		{USAPI, "https://api.getlago.com"},
		{USAPI, "https://api.getlago.com/api/v1"},
		{EUAPI, "https://api.eu.getlago.com"},
		{EUAPI, "https://api.eu.getlago.com/api/v1"},
	} {
		fromRegion, err := NormalizeBaseURL(testCase.shorthand, false)
		if err != nil {
			t.Fatal(err)
		}
		fromFlag, err := NormalizeBaseURL(testCase.explicit, false)
		if err != nil {
			t.Fatal(err)
		}
		if fromRegion.String() != fromFlag.String() {
			t.Errorf("region shorthand %q resolved to %q but --api-url %q resolved to %q",
				testCase.shorthand, fromRegion, testCase.explicit, fromFlag)
		}
	}
}

// Pasting the dashboard URL is the mistake with the worst failure mode: the app host
// answers, so without this the CLI sends a live API key to the frontend and reports a
// confusing parse error. It must refuse and name the api.* host to use instead.
func TestAppURLIsRefusedByNameForEveryDeployment(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct{ raw, wantHost string }{
		{"https://app.getlago.com", "api.getlago.com"},
		{"https://app.getlago.com/", "api.getlago.com"},
		{"https://app.getlago.com/customers", "api.getlago.com"},
		{"https://app.eu.getlago.com", "api.eu.getlago.com"},
		{"https://APP.GETLAGO.COM", "api.getlago.com"},
		{"https://app.lago.acme.test", "api.lago.acme.test"},
		{"https://app.lago.acme.test:8443", "api.lago.acme.test"},
	} {
		_, err := NormalizeBaseURL(testCase.raw, false)
		if err == nil {
			t.Errorf("NormalizeBaseURL(%q) accepted the dashboard host", testCase.raw)
			continue
		}
		if apperr.ExitCode(err) != apperr.ExitUsage {
			t.Errorf("%q exit code = %d, want %d", testCase.raw, apperr.ExitCode(err), apperr.ExitUsage)
		}
		var appErr *apperr.Error
		if !asAppError(err, &appErr) {
			t.Fatalf("%q did not return a structured error: %v", testCase.raw, err)
		}
		if !strings.Contains(appErr.Suggestion, testCase.wantHost) {
			t.Errorf("%q suggestion does not name %q: %s", testCase.raw, testCase.wantHost, appErr.Suggestion)
		}
	}
}

// A host that merely contains "app" is not a dashboard. `apps.acme.com` and a
// single-domain self-hosted deployment must both still work.
func TestOnlyAnAppSubdomainIsTreatedAsTheDashboard(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{
		"https://apps.acme.com/lago",
		"https://lago.acme.com",
		"https://application.acme.com",
		"https://acme.com/app",
	} {
		if _, err := NormalizeBaseURL(raw, false); err != nil {
			t.Errorf("NormalizeBaseURL(%q) was refused as a dashboard: %v", raw, err)
		}
	}
}

// Inputs that cannot become a base URL must fail with a usage error that names the
// three deployment targets, not a bare "invalid URL".
func TestNormalizeBaseURLRejectsUnusableInput(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name     string
		raw      string
		insecure bool
		wantExit int
		wantText string
	}{
		{name: "empty", raw: "", wantExit: apperr.ExitAuth, wantText: "not configured"},
		{name: "whitespace only", raw: "   ", wantExit: apperr.ExitAuth, wantText: "not configured"},
		{name: "bare host with no scheme", raw: "api.getlago.com", wantExit: apperr.ExitUsage, wantText: "api.eu.getlago.com"},
		{name: "scheme with no host", raw: "https://", wantExit: apperr.ExitUsage, wantText: "absolute URL"},
		{name: "plain http without --insecure", raw: "http://lago.acme.test", wantExit: apperr.ExitUsage, wantText: "HTTPS"},
		{name: "localhost http without --insecure", raw: "http://localhost:3000", wantExit: apperr.ExitUsage, wantText: "insecure"},
		{name: "non-http scheme", raw: "ftp://lago.acme.test", wantExit: apperr.ExitUsage, wantText: "HTTPS"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			_, err := NormalizeBaseURL(testCase.raw, testCase.insecure)
			if err == nil {
				t.Fatalf("NormalizeBaseURL(%q) was accepted", testCase.raw)
			}
			if apperr.ExitCode(err) != testCase.wantExit {
				t.Errorf("exit code = %d, want %d (%v)", apperr.ExitCode(err), testCase.wantExit, err)
			}
			var appErr *apperr.Error
			if !asAppError(err, &appErr) {
				t.Fatalf("not a structured error: %v", err)
			}
			combined := appErr.Message + " " + appErr.Suggestion
			if !strings.Contains(combined, testCase.wantText) {
				t.Errorf("error does not mention %q: %s", testCase.wantText, combined)
			}
		})
	}
}

// A path pasted straight out of the API reference already carries /api/v1. The base URL
// carries it too, so one of them has to go: QA's /api/v1/api/v1 request is this bug.
func TestRequestPathDropsARedundantPrefix(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct{ raw, want string }{
		{"/customers", "/customers"},
		{"customers", "customers"},
		{"/api/v1/customers", "/customers"},
		{"api/v1/customers", "/customers"},
		{"/api/v1/", "/"},
		{"/api/v1", "/"},
		{"/api/v1/invoices/inv_1/finalize", "/invoices/inv_1/finalize"},
		// Only the leading occurrence is a paste artefact.
		{"/customers/api/v1", "/customers/api/v1"},
		// A path that merely starts with similar text is untouched.
		{"/api/v10/customers", "/api/v10/customers"},
		{"/api/v2/customers", "/api/v2/customers"},
		{"/apiv1/customers", "/apiv1/customers"},
	} {
		if got := NormalizeRequestPath(testCase.raw); got != testCase.want {
			t.Errorf("NormalizeRequestPath(%q) = %q, want %q", testCase.raw, got, testCase.want)
		}
	}
}

// End to end: whichever base shape and whichever path shape an operator combines, the
// request must land on exactly one /api/v1, on every deployment target.
func TestResolvedRequestURLNeverDoublesThePrefix(t *testing.T) {
	t.Parallel()
	for _, base := range []string{
		"https://api.getlago.com", "https://api.getlago.com/api/v1",
		"https://api.eu.getlago.com", "https://api.eu.getlago.com/api/v1/",
		"https://lago.acme.test:8443", "https://tools.acme.com/lago/api/v1",
	} {
		client, err := New(Config{BaseURL: base, APIKey: "fake-key"})
		if err != nil {
			t.Fatalf("New(%q) failed: %v", base, err)
		}
		for _, path := range []string{"/customers", "/api/v1/customers"} {
			resolved, err := client.resolve(path, url.Values{"page": {"2"}})
			if err != nil {
				t.Fatalf("%s + %s: %v", base, path, err)
			}
			if strings.Count(resolved.Path, "/api/v1") != 1 {
				t.Errorf("base %q with path %q resolved to %q", base, path, resolved.Path)
			}
			if !strings.HasSuffix(resolved.Path, "/api/v1/customers") {
				t.Errorf("base %q with path %q resolved to %q, want it to end in /api/v1/customers", base, path, resolved.Path)
			}
			if resolved.Query().Get("page") != "2" {
				t.Errorf("base %q lost the query string: %s", base, resolved)
			}
		}
	}
}

func asAppError(err error, target **apperr.Error) bool {
	appErr, ok := err.(*apperr.Error)
	if ok {
		*target = appErr
	}
	return ok
}

// A bare 404 has to become a message naming what was not found, whether the request
// addressed one identifier or several.
func TestDescribeNotFoundNamesEveryIdentifier(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name     string
		subjects []Subject
		want     []string
	}{
		{
			name:     "one identifier reads as a sentence",
			subjects: []Subject{{Kind: "subscription", Value: "ai_plan_gpt4_tokens"}},
			want:     []string{`no subscription "ai_plan_gpt4_tokens" exists`},
		},
		{
			name: "several identifiers say which set contains the miss",
			subjects: []Subject{
				{Kind: "customer", Value: "cus_1"},
				{Kind: "plan", Value: "quickstart"},
			},
			want: []string{"not found:", `customer "cus_1"`, `plan "quickstart"`, "one of these"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got, ok := describeNotFound("", testCase.subjects)
			if !ok {
				t.Fatalf("describeNotFound(%+v) declined to describe", testCase.subjects)
			}
			for _, want := range testCase.want {
				if !strings.Contains(got, want) {
					t.Errorf("describeNotFound(%+v) = %q, missing %q", testCase.subjects, got, want)
				}
			}
		})
	}
}

// QA run 3: `invoices create` with an unknown add-on code answered 404 with the Lago code
// add_on_not_found, and the CLI said `no customer "x" exists` because the customer was
// the only identifier it knew about. The Lago code names what is missing; the subjects
// only supply the value when they agree with it.
func TestQA_E5_NotFoundNamesTheResourceFromTheLagoCode(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name     string
		code     string
		subjects []Subject
		want     string
		reject   []string
	}{
		{
			name:     "code names a resource no subject carries",
			code:     "add_on_not_found",
			subjects: []Subject{{Kind: "customer", Value: "qa_cust"}},
			want:     "no matching add on exists in this organization",
			reject:   []string{"customer", "qa_cust"},
		},
		{
			name:     "code names one of several subjects",
			code:     "plan_not_found",
			subjects: []Subject{{Kind: "customer", Value: "cus_1"}, {Kind: "plan", Value: "quickstart"}},
			want:     `no plan "quickstart" exists in this organization`,
			reject:   []string{"cus_1"},
		},
		{
			name:     "multi-word kinds match snake_case codes",
			code:     "billable_metric_not_found",
			subjects: []Subject{{Kind: "billable metric", Value: "requests"}},
			want:     `no billable metric "requests" exists in this organization`,
		},
		{
			name:     "no code falls back to the subjects",
			code:     "",
			subjects: []Subject{{Kind: "subscription", Value: "sub_1"}},
			want:     `no subscription "sub_1" exists in this organization`,
		},
		{
			name:     "a code of another shape falls back to the subjects",
			code:     "not_found",
			subjects: []Subject{{Kind: "customer", Value: "cus_1"}},
			want:     `no customer "cus_1" exists in this organization`,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got, ok := describeNotFound(testCase.code, testCase.subjects)
			if !ok {
				t.Fatalf("describeNotFound declined to describe %q %+v", testCase.code, testCase.subjects)
			}
			if got != testCase.want {
				t.Errorf("describeNotFound(%q, %+v) = %q, want %q", testCase.code, testCase.subjects, got, testCase.want)
			}
			for _, bad := range testCase.reject {
				if strings.Contains(got, bad) {
					t.Errorf("message %q names %q, which the Lago code rules out", got, bad)
				}
			}
		})
	}
	// A 404 with a resource code and no subjects at all is still described.
	if got, ok := describeNotFound("wallet_not_found", nil); !ok || got != "no matching wallet exists in this organization" {
		t.Errorf("code without subjects = %q, %v", got, ok)
	}
	// A 404 with neither is left to the API's own message.
	if _, ok := describeNotFound("", nil); ok {
		t.Error("nothing to say, yet describeNotFound described")
	}
}

// responseError only rewrites a 404, and only when it knows which identifiers were
// addressed. Everything else keeps the API's own message.
func TestResponseErrorOnlyRewritesAKnownNotFound(t *testing.T) {
	t.Parallel()
	notFound := &Response{Status: 404, Body: []byte(`{"status":404,"error":"Not Found","code":"not_found"}`)}
	validation := &Response{Status: 422, Body: []byte(`{"status":422,"error":"code is invalid"}`)}
	subjects := []Subject{{Kind: "plan", Value: "quickstart"}}

	if err := responseError(notFound, subjects); !strings.Contains(err.Error(), `no plan "quickstart"`) {
		t.Errorf("a 404 with subjects was not rewritten: %v", err)
	}
	if err := responseError(notFound, nil); err.Error() != "Not Found" {
		t.Errorf("a 404 with no subjects should keep the API message, got %v", err)
	}
	if err := responseError(validation, subjects); !strings.Contains(err.Error(), "code is invalid") {
		t.Errorf("a 422 was rewritten as not-found: %v", err)
	}
}

// QA S-16, N-9: `https://api.getlago.com@evil.example` parses with a host of
// evil.example and userinfo of api.getlago.com, so the printed host and the dialled host
// disagreed. Userinfo is refused outright, and a pasted password never reaches the error.
func TestQA_S16_UserinfoInBaseURLIsRejected(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{
		"https://api.getlago.com@evil.example",
		"https://user:lago_test_FAKEpassword@api.getlago.com",
		"https://token@billing.acme.test/api/v1",
		"https://:lago_test_FAKEpassword@api.eu.getlago.com",
	} {
		_, err := NormalizeBaseURL(raw, false)
		if err == nil {
			t.Errorf("%s was accepted", raw)
			continue
		}
		if apperr.ExitCode(err) != apperr.ExitUsage {
			t.Errorf("%s: exit code = %d, want %d", raw, apperr.ExitCode(err), apperr.ExitUsage)
		}
		if !strings.Contains(err.Error(), "embeds credentials") {
			t.Errorf("%s: error does not name the problem: %v", raw, err)
		}
		if strings.Contains(err.Error(), "FAKEpassword") {
			t.Errorf("%s: the password was echoed: %v", raw, err)
		}
	}
}

// QA run 4: voiding an already voided invoice answers 405, which Lago uses for "the
// resource's state forbids this", not for a wrong HTTP verb. The suggestion says so
// instead of sending the operator to check command flags.
func TestQA_L2g_MethodNotAllowedExplainsResourceState(t *testing.T) {
	t.Parallel()
	code, suggestion := classify(405)
	if code != apperr.ExitValidation {
		t.Errorf("405 exit code = %d, want %d", code, apperr.ExitValidation)
	}
	if !strings.Contains(suggestion, "state") || strings.Contains(suggestion, "command flags") {
		t.Errorf("405 suggestion does not explain resource state: %q", suggestion)
	}
}
