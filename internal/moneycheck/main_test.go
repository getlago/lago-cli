package main

import (
	"go/ast"
	"go/parser"
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
