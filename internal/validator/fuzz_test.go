package validator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/13rac1/gowasm-bindgen/internal/parser"
)

// FuzzValidateFunctions tests validation with arbitrary Go source that parses successfully.
// Run with: go test -fuzz=FuzzValidateFunctions -fuzztime=60s ./internal/validator/
func FuzzValidateFunctions(f *testing.F) {
	// Seed corpus with edge cases for validation
	seeds := []string{
		// Valid basic function
		`package main; func Greet(name string) string { return name }`,
		// Multiple errors in return (invalid)
		`package main; func Bad() (int, error, error) { return 0, nil, nil }`,
		// Error not last (invalid)
		`package main; func BadOrder() (error, int) { return nil, 0 }`,
		// Nested callback (invalid)
		`package main; func Nested(cb func(func())) {}`,
		// Callback with return (invalid)
		`package main; func Filter(predicate func(string) bool) {}`,
		// Non-string map key (invalid)
		`package main; func IntKey(m map[int]string) {}`,
		// Valid callback
		`package main; func ForEach(items []string, cb func(string, int)) {}`,
		// Anonymous/embedded field (invalid)
		`package main; type Base struct { ID int "json:\"id\"" }; type Ext struct { Base }; func Get() Ext { return Ext{} }`,
		// Deep nesting
		`package main; func Deep(x [][][][][]string) {}`,
		// Multiple map nesting
		`package main; func DeepMap(m map[string]map[string]map[string]int) {}`,
		// Pointer to slice
		`package main; func PtrSlice(p *[]string) {}`,
		// Slice of pointers
		`package main; func SlicePtr(s []*string) {}`,
		// Struct with all field types
		`package main
type Complex struct {
	Name    string              "json:\"name\""
	Age     int                 "json:\"age\""
	Tags    []string            "json:\"tags\""
	Meta    map[string]string   "json:\"meta\""
	Active  bool                "json:\"active\""
}
func GetComplex() Complex { return Complex{} }`,
		// Interface return (becomes any)
		`package main; func GetAny() interface{} { return nil }`,
		// Byte slice (typed array)
		`package main; func Hash(data []byte) []byte { return data }`,
		// Int slice (typed array)
		`package main; func Numbers(nums []int32) []int32 { return nums }`,
		// Float slice
		`package main; func Floats(f []float64) []float64 { return f }`,
		// Recursive type
		`package main; type Node struct { Value int "json:\"value\""; Children []*Node "json:\"children\"" }; func Tree() *Node { return nil }`,
		// Empty struct
		`package main; type Empty struct {}; func GetEmpty() Empty { return Empty{} }`,
		// Many parameters
		`package main; func Many(a, b, c, d, e, f, g, h int) int { return a }`,
		// Mixed named/unnamed (shouldn't happen in real code but worth testing)
		`package main; func Mixed(a int, _ string) int { return a }`,
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

		// Parse first - skip if parsing fails
		parsed, err := parser.ParseSourceFile(tmpFile)
		if err != nil {
			return // Invalid Go source, not interesting for validator fuzzing
		}

		// Validate - should not panic regardless of input
		// Validation errors are expected and fine
		_ = ValidateFunctions(parsed)
	})
}
