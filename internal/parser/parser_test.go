package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestParseTestFiles(t *testing.T) {
	tests := []struct {
		name        string
		patterns    []string
		wantErr     bool
		wantFiles   int
		description string
	}{
		{
			name:        "valid test file",
			patterns:    []string{"../../testdata/simple/*_test.go"},
			wantErr:     false,
			wantFiles:   1,
			description: "should parse simple_test.go",
		},
		{
			name:        "no matching files",
			patterns:    []string{"../../testdata/nonexistent/*_test.go"},
			wantErr:     true,
			wantFiles:   0,
			description: "should error when no files match",
		},
		{
			name:        "invalid pattern",
			patterns:    []string{"[invalid"},
			wantErr:     true,
			wantFiles:   0,
			description: "should error on invalid glob pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, fset, err := ParseTestFiles(tt.patterns)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTestFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(files) != tt.wantFiles {
					t.Errorf("ParseTestFiles() got %d files, want %d", len(files), tt.wantFiles)
				}
				if fset == nil {
					t.Error("ParseTestFiles() returned nil FileSet")
				}
			}
		})
	}
}

func TestFindTestFunctions(t *testing.T) {
	src := `
package test

import "testing"

func TestOne(t *testing.T) {}
func TestTwo(t *testing.T) {}
func BenchmarkSomething(b *testing.B) {}
func helperFunc() {}
func ExampleSomething() {}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse test source: %v", err)
	}

	testFuncs := FindTestFunctions([]*ast.File{file})

	if len(testFuncs) != 2 {
		t.Errorf("FindTestFunctions() found %d functions, want 2", len(testFuncs))
	}

	// Verify we found the right functions
	names := make(map[string]bool)
	for _, fn := range testFuncs {
		names[fn.Name.Name] = true
	}

	if !names["TestOne"] {
		t.Error("FindTestFunctions() did not find TestOne")
	}
	if !names["TestTwo"] {
		t.Error("FindTestFunctions() did not find TestTwo")
	}
	if names["BenchmarkSomething"] {
		t.Error("FindTestFunctions() incorrectly found BenchmarkSomething")
	}
}

func TestFindWASMCalls(t *testing.T) {
	src := `
package test

import (
	"syscall/js"
	"testing"
)

func someFunc(this js.Value, args []js.Value) interface{} {
	return nil
}

func TestWASMCall(t *testing.T) {
	result := someFunc(js.Null(), []js.Value{
		js.ValueOf("hello"),
		js.ValueOf(42),
	})
	_ = result
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse test source: %v", err)
	}

	testFuncs := FindTestFunctions([]*ast.File{file})
	if len(testFuncs) != 1 {
		t.Fatalf("expected 1 test function, got %d", len(testFuncs))
	}

	calls, rejections := FindWASMCalls(testFuncs[0], fset)

	if len(rejections) > 0 {
		t.Fatalf("FindWASMCalls() had %d rejections: %v", len(rejections), rejections)
	}

	if len(calls) != 1 {
		t.Fatalf("FindWASMCalls() found %d calls, want 1", len(calls))
	}

	call := calls[0]

	if call.FuncName != "someFunc" {
		t.Errorf("FuncName = %q, want %q", call.FuncName, "someFunc")
	}

	if call.ResultVar != "result" {
		t.Errorf("ResultVar = %q, want %q", call.ResultVar, "result")
	}

	if call.TestFunc != "TestWASMCall" {
		t.Errorf("TestFunc = %q, want %q", call.TestFunc, "TestWASMCall")
	}

	if len(call.Args) != 2 {
		t.Fatalf("got %d arguments, want 2", len(call.Args))
	}

	if call.Args[0].GoType != "string" {
		t.Errorf("Args[0].GoType = %q, want %q", call.Args[0].GoType, "string")
	}

	if call.Args[1].GoType != "number" {
		t.Errorf("Args[1].GoType = %q, want %q", call.Args[1].GoType, "number")
	}
}

