package extractor

// FunctionSignature represents a fully extracted WASM function signature
type FunctionSignature struct {
	Name       string      // Function name
	Params     []Parameter // Parameters with names and types
	Returns    ReturnType  // Return type info
	Examples   []Example   // From test cases
	Doc        string      // Doc comment from test function
	TestFunc   string      // Source test function name
	SourceFile string      // Source file path
	Line       int         // Line number
}

// Parameter represents a function parameter
type Parameter struct {
	Name string // Param name (from struct field or arg0, arg1, etc.)
	Type string // TypeScript type
}

// ReturnType represents the return type of a function
type ReturnType struct {
	Type    string  // TypeScript type string
	IsUnion bool    // Has union variants (e.g., success|error)
	Fields  []Field // For object returns
}

// Field represents a field in an object return type
type Field struct {
	Name string
	Type string
}

// Example represents a test case example
type Example struct {
	Name   string   // Test case name
	Args   []string // Argument values as TS literals
	Result string   // Expected result (if extractable)
}
