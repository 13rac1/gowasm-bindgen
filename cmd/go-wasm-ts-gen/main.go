package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/13rac1/go-wasm-ts-gen/internal/extractor"
	"github.com/13rac1/go-wasm-ts-gen/internal/generator"
	"github.com/13rac1/go-wasm-ts-gen/internal/parser"
	"github.com/13rac1/go-wasm-ts-gen/internal/validator"
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

	// Validate flags
	if len(tests) == 0 {
		return fmt.Errorf("--tests is required\n\n" +
			"Usage: go-wasm-ts-gen --tests 'path/*_test.go' --output types.d.ts")
	}
	if output == "" {
		return fmt.Errorf("--output is required\n\n" +
			"Usage: go-wasm-ts-gen --tests 'path/*_test.go' --output types.d.ts")
	}

	// Parse test files
	fmt.Println("Parsing test files...")
	files, fset, err := parser.ParseTestFiles(tests)
	if err != nil {
		return fmt.Errorf("parsing test files: %w", err)
	}

	// Extract signatures
	sigs, rejections, err := extractor.ExtractSignatures(files, fset)
	if err != nil {
		return fmt.Errorf("extracting signatures: %w", err)
	}

	// Fail on malformed WASM patterns
	if len(rejections) > 0 {
		fmt.Printf("\nerror: found %d malformed WASM call pattern(s):\n", len(rejections))
		for _, r := range rejections {
			fmt.Printf("  %s:%d: %s (%s)\n", r.SourceFile, r.Line, r.Reason, r.FuncName)
		}
		fmt.Println("\nExpected pattern:")
		fmt.Println("  result := funcName(js.Null(), []js.Value{js.ValueOf(arg), ...})")
		return fmt.Errorf("malformed WASM call patterns detected")
	}

	// Print found signatures
	fmt.Printf("\nFound %d WASM function(s):\n", len(sigs))
	for _, sig := range sigs {
		fmt.Printf("\n  %s (%s:%d)\n", sig.Name, sig.SourceFile, sig.Line)
		fmt.Println("    Parameters:")
		for _, p := range sig.Params {
			fmt.Printf("      - %s: %s\n", p.Name, p.Type)
		}
		fmt.Printf("    Return type: %s\n", sig.Returns.Type)
	}
	fmt.Println()

	if len(sigs) == 0 {
		return fmt.Errorf("no WASM function signatures found\n\n" +
			"Expected pattern:\n" +
			"  result := funcName(js.Null(), []js.Value{js.ValueOf(arg), ...})\n\n" +
			"Checklist:\n" +
			"  - Test functions start with 'Test'\n" +
			"  - First argument is js.Null()\n" +
			"  - Second argument is []js.Value{...} literal")
	}

	// Validate signatures (always runs, fails on issues)
	if err := validator.Validate(sigs); err != nil {
		return err
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
