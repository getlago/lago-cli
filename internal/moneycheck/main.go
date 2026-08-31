package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

var monetaryWords = []string{"amount", "balance", "credit", "currencyunit", "fee", "minorunit", "money", "price", "subtotal", "tax", "total"}

func main() {
	root := "."
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") && os.Args[1] != "./..." {
		root = os.Args[1]
	}
	files := token.NewFileSet()
	violations := 0
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
							position := files.Position(typed.Pos())
							fmt.Fprintf(os.Stderr, "%s:%d: monetary field %s must use int64 MinorUnits or exact Decimal, never float\n", path, position.Line, name.Name)
							violations++
						}
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

func isFloat(expression ast.Expr) bool {
	identifier, ok := expression.(*ast.Ident)
	return ok && (identifier.Name == "float32" || identifier.Name == "float64")
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
