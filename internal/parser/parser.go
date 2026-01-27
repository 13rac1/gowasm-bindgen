// Package parser extracts function signatures and types from Go source files.
package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"unicode"
)

// primitiveTypes is a lookup table for Go primitive types.
// Defined at package level to avoid allocation on each isPrimitive call.
var primitiveTypes = map[string]bool{
	"string":     true,
	"int":        true,
	"int8":       true,
	"int16":      true,
	"int32":      true,
	"int64":      true,
	"uint":       true,
	"uint8":      true,
	"uint16":     true,
	"uint32":     true,
	"uint64":     true,
	"float32":    true,
	"float64":    true,
	"bool":       true,
	"byte":       true,
	"rune":       true,
	"complex64":  true,
	"complex128": true,
}

// ParseSourceFile parses a Go source file and extracts exported functions and types
func ParseSourceFile(path string) (*ParsedFile, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	result := &ParsedFile{
		Package:   file.Name.Name,
		Functions: []GoFunction{},
		Types:     make(map[string]*GoType),
	}

	// First pass: collect all type definitions
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if isExported(typeSpec.Name.Name) {
						goType := resolveType(typeSpec.Type, result.Types)
						goType.Name = typeSpec.Name.Name
						result.Types[typeSpec.Name.Name] = &goType
					}
				}
			}
		}
	}

	// Second pass: collect exported functions
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			// Only exported functions (no methods)
			if funcDecl.Recv == nil && isExported(funcDecl.Name.Name) {
				fn := extractFunction(funcDecl, result.Types)
				result.Functions = append(result.Functions, fn)
			}
		}
	}

	return result, nil
}

// extractFunction extracts function signature from AST
func extractFunction(fn *ast.FuncDecl, types map[string]*GoType) GoFunction {
	function := GoFunction{
		Name:    fn.Name.Name,
		Params:  []GoParameter{},
		Returns: []GoType{},
		Doc:     extractDocComment(fn.Doc),
	}

	// Extract parameters
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			paramType := resolveType(field.Type, types)
			for _, name := range field.Names {
				function.Params = append(function.Params, GoParameter{
					Name: name.Name,
					Type: paramType,
				})
			}
		}
	}

	// Extract return types
	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			returnType := resolveType(field.Type, types)
			function.Returns = append(function.Returns, returnType)
		}
	}

	return function
}

// resolveType converts an AST type expression to GoType
func resolveType(expr ast.Expr, types map[string]*GoType) GoType {
	return resolveTypeWithVisited(expr, types, make(map[string]bool))
}

