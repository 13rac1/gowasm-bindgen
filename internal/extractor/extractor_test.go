package extractor

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	goparser "github.com/13rac1/gowasm-bindgen/internal/parser"
)

func TestExtractSignatures(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantLen int
	}{
		{
			name:    "table driven test",
			pattern: "../../testdata/table/table_test.go",
			wantLen: 2,
		},
		{
			name:    "returns test",
			pattern: "../../testdata/returns/returns_test.go",
			wantLen: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, tt.pattern, nil, parser.ParseComments)
			if err != nil {
				t.Fatalf("failed to parse file: %v", err)
			}

			sigs, _ := ExtractSignatures([]*ast.File{file}, fset)

			if len(sigs) != tt.wantLen {
				t.Errorf("ExtractSignatures() got %d signatures, want %d", len(sigs), tt.wantLen)
			}
		})
	}
}

func TestExtractParameters_TableDriven(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "../../testdata/table/table_test.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	testFuncs := goparser.FindTestFunctions([]*ast.File{file})
	if len(testFuncs) == 0 {
		t.Fatal("no test functions found")
	}

	// Find TestWithTableDriven
	var fn *ast.FuncDecl
	for _, f := range testFuncs {
		if f.Name.Name == "TestWithTableDriven" {
			fn = f
			break
		}
	}
	if fn == nil {
		t.Fatal("TestWithTableDriven not found")
	}

	calls, _ := goparser.FindWASMCalls(fn, fset)
	if len(calls) == 0 {
		t.Fatal("no WASM calls found")
	}

	params := ExtractParameters(calls[0], fn)

	want := []Parameter{
		{Name: "input", Type: "string"},
		{Name: "count", Type: "number"},
		{Name: "enabled", Type: "boolean"},
	}

	if len(params) != len(want) {
		t.Fatalf("got %d params, want %d", len(params), len(want))
	}

	for i, p := range params {
		if p.Name != want[i].Name {
			t.Errorf("param[%d].Name = %s, want %s", i, p.Name, want[i].Name)
		}
		if p.Type != want[i].Type {
			t.Errorf("param[%d].Type = %s, want %s", i, p.Type, want[i].Type)
		}
	}
}

func TestExtractParameters_NoTable(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "../../testdata/table/table_test.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	testFuncs := goparser.FindTestFunctions([]*ast.File{file})

	// Find TestWithoutTable
	var fn *ast.FuncDecl
	for _, f := range testFuncs {
		if f.Name.Name == "TestWithoutTable" {
			fn = f
			break
		}
	}
	if fn == nil {
		t.Fatal("TestWithoutTable not found")
	}

	calls, _ := goparser.FindWASMCalls(fn, fset)
	if len(calls) == 0 {
		t.Fatal("no WASM calls found")
	}

	params := ExtractParameters(calls[0], fn)

	// Should fall back to arg0, arg1
	if len(params) != 2 {
		t.Fatalf("got %d params, want 2", len(params))
	}

	if params[0].Name != "arg0" {
		t.Errorf("param[0].Name = %s, want arg0", params[0].Name)
	}
	if params[1].Name != "arg1" {
		t.Errorf("param[1].Name = %s, want arg1", params[1].Name)
	}
}

func TestFindTableStruct(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "../../testdata/table/table_test.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	testFuncs := goparser.FindTestFunctions([]*ast.File{file})

	// Find TestWithTableDriven
	var fn *ast.FuncDecl
	for _, f := range testFuncs {
		if f.Name.Name == "TestWithTableDriven" {
			fn = f
			break
		}
	}
	if fn == nil {
		t.Fatal("TestWithTableDriven not found")
	}

	tableStruct := FindTableStruct(fn)
	if tableStruct == nil {
		t.Fatal("FindTableStruct() returned nil")
	}

	if tableStruct.Fields == nil || len(tableStruct.Fields.List) == 0 {
		t.Error("table struct has no fields")
	}
}

func TestExtractReturnType_Object(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "../../testdata/returns/returns_test.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	testFuncs := goparser.FindTestFunctions([]*ast.File{file})

	// Find TestReturnsObject
	var fn *ast.FuncDecl
	for _, f := range testFuncs {
		if f.Name.Name == "TestReturnsObject" {
			fn = f
			break
		}
	}
	if fn == nil {
		t.Fatal("TestReturnsObject not found")
	}

	calls, _ := goparser.FindWASMCalls(fn, fset)
	if len(calls) == 0 {
		t.Fatal("no WASM calls found")
	}

	ret := ExtractReturnType(fn, calls[0].ResultVar)

	if len(ret.Fields) == 0 {
		t.Error("expected return type to have fields")
	}

	// Check for valid and hash fields
	foundValid := false
	foundHash := false
	for _, field := range ret.Fields {
		if field.Name == "valid" && field.Type == "boolean" {
			foundValid = true
		}
		if field.Name == "hash" && field.Type == "string" {
			foundHash = true
		}
	}

	if !foundValid {
		t.Error("expected 'valid' field with type boolean")
	}
	if !foundHash {
		t.Error("expected 'hash' field with type string")
	}
}

