package transport

import (
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/getlago/lago-cli/internal/apperr"
)

// A v2 request must swap the one /api/v1 segment for /api/v2 on every deployment shape
// NormalizeBaseURL accepts, and a path pasted from the v2 reference must select v2 on
// its own, so `lago api GET /api/v2/products` works without a flag.
func TestV2RequestsResolveUnderAPIV2OnEveryDeployment(t *testing.T) {
	t.Parallel()
	for _, base := range []string{
		"https://api.getlago.com", "https://api.getlago.com/api/v1",
		"https://api.eu.getlago.com/api/v1/", "https://lago.acme.test:8443",
		"https://tools.acme.com/lago/api/v1",
	} {
		client, err := New(Config{BaseURL: base, APIKey: "fake-key"})
		if err != nil {
			t.Fatalf("New(%q) failed: %v", base, err)
		}
		normalized, _ := NormalizeBaseURL(base, false)
		want := normalized.Path[:len(normalized.Path)-len(APIPrefix)] + "/api/v2/products"
		for _, tc := range []struct{ version, path string }{
			{"v2", "/products"},
			{"", "/api/v2/products"},
			{"v2", "/api/v2/products"},
			{"", "api/v2/products"},
		} {
			resolved, err := client.resolve(tc.version, tc.path, url.Values{"limit": {"5"}})
			if err != nil {
				t.Fatalf("%s + %+v: %v", base, tc, err)
			}
			if resolved.Path != want {
				t.Errorf("base %q with %+v resolved to %q, want %q", base, tc, resolved.Path, want)
			}
			if resolved.Query().Get("limit") != "5" {
				t.Errorf("base %q lost the query string: %s", base, resolved)
			}
		}
		v1, err := client.resolve("", "/customers", nil)
		if err != nil || v1.Path != normalized.Path+"/customers" {
			t.Errorf("v1 request on %q resolved to %v (%v)", base, v1, err)
		}
	}
}

// API v2 errors carry a code and field details but no message. The message is the
// humanized code (the field lines are printed from the details, as for v1), the details
// keep the v1 field-to-codes shape, and the raw entries are kept for JSON output.
func TestV2ValidationErrorIsDescribed(t *testing.T) {
	t.Parallel()
	err := responseError(&Response{Status: http.StatusUnprocessableEntity, Body: []byte(`{"code":"validation_errors","details":[{"code":"value_is_mandatory","field":"name","meta":{}},{"code":"value_already_exist","field":"code","meta":{}}]}`)}, nil)
	var appError *apperr.Error
	if !asAppError(err, &appError) {
		t.Fatalf("error is %T", err)
	}
	if appError.Message != "validation errors" {
		t.Errorf("message = %q", appError.Message)
	}
	if appError.Code != "validation_errors" {
		t.Errorf("code = %q", appError.Code)
	}
	want := map[string]any{"name": []any{"value_is_mandatory"}, "code": []any{"value_already_exist"}}
	if !reflect.DeepEqual(appError.Details, want) {
		t.Errorf("details = %#v, want %#v", appError.Details, want)
	}
	if len(appError.FieldErrors) != 2 || appError.FieldErrors[0]["field"] != "name" || appError.FieldErrors[0]["meta"] != nil {
		t.Errorf("field errors = %#v", appError.FieldErrors)
	}
}

// A declined card is the case the meta exists for: the provider's own code must reach
// the message and the JSON payload, not just "provider error".
func TestV2ProviderErrorKeepsTheProviderCode(t *testing.T) {
	t.Parallel()
	body := []byte(`{"code":"provider_error","details":[{"code":"provider_error","field":"base","meta":{"provider_name":"stripe","provider_code":"stripe_eu","provider_error_code":"card_declined"}}]}`)
	err := responseError(&Response{Status: http.StatusUnprocessableEntity, Body: body}, nil)
	var appError *apperr.Error
	if !asAppError(err, &appError) {
		t.Fatalf("error is %T", err)
	}
	if appError.Message != "provider error: card_declined" {
		t.Errorf("message = %q", appError.Message)
	}
	meta, _ := appError.FieldErrors[0]["meta"].(map[string]any)
	if meta["provider_name"] != "stripe" || meta["provider_error_code"] != "card_declined" {
		t.Errorf("field errors = %#v", appError.FieldErrors)
	}
	encoded := string(apperr.Encode(err))
	for _, want := range []string{`"field_errors"`, `"provider_error_code": "card_declined"`, `"details"`} {
		if !strings.Contains(encoded, want) {
			t.Errorf("JSON error is missing %s:\n%s", want, encoded)
		}
	}
}

func TestV2ErrorWithoutDetailsUsesTheCode(t *testing.T) {
	t.Parallel()
	err := responseError(&Response{Status: http.StatusForbidden, Body: []byte(`{"code":"feature_unavailable","details":[]}`)}, nil)
	var appError *apperr.Error
	if !asAppError(err, &appError) || appError.Message != "feature unavailable" {
		t.Fatalf("error = %#v", err)
	}
}

// The v2 404 names the resource in details[].meta.resource. It must produce the same
// "no product \"x\" exists" message the v1 `<resource>_not_found` code does.
func TestV2NotFoundNamesTheIdentifier(t *testing.T) {
	t.Parallel()
	body := []byte(`{"code":"resource_not_found","details":[{"code":"resource_not_found","field":"base","meta":{"resource":"product"}}]}`)
	err := responseError(&Response{Status: http.StatusNotFound, Body: body}, []Subject{{Kind: "product", Value: "compute"}})
	var appError *apperr.Error
	if !asAppError(err, &appError) {
		t.Fatalf("error is %T", err)
	}
	if appError.Message != `no product "compute" exists in this organization` {
		t.Errorf("message = %q", appError.Message)
	}
	if got := v2NotFoundCode("resource_not_found", nil); got != "resource_not_found" {
		t.Errorf("v2NotFoundCode without details = %q", got)
	}
}
