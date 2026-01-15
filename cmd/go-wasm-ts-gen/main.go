package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/13rac1/go-wasm-ts-gen/internal/extractor"
	"github.com/13rac1/go-wasm-ts-gen/internal/generator"
	"github.com/13rac1/go-wasm-ts-gen/internal/parser"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse flags
	var tests stringSlice
	var output string

	flag.Var(&tests, "tests", "glob pattern for test files (can be repeated)")
	flag.StringVar(&output, "output", "", "output .d.ts file path")
	flag.Parse()

	// Validate
	if len(tests) == 0 {
		return fmt.Errorf("--tests is required")
	}
	if output == "" {
		return fmt.Errorf("--output is required")
	}

	// Parse test files
	files, fset, err := parser.ParseTestFiles(tests)
	if err != nil {
		return fmt.Errorf("parsing test files: %w", err)
	}

	// Extract signatures
	sigs, err := extractor.ExtractSignatures(files, fset)
	if err != nil {
		return fmt.Errorf("extracting signatures: %w", err)
	}

	if len(sigs) == 0 {
		return fmt.Errorf("no WASM function signatures found in test files")
	}

	// Generate TypeScript
	dts := generator.Generate(sigs)

	// Create output directory if needed
	if dir := filepath.Dir(output); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
	}

	// Write output
	if err := os.WriteFile(output, []byte(dts), 0644); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	fmt.Printf("Generated %s with %d function(s)\n", output, len(sigs))
	return nil
}

// stringSlice allows repeated flag values
type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}
