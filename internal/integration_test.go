package internal_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/13rac1/go-wasm-ts-gen/internal/extractor"
	"github.com/13rac1/go-wasm-ts-gen/internal/generator"
	"github.com/13rac1/go-wasm-ts-gen/internal/parser"
)

// TestFullPipeline_Primitives tests the complete pipeline with primitive type parameters
func TestFullPipeline_Primitives(t *testing.T) {
	// Parse
	pattern := filepath.Join("..", "testdata", "integration", "primitives", "*_test.go")
	files, fset, err := parser.ParseTestFiles([]string{pattern})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Extract
	sigs, err := extractor.ExtractSignatures(files, fset)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}

	if len(sigs) == 0 {
		t.Fatal("expected at least one signature, got none")
	}

	// Generate
	output := generator.Generate(sigs)

	// Verify generated TypeScript
	if !strings.Contains(output, "processPrimitives") {
		t.Errorf("expected processPrimitives function, got:\n%s", output)
	}

	if !strings.Contains(output, "str: string") {
		t.Errorf("expected str: string parameter, got:\n%s", output)
	}

	if !strings.Contains(output, "num: number") {
		t.Errorf("expected num: number parameter, got:\n%s", output)
	}

	if !strings.Contains(output, "flag: boolean") {
		t.Errorf("expected flag: boolean parameter, got:\n%s", output)
	}
}

// TestFullPipeline_Arrays tests the complete pipeline with array type parameters
func TestFullPipeline_Arrays(t *testing.T) {
	// Parse
	pattern := filepath.Join("..", "testdata", "integration", "arrays", "*_test.go")
	files, fset, err := parser.ParseTestFiles([]string{pattern})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Extract
	sigs, err := extractor.ExtractSignatures(files, fset)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}

	if len(sigs) == 0 {
		t.Fatal("expected at least one signature, got none")
	}

	// Generate
	output := generator.Generate(sigs)

	// Verify generated TypeScript
	if !strings.Contains(output, "processArrays") {
		t.Errorf("expected processArrays function, got:\n%s", output)
	}

	if !strings.Contains(output, "strings: string[]") {
		t.Errorf("expected strings: string[] parameter, got:\n%s", output)
	}

	if !strings.Contains(output, "numbers: number[]") {
		t.Errorf("expected numbers: number[] parameter, got:\n%s", output)
	}
}

// TestFullPipeline_Objects tests the complete pipeline with object return type
func TestFullPipeline_Objects(t *testing.T) {
	// Parse
	pattern := filepath.Join("..", "testdata", "integration", "objects", "*_test.go")
	files, fset, err := parser.ParseTestFiles([]string{pattern})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Extract
	sigs, err := extractor.ExtractSignatures(files, fset)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}

	if len(sigs) == 0 {
		t.Fatal("expected at least one signature, got none")
	}

	// Generate
	output := generator.Generate(sigs)

	// Verify generated TypeScript
	if !strings.Contains(output, "getObject") {
		t.Errorf("expected getObject function, got:\n%s", output)
	}

	// Check for object return type fields
	if !strings.Contains(output, "valid: boolean") {
		t.Errorf("expected valid: boolean field, got:\n%s", output)
	}

	if !strings.Contains(output, "hash: string") {
		t.Errorf("expected hash: string field, got:\n%s", output)
	}

	if !strings.Contains(output, "count: number") {
		t.Errorf("expected count: number field, got:\n%s", output)
	}
}

// TestFullPipeline_Unions tests the complete pipeline with union return type
func TestFullPipeline_Unions(t *testing.T) {
	// Parse
	pattern := filepath.Join("..", "testdata", "integration", "unions", "*_test.go")
	files, fset, err := parser.ParseTestFiles([]string{pattern})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Extract
	sigs, err := extractor.ExtractSignatures(files, fset)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}

	if len(sigs) == 0 {
		t.Fatal("expected at least one signature, got none")
	}

	// Generate
	output := generator.Generate(sigs)

	// Verify generated TypeScript
	if !strings.Contains(output, "validate") {
		t.Errorf("expected validate function, got:\n%s", output)
	}

	// Check for union type with error and success variants
	if !strings.Contains(output, "error: string") {
		t.Errorf("expected error: string field, got:\n%s", output)
	}

	if !strings.Contains(output, "success: boolean") {
		t.Errorf("expected success: boolean field, got:\n%s", output)
	}

	// Check for union operator
	if !strings.Contains(output, "|") {
		t.Errorf("expected union type (|), got:\n%s", output)
	}
}

