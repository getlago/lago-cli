package main

import "testing"

func TestCommandNamingHelpers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, got, want string
	}{
		{"snake case", normalizeName("billable_metrics"), "billable-metrics"},
		{"camel case", normalizeName("invoicePreview"), "invoice-preview"},
		{"plural", singular("customers"), "customer"},
		{"summary", derivedAction("createEvent", "events", "/events", "Send usage events"), "send"},
		{"custom path", derivedAction("invoicePreview", "invoices", "/invoices/preview", "Create an invoice preview"), "preview"},
		{"qualified", qualifiedAction("findAllAppliedCoupons", "coupons", false), "list-applied"},
		{"qualified keeps the noun behind a qualifier", qualifiedAction("findAllAppliedCoupons", "coupons", true), "list-applied-coupons"},
		{"qualified bare verb keeps its short name", qualifiedAction("applyCoupon", "coupons", true), "apply"},
		{"scoped create names what it creates", derivedAction("createCustomerWallet", "wallets", "/customers/{external_customer_id}/wallets", "Create a customer wallet"), "create-customer-wallet"},
		{"scoped destroy names what it destroys", derivedAction("destroySubscriptionEntitlement", "entitlements", "/subscriptions/{external_id}/entitlements/{feature_code}", "Remove"), "destroy-subscription-entitlement"},
		{"scoped bare verb", derivedAction("createEntitlement", "entitlements", "/plans/{code}/entitlements", "Create"), "create"},
		{"scoped apply", derivedAction("applyCoupon", "coupons", "/applied_coupons", "Apply a coupon"), "apply"},
	}
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("%s = %q, want %q", test.name, test.got, test.want)
		}
	}
}

func TestGeneratorRejectsMissingPaths(t *testing.T) {
	t.Parallel()
	if _, err := (generator{document: map[string]any{}}).operations(); err == nil {
		t.Fatal("missing paths unexpectedly generated")
	}
}

// The terse-output classification is a generator rule, so it is tested here rather than
// re-asserted at 109 call sites. Every write is terse; the excluded cases are the ones
// that would be wrong to reduce: read-shaped POSTs whose body is the answer, and bulk
// ingestion whose output is a summary.
func TestMutationClassification(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		method, action string
		want           bool
	}{
		{"POST", "create", true},
		{"PUT", "update", true},
		{"PATCH", "update", true},
		{"POST", "create-plan-charge", true},
		{"PUT", "update-subscription-alert", true},
		{"PATCH", "update-subscription", true},
		{"DELETE", "delete", true}, // QA X-3: deletes print identifiers plus status
		{"DELETE", "terminate", true},
		{"DELETE", "destroy-plan-charge", true},
		{"POST", "apply", true},
		{"PUT", "finalize", true}, // the new state is the status column
		{"POST", "void", true},
		{"POST", "execute", true},
		{"PUT", "refresh", true},
		{"PATCH", "merge-plan-metadata", true},
		{"POST", "created", true}, // any write not excluded is terse
		{"POST", "updates", true},

		{"GET", "create", false}, // a read is never terse, whatever it is called
		{"HEAD", "create", false},
		{"POST", "preview", false}, // invoices preview: the body is the answer
		{"POST", "estimate", false},
		{"POST", "estimate-fees", false},
		{"POST", "estimate-instant-fees", false},
		{"POST", "batch-estimate-instant-fees", false},
		{"POST", "evaluate-expression", false},
		{"POST", "send", false}, // events send: the summary is the answer
		{"POST", "batch", false},
		{"POST", "download", false},
		{"POST", "checkout-url", false},
		{"POST", "payment-url", false},
		{"POST", "wallet-transaction-payment-url", false},
	} {
		if got := mutationOperation(test.method, test.action); got != test.want {
			t.Errorf("mutationOperation(%q, %q) = %v, want %v", test.method, test.action, got, test.want)
		}
	}
}

// QA M-subscriptions-u: `ending_at` is `type: [string, 'null']` and listed under
// `required`. The union must split into its base type and a nullable flag.
func TestQA_L2f_NullableTypeSplitsUnions(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		schema   map[string]any
		wantType string
		wantNull bool
	}{
		{map[string]any{"type": "integer"}, "integer", false},
		{map[string]any{"type": []any{"integer", "null"}}, "integer", true},
		{map[string]any{"type": []any{"null", "string"}}, "string", true},
		{map[string]any{"type": []any{"boolean"}}, "boolean", false},
		{map[string]any{"type": []any{"integer", "string"}}, "", false},
		{map[string]any{"type": []any{"integer", "string", "null"}}, "", true},
		{map[string]any{"type": "null"}, "null", true},
		{map[string]any{"properties": map[string]any{}}, "", false},
		{nil, "", false},
	} {
		gotType, gotNull := nullableType(test.schema)
		if gotType != test.wantType || gotNull != test.wantNull {
			t.Errorf("nullableType(%v) = (%q, %v), want (%q, %v)", test.schema, gotType, gotNull, test.wantType, test.wantNull)
		}
	}
}

