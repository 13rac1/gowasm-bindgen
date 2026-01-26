package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

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

//nolint:gocyclo // CLI entry point has linear flow that benefits from being in one function
func run() error {
	// Parse flags
	var outputDir string
	var noBuild bool
	var compiler string
	var mode string
	var className string
	var optimize bool
	var verbose bool

	flag.CommandLine.SetInterspersed(true) // Allow flags after positional arguments
	flag.StringVarP(&outputDir, "output", "o", "generated", "Output directory for all artifacts")
	flag.BoolVar(&noBuild, "no-build", false, "Skip WASM compilation (generate only)")
	flag.StringVar(&compiler, "compiler", "tinygo", "Compiler: 'tinygo' or 'go'")
	flag.StringVarP(&mode, "mode", "m", "worker", "Generation mode: 'sync' or 'worker'")
	flag.StringVarP(&className, "class-name", "c", "", "TypeScript class name (default: Go<DirName>)")
	flag.BoolVar(&optimize, "optimize", true, "Enable size optimizations (tinygo only)")
	flag.BoolVarP(&verbose, "verbose", "v", false, "Enable verbose debug output")
	flag.Parse()

	// Validate flags
	usage := "Usage: gowasm-bindgen <source.go> [-o generated] [--no-build] [--compiler tinygo|go] [-m sync|worker] [-c ClassName]"
	if flag.NArg() == 0 {
		return fmt.Errorf("missing source file argument\n\n%s", usage)
	}
	if mode != "sync" && mode != "worker" {
		return fmt.Errorf("--mode must be 'sync' or 'worker', got %q\n\n%s", mode, usage)
	}
	if compiler != "tinygo" && compiler != "go" {
		return fmt.Errorf("--compiler must be 'tinygo' or 'go', got %q\n\n%s", compiler, usage)
	}

	sourceFile := flag.Arg(0)
	sourceDir := filepath.Dir(sourceFile)
	dirName := filepath.Base(sourceDir)
	if dirName == "." || dirName == "" {
		dirName = "main"
	}

	// Derive class name (can be overridden with --class-name)
	if className == "" {
		className = generator.DeriveClassName(dirName)
	}

	// Derive output paths
	tsFilename := toKebabCase(className) + ".ts"
	tsOutput := filepath.Join(outputDir, tsFilename)
	goOutput := filepath.Join(sourceDir, "bindings_gen.go")
	wasmFile := filepath.Join(outputDir, dirName+".wasm")
	wasmURL := dirName + ".wasm"

	if verbose {
		fmt.Fprintf(os.Stderr, "[DEBUG] Source file: %s\n", sourceFile)
		fmt.Fprintf(os.Stderr, "[DEBUG] Source dir: %s\n", sourceDir)
		fmt.Fprintf(os.Stderr, "[DEBUG] Output dir: %s\n", outputDir)
		fmt.Fprintf(os.Stderr, "[DEBUG] TS output: %s\n", tsOutput)
		fmt.Fprintf(os.Stderr, "[DEBUG] Go output: %s\n", goOutput)
		fmt.Fprintf(os.Stderr, "[DEBUG] WASM file: %s\n", wasmFile)
		fmt.Fprintf(os.Stderr, "[DEBUG] Mode: %s\n", mode)
		fmt.Fprintf(os.Stderr, "[DEBUG] Class name: %s\n", className)
		fmt.Fprintf(os.Stderr, "[DEBUG] Compiler: %s\n", compiler)
		fmt.Fprintf(os.Stderr, "[DEBUG] Optimize: %v\n", optimize)
		fmt.Fprintf(os.Stderr, "[DEBUG] No build: %v\n", noBuild)
	}

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

	// Create output directory
	if err := os.MkdirAll(outputDir, 0750); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Generate Go bindings
	fmt.Printf("\nGenerating Go bindings...\n")
	workerMode := mode == "worker"
	bindingsCode := generator.GenerateGoBindings(parsed, workerMode)

	if err := os.WriteFile(goOutput, []byte(bindingsCode), 0600); err != nil {
		return fmt.Errorf("writing Go bindings: %w", err)
	}
	fmt.Printf("Generated %s\n", goOutput)

	// Generate TypeScript client
	if mode == "sync" {
		if verbose {
			fmt.Fprintf(os.Stderr, "[DEBUG] Generating sync mode client\n")
		}
		if err := generateSyncOutput(parsed, tsOutput, className); err != nil {
			return err
		}
	} else {
		if verbose {
			fmt.Fprintf(os.Stderr, "[DEBUG] Generating worker mode client\n")
		}
		if err := generateWorkerOutput(parsed, tsOutput, wasmURL, className); err != nil {
			return err
		}
	}

	// Stop here if --no-build
	if noBuild {
		return nil
	}

	// Copy wasm_exec.js
	fmt.Printf("\nCopying wasm_exec.js...\n")
	if err := copyWasmExec(compiler, outputDir); err != nil {
		return err
	}

	// Compile WASM
	fmt.Printf("\nCompiling WASM with %s...\n", compiler)
	if err := compileWasm(sourceDir, wasmFile, compiler, optimize); err != nil {
		return fmt.Errorf("compiling WASM: %w", err)
	}

	fmt.Printf("\nBuild complete!\n")
	fmt.Printf("  %s\n", tsOutput)
	if mode == "worker" {
		fmt.Printf("  %s\n", filepath.Join(outputDir, "worker.js"))
	}
	fmt.Printf("  %s\n", filepath.Join(outputDir, "wasm_exec.js"))
	fmt.Printf("  %s\n", wasmFile)

	return nil
}

