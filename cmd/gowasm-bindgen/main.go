package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	flag.StringVar(&output, "output", "", "output TypeScript file path (e.g., client.ts)")
	flag.StringVar(&goOutput, "go-output", "", "output Go bindings file path (e.g., bindings_gen.go)")
	flag.BoolVar(&sync, "sync", false, "generate synchronous API (default: worker-based async)")
	flag.Parse()

	// Validate flags
	if flag.NArg() == 0 {
		return fmt.Errorf("missing source file argument\n\n" +
			"Usage: gowasm-bindgen <source.go> --output client.ts [--go-output bindings_gen.go] [--sync]")
	}
	if output == "" {
		return fmt.Errorf("--output is required\n\n" +
			"Usage: gowasm-bindgen <source.go> --output client.ts [--go-output bindings_gen.go] [--sync]")
	}

	sourceFile := flag.Arg(0)

	// Check if source file exists
	if _, err := os.Stat(sourceFile); err != nil {
		return fmt.Errorf("source file not found: %s", sourceFile)
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

	if len(parsed.Functions) == 0 {
		return fmt.Errorf("no exported functions found in %s\n\n"+
			"Functions must be exported (start with uppercase letter) and have no receiver", sourceFile)
	}

	// Validate functions
	if err := validator.ValidateFunctions(parsed); err != nil {
		return err
	}

	// Create output directory if needed
	if dir := filepath.Dir(output); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
	}

	// Generate Go bindings if requested
	if goOutput != "" {
		fmt.Printf("\nGenerating Go bindings...\n")
		bindingsCode := generator.GenerateGoBindings(parsed)

		if err := os.WriteFile(goOutput, []byte(bindingsCode), 0644); err != nil {
			return fmt.Errorf("writing Go bindings: %w", err)
		}

		fmt.Printf("Generated %s\n", goOutput)
	}

	// Generate TypeScript client
	if sync {
		return generateSyncOutput(parsed, output)
	}

	return generateWorkerOutput(parsed, output)
}

func generateSyncOutput(parsed *parser.ParsedFile, output string) error {
	// Generate TypeScript class-based client
	content := generator.Generate(parsed)

	// Write output
	if err := os.WriteFile(output, []byte(content), 0644); err != nil {
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

func generateWorkerOutput(parsed *parser.ParsedFile, output string) error {
	// Check for callbacks - they cannot work in worker mode
	if funcs := functionsWithCallbacks(parsed); len(funcs) > 0 {
		return fmt.Errorf("function %q has callback parameter which cannot work in Worker mode.\n"+
			"Callbacks require --sync mode because Web Workers cannot serialize functions.\n\n"+
			"To use callbacks, regenerate with: gowasm-bindgen --sync --output %s ...", funcs[0], output)
	}

	outputDir := filepath.Dir(output)

	// Generate worker.js
	workerPath := filepath.Join(outputDir, "worker.js")
	if err := os.WriteFile(workerPath, []byte(generator.GenerateWorker()), 0644); err != nil {
		return fmt.Errorf("writing worker: %w", err)
	}

	// Generate client.ts
	clientContent := generator.GenerateClient(parsed)
	if err := os.WriteFile(output, []byte(clientContent), 0644); err != nil {
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

// functionsWithCallbacks returns names of functions that have callback parameters
func functionsWithCallbacks(parsed *parser.ParsedFile) []string {
	var funcs []string
	for _, fn := range parsed.Functions {
		for _, param := range fn.Params {
			if param.Type.Kind == parser.KindFunction {
				funcs = append(funcs, fn.Name)
				break
			}
		}
	}
	return funcs
}
