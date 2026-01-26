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
	var tsOutput string
	var goOutput string
	var mode string
	var verbose bool
	var wasmURL string
	var className string

	flag.StringVarP(&tsOutput, "ts-output", "t", "", "TypeScript output file (e.g., client.ts)")
	flag.StringVarP(&goOutput, "go-output", "g", "", "Go bindings output file (e.g., bindings_gen.go)")
	flag.StringVarP(&mode, "mode", "m", "worker", "generation mode: 'sync' or 'worker'")
	flag.BoolVarP(&verbose, "verbose", "v", false, "enable verbose debug output")
	flag.StringVarP(&wasmURL, "wasm-url", "w", "", "WASM URL in generated fetch() (default: <dirname>.wasm)")
	flag.StringVarP(&className, "class-name", "c", "", "TypeScript class name (default: Go<DirName>)")
	flag.Parse()

	// Validate flags
	usage := "Usage: gowasm-bindgen <source.go> -t client.ts [-g bindings_gen.go] [-m sync|worker] [-w app.wasm] [-c ClassName]"
	if flag.NArg() == 0 {
		return fmt.Errorf("missing source file argument\n\n%s", usage)
	}
	if tsOutput == "" {
		return fmt.Errorf("-t/--ts-output is required\n\n%s", usage)
	}
	if mode != "sync" && mode != "worker" {
		return fmt.Errorf("--mode must be 'sync' or 'worker', got %q\n\n%s", mode, usage)
	}

	sourceFile := flag.Arg(0)

	if verbose {
		fmt.Fprintf(os.Stderr, "[DEBUG] Source file: %s\n", sourceFile)
		fmt.Fprintf(os.Stderr, "[DEBUG] TS output: %s\n", tsOutput)
		fmt.Fprintf(os.Stderr, "[DEBUG] Mode: %s\n", mode)
		fmt.Fprintf(os.Stderr, "[DEBUG] WASM URL: %s\n", wasmURL)
		fmt.Fprintf(os.Stderr, "[DEBUG] Class name: %s\n", className)
	}

	// Check if source file exists
	if _, err := os.Stat(sourceFile); err != nil {
		return fmt.Errorf("source file not found: %s", sourceFile)
	}

	// Derive default WASM URL from directory name if not specified
	dirName := filepath.Base(filepath.Dir(sourceFile))
	if wasmURL == "" {
		if dirName == "." || dirName == "" {
			wasmURL = "main.wasm"
		} else {
			wasmURL = dirName + ".wasm"
		}
	}

	// Derive default class name from directory name if not specified
	if className == "" {
		className = generator.DeriveClassName(dirName)
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
	if dir := filepath.Dir(tsOutput); dir != "." {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
	}

	// Generate Go bindings if requested
	// workerMode is true when using worker mode (default)
	if goOutput != "" {
		fmt.Printf("\nGenerating Go bindings...\n")
		workerMode := mode == "worker"
		bindingsCode := generator.GenerateGoBindings(parsed, workerMode)

		if err := os.WriteFile(goOutput, []byte(bindingsCode), 0600); err != nil {
			return fmt.Errorf("writing Go bindings: %w", err)
		}

		fmt.Printf("Generated %s\n", goOutput)
	}

	// Generate TypeScript client
	if mode == "sync" {
		if verbose {
			fmt.Fprintf(os.Stderr, "[DEBUG] Generating sync mode client\n")
		}
		return generateSyncOutput(parsed, tsOutput, className)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[DEBUG] Generating worker mode client\n")
	}
	return generateWorkerOutput(parsed, tsOutput, wasmURL, className)
}

func generateSyncOutput(parsed *parser.ParsedFile, output, className string) error {
	// Generate TypeScript class-based client
	content := generator.Generate(parsed, filepath.Base(output), className)

	// Write output
	if err := os.WriteFile(output, []byte(content), 0600); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	fmt.Printf("\nGenerated %s with %d function(s) (sync mode)\n", output, len(parsed.Functions))
	fmt.Println("\nUsage:")
	fmt.Printf("  import { %s } from './client';\n", className)
	fmt.Printf("  const wasm = await %s.init('./example.wasm');\n", className)
	if len(parsed.Functions) > 0 {
		exampleFunc := lowerFirst(parsed.Functions[0].Name)
		fmt.Printf("  const result = wasm.%s(...);\n", exampleFunc)
	}
	return nil
}

func generateWorkerOutput(parsed *parser.ParsedFile, output, wasmPath, className string) error {
	outputDir := filepath.Dir(output)

	// Generate worker.js
	workerPath := filepath.Join(outputDir, "worker.js")
	if err := os.WriteFile(workerPath, []byte(generator.GenerateWorker(wasmPath)), 0600); err != nil {
		return fmt.Errorf("writing worker: %w", err)
	}

	// Generate client.ts
	clientContent := generator.GenerateClient(parsed, filepath.Base(output), className)
	if err := os.WriteFile(output, []byte(clientContent), 0600); err != nil {
		return fmt.Errorf("writing client: %w", err)
	}

	fmt.Printf("\nGenerated %s (Web Worker entry point)\n", workerPath)
	fmt.Printf("Generated %s with %d function(s) (worker mode)\n", output, len(parsed.Functions))
	fmt.Println("\nUsage:")
	fmt.Printf("  import { %s } from './client';\n", className)
	fmt.Printf("  const wasm = await %s.init('./worker.js');\n", className)
	if len(parsed.Functions) > 0 {
		exampleFunc := lowerFirst(parsed.Functions[0].Name)
		fmt.Printf("  const result = await wasm.%s(...);\n", exampleFunc)
	}
	fmt.Printf("  wasm.terminate();\n")
	return nil
}

// lowerFirst converts first letter to lowercase
func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}