func generateSyncOutput(parsed *parser.ParsedFile, output, className string) error {
	// Generate TypeScript class-based client
	content := generator.Generate(parsed, filepath.Base(output), className)

	// Write output
	if err := os.WriteFile(output, []byte(content), 0600); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	// Derive import path (strip .ts extension)
	importPath := "./" + strings.TrimSuffix(filepath.Base(output), ".ts")

	fmt.Printf("\nGenerated %s with %d function(s) (sync mode)\n", output, len(parsed.Functions))
	fmt.Println("\nUsage:")
	fmt.Printf("  import { %s } from '%s';\n", className, importPath)
	fmt.Printf("  const wasm = await %s.init('./<name>.wasm');\n", className)
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

	// Derive import path (strip .ts extension)
	importPath := "./" + strings.TrimSuffix(filepath.Base(output), ".ts")

	fmt.Printf("\nGenerated %s (Web Worker entry point)\n", workerPath)
	fmt.Printf("Generated %s with %d function(s) (worker mode)\n", output, len(parsed.Functions))
	fmt.Println("\nUsage:")
	fmt.Printf("  import { %s } from '%s';\n", className, importPath)
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

// toKebabCase converts "GoMain" to "go-main"
func toKebabCase(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '-')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// copyWasmExec copies the wasm_exec.js runtime from the compiler installation
func copyWasmExec(compiler, destDir string) error {
	srcPath, err := getWasmExecPath(compiler)
	if err != nil {
		return err
	}
	if _, err := os.Stat(srcPath); err != nil {
		return fmt.Errorf("wasm_exec.js not found at %s: %w", srcPath, err)
	}
	destPath := filepath.Join(destDir, "wasm_exec.js")
	if err := copyFile(srcPath, destPath); err != nil {
		return fmt.Errorf("copying wasm_exec.js: %w", err)
	}
	fmt.Printf("Copied %s\n", destPath)
	return nil
}

// getWasmExecPath returns the path to wasm_exec.js for the given compiler
func getWasmExecPath(compiler string) (string, error) {
	if compiler == "tinygo" {
		out, err := exec.Command("tinygo", "env", "TINYGOROOT").Output()
		if err != nil {
			return "", fmt.Errorf("tinygo not found: %w", err)
		}
		return filepath.Join(strings.TrimSpace(string(out)), "targets", "wasm_exec.js"), nil
	}
	out, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return "", fmt.Errorf("go not found: %w", err)
	}
	return filepath.Join(strings.TrimSpace(string(out)), "lib", "wasm", "wasm_exec.js"), nil
}

// compileWasm compiles the Go source to WASM
func compileWasm(sourceDir, outputFile, compiler string, optimize bool) error {
	// Make output path absolute since we'll change to sourceDir
	if !filepath.IsAbs(outputFile) {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}
		outputFile = filepath.Join(cwd, outputFile)
	}

	var cmd *exec.Cmd
	if compiler == "tinygo" {
		args := []string{"build", "-o", outputFile, "-target", "wasm"}
		if optimize {
			args = append(args, "-opt=z", "-no-debug", "-panic=trap")
		}
		args = append(args, ".")
		cmd = exec.Command("tinygo", args...) //nolint:gosec // args are validated
	} else {
		cmd = exec.Command("go", "build", "-o", outputFile, ".") //nolint:gosec // args are validated
		cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	}
	cmd.Dir = sourceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running %s: %w", compiler, err)
	}
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src) //nolint:gosec // src is from trusted source (compiler path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", src, err)
	}
	if err := os.WriteFile(dst, data, 0600); err != nil {
		return fmt.Errorf("writing %s: %w", dst, err)
	}
	return nil
}
