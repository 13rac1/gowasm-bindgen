package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"

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
	var output string
	var goOutput string
	var sync bool
	var verbose bool
	var wasmPath string

	flag.StringVar(&output, "output", "", "output TypeScript file path (e.g., client.ts)")
	flag.StringVar(&goOutput, "go-output", "", "output Go bindings file path (e.g., bindings_gen.go)")
	flag.BoolVar(&sync, "sync", false, "generate synchronous API (default: worker-based async)")
	flag.BoolVar(&verbose, "verbose", false, "enable verbose debug output")
	flag.StringVar(&wasmPath, "wasm-path", "", "WASM file path in generated worker.js (default: <dirname>.wasm)")
	flag.Parse()

	// Validate flags
	usage := "Usage: gowasm-bindgen <source.go> --output client.ts [--go-output bindings_gen.go] [--sync] [--wasm-path module.wasm] [--verbose]"
	if flag.NArg() == 0 {
		return fmt.Errorf("missing source file argument\n\n%s", usage)
	}
	if output == "" {
		return fmt.Errorf("--output is required\n\n%s", usage)
	}

	sourceFile := flag.Arg(0)

	if verbose {
		fmt.Fprintf(os.Stderr, "[DEBUG] Source file: %s\n", sourceFile)
		fmt.Fprintf(os.Stderr, "[DEBUG] Output: %s\n", output)
		fmt.Fprintf(os.Stderr, "[DEBUG] Sync mode: %v\n", sync)
		fmt.Fprintf(os.Stderr, "[DEBUG] WASM path: %s\n", wasmPath)
	}

	// Check if source file exists
	if _, err := os.Stat(sourceFile); err != nil {
		return fmt.Errorf("source file not found: %s", sourceFile)
	}

	// Derive default WASM path from directory name if not specified
	if wasmPath == "" {
		dirName := filepath.Base(filepath.Dir(sourceFile))
		if dirName == "." || dirName == "" {
			wasmPath = "main.wasm"
		} else {
			wasmPath = dirName + ".wasm"
		}
	}

	// Parse source file
	fmt.Printf("Parsing %s...\n", sourceFile)
	parsed, err := parser.ParseSourceFile(sourceFile)
	if err != nil {
		return fmt.Errorf("parsing source file: %w", err)
	}

	fmt.Printf("Package: %s\n", parsed.Package)
	fmt.Printf("Found %d exported function(s):\n", len(parsed.Functions))
	for _, fn := range parsed.Functions {
		fmt.Printf("  - %s\n", fn.Name)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[DEBUG] Found %d types\n", len(parsed.Types))
		for _, fn := range parsed.Functions {
			fmt.Fprintf(os.Stderr, "[DEBUG] Function %s: %d params, %d returns\n", fn.Name, len(fn.Params), len(fn.Returns))
		}
	}

	if len(parsed.Functions) == 0 {
		return fmt.Errorf("no exported functions found in %s\n\n"+
			"Functions must be exported (start with uppercase letter) and have no receiver", sourceFile)
	}

	// Check for select {} in main (required for WASM to stay alive)
	if parsed.Package == "main" {
		hasSelect, err := parser.HasSelectInMain(sourceFile)
		if err != nil {
			return fmt.Errorf("checking for select {}: %w", err)
		}
		if !hasSelect {
			return fmt.Errorf("main() does not contain 'select {}' - " +
				"WASM modules require this to block forever and receive JavaScript calls - " +
				"add 'select {}' at the end of your main() function")
		}
	}

	// Validate functions
	if err := validator.ValidateFunctions(parsed); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Create output directory if needed
	if dir := filepath.Dir(output); dir != "." {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
	}

	// Generate Go bindings if requested
	// workerMode is true when NOT using sync mode (default is worker mode)
	if goOutput != "" {
		fmt.Printf("\nGenerating Go bindings...\n")
		workerMode := !sync
		bindingsCode := generator.GenerateGoBindings(parsed, workerMode)

		if err := os.WriteFile(goOutput, []byte(bindingsCode), 0600); err != nil {
			return fmt.Errorf("writing Go bindings: %w", err)
		}

		fmt.Printf("Generated %s\n", goOutput)
	}

	// Generate TypeScript client
	if sync {
		if verbose {
			fmt.Fprintf(os.Stderr, "[DEBUG] Generating sync mode client\n")
		}
		return generateSyncOutput(parsed, output)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[DEBUG] Generating worker mode client\n")
	}
	return generateWorkerOutput(parsed, output, wasmPath)
}

func generateSyncOutput(parsed *parser.ParsedFile, output string) error {
	// Generate TypeScript class-based client
	content := generator.Generate(parsed)

	// Write output
	if err := os.WriteFile(output, []byte(content), 0600); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	fmt.Printf("\nGenerated %s with %d function(s) (sync mode)\n", output, len(parsed.Functions))
	fmt.Println("\nUsage:")
	fmt.Printf("  import { %s } from './client';\n", toClassName(parsed.Package))
	fmt.Printf("  const wasm = await %s.init('./example.wasm');\n", toClassName(parsed.Package))
	if len(parsed.Functions) > 0 {
		exampleFunc := lowerFirst(parsed.Functions[0].Name)
		fmt.Printf("  const result = wasm.%s(...);\n", exampleFunc)
	}
	return nil
}

func generateWorkerOutput(parsed *parser.ParsedFile, output, wasmPath string) error {
	outputDir := filepath.Dir(output)

	// Generate worker.js
	workerPath := filepath.Join(outputDir, "worker.js")
	if err := os.WriteFile(workerPath, []byte(generator.GenerateWorker(wasmPath)), 0600); err != nil {
		return fmt.Errorf("writing worker: %w", err)
	}

	// Generate client.ts
	clientContent := generator.GenerateClient(parsed)
	if err := os.WriteFile(output, []byte(clientContent), 0600); err != nil {
		return fmt.Errorf("writing client: %w", err)
	}

	fmt.Printf("\nGenerated %s (Web Worker entry point)\n", workerPath)
	fmt.Printf("Generated %s with %d function(s) (worker mode)\n", output, len(parsed.Functions))
	fmt.Println("\nUsage:")
	fmt.Printf("  import { %s } from './client';\n", toClassName(parsed.Package))
	fmt.Printf("  const wasm = await %s.init('./worker.js');\n", toClassName(parsed.Package))
	if len(parsed.Functions) > 0 {
		exampleFunc := lowerFirst(parsed.Functions[0].Name)
		fmt.Printf("  const result = await wasm.%s(...);\n", exampleFunc)
	}
	fmt.Printf("  wasm.terminate();\n")
	return nil
}

// toClassName converts a Go package name to a TypeScript class name.
func toClassName(packageName string) string {
	if packageName == "" {
		return "Wasm"
	}
	return strings.ToUpper(packageName[:1]) + packageName[1:]
}

// lowerFirst converts first letter to lowercase
func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}
