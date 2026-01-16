package parser

// TypeKind represents the category of a Go type
type TypeKind int

const (
	KindPrimitive TypeKind = iota
	KindSlice
	KindArray
	KindMap
	KindStruct
	KindPointer
	KindError
	KindUnsupported
)

// GoType represents a parsed Go type with full structural information
type GoType struct {
	Name    string    // Type name (e.g., "string", "User", "[]string")
	Kind    TypeKind  // Category of the type
	Elem    *GoType   // Element type for slices/arrays/pointers
	Key     *GoType   // Key type for maps
	Value   *GoType   // Value type for maps
	Fields  []GoField // Fields for struct types
	IsError bool      // True if this is the error type
}

// GoField represents a single field in a struct
type GoField struct {
	Name    string // Field name
	Type    GoType // Field type
	JSONTag string // JSON tag value (if present)
}

// GoFunction represents a parsed exported function
type GoFunction struct {
	Name    string        // Function name
	Params  []GoParameter // Function parameters
	Returns []GoType      // Return types
	Doc     string        // Documentation comment
}

// GoParameter represents a single function parameter
type GoParameter struct {
	Name string // Parameter name
	Type GoType // Parameter type
}

// ParsedFile represents a parsed Go source file
type ParsedFile struct {
	Package   string             // Package name
	Functions []GoFunction       // Exported functions
	Types     map[string]*GoType // Type definitions in the file
}
