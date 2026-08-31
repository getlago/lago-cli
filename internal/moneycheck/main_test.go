package main

import (
	"bytes"
	"errors"
	"go/ast"
	"go/parser"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonetaryNameAndFloatDetection(t *testing.T) {
	t.Parallel()
	if !monetaryName("TotalAmountCents") || monetaryName("AttemptCount") {
		t.Fatal("monetary name classification changed")
	}
	if !isFloat(ast.NewIdent("float64")) || isFloat(ast.NewIdent("int64")) {
		t.Fatal("float type classification changed")
	}
	expression, err := parser.ParseExpr("100 * 1.5")
	if err != nil || !isFloatingExpression(expression) {
		t.Fatalf("floating expression was not detected: %v", err)
	}
}

// TestCanaryIsRejected runs the checker against internal/moneycheck/canary, which
// holds one deliberate defect per class the checker must catch. If this test passes
// silently the checker has regressed into a no-op.
func TestCanaryIsRejected(t *testing.T) {
	t.Parallel()
	binary := filepath.Join(t.TempDir(), "moneycheck")
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build moneycheck: %v\n%s", err, output)
	}

	command := exec.Command(binary, "./canary")
	command.Env = append(os.Environ(), "MONEYCHECK_BUILD_TAGS=moneycheck_canary")
	var stderr bytes.Buffer
	command.Stderr = &stderr
	err := command.Run()

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() == 0 {
		t.Fatalf("moneycheck accepted the canary; it must exit non-zero\nstderr: %s", stderr.String())
	}

	for _, want := range []string{
		"AmountCents", "TotalDueCents", "FeeByCurrency", "Cents",
		"Rate", "Charge", "Discount", "Money", "totalAmount", "TotalAmount",
	} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("moneycheck did not flag %s\nstderr: %s", want, stderr.String())
		}
	}
	for _, unwanted := range []string{"VatAmount", "AttemptCount"} {
		if strings.Contains(stderr.String(), unwanted) {
			t.Errorf("moneycheck falsely flagged %s", unwanted)
		}
	}
}