func TestFindWASMCallsMultiple(t *testing.T) {
	src := `
package test

import (
	"syscall/js"
	"testing"
)

func funcA(this js.Value, args []js.Value) interface{} { return nil }
func funcB(this js.Value, args []js.Value) interface{} { return nil }

func TestMultipleCalls(t *testing.T) {
	result1 := funcA(js.Null(), []js.Value{js.ValueOf("test")})
	result2 := funcB(js.Null(), []js.Value{js.ValueOf(123)})
	_, _ = result1, result2
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse test source: %v", err)
	}

	testFuncs := FindTestFunctions([]*ast.File{file})
	if len(testFuncs) != 1 {
		t.Fatalf("expected 1 test function, got %d", len(testFuncs))
	}

	calls, rejections := FindWASMCalls(testFuncs[0], fset)

	if len(rejections) > 0 {
		t.Fatalf("FindWASMCalls() had %d rejections: %v", len(rejections), rejections)
	}

	if len(calls) != 2 {
		t.Fatalf("FindWASMCalls() found %d calls, want 2", len(calls))
	}

	if calls[0].FuncName != "funcA" {
		t.Errorf("calls[0].FuncName = %q, want %q", calls[0].FuncName, "funcA")
	}

	if calls[1].FuncName != "funcB" {
		t.Errorf("calls[1].FuncName = %q, want %q", calls[1].FuncName, "funcB")
	}
}

func TestGoTypeToTS(t *testing.T) {
	tests := []struct {
		name     string
		goType   string
		expected string
	}{
		{"string", "string", "string"},
		{"int", "int", "number"},
		{"int8", "int8", "number"},
		{"int16", "int16", "number"},
		{"int32", "int32", "number"},
		{"int64", "int64", "number"},
		{"uint", "uint", "number"},
		{"uint8", "uint8", "number"},
		{"uint16", "uint16", "number"},
		{"uint32", "uint32", "number"},
		{"uint64", "uint64", "number"},
		{"float32", "float32", "number"},
		{"float64", "float64", "number"},
		{"bool", "bool", "boolean"},
		{"unknown", "MyCustomType", "any"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := "package test\nvar x " + tt.goType
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", src, 0)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			// Extract the type from the var declaration
			var typeExpr ast.Expr
			ast.Inspect(file, func(n ast.Node) bool {
				if vs, ok := n.(*ast.ValueSpec); ok && len(vs.Names) > 0 {
					typeExpr = vs.Type
					return false
				}
				return true
			})

			if typeExpr == nil {
				t.Fatal("failed to find type expression")
			}

			result := GoTypeToTS(typeExpr)
			if result != tt.expected {
				t.Errorf("GoTypeToTS(%q) = %q, want %q", tt.goType, result, tt.expected)
			}
		})
	}
}

func TestGoTypeToTSArrays(t *testing.T) {
	tests := []struct {
		name     string
		goType   string
		expected string
	}{
		{"string array", "[]string", "string[]"},
		{"int array", "[]int", "number[]"},
		{"bool array", "[]bool", "boolean[]"},
		{"nested array", "[][]string", "string[][]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := "package test\nvar x " + tt.goType
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", src, 0)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			var typeExpr ast.Expr
			ast.Inspect(file, func(n ast.Node) bool {
				if vs, ok := n.(*ast.ValueSpec); ok && len(vs.Names) > 0 {
					typeExpr = vs.Type
					return false
				}
				return true
			})

			if typeExpr == nil {
				t.Fatal("failed to find type expression")
			}

			result := GoTypeToTS(typeExpr)
			if result != tt.expected {
				t.Errorf("GoTypeToTS(%q) = %q, want %q", tt.goType, result, tt.expected)
			}
		})
	}
}

func TestGoTypeToTSMaps(t *testing.T) {
	tests := []struct {
		name     string
		goType   string
		expected string
	}{
		{"string map", "map[string]string", "{[key: string]: string}"},
		{"int map", "map[string]int", "{[key: string]: number}"},
		{"bool map", "map[string]bool", "{[key: string]: boolean}"},
		{"unsupported key", "map[int]string", "any"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := "package test\nvar x " + tt.goType
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", src, 0)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			var typeExpr ast.Expr
			ast.Inspect(file, func(n ast.Node) bool {
				if vs, ok := n.(*ast.ValueSpec); ok && len(vs.Names) > 0 {
					typeExpr = vs.Type
					return false
				}
				return true
			})

			if typeExpr == nil {
				t.Fatal("failed to find type expression")
			}

			result := GoTypeToTS(typeExpr)
			if result != tt.expected {
				t.Errorf("GoTypeToTS(%q) = %q, want %q", tt.goType, result, tt.expected)
			}
		})
	}
}

func TestInferTypeFromLiteral(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"string literal", `"hello"`, "string"},
		{"int literal", `42`, "number"},
		{"float literal", `3.14`, "number"},
		{"bool true", `true`, "boolean"},
		{"bool false", `false`, "boolean"},
		{"string array", `[]string{"a", "b"}`, "string[]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := "package test\nvar x = " + tt.code
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", src, 0)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			var valueExpr ast.Expr
			ast.Inspect(file, func(n ast.Node) bool {
				if vs, ok := n.(*ast.ValueSpec); ok && len(vs.Values) > 0 {
					valueExpr = vs.Values[0]
					return false
				}
				return true
			})

			if valueExpr == nil {
				t.Fatal("failed to find value expression")
			}

			result := InferTypeFromLiteral(valueExpr)
			if result != tt.expected {
				t.Errorf("InferTypeFromLiteral(%q) = %q, want %q", tt.code, result, tt.expected)
			}
		})
	}
}

func TestFindWASMCalls_Rejections(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		wantReason string
	}{
		{"wrong_arg_count", "../../testdata/errors/wrong_arg_count/*_test.go", "expected exactly 2"},
		{"wrong_first_arg", "../../testdata/errors/wrong_first_arg/*_test.go", "not js.Null()"},
		{"wrong_second_arg", "../../testdata/errors/wrong_second_arg/*_test.go", "not []js.Value"},
		{"selector_func", "../../testdata/errors/selector_func/*_test.go", "method/selector"},
		{"no_assignment", "../../testdata/errors/no_assignment/*_test.go", "not assigned"},
		{"wrong_slice_type", "../../testdata/errors/wrong_slice_type/*_test.go", "not []js.Value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, fset, err := ParseTestFiles([]string{tt.pattern})
			if err != nil {
				t.Fatalf("failed to parse test files: %v", err)
			}

			testFuncs := FindTestFunctions(files)
			if len(testFuncs) == 0 {
				t.Fatal("no test functions found")
			}

			var allRejections []RejectedCall
			for _, fn := range testFuncs {
				_, rejections := FindWASMCalls(fn, fset)
				allRejections = append(allRejections, rejections...)
			}

			if len(allRejections) == 0 {
				t.Error("expected rejection, got none")
				return
			}

			found := false
			for _, r := range allRejections {
				if strings.Contains(r.Reason, tt.wantReason) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected reason containing %q, got %v", tt.wantReason, allRejections)
			}
		})
	}
}

func TestFindWASMCallsNoMatches(t *testing.T) {
	src := `
package test

import "testing"

func TestNormal(t *testing.T) {
	x := 42
	y := "hello"
	_ = x + len(y)
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse test source: %v", err)
	}

	testFuncs := FindTestFunctions([]*ast.File{file})
	if len(testFuncs) != 1 {
		t.Fatalf("expected 1 test function, got %d", len(testFuncs))
	}

	calls, rejections := FindWASMCalls(testFuncs[0], fset)

	if len(calls) != 0 {
		t.Errorf("FindWASMCalls() found %d calls, want 0", len(calls))
	}

	if len(rejections) != 0 {
		t.Errorf("FindWASMCalls() had %d rejections, want 0", len(rejections))
	}
}