// QA L-2f, M-optional-body: Body.Required follows requestBody.required, whose OpenAPI
// default is false. A nullable field listed under `required` is not a required flag.
func TestQA_MOptionalBody_BodyRequiredFollowsRequestBodyRequired(t *testing.T) {
	t.Parallel()
	schema := map[string]any{
		"type":     "object",
		"required": []any{"thing"},
		"properties": map[string]any{
			"thing": map[string]any{
				"type":     "object",
				"required": []any{"code", "ending_at"},
				"properties": map[string]any{
					"code":      map[string]any{"type": "string"},
					"ending_at": map[string]any{"type": []any{"string", "null"}},
				},
			},
		},
	}
	document := map[string]any{
		"openapi": "3.1.0",
		"info":    map[string]any{"version": "test"},
		"paths": map[string]any{
			"/things": map[string]any{
				"post": map[string]any{"operationId": "createThing", "tags": []any{"things"}, "summary": "Create", "requestBody": map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": schema}}}},
				"put":  map[string]any{"operationId": "updateThing", "tags": []any{"things"}, "summary": "Update", "requestBody": map[string]any{"required": false, "content": map[string]any{"application/json": map[string]any{"schema": schema}}}},
			},
			"/things/{code}/void": map[string]any{
				"post": map[string]any{"operationId": "voidThing", "tags": []any{"things"}, "summary": "Void", "parameters": []any{map[string]any{"name": "code", "in": "path", "required": true, "schema": map[string]any{"type": "string"}}}, "requestBody": map[string]any{"content": map[string]any{"application/json": map[string]any{"schema": schema}}}},
			},
		},
	}
	operations, err := (generator{document: document}).operations()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"createThing": true, "updateThing": false, "voidThing": false}
	for _, operation := range operations {
		wantRequired, known := want[operation.OperationID]
		if !known {
			t.Fatalf("unexpected operation %s", operation.OperationID)
		}
		if operation.Body == nil || operation.Body.Required != wantRequired {
			t.Errorf("%s body required = %v, want %v", operation.OperationID, operation.Body != nil && operation.Body.Required, wantRequired)
		}
		for _, field := range operation.Body.Fields {
			switch field.Flag {
			case "code":
				if !field.Required || field.Nullable {
					t.Errorf("%s --code required=%v nullable=%v", operation.OperationID, field.Required, field.Nullable)
				}
			case "ending-at":
				if field.Required || !field.Nullable {
					t.Errorf("%s --ending-at must be optional and nullable, got required=%v nullable=%v", operation.OperationID, field.Required, field.Nullable)
				}
			}
		}
	}
}

// QA type coercion: `coupons create --amount-cents 1` sent `"1"` because the field is
// `type: [integer, 'null']` and the generator only read a plain string type. Unions
// resolve to their non-null member; a oneOf/anyOf resolves only when its scalar members
// agree, so `event.timestamp` (`oneOf [integer, string]`) stays a string.
func TestQA_TypeCoercion_SchemaTypeResolvesUnions(t *testing.T) {
	t.Parallel()
	g := generator{document: map[string]any{"components": map[string]any{"schemas": map[string]any{
		"Cents": map[string]any{"type": "integer"},
	}}}}
	for _, test := range []struct {
		schema map[string]any
		want   string
	}{
		{map[string]any{"type": "integer"}, "integer"},
		{map[string]any{"type": []any{"integer", "null"}}, "integer"},
		{map[string]any{"type": []any{"null", "number"}}, "number"},
		{map[string]any{"type": []any{"boolean", "null"}}, "boolean"},
		{map[string]any{"type": []any{"object", "null"}, "properties": map[string]any{"a": map[string]any{"type": "string"}}}, "object"},
		{map[string]any{"type": []any{"array", "null"}, "items": map[string]any{"type": "string"}}, "array"},
		{map[string]any{"oneOf": []any{map[string]any{"type": "integer"}, map[string]any{"type": "string"}}}, "string"},
		{map[string]any{"anyOf": []any{map[string]any{"type": "integer"}, map[string]any{"$ref": "#/components/schemas/Cents"}}}, "integer"},
		{map[string]any{"oneOf": []any{map[string]any{"type": "integer"}, map[string]any{"type": "null"}}}, "integer"},
		{map[string]any{"oneOf": []any{map[string]any{"type": "object", "properties": map[string]any{}}, map[string]any{"type": "integer"}}}, "string"},
		{map[string]any{"type": "string", "pattern": "^[0-9]+.?[0-9]*$"}, "string"},
		{map[string]any{}, "string"},
		{nil, "string"},
	} {
		if got := g.schemaType(test.schema); got != test.want {
			t.Errorf("schemaType(%v) = %q, want %q", test.schema, got, test.want)
		}
	}
}
