package contract

import (
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
		if field.Flag == "subscription-ending-at" && !field.Nullable {
			t.Error("subscription.ending_at is not recorded as nullable")
		}
	}
}
