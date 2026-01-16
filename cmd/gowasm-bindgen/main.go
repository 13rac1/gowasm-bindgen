package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/13rac1/gowasm-bindgen/internal/extractor"
	"github.com/13rac1/gowasm-bindgen/internal/generator"
	"github.com/13rac1/gowasm-bindgen/internal/parser"
	"github.com/13rac1/gowasm-bindgen/internal/validator"
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
	var sync bool

	flag.Var(&tests, "tests", "glob pattern for test files (can be repeated)")
	flag.StringVar(&output, "output", "", "output client.ts file path")
	flag.BoolVar(&sync, "sync", false, "generate synchronous API (default: worker-based async)")
	flag.Parse()

	// Validate flags
	if len(tests) == 0 {
		return fmt.Errorf("--tests is required\n\n" +
			"Usage: gowasm-bindgen --tests 'path/*_test.go' --output client.ts")
	}
	if output == "" {
		return fmt.Errorf("--output is required\n\n" +
			"Usage: gowasm-bindgen --tests 'path/*_test.go' --output client.ts")
	}

	// Parse test files
	fmt.Println("Parsing test files...")
	files, fset, err := parser.ParseTestFiles(tests)
	if err != nil {
		return fmt.Errorf("parsing test files: %w", err)
	}

	// Extract package name
	packageName := parser.GetPackageName(files)
	fmt.Printf("Package: %s\n", packageName)

	// Extract signatures
	sigs, rejections := extractor.ExtractSignatures(files, fset)

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

	// Create output directory if needed
	if dir := filepath.Dir(output); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
	}

	if sync {
		// Sync mode: generate class-based client.ts only
		return generateSyncOutput(sigs, packageName, output)
	}

	// Default mode (worker): generate client.ts and worker.js
	return generateWorkerOutput(sigs, packageName, output)
}

func generateSyncOutput(sigs []extractor.FunctionSignature, packageName string, output string) error {
	// Generate TypeScript class-based client
	content := generator.Generate(sigs, packageName)

	// Write output
	if err := os.WriteFile(output, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	fmt.Printf("Generated %s with %d function(s) (sync mode)\n", output, len(sigs))
	fmt.Println("\nUsage:")
	fmt.Printf("  import { %s } from './client';\n", toClassName(packageName))
	fmt.Printf("  const wasm = await %s.init('./example.wasm');\n", toClassName(packageName))
	fmt.Printf("  const result = wasm.greet('World');\n")
	return nil
}

func generateWorkerOutput(sigs []extractor.FunctionSignature, packageName string, output string) error {
	outputDir := filepath.Dir(output)

	// Generate worker.js
	workerPath := filepath.Join(outputDir, "worker.js")
	if err := os.WriteFile(workerPath, []byte(generator.GenerateWorker()), 0644); err != nil {
		return fmt.Errorf("writing worker: %w", err)
	}

	// Generate client.ts
	clientContent := generator.GenerateClient(sigs, packageName)
	if err := os.WriteFile(output, []byte(clientContent), 0644); err != nil {
		return fmt.Errorf("writing client: %w", err)
	}

	fmt.Printf("Generated %s (Web Worker entry point)\n", workerPath)
	fmt.Printf("Generated %s with %d function(s) (worker mode)\n", output, len(sigs))
	fmt.Println("\nUsage:")
	fmt.Printf("  import { %s } from './client';\n", toClassName(packageName))
	fmt.Printf("  const wasm = await %s.init('./worker.js');\n", toClassName(packageName))
	fmt.Printf("  const result = await wasm.greet('World');\n")
	fmt.Printf("  wasm.terminate();\n")
	return nil
}

// toClassName converts a Go package name to a TypeScript class name.
func toClassName(packageName string) string {
	if packageName == "" {
		return "Wasm"
	}
	// Import strings for ToUpper is already there
	return strings.ToUpper(packageName[:1]) + packageName[1:]
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