func TestExtractReturnType_Primitive(t *testing.T) {
	tests := []struct {
		name     string
		testFunc string
		wantType string
	}{
		{
			name:     "string return",
			testFunc: "TestReturnsString",
			wantType: "string",
		},
		{
			name:     "number return",
			testFunc: "TestReturnsNumber",
			wantType: "number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "../../testdata/returns/returns_test.go", nil, parser.ParseComments)
			if err != nil {
				t.Fatalf("failed to parse file: %v", err)
			}

			testFuncs := goparser.FindTestFunctions([]*ast.File{file})

			var fn *ast.FuncDecl
			for _, f := range testFuncs {
				if f.Name.Name == tt.testFunc {
					fn = f
					break
				}
			}
			if fn == nil {
				t.Fatalf("%s not found", tt.testFunc)
			}

			calls, _ := goparser.FindWASMCalls(fn, fset)
			if len(calls) == 0 {
				t.Fatal("no WASM calls found")
			}

			ret := ExtractReturnType(fn, calls[0].ResultVar)

			if ret.Type != tt.wantType {
				t.Errorf("ExtractReturnType() type = %s, want %s", ret.Type, tt.wantType)
			}
		})
	}
}

func TestDetectUnionType(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "../../testdata/returns/returns_test.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	testFuncs := goparser.FindTestFunctions([]*ast.File{file})

	// Find TestReturnsUnion
	var fn *ast.FuncDecl
	for _, f := range testFuncs {
		if f.Name.Name == "TestReturnsUnion" {
			fn = f
			break
		}
	}
	if fn == nil {
		t.Fatal("TestReturnsUnion not found")
	}

	calls, _ := goparser.FindWASMCalls(fn, fset)
	if len(calls) == 0 {
		t.Fatal("no WASM calls found")
	}

	resultVars := findResultVariables(fn, calls[0].ResultVar)
	isUnion, variants := DetectUnionType(fn, resultVars)

	if !isUnion {
		t.Error("expected union type to be detected")
	}

	if len(variants) != 2 {
		t.Errorf("expected 2 union variants, got %d", len(variants))
	}
}

func TestExtractExamples(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "../../testdata/table/table_test.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	testFuncs := goparser.FindTestFunctions([]*ast.File{file})

	// Find TestWithTableDriven
	var fn *ast.FuncDecl
	for _, f := range testFuncs {
		if f.Name.Name == "TestWithTableDriven" {
			fn = f
			break
		}
	}
	if fn == nil {
		t.Fatal("TestWithTableDriven not found")
	}

	tableStruct := FindTableStruct(fn)
	if tableStruct == nil {
		t.Fatal("table struct not found")
	}

	examples := ExtractExamples(fn, tableStruct)

	if len(examples) == 0 {
		t.Error("expected examples to be extracted")
	}

	for i, ex := range examples {
		if ex.Name == "" {
			t.Errorf("example[%d] has empty name", i)
		}
		if len(ex.Args) == 0 {
			t.Errorf("example[%d] has no args", i)
		}
	}
}

func TestGenerateArgName(t *testing.T) {
	tests := []struct {
		index int
		want  string
	}{
		{0, "arg0"},
		{1, "arg1"},
		{9, "arg9"},
		{10, "arg10"},
		{11, "arg11"},
		{99, "arg99"},
		{100, "arg100"},
	}

	for _, tt := range tests {
		got := generateArgName(tt.index)
		if got != tt.want {
			t.Errorf("generateArgName(%d) = %q, want %q", tt.index, got, tt.want)
		}
	}
}

func TestIsSimpleIdentifier(t *testing.T) {
	tests := []struct {
		expr string
		want bool
	}{
		{"userName", true},
		{"input", true},
		{"x", true},
		{"_private", true},
		{"", false},
		{"tt.field", false},   // selector
		{`"hello"`, false},    // string literal
		{"123", false},        // number literal
		{"{...}", false},      // composite literal
		{"'x'", false},        // rune literal
		{"obj.method", false}, // selector
	}

	for _, tt := range tests {
		got := isSimpleIdentifier(tt.expr)
		if got != tt.want {
			t.Errorf("isSimpleIdentifier(%q) = %v, want %v", tt.expr, got, tt.want)
		}
	}
}

func TestInferParamName(t *testing.T) {
	tests := []struct {
		expr  string
		index int
		want  string
	}{
		{"userName", 0, "userName"},
		{"input", 1, "input"},
		{`"hello"`, 0, "arg0"},  // literal fallback
		{"tt.field", 2, "arg2"}, // selector fallback
		{"123", 3, "arg3"},      // number fallback
	}

	for _, tt := range tests {
		got := inferParamName(tt.expr, tt.index)
		if got != tt.want {
			t.Errorf("inferParamName(%q, %d) = %q, want %q", tt.expr, tt.index, got, tt.want)
		}
	}
}

func TestExtractParameters_ManyParams(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "../../testdata/errors/many_params/many_params_test.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	testFuncs := goparser.FindTestFunctions([]*ast.File{file})
	if len(testFuncs) == 0 {
		t.Fatal("no test functions found")
	}

	fn := testFuncs[0]
	calls, _ := goparser.FindWASMCalls(fn, fset)
	if len(calls) == 0 {
		t.Fatal("no WASM calls found")
	}

	params := ExtractParameters(calls[0], fn)

	// Should have 15 fallback params (arg0 through arg14)
	if len(params) != 15 {
		t.Fatalf("got %d params, want 15", len(params))
	}

	// Verify param names are correct, especially for index >= 10
	for i, p := range params {
		want := generateArgName(i)
		if p.Name != want {
			t.Errorf("param[%d].Name = %q, want %q", i, p.Name, want)
		}
	}
}
