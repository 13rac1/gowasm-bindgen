package parser

import (
	"os"
	"path/filepath"
	"testing"
)

// FuzzParseSourceFile tests the parser with arbitrary Go source code.
// Run with: go test -fuzz=FuzzParseSourceFile -fuzztime=60s ./internal/parser/
func FuzzParseSourceFile(f *testing.F) {
	// Seed corpus with valid Go source patterns
	seeds := []string{
		// Basic function
		`package main; func Greet(name string) string { return name }`,
		// Multiple parameters
		`package main; func Add(a, b int) int { return a + b }`,
		// Struct return
		`package main; type User struct { Name string "json:\"name\"" }; func GetUser() User { return User{} }`,
		// Error return
		`package main; func Fetch(url string) (string, error) { return "", nil }`,
		// Slice parameter
		`package main; func Sum(nums []int) int { return 0 }`,
		// Map parameter
		`package main; func Lookup(m map[string]int, key string) int { return m[key] }`,
		// Callback parameter
		`package main; func ForEach(items []string, cb func(string, int)) {}`,
		// Pointer types
		`package main; func Deref(p *string) string { return *p }`,
		// Deeply nested types
		`package main; func Nested(x [][][]string) {}`,
		`package main; func DeepMap(m map[string]map[string][]int) {}`,
		// Multiple pointer indirection
		`package main; func MultiPtr(p ***string) {}`,
		// Empty function
		`package main; func Empty() {}`,
		// Multiple returns
		`package main; func Multi() (int, string, bool) { return 0, "", false }`,
		// Byte slice (typed array)
		`package main; func Hash(data []byte) []byte { return data }`,
		// Interface return
		`package main; func Any() interface{} { return nil }`,
		// Complex comments
		"package main\n// Greet returns a greeting.\n// It takes a name parameter.\nfunc Greet(name string) string { return name }",
		// Recursive struct
		`package main; type Node struct { Value int "json:\"value\""; Next *Node "json:\"next\"" }; func GetNode() *Node { return nil }`,
		// Anonymous field (should be handled)
		`package main; type Base struct { ID int "json:\"id\"" }; type Extended struct { Base; Name string "json:\"name\"" }`,
		// Select in main
		`package main; func main() { select {} }`,
		// Select with cases (invalid for WASM)
		`package main; func main() { ch := make(chan int); select { case <-ch: } }`,
	}

	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// Skip empty input
		if len(data) == 0 {
			return
		}

		// Write to temp file
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "fuzz.go")
		if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}

		// Parse - should not panic regardless of input
		// Errors are expected for invalid Go source
		_, _ = ParseSourceFile(tmpFile)
	})
}

// FuzzHasSelectInMain tests select{} detection with various main() implementations.
func FuzzHasSelectInMain(f *testing.F) {
	seeds := []string{
		// Valid select{}
		`package main; func main() { select {} }`,
		// Select with cases
		`package main; func main() { ch := make(chan int); select { case <-ch: } }`,
		// Nested in if
		`package main; func main() { if true { select {} } }`,
		// Nested in for
		`package main; func main() { for { select {} } }`,
		// In goroutine (shouldn't count)
		`package main; func main() { go func() { select {} }() }`,
		// Multiple selects
		`package main; func main() { select {}; select {} }`,
		// No select
		`package main; func main() {}`,
		// Select in other function
		`package main; func other() { select {} }; func main() {}`,
		// Complex control flow
		`package main; func main() { if true { for { if false { select {} } } } }`,
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

		// Should not panic
		_, _ = HasSelectInMain(tmpFile)
	})
}
