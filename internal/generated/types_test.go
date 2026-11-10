package generated

import "testing"

func TestMustDecodeOperationsPanicsOnACorruptTable(t *testing.T) {
	t.Parallel()
	defer func() {
		if recover() == nil {
			t.Fatal("a corrupt operation table decoded without panicking")
		}
	}()
	mustDecodeOperations("{not json")
}

func TestMustDecodeOperationsDecodesATable(t *testing.T) {
	t.Parallel()
	operations := mustDecodeOperations(`[{"resource":"products","action":"list","api_version":"v2","cursor_paginated":true}]`)
	if len(operations) != 1 || operations[0].APIVersion != "v2" || !operations[0].CursorPaginated {
		t.Fatalf("decoded %+v", operations)
	}
}
