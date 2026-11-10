package transport

import (
	"net/http"
	"net/url"
	"reflect"
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

// API v2 errors carry a code and field details but no message. The CLI must still print
// what failed instead of the raw body, and keep the v1 details shape for JSON output.
func TestV2ValidationErrorIsDescribed(t *testing.T) {
	t.Parallel()
	err := responseError(&Response{Status: http.StatusUnprocessableEntity, Body: []byte(`{"code":"validation_errors","details":[{"code":"value_is_mandatory","field":"name","meta":{}},{"code":"value_already_exist","field":"code","meta":{}}]}`)}, nil)
	var appError *apperr.Error
	if !asAppError(err, &appError) {
		t.Fatalf("error is %T", err)
	}
	if appError.Message != "validation errors: name: value_is_mandatory, code: value_already_exist" {
		t.Errorf("message = %q", appError.Message)
	}
	if appError.Code != "validation_errors" {
		t.Errorf("code = %q", appError.Code)
	}
	want := map[string]any{"name": []any{"value_is_mandatory"}, "code": []any{"value_already_exist"}}
	if !reflect.DeepEqual(appError.Details, want) {
		t.Errorf("details = %#v, want %#v", appError.Details, want)
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
