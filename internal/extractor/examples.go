package extractor

import (
	"go/ast"
	"strings"
)

// ExtractExamples extracts test case examples from table-driven tests
func ExtractExamples(fn *ast.FuncDecl, tableStruct *ast.StructType) []Example {
	if tableStruct == nil {
		return nil
	}

	testCases := findTestCases(fn)
	if len(testCases) == 0 {
		return nil
	}

	var examples []Example
	for _, tc := range testCases {
		example := extractExample(tc, tableStruct)
		if example.Name != "" {
			examples = append(examples, example)
		}
	}

	return examples
}

// findTestCases finds test case composite literals in the table
func findTestCases(fn *ast.FuncDecl) []*ast.CompositeLit {
	var testCases []*ast.CompositeLit

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		// Look for: tests := []struct{...}{...}
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		// Check if LHS is "tests"
		if len(assign.Lhs) != 1 {
			return true
		}

		lhsIdent, ok := assign.Lhs[0].(*ast.Ident)
		if !ok || lhsIdent.Name != "tests" {
			return true
		}

		// Get RHS composite literal
		if len(assign.Rhs) != 1 {
			return true
		}

		comp, ok := assign.Rhs[0].(*ast.CompositeLit)
		if !ok {
			return true
		}

		// Extract test case literals
		for _, elt := range comp.Elts {
			if caseLit, ok := elt.(*ast.CompositeLit); ok {
				testCases = append(testCases, caseLit)
			}
		}

		return false
	})

	return testCases
}

// extractExample extracts a single example from a test case literal
func extractExample(caseLit *ast.CompositeLit, tableStruct *ast.StructType) Example {
	fields := extractFieldValues(caseLit)

	// Get the "name" field
	name, ok := fields["name"]
	if !ok {
		return Example{}
	}

	// Extract argument values (skip metadata fields)
	args := extractArgumentValues(fields, tableStruct)

	// Try to get expected result from "want" field
	result := fields["want"]

	return Example{
		Name:   name,
		Args:   args,
		Result: result,
	}
}

// extractFieldValues extracts field values from a composite literal
func extractFieldValues(comp *ast.CompositeLit) map[string]string {
	fields := make(map[string]string)

	for _, elt := range comp.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}

		value := formatLiteralValue(kv.Value)
		fields[key.Name] = value
	}

	return fields
}

// extractArgumentValues extracts argument values, skipping metadata fields
func extractArgumentValues(fields map[string]string, tableStruct *ast.StructType) []string {
	metadataFields := map[string]bool{
		"name":    true,
		"want":    true,
		"wantErr": true,
	}

	var args []string

	// Iterate through struct fields in order
	for _, field := range tableStruct.Fields.List {
		for _, fieldName := range field.Names {
			name := fieldName.Name
			if metadataFields[name] {
				continue
			}

			if value, ok := fields[name]; ok {
				args = append(args, value)
			}
		}
	}

	return args
}

// formatLiteralValue formats a literal expression as a TypeScript value
func formatLiteralValue(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		// String, number, or boolean literal
		return e.Value

	case *ast.Ident:
		// Boolean or constant
		if e.Name == "true" || e.Name == "false" {
			return e.Name
		}
		return e.Name

	case *ast.CompositeLit:
		// Array or object literal
		return formatCompositeLiteral(e)

	case *ast.UnaryExpr:
		// Handle negative numbers
		if e.Op.String() == "-" {
			if lit, ok := e.X.(*ast.BasicLit); ok {
				return "-" + lit.Value
			}
		}
		return "..."

	default:
		return "..."
	}
}

// formatCompositeLiteral formats a composite literal as TypeScript
func formatCompositeLiteral(comp *ast.CompositeLit) string {
	// Check if it's an array
	if _, ok := comp.Type.(*ast.ArrayType); ok {
		return formatArrayLiteral(comp)
	}

	// Otherwise treat as object
	return formatObjectLiteral(comp)
}

// formatArrayLiteral formats an array literal
func formatArrayLiteral(comp *ast.CompositeLit) string {
	if len(comp.Elts) == 0 {
		return "[]"
	}

	var elements []string
	for _, elt := range comp.Elts {
		elements = append(elements, formatLiteralValue(elt))
	}

	return "[" + strings.Join(elements, ", ") + "]"
}

// formatObjectLiteral formats an object literal
func formatObjectLiteral(comp *ast.CompositeLit) string {
	if len(comp.Elts) == 0 {
		return "{}"
	}

	var pairs []string
	for _, elt := range comp.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			key := formatLiteralValue(kv.Key)
			value := formatLiteralValue(kv.Value)
			pairs = append(pairs, key+": "+value)
		}
	}

	if len(pairs) == 0 {
		return "{}"
	}

	return "{" + strings.Join(pairs, ", ") + "}"
}
