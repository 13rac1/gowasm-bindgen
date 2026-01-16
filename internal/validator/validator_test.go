package validator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/13rac1/gowasm-bindgen/internal/extractor"
)

func TestValidate_Success(t *testing.T) {
	sigs := []extractor.FunctionSignature{
		{
			Name:       "greet",
			SourceFile: "test.go",
			Line:       10,
			Params: []extractor.Parameter{
				{Name: "name", Type: "string"},
			},
			Returns: extractor.ReturnType{Type: "string"},
		},
	}

	err := Validate(sigs)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidate_AnyReturnType(t *testing.T) {
	sigs := []extractor.FunctionSignature{
		{
			Name:       "process",
			SourceFile: "wasm/test.go",
			Line:       25,
			Params: []extractor.Parameter{
				{Name: "data", Type: "string"},
			},
			Returns: extractor.ReturnType{Type: "any"},
		},
	}

	err := Validate(sigs)
	if err == nil {
		t.Fatal("expected error for 'any' return type")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "return type inferred as 'any'") {
		t.Errorf("expected error about 'any' return type, got: %s", errStr)
	}
	if !strings.Contains(errStr, "wasm/test.go:25") {
		t.Errorf("expected file:line in error, got: %s", errStr)
	}
}

func TestValidate_FallbackParamsAllowed(t *testing.T) {
	// Fallback param names (arg0, arg1) are now acceptable
	sigs := []extractor.FunctionSignature{
		{
			Name:       "compute",
			SourceFile: "wasm/calc.go",
			Line:       15,
			Params: []extractor.Parameter{
				{Name: "arg0", Type: "number"},
				{Name: "arg1", Type: "number"},
			},
			Returns: extractor.ReturnType{Type: "number"},
		},
	}

	err := Validate(sigs)
	if err != nil {
		t.Errorf("expected no error for fallback param names, got: %v", err)
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	sigs := []extractor.FunctionSignature{
		{
			Name:       "func1",
			SourceFile: "test1.go",
			Line:       10,
			Params: []extractor.Parameter{
				{Name: "arg0", Type: "string"},
			},
			Returns: extractor.ReturnType{Type: "any"},
		},
		{
			Name:       "func2",
			SourceFile: "test2.go",
			Line:       20,
			Params: []extractor.Parameter{
				{Name: "arg0", Type: "number"},
			},
			Returns: extractor.ReturnType{Type: "any"},
		},
	}

	err := Validate(sigs)
	if err == nil {
		t.Fatal("expected errors")
	}

	verr, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	// func1 has 1 error (any return)
	// func2 has 1 error (any return)
	if len(verr.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d: %v", len(verr.Errors), verr.Errors)
	}
}

func TestValidate_ErrorFixtures(t *testing.T) {
	// Only test the "any" return type error - other fixtures now pass validation
	files, fset, err := parseTestFile("../../testdata/errors/no_return_type/no_return_test.go")
	if err != nil {
		t.Fatalf("failed to parse test file: %v", err)
	}

	sigs, _ := extractor.ExtractSignatures(files, fset)

	if len(sigs) == 0 {
		t.Fatal("no signatures found")
	}

	err = Validate(sigs)
	if err == nil {
		t.Fatal("expected validation to fail for 'any' return type")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "return type inferred as 'any'") {
		t.Errorf("expected error about 'any' return type, got: %s", errStr)
	}
}

// parseTestFile is a helper to parse a single test file
func parseTestFile(pattern string) ([]*ast.File, *token.FileSet, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, pattern, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	return []*ast.File{file}, fset, nil
}
