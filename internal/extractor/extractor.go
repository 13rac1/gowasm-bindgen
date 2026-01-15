package extractor

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/13rac1/go-wasm-ts-gen/internal/parser"
)

// ExtractSignatures extracts function signatures from parsed test files
func ExtractSignatures(files []*ast.File, fset *token.FileSet) ([]FunctionSignature, error) {
	testFuncs := parser.FindTestFunctions(files)

	var signatures []FunctionSignature

	for _, fn := range testFuncs {
		calls := parser.FindWASMCalls(fn, fset)

		for _, call := range calls {
			sig := extractSignature(call, fn, fset)
			signatures = append(signatures, sig)
		}
	}

	return signatures, nil
}

// extractSignature extracts a complete function signature from a WASM call
func extractSignature(call parser.WASMCall, fn *ast.FuncDecl, fset *token.FileSet) FunctionSignature {
	params := ExtractParameters(call, fn)
	returns := ExtractReturnType(fn, call.ResultVar)
	examples := ExtractExamples(fn, FindTableStruct(fn))
	doc := extractDocComment(fn)

	return FunctionSignature{
		Name:       call.FuncName,
		Params:     params,
		Returns:    returns,
		Examples:   examples,
		Doc:        doc,
		TestFunc:   call.TestFunc,
		SourceFile: call.SourceFile,
		Line:       call.Line,
	}
}

// ExtractParameters extracts parameter names and types for a WASM call
func ExtractParameters(call parser.WASMCall, fn *ast.FuncDecl) []Parameter {
	tableStruct := FindTableStruct(fn)
	if tableStruct == nil {
		return extractParametersFallback(call)
	}

	return extractParametersFromTable(call, tableStruct)
}

// extractParametersFromTable maps js.ValueOf(tt.field) to struct fields
func extractParametersFromTable(call parser.WASMCall, tableStruct *ast.StructType) []Parameter {
	var params []Parameter

	for _, arg := range call.Args {
		// Extract field name from tt.fieldName pattern
		fieldName := extractTableFieldName(arg.Expression)
		if fieldName == "" {
			// Fallback: use arg type directly
			params = append(params, Parameter{
				Name: generateArgName(len(params)),
				Type: arg.GoType,
			})
			continue
		}

		// Find field in table struct
		field := findStructField(tableStruct, fieldName)
		if field == nil {
			// Field not found, use inferred type
			params = append(params, Parameter{
				Name: fieldName,
				Type: arg.GoType,
			})
			continue
		}

		// Get type from struct field
		tsType := parser.GoTypeToTS(field.Type)
		params = append(params, Parameter{
			Name: fieldName,
			Type: tsType,
		})
	}

	return params
}

// extractParametersFallback generates arg0, arg1, etc. when no table struct found
func extractParametersFallback(call parser.WASMCall) []Parameter {
	params := make([]Parameter, len(call.Args))
	for i, arg := range call.Args {
		params[i] = Parameter{
			Name: generateArgName(i),
			Type: arg.GoType,
		}
	}
	return params
}

// FindTableStruct finds the table-driven test struct in a function
func FindTableStruct(fn *ast.FuncDecl) *ast.StructType {
	var tableStruct *ast.StructType

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		// Look for: tests := []struct{...}{...}
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		// Check if LHS is "tests" variable
		if len(assign.Lhs) != 1 {
			return true
		}

		lhsIdent, ok := assign.Lhs[0].(*ast.Ident)
		if !ok || lhsIdent.Name != "tests" {
			return true
		}

		// Check if RHS is a composite literal
		if len(assign.Rhs) != 1 {
			return true
		}

		comp, ok := assign.Rhs[0].(*ast.CompositeLit)
		if !ok {
			return true
		}

		// Extract struct type from []struct{...}
		arrayType, ok := comp.Type.(*ast.ArrayType)
		if !ok {
			return true
		}

		structType, ok := arrayType.Elt.(*ast.StructType)
		if !ok {
			return true
		}

		tableStruct = structType
		return false // Stop searching
	})

	return tableStruct
}

// extractTableFieldName extracts field name from "tt.fieldName" expression
func extractTableFieldName(expr string) string {
	// Handle "tt.fieldName" pattern
	if strings.HasPrefix(expr, "tt.") {
		return strings.TrimPrefix(expr, "tt.")
	}
	return ""
}

// findStructField finds a field by name in a struct type
func findStructField(structType *ast.StructType, name string) *ast.Field {
	for _, field := range structType.Fields.List {
		for _, fieldName := range field.Names {
			if fieldName.Name == name {
				return field
			}
		}
	}
	return nil
}

// generateArgName generates argN parameter names
func generateArgName(index int) string {
	return fmt.Sprintf("arg%d", index)
}

// extractDocComment extracts the doc comment from a function declaration
func extractDocComment(fn *ast.FuncDecl) string {
	if fn.Doc == nil {
		return ""
	}

	var lines []string
	for _, comment := range fn.Doc.List {
		text := comment.Text
		// Remove leading // or /*
		text = strings.TrimPrefix(text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)
		lines = append(lines, text)
	}

	return strings.Join(lines, "\n")
}