// resolveTypeWithVisited is the internal implementation that tracks visited types
// to detect cycles in type definitions (e.g., type A B; type B A).
func resolveTypeWithVisited(expr ast.Expr, types map[string]*GoType, visited map[string]bool) GoType {
	switch t := expr.(type) {
	case *ast.Ident:
		// Check for error type
		if t.Name == "error" {
			return GoType{
				Name:    "error",
				Kind:    KindError,
				IsError: true,
			}
		}

		// Check for known primitives
		if isPrimitive(t.Name) {
			return GoType{
				Name: t.Name,
				Kind: KindPrimitive,
			}
		}

		// Detect cycles in type references
		if visited[t.Name] {
			return GoType{
				Name: t.Name,
				Kind: KindUnsupported,
			}
		}

		// Check if this is a defined type in the file
		if knownType, ok := types[t.Name]; ok {
			visited[t.Name] = true
			return *knownType
		}

		// Unknown type, treat as primitive
		return GoType{
			Name: t.Name,
			Kind: KindPrimitive,
		}

	case *ast.ArrayType:
		elemType := resolveTypeWithVisited(t.Elt, types, visited)
		if t.Len == nil {
			// Slice
			return GoType{
				Name: "[]" + elemType.Name,
				Kind: KindSlice,
				Elem: &elemType,
			}
		}
		// Array with length
		return GoType{
			Name: "[]" + elemType.Name,
			Kind: KindArray,
			Elem: &elemType,
		}

	case *ast.MapType:
		keyType := resolveTypeWithVisited(t.Key, types, visited)
		valueType := resolveTypeWithVisited(t.Value, types, visited)
		return GoType{
			Name:  fmt.Sprintf("map[%s]%s", keyType.Name, valueType.Name),
			Kind:  KindMap,
			Key:   &keyType,
			Value: &valueType,
		}

	case *ast.StarExpr:
		elemType := resolveTypeWithVisited(t.X, types, visited)
		return GoType{
			Name: "*" + elemType.Name,
			Kind: KindPointer,
			Elem: &elemType,
		}

	case *ast.StructType:
		structType := GoType{
			Name:   "struct",
			Kind:   KindStruct,
			Fields: []GoField{},
		}

		if t.Fields != nil {
			for _, field := range t.Fields.List {
				fieldType := resolveTypeWithVisited(field.Type, types, visited)
				jsonTag := extractJSONTag(field.Tag)

				if len(field.Names) == 0 {
					// Anonymous/embedded field - add with empty name for validator to catch
					structType.Fields = append(structType.Fields, GoField{
						Name: "",
						Type: fieldType,
					})
				} else {
					for _, name := range field.Names {
						structType.Fields = append(structType.Fields, GoField{
							Name:    name.Name,
							Type:    fieldType,
							JSONTag: jsonTag,
						})
					}
				}
			}
		}

		return structType

	case *ast.SelectorExpr:
		// Handle qualified identifiers (e.g., time.Time, sql.NullString)
		if x, ok := t.X.(*ast.Ident); ok {
			return GoType{
				Name: x.Name + "." + t.Sel.Name,
				Kind: KindUnsupported,
			}
		}
		return GoType{
			Name: "unknown selector",
			Kind: KindUnsupported,
		}

	case *ast.FuncType:
		// Parse callback parameters
		var params []GoType
		if t.Params != nil {
			for _, field := range t.Params.List {
				paramType := resolveTypeWithVisited(field.Type, types, visited)
				// Functions can have unnamed params like func(string, int)
				if len(field.Names) == 0 {
					params = append(params, paramType)
				} else {
					for range field.Names {
						params = append(params, paramType)
					}
				}
			}
		}

		// Check for return values
		hasReturns := t.Results != nil && len(t.Results.List) > 0

		return GoType{
			Name:           "func",
			Kind:           KindFunction,
			CallbackParams: params,
			IsVoid:         !hasReturns,
		}

	case *ast.ChanType:
		return GoType{
			Name: "chan",
			Kind: KindUnsupported,
		}

	case *ast.InterfaceType:
		return GoType{
			Name: "interface",
			Kind: KindUnsupported,
		}

	default:
		return GoType{
			Name: fmt.Sprintf("unknown(%T)", expr),
			Kind: KindUnsupported,
		}
	}
}

// extractJSONTag extracts the JSON tag value from a field tag
func extractJSONTag(tag *ast.BasicLit) string {
	if tag == nil {
		return ""
	}

	// Parse tag string (remove backticks)
	tagStr := strings.Trim(tag.Value, "`")
	tags := reflect.StructTag(tagStr)
	jsonTag := tags.Get("json")

	// Extract just the name, ignore options like omitempty
	if idx := strings.Index(jsonTag, ","); idx != -1 {
		return jsonTag[:idx]
	}

	return jsonTag
}

// extractDocComment extracts documentation from comment group
func extractDocComment(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}

	var lines []string
	for _, comment := range doc.List {
		text := comment.Text
		text = strings.TrimPrefix(text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)
		if text != "" {
			lines = append(lines, text)
		}
	}

	return strings.Join(lines, "\n")
}

// isExported checks if a name is exported (starts with uppercase)
func isExported(name string) bool {
	if name == "" {
		return false
	}
	return unicode.IsUpper(rune(name[0]))
}

// isPrimitive checks if a type name is a Go primitive
func isPrimitive(name string) bool {
	return primitiveTypes[name]
}

// HasSelectInMain checks if a Go source file has a main function containing select {}.
// This is required for WASM modules to stay alive and receive JavaScript calls.
func HasSelectInMain(path string) (bool, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return false, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	// Find main function
	var mainFunc *ast.FuncDecl
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if funcDecl.Name.Name == "main" && funcDecl.Recv == nil {
			mainFunc = funcDecl
			break
		}
	}
	if mainFunc == nil {
		return false, nil
	}

	// Use ast.Inspect to find empty select statements
	found := false
	ast.Inspect(mainFunc.Body, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectStmt); ok {
			if sel.Body == nil || len(sel.Body.List) == 0 {
				found = true
				return false // stop inspection
			}
		}
		return true // continue
	})
	return found, nil
}