// TestFullPipeline_Multiple tests the complete pipeline with multiple functions
func TestFullPipeline_Multiple(t *testing.T) {
	// Parse
	pattern := filepath.Join("..", "testdata", "integration", "multiple", "*_test.go")
	files, fset, err := parser.ParseTestFiles([]string{pattern})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Extract
	sigs, err := extractor.ExtractSignatures(files, fset)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}

	if len(sigs) < 3 {
		t.Fatalf("expected at least 3 signatures, got %d", len(sigs))
	}

	// Generate
	output := generator.Generate(sigs)

	// Verify all three functions are present
	expectedFuncs := []string{"addNumbers", "greet", "checkFlag"}
	for _, funcName := range expectedFuncs {
		if !strings.Contains(output, funcName) {
			t.Errorf("expected %s function, got:\n%s", funcName, output)
		}
	}

	// Verify addNumbers signature
	if !strings.Contains(output, "a: number") || !strings.Contains(output, "b: number") {
		t.Errorf("expected addNumbers(a: number, b: number), got:\n%s", output)
	}

	// Verify greet signature
	if !strings.Contains(output, "userName: string") {
		t.Errorf("expected greet(userName: string), got:\n%s", output)
	}

	// Verify return types
	if !strings.Contains(output, ": number") {
		t.Errorf("expected number return type for addNumbers, got:\n%s", output)
	}

	if !strings.Contains(output, ": string") {
		t.Errorf("expected string return type for greet, got:\n%s", output)
	}

	if !strings.Contains(output, ": boolean") {
		t.Errorf("expected boolean return type for checkFlag, got:\n%s", output)
	}
}

// TestFullPipeline_AllIntegration runs all integration fixtures together
func TestFullPipeline_AllIntegration(t *testing.T) {
	// Parse all integration test files
	patterns := []string{
		filepath.Join("..", "testdata", "integration", "primitives", "*_test.go"),
		filepath.Join("..", "testdata", "integration", "arrays", "*_test.go"),
		filepath.Join("..", "testdata", "integration", "objects", "*_test.go"),
		filepath.Join("..", "testdata", "integration", "unions", "*_test.go"),
		filepath.Join("..", "testdata", "integration", "multiple", "*_test.go"),
	}

	files, fset, err := parser.ParseTestFiles(patterns)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Extract
	sigs, err := extractor.ExtractSignatures(files, fset)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}

	if len(sigs) < 6 {
		t.Fatalf("expected at least 6 signatures from all fixtures, got %d", len(sigs))
	}

	// Generate
	output := generator.Generate(sigs)

	// Verify header is present
	if !strings.Contains(output, "Auto-generated TypeScript declarations") {
		t.Errorf("expected header comment, got:\n%s", output)
	}

	// Verify declare global structure
	if !strings.Contains(output, "declare global") {
		t.Errorf("expected declare global, got:\n%s", output)
	}

	if !strings.Contains(output, "interface Window") {
		t.Errorf("expected interface Window, got:\n%s", output)
	}

	// Verify export statement
	if !strings.Contains(output, "export {}") {
		t.Errorf("expected export {}, got:\n%s", output)
	}

	// Spot check a few functions
	expectedFuncs := []string{"processPrimitives", "processArrays", "getObject", "validate", "addNumbers"}
	for _, funcName := range expectedFuncs {
		if !strings.Contains(output, funcName) {
			t.Errorf("expected %s function in combined output", funcName)
		}
	}
}
