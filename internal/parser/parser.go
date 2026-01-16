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

		// Check if this is a defined type in the file
		if knownType, ok := types[t.Name]; ok {
			return *knownType
		}

		// Unknown type, treat as primitive
		return GoType{
			Name: t.Name,
			Kind: KindPrimitive,
		}

	case *ast.ArrayType:
		elemType := resolveType(t.Elt, types)
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
		keyType := resolveType(t.Key, types)
		valueType := resolveType(t.Value, types)
		return GoType{
			Name:  fmt.Sprintf("map[%s]%s", keyType.Name, valueType.Name),
			Kind:  KindMap,
			Key:   &keyType,
			Value: &valueType,
		}

	case *ast.StarExpr:
		elemType := resolveType(t.X, types)
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
				fieldType := resolveType(field.Type, types)
				jsonTag := extractJSONTag(field.Tag)

				for _, name := range field.Names {
					structType.Fields = append(structType.Fields, GoField{
						Name:    name.Name,
						Type:    fieldType,
						JSONTag: jsonTag,
					})
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
		return GoType{
			Name: "func",
			Kind: KindUnsupported,
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
	primitives := map[string]bool{
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
	return primitives[name]
}
