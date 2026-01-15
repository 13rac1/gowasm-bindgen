package parser

import (
	"go/ast"
	"strings"
)

// GoTypeToTS converts a Go type expression to TypeScript type string
func GoTypeToTS(expr ast.Expr) string {
	if expr == nil {
		return "any"
	}

	switch t := expr.(type) {
	case *ast.Ident:
		return identToTS(t.Name)

	case *ast.ArrayType:
		elemType := GoTypeToTS(t.Elt)
		return elemType + "[]"

	case *ast.MapType:
		// Only support map[string]T for now
		if keyIdent, ok := t.Key.(*ast.Ident); ok && keyIdent.Name == "string" {
			valueType := GoTypeToTS(t.Value)
			return "{[key: string]: " + valueType + "}"
		}
		return "any"

	case *ast.SelectorExpr:
		// Handle qualified identifiers like js.Value
		if xIdent, ok := t.X.(*ast.Ident); ok {
			return xIdent.Name + "." + t.Sel.Name
		}
		return "any"

	default:
		return "any"
	}
}

// identToTS converts a Go identifier to TypeScript type
func identToTS(name string) string {
	switch name {
	case "string":
		return "string"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	default:
		return "any"
	}
}

// InferTypeFromLiteral attempts to infer Go type from a literal expression
func InferTypeFromLiteral(expr ast.Expr) string {
	if expr == nil {
		return "any"
	}

	switch lit := expr.(type) {
	case *ast.BasicLit:
		switch lit.Kind.String() {
		case "INT":
			return "number"
		case "FLOAT":
			return "number"
		case "STRING":
			return "string"
		}

	case *ast.Ident:
		if lit.Name == "true" || lit.Name == "false" {
			return "boolean"
		}

	case *ast.CompositeLit:
		// Array or slice literal
		if _, ok := lit.Type.(*ast.ArrayType); ok {
			if len(lit.Elts) > 0 {
				elemType := InferTypeFromLiteral(lit.Elts[0])
				return elemType + "[]"
			}
			return "any[]"
		}
		// Map literal
		if _, ok := lit.Type.(*ast.MapType); ok {
			return "any"
		}
	}

	return "any"
}

// FormatTSType formats a TypeScript type string, handling arrays and objects
func FormatTSType(goType string) string {
	// Clean up any extra whitespace
	return strings.TrimSpace(goType)
}
