package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

var monetaryWords = []string{
	"amount", "balance", "cents", "charge", "cost", "coupon", "credit", "currencyunit",
	"discount", "fee", "invoice", "minorunit", "money", "payment", "price", "rate",
	"refund", "subtotal", "tax", "total", "vat", "wallet",
}

// buildContext controls which files the checker considers part of the build.
// MONEYCHECK_BUILD_TAGS opts extra tags in; TestCanaryIsRejected uses it to reach
// the deliberately-defective canary package that `make lint` must not see.
var buildContext = newBuildContext()

func newBuildContext() build.Context {
	context := build.Default
	for _, tag := range strings.Fields(os.Getenv("MONEYCHECK_BUILD_TAGS")) {
		context.BuildTags = append(context.BuildTags, tag)
	}
	return context
}

func main() {
	root := "."
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") && os.Args[1] != "./..." {
		root = os.Args[1]
	}
	files := token.NewFileSet()
	violations := 0
	// root is argv[1], typed by the maintainer running `make lint`. The checker only reads
	// Go source under it and writes nothing, so a traversal here reaches files the same
	// person could already open.
	// #nosec G703 -- maintainer-supplied scan root in a read-only linter.
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "vendor" || entry.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		// Skip files the normal build excludes by tag. This is how the deliberate
		// defects in internal/moneycheck/canary stay out of `make lint` while still
		// being reachable by TestCanaryIsRejected, which scans them directly.
		if included, matchErr := buildContext.MatchFile(filepath.Dir(path), filepath.Base(path)); matchErr == nil && !included {
			return nil
		}
		file, parseErr := parser.ParseFile(files, path, nil, 0)
		if parseErr != nil {
			return parseErr
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch typed := node.(type) {
			case *ast.Field:
				if isFloat(typed.Type) {
					for _, name := range typed.Names {
						if monetaryName(name.Name) {
							report(path, files.Position(typed.Pos()).Line, "field", name.Name, &violations)
						}
					}
				}
			case *ast.TypeSpec:
				// type Money float64
				if isFloat(typed.Type) && monetaryName(typed.Name.Name) {
					report(path, files.Position(typed.Pos()).Line, "type", typed.Name.Name, &violations)
				}
			case *ast.ValueSpec:
				// var totalAmount float64 = 12.34
				for _, name := range typed.Names {
					if !monetaryName(name.Name) {
						continue
					}
					if isFloat(typed.Type) {
						report(path, files.Position(typed.Pos()).Line, "declaration", name.Name, &violations)
						continue
					}
					for _, value := range typed.Values {
						if isFloatingExpression(value) {
							report(path, files.Position(typed.Pos()).Line, "declaration", name.Name, &violations)
							break
						}
					}
				}
			case *ast.FuncDecl:
				// func TotalAmount() float64 — results are unnamed Fields, so the
				// Field case above never sees them. Attribute them to the function.
				if typed.Type.Results == nil || !monetaryName(typed.Name.Name) {
					break
				}
				for _, result := range typed.Type.Results.List {
					if isFloat(result.Type) && len(result.Names) == 0 {
						report(path, files.Position(typed.Pos()).Line, "return value of", typed.Name.Name, &violations)
					}
				}
			case *ast.AssignStmt:
				for index, left := range typed.Lhs {
					identifier, ok := left.(*ast.Ident)
					if ok && monetaryName(identifier.Name) && index < len(typed.Rhs) && isFloatingExpression(typed.Rhs[index]) {
						position := files.Position(typed.Pos())
						fmt.Fprintf(os.Stderr, "%s:%d: monetary assignment %s uses floating-point arithmetic\n", path, position.Line, identifier.Name)
						violations++
					}
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "moneycheck:", err)
		os.Exit(1)
	}
	if violations > 0 {
		os.Exit(1)
	}
	fmt.Println("moneycheck: no floating-point monetary fields")
}

func report(path string, line int, kind, name string, violations *int) {
	fmt.Fprintf(os.Stderr, "%s:%d: monetary %s %s must use int64 MinorUnits or exact Decimal, never float\n", path, line, kind, name)
	*violations++
}

func isFloatingExpression(expression ast.Expr) bool {
	switch typed := expression.(type) {
	case *ast.BasicLit:
		return typed.Kind == token.FLOAT
	case *ast.CallExpr:
		return isFloat(typed.Fun)
	case *ast.BinaryExpr:
		return isFloatingExpression(typed.X) || isFloatingExpression(typed.Y)
	case *ast.ParenExpr:
		return isFloatingExpression(typed.X)
	default:
		return false
	}
}

// isFloat reports whether an expression denotes a floating-point type, looking
// through the wrappers an API struct actually uses: *float64 for optional JSON
// fields, []float64 for repeated ones, and map[string]float64 for keyed ones.
func isFloat(expression ast.Expr) bool {
	switch typed := expression.(type) {
	case *ast.Ident:
		return typed.Name == "float32" || typed.Name == "float64"
	case *ast.StarExpr:
		return isFloat(typed.X)
	case *ast.ArrayType:
		return isFloat(typed.Elt)
	case *ast.MapType:
		return isFloat(typed.Key) || isFloat(typed.Value)
	case *ast.Ellipsis:
		return isFloat(typed.Elt)
	case *ast.ParenExpr:
		return isFloat(typed.X)
	}
	return false
}

func monetaryName(name string) bool {
	normalized := strings.ToLower(name)
	for _, word := range monetaryWords {
		if strings.Contains(normalized, word) {
			return true
		}
	}
	return false
}
