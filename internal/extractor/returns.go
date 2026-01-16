package extractor

import (
	"go/ast"
	"strings"
)

// ExtractReturnType infers return type from result usage in test
func ExtractReturnType(fn *ast.FuncDecl, resultVar string) ReturnType {
	if resultVar == "" {
		return ReturnType{Type: "void"}
	}

	// Find all variables that are type assertions of the result
	resultVars := findResultVariables(fn, resultVar)

	isUnion, unionVariants := DetectUnionType(fn, resultVars)
	if isUnion {
		return ReturnType{
			Type: strings.Join(unionVariants, " | "),
		}
	}

	fields := extractReturnFields(fn, resultVars)
	if len(fields) > 0 {
		return ReturnType{
			Type:   formatObjectType(fields),
			Fields: fields,
		}
	}

	// Check for primitive return type
	primitiveType := detectPrimitiveReturn(fn, resultVars)
	if primitiveType != "" {
		return ReturnType{Type: primitiveType}
	}

	return ReturnType{Type: "any"}
}

// findResultVariables finds all variables that reference the result (including type assertions)
func findResultVariables(fn *ast.FuncDecl, resultVar string) []string {
	vars := []string{resultVar}
	seen := make(map[string]bool)
	seen[resultVar] = true

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		// Look for: jsResult := result.(js.Value)
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		if len(assign.Rhs) != 1 {
			return true
		}

		// Check if RHS is a type assertion
		typeAssert, ok := assign.Rhs[0].(*ast.TypeAssertExpr)
		if !ok {
			return true
		}

		// Check if X is the result variable
		xIdent, ok := typeAssert.X.(*ast.Ident)
		if !ok || xIdent.Name != resultVar {
			return true
		}

		// Add the LHS variable
		if len(assign.Lhs) == 1 {
			if lhsIdent, ok := assign.Lhs[0].(*ast.Ident); ok {
				if !seen[lhsIdent.Name] {
					vars = append(vars, lhsIdent.Name)
					seen[lhsIdent.Name] = true
				}
			}
		}

		return true
	})

	return vars
}

// DetectUnionType checks for error/success union patterns
func DetectUnionType(fn *ast.FuncDecl, resultVars []string) (bool, []string) {
	if len(resultVars) == 0 {
		return false, nil
	}

	var hasErrorBranch, hasSuccessBranch bool
	var errorFields, successFields []Field

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}

		// Check for: if !result.Get("error").IsUndefined()
		for _, resultVar := range resultVars {
			if isErrorCheckPattern(ifStmt.Cond, resultVar) {
				hasErrorBranch = true
				errorFields = extractReturnFieldsFromNode(ifStmt.Body, resultVars)
				if ifStmt.Else != nil {
					hasSuccessBranch = true
					if block, ok := ifStmt.Else.(*ast.BlockStmt); ok {
						successFields = extractReturnFieldsFromNode(block, resultVars)
					}
				}
				return false
			}
		}

		return true
	})

	if !hasErrorBranch || !hasSuccessBranch {
		return false, nil
	}

	variants := []string{
		formatObjectType(errorFields),
		formatObjectType(successFields),
	}

	return true, variants
}

// isErrorCheckPattern checks if condition matches: !result.Get("error").IsUndefined()
func isErrorCheckPattern(cond ast.Expr, resultVar string) bool {
	unary, ok := cond.(*ast.UnaryExpr)
	if !ok || unary.Op.String() != "!" {
		return false
	}

	call, ok := unary.X.(*ast.CallExpr)
	if !ok {
		return false
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "IsUndefined" {
		return false
	}

	// Check for result.Get("error")
	getCall, ok := sel.X.(*ast.CallExpr)
	if !ok {
		return false
	}

	getSel, ok := getCall.Fun.(*ast.SelectorExpr)
	if !ok || getSel.Sel.Name != "Get" {
		return false
	}

	ident, ok := getSel.X.(*ast.Ident)
	if !ok || ident.Name != resultVar {
		return false
	}

	// Check if argument is "error"
	if len(getCall.Args) == 0 {
		return false
	}

	lit, ok := getCall.Args[0].(*ast.BasicLit)
	if !ok {
		return false
	}

	return strings.Trim(lit.Value, `"'`) == "error"
}

// extractReturnFields extracts fields from result.Get("field").Type() patterns
func extractReturnFields(fn *ast.FuncDecl, resultVars []string) []Field {
	return extractReturnFieldsFromNode(fn.Body, resultVars)
}

// extractReturnFieldsFromNode extracts fields from a specific AST node
func extractReturnFieldsFromNode(node ast.Node, resultVars []string) []Field {
	var fields []Field
	seen := make(map[string]bool)

	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		for _, resultVar := range resultVars {
			field := extractFieldAccess(call, resultVar)
			if field.Name != "" && !seen[field.Name] {
				fields = append(fields, field)
				seen[field.Name] = true
			}
		}

		return true
	})

	return fields
}

// extractFieldAccess extracts field name and type from result.Get("field").Type()
func extractFieldAccess(call *ast.CallExpr, resultVar string) Field {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return Field{}
	}

	// Get the type from method name (Bool, String, Int, etc.)
	typeName := sel.Sel.Name
	tsType := methodToTSType(typeName)
	if tsType == "" {
		return Field{}
	}

	// Check if X is result.Get("fieldName")
	getCall, ok := sel.X.(*ast.CallExpr)
	if !ok {
		return Field{}
	}

	getSel, ok := getCall.Fun.(*ast.SelectorExpr)
	if !ok || getSel.Sel.Name != "Get" {
		return Field{}
	}

	ident, ok := getSel.X.(*ast.Ident)
	if !ok || ident.Name != resultVar {
		return Field{}
	}

	// Extract field name from Get("fieldName")
	if len(getCall.Args) == 0 {
		return Field{}
	}

	lit, ok := getCall.Args[0].(*ast.BasicLit)
	if !ok {
		return Field{}
	}

	fieldName := strings.Trim(lit.Value, `"'`)

	return Field{
		Name: fieldName,
		Type: tsType,
	}
}

// detectPrimitiveReturn detects primitive returns like result.String()
func detectPrimitiveReturn(fn *ast.FuncDecl, resultVars []string) string {
	var primitiveType string

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		// Check if ident matches any result variable
		for _, resultVar := range resultVars {
			if ident.Name == resultVar {
				tsType := methodToTSType(sel.Sel.Name)
				if tsType != "" {
					primitiveType = tsType
					return false
				}
			}
		}

		return true
	})

	return primitiveType
}

// methodToTSType converts js.Value method name to TypeScript type
func methodToTSType(method string) string {
	switch method {
	case "Bool":
		return "boolean"
	case "String":
		return "string"
	case "Int", "Float":
		return "number"
	default:
		return ""
	}
}

// formatObjectType formats fields into a TypeScript object type
func formatObjectType(fields []Field) string {
	if len(fields) == 0 {
		return "any"
	}

	var parts []string
	for _, field := range fields {
		parts = append(parts, field.Name+": "+field.Type)
	}

	return "{" + strings.Join(parts, ", ") + "}"
}
