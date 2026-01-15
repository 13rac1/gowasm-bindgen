package validator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/13rac1/go-wasm-ts-gen/internal/extractor"
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

func TestValidate_FallbackParams(t *testing.T) {
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
	if err == nil {
		t.Fatal("expected error for fallback param names")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "fallback param names") {
		t.Errorf("expected error about fallback params, got: %s", errStr)
	}
	if !strings.Contains(errStr, "wasm/calc.go:15") {
		t.Errorf("expected file:line in error, got: %s", errStr)
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
			Returns: extractor.ReturnType{Type: "string"},
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

	// func1 has 2 errors (any return + fallback params)
	// func2 has 1 error (fallback params)
	if len(verr.Errors) != 3 {
		t.Errorf("expected 3 errors, got %d: %v", len(verr.Errors), verr.Errors)
	}
}

func TestHasFallbackParams(t *testing.T) {
	tests := []struct {
		name   string
		params []extractor.Parameter
		want   bool
	}{
		{
			name:   "named params",
			params: []extractor.Parameter{{Name: "input", Type: "string"}},
			want:   false,
		},
		{
			name:   "arg0",
			params: []extractor.Parameter{{Name: "arg0", Type: "string"}},
			want:   true,
		},
		{
			name:   "arg10",
			params: []extractor.Parameter{{Name: "arg10", Type: "string"}},
			want:   true,
		},
		{
			name:   "arg prefix but not fallback",
			params: []extractor.Parameter{{Name: "argument", Type: "string"}},
			want:   false,
		},
		{
			name:   "mixed",
			params: []extractor.Parameter{{Name: "name", Type: "string"}, {Name: "arg1", Type: "number"}},
			want:   true,
		},
		{
			name:   "empty",
			params: []extractor.Parameter{},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasFallbackParams(tt.params)
			if got != tt.want {
				t.Errorf("hasFallbackParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidate_ErrorFixtures(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		wantErrMsg  string
		wantErrMsgs []string // for multiple expected errors
	}{
		{
			name:       "no_table_struct",
			pattern:    "../../testdata/errors/no_table_struct/no_table_test.go",
			wantErrMsg: "fallback param names",
		},
		{
			name:    "no_return_type",
			pattern: "../../testdata/errors/no_return_type/no_return_test.go",
			wantErrMsgs: []string{
				"return type inferred as 'any'",
				"fallback param names",
			},
		},
		{
			name:       "many_params",
			pattern:    "../../testdata/errors/many_params/many_params_test.go",
			wantErrMsg: "fallback param names",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the test file
			files, fset, err := parseTestFile(tt.pattern)
			if err != nil {
				t.Fatalf("failed to parse test file: %v", err)
			}

			// Extract signatures
			sigs, err := extractor.ExtractSignatures(files, fset)
			if err != nil {
				t.Fatalf("failed to extract signatures: %v", err)
			}

			if len(sigs) == 0 {
				t.Fatal("no signatures found")
			}

			// Validate - should fail
			err = Validate(sigs)
			if err == nil {
				t.Fatal("expected validation to fail")
			}

			errStr := err.Error()

			// Check for expected error messages
			if tt.wantErrMsg != "" {
				if !strings.Contains(errStr, tt.wantErrMsg) {
					t.Errorf("expected error containing %q, got: %s", tt.wantErrMsg, errStr)
				}
			}

			for _, msg := range tt.wantErrMsgs {
				if !strings.Contains(errStr, msg) {
					t.Errorf("expected error containing %q, got: %s", msg, errStr)
				}
			}
		})
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
