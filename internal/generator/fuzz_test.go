package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/13rac1/gowasm-bindgen/internal/parser"
	"github.com/13rac1/gowasm-bindgen/internal/validator"
)

// FuzzGenerate tests code generation with arbitrary valid Go source.
// Run with: go test -fuzz=FuzzGenerate -fuzztime=60s ./internal/generator/
func FuzzGenerate(f *testing.F) {
	// Seed corpus with valid Go source that should generate code
	seeds := []string{
		// Basic function
		`package main; func Greet(name string) string { return name }`,
		// Multiple parameters
		`package main; func Add(a, b int) int { return a + b }`,
		// Struct return
		`package main
type User struct {
	Name string "json:\"name\""
	Age  int    "json:\"age\""
}
func GetUser() User { return User{} }`,
		// Error return
		`package main; func Fetch(url string) (string, error) { return "", nil }`,
		// Slice parameter and return
		`package main; func Double(nums []int) []int { return nums }`,
		// Map parameter
		`package main; func Get(m map[string]int, key string) int { return m[key] }`,
		// Callback parameter
		`package main; func ForEach(items []string, cb func(string, int)) {}`,
		// Byte slice (typed array)
		`package main; func Hash(data []byte) []byte { return data }`,
		// Nested types
		`package main; func Nested(x [][]string) [][]string { return x }`,
		// Pointer return
		`package main; func GetPtr() *string { s := "hi"; return &s }`,
		// Multiple functions
		`package main
func One() int { return 1 }
func Two() int { return 2 }
func Three() int { return 3 }`,
		// Complex struct
		`package main
type Config struct {
	Host    string            "json:\"host\""
	Port    int               "json:\"port\""
	Tags    []string          "json:\"tags\""
	Options map[string]string "json:\"options\""
	Enabled bool              "json:\"enabled\""
}
func LoadConfig() Config { return Config{} }`,
		// Recursive struct
		`package main
type Node struct {
	Value    int     "json:\"value\""
	Children []*Node "json:\"children\""
}
func BuildTree() *Node { return nil }`,
		// Interface return
		`package main; func GetAny() interface{} { return nil }`,
		// Empty function
		`package main; func Noop() {}`,
		// Float arrays
		`package main; func Floats(f []float64) []float64 { return f }`,
		// Multiple returns (non-error)
		`package main; func Pair() (int, string) { return 0, "" }`,
		// Struct with nested struct field
		`package main
type Inner struct { X int "json:\"x\"" }
type Outer struct { I Inner "json:\"i\"" }
func GetOuter() Outer { return Outer{} }`,
		// Special characters in doc comments
		"package main\n// Greet says \"hello\" to <name> & returns a greeting.\nfunc Greet(name string) string { return name }",
	}

	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}

		// Write to temp file
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "fuzz.go")
		if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}

		// Parse
		parsed, err := parser.ParseSourceFile(tmpFile)
		if err != nil {
			return // Invalid Go source
		}

		// Validate
		if err := validator.ValidateFunctions(parsed); err != nil {
			return // Invalid for generation
		}

		// No functions to generate
		if len(parsed.Functions) == 0 {
			return
		}

		// Generate sync mode TypeScript - should not panic
		_ = Generate(parsed, "test.ts", "TestWasm")

		// Generate worker mode TypeScript - should not panic
		_ = GenerateClient(parsed, "test.ts", "TestWasm")

		// Generate Go bindings (sync mode) - should not panic
		_ = GenerateGoBindings(parsed, false)

		// Generate Go bindings (worker mode) - should not panic
		_ = GenerateGoBindings(parsed, true)

		// Generate worker.js - should not panic
		_ = GenerateWorker("test.wasm")
	})
}

// FuzzGenerateGoBindings tests Go binding generation specifically.
func FuzzGenerateGoBindings(f *testing.F) {
	seeds := []string{
		`package main; func Greet(name string) string { return name }`,
		`package main; func Hash(data []byte) []byte { return data }`,
		`package main
type User struct { Name string "json:\"name\"" }
func GetUser() User { return User{} }`,
		`package main; func ForEach(items []string, cb func(string, int)) {}`,
		`package main; func Fetch(url string) (string, error) { return "", nil }`,
	}

	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}

		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "fuzz.go")
		if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}

		parsed, err := parser.ParseSourceFile(tmpFile)
		if err != nil {
			return
		}

		if err := validator.ValidateFunctions(parsed); err != nil {
			return
		}

		if len(parsed.Functions) == 0 {
			return
		}

		// Test both modes - should not panic
		_ = GenerateGoBindings(parsed, false)
		_ = GenerateGoBindings(parsed, true)
	})
}
