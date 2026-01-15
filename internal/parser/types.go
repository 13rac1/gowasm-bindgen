package parser

// WASMCall represents a detected WASM function call in test code
type WASMCall struct {
	FuncName   string     // e.g., "merkleHashLeaf"
	Args       []Argument // Arguments passed via js.ValueOf()
	ResultVar  string     // Variable name holding result (e.g., "result")
	TestFunc   string     // Parent test function name
	SourceFile string     // Source file path
	Line       int        // Line number of the call
}

// Argument represents a single argument to a WASM function
type Argument struct {
	Expression string // The expression passed to js.ValueOf()
	GoType     string // Resolved Go type (string, int, []string, etc.)
}
