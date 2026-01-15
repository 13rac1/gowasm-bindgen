package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// ParseTestFiles parses Go test files matching the given patterns and returns the AST
func ParseTestFiles(patterns []string) ([]*ast.File, *token.FileSet, error) {
	fset := token.NewFileSet()
	var files []*ast.File

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid pattern %q: %w", pattern, err)
		}

		for _, path := range matches {
			file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse %s: %w", path, err)
			}
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		return nil, nil, fmt.Errorf("no files matched patterns: %v", patterns)
	}

	return files, fset, nil
}

// FindTestFunctions finds all Test* functions in the given AST files
func FindTestFunctions(files []*ast.File) []*ast.FuncDecl {
	var testFuncs []*ast.FuncDecl

	for _, file := range files {
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if strings.HasPrefix(fn.Name.Name, "Test") {
				testFuncs = append(testFuncs, fn)
			}

			return true
		})
	}

	return testFuncs
}

// FindWASMCalls finds WASM function calls in a test function body
// Pattern: funcName(js.Null(), []js.Value{...})
func FindWASMCalls(fn *ast.FuncDecl, fset *token.FileSet) []WASMCall {
	var calls []WASMCall

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		// Look for assignment statements
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		// Check if RHS is a function call
		if len(assign.Rhs) != 1 {
			return true
		}

		call, ok := assign.Rhs[0].(*ast.CallExpr)
		if !ok {
			return true
		}

		// Extract result variable name
		resultVar := ""
		if len(assign.Lhs) == 1 {
			if ident, ok := assign.Lhs[0].(*ast.Ident); ok {
				resultVar = ident.Name
			}
		}

		// Try to match the WASM call pattern
		wasmCall := matchWASMCallPattern(call, fset)
		if wasmCall == nil {
			return true
		}

		wasmCall.ResultVar = resultVar
		wasmCall.TestFunc = fn.Name.Name
		wasmCall.SourceFile = fset.Position(fn.Pos()).Filename
		wasmCall.Line = fset.Position(call.Pos()).Line

		calls = append(calls, *wasmCall)
		return true
	})

	return calls
}

// matchWASMCallPattern checks if a call expression matches the WASM pattern:
// funcName(js.Null(), []js.Value{...})
func matchWASMCallPattern(call *ast.CallExpr, fset *token.FileSet) *WASMCall {
	// Must be a simple function call (identifier)
	funcIdent, ok := call.Fun.(*ast.Ident)
	if !ok {
		return nil
	}

	// Must have exactly 2 arguments
	if len(call.Args) != 2 {
		return nil
	}

	// First argument should be js.Null()
	if !isJsNullCall(call.Args[0]) {
		return nil
	}

	// Second argument should be []js.Value{...}
	args := extractJsValueSlice(call.Args[1], fset)
	if args == nil {
		return nil
	}

	return &WASMCall{
		FuncName: funcIdent.Name,
		Args:     args,
	}
}

// isJsNullCall checks if an expression is js.Null()
func isJsNullCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	x, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	return x.Name == "js" && sel.Sel.Name == "Null"
}

// extractJsValueSlice extracts arguments from []js.Value{...} composite literal
func extractJsValueSlice(expr ast.Expr, fset *token.FileSet) []Argument {
	comp, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}

	// Check if type is []js.Value
	if !isJsValueSliceType(comp.Type) {
		return nil
	}

	var args []Argument

	for _, elt := range comp.Elts {
		arg := extractArgument(elt, fset)
		args = append(args, arg)
	}

	return args
}

// isJsValueSliceType checks if a type is []js.Value
func isJsValueSliceType(typ ast.Expr) bool {
	arrayType, ok := typ.(*ast.ArrayType)
	if !ok {
		return false
	}

	sel, ok := arrayType.Elt.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	x, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	return x.Name == "js" && sel.Sel.Name == "Value"
}

// extractArgument extracts argument info from js.ValueOf(...) call
func extractArgument(expr ast.Expr, fset *token.FileSet) Argument {
	// Handle js.ValueOf(x)
	call, ok := expr.(*ast.CallExpr)
	if ok && isJsValueOfCall(call) && len(call.Args) > 0 {
		innerExpr := call.Args[0]
		return Argument{
			Expression: exprToString(innerExpr, fset),
			GoType:     InferTypeFromLiteral(innerExpr),
		}
	}

	// Fallback: use the expression directly
	return Argument{
		Expression: exprToString(expr, fset),
		GoType:     InferTypeFromLiteral(expr),
	}
}

// isJsValueOfCall checks if an expression is js.ValueOf(...)
func isJsValueOfCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	x, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	return x.Name == "js" && sel.Sel.Name == "ValueOf"
}

// exprToString converts an AST expression to its source code string
func exprToString(expr ast.Expr, fset *token.FileSet) string {
	if expr == nil {
		return ""
	}

	start := expr.Pos()
	end := expr.End()

	if !start.IsValid() || !end.IsValid() {
		return formatExprFallback(expr)
	}

	startPos := fset.Position(start)
	endPos := fset.Position(end)

	// For simple cases, reconstruct the expression
	if startPos.Filename != endPos.Filename {
		return formatExprFallback(expr)
	}

	// Read from source would require the file content
	// For now, use a simple formatter
	return formatExprFallback(expr)
}

// formatExprFallback formats an expression without source access
func formatExprFallback(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return e.Value
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		// Handle tt.field pattern
		if x, ok := e.X.(*ast.Ident); ok {
			return x.Name + "." + e.Sel.Name
		}
		return "..."
	case *ast.CompositeLit:
		return "{...}" // Simplified
	default:
		return "..."
	}
}
