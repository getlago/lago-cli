package contract

import (
	"strings"
	"testing"

	"github.com/getlago/lago-cli/internal/generated"
)

func findOperation(t *testing.T, operationID string) generated.Operation {
	t.Helper()
	for _, operation := range generated.Operations {
		if operation.OperationID == operationID {
			return operation
		}
	}
	t.Fatalf("operation %s is not in the generated table", operationID)
	return generated.Operation{}
}

// QA M-subscriptions-u: a nullable field must never be a required flag. Walks every
// generated body so the trap cannot come back through another endpoint.
func TestQA_L2f_NoFieldIsBothRequiredAndNullable(t *testing.T) {
	t.Parallel()
	nullable := 0
	for _, operation := range generated.Operations {
		if operation.Body == nil {
			continue
		}
		for _, field := range operation.Body.Fields {
			if field.Nullable {
				nullable++
			}
			if field.Required && field.Nullable {
				t.Errorf("%s --%s is both required and nullable", operation.OperationID, field.Flag)
			}
		}
	}
	if nullable == 0 {
		t.Fatal("no nullable field was recorded; the union handling is not running")
	}
}

// QA L-2f, M-optional-body: bodies are optional exactly where the spec says so.
func TestQA_MOptionalBody_OptionalBodiesMatchTheSpec(t *testing.T) {
	t.Parallel()
	for operationID, want := range map[string]bool{
		"voidInvoice":        false,
		"retryPayment":       false,
		"createCustomer":     true,
		"createSubscription": true,
		"createPlan":         true,
	} {
		operation := findOperation(t, operationID)
		if operation.Body == nil {
			t.Errorf("%s has no body", operationID)
			continue
		}
		if operation.Body.Required != want {
			t.Errorf("%s body required = %v, want %v", operationID, operation.Body.Required, want)
		}
	}
	required, optional := 0, 0
	for _, operation := range generated.Operations {
		if operation.Body == nil {
			continue
		}
		if operation.Body.Required {
			required++
		} else {
			optional++
		}
	}
	if required < 60 || optional < 5 {
		t.Errorf("required=%d optional=%d bodies; the spec has 65 required and 10 optional", required, optional)
	}
}

func TestQA_MSubscriptionsU_UpdateSubscriptionDoesNotRequireEndingAt(t *testing.T) {
	t.Parallel()
	operation := findOperation(t, "updateSubscription")
	for _, field := range operation.Body.Fields {
		if field.Required {
			t.Errorf("subscriptions update still requires --%s", field.Flag)
		}
		if field.Flag == "ending-at" && !field.Nullable {
			t.Error("subscription.ending_at is not recorded as nullable")
		}
	}
}

// QA type coercion: every minor-unit amount in a request body is an integer. The one
// spec exception is `precise_total_amount_cents` on events, a decimal string by design,
// which is why the check keys off the `precise_` prefix rather than a hand list.
func TestQA_TypeCoercion_MonetaryFieldsAreIntegers(t *testing.T) {
	t.Parallel()
	integers := 0
	for _, operation := range generated.Operations {
		if operation.Body == nil {
			continue
		}
		for _, field := range operation.Body.Fields {
			last := field.Path[len(field.Path)-1]
			if !strings.HasSuffix(last, "_cents") || strings.HasPrefix(last, "precise_") {
				continue
			}
			if field.Type != "integer" {
				t.Errorf("%s --%s is %q, want integer", operation.OperationID, field.Flag, field.Type)
			}
			integers++
		}
	}
	if integers < 20 {
		t.Fatalf("only %d *_cents fields found; the walk is not reaching bodies", integers)
	}
}
