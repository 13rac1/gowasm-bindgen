package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/13rac1/gowasm-bindgen/internal/generator"
	"github.com/13rac1/gowasm-bindgen/internal/parser"
	"github.com/13rac1/gowasm-bindgen/internal/validator"
)

// Config holds CLI configuration for testability.
type Config struct {
	SourceFile string
	OutputDir  string
	NoBuild    bool
	Compiler   string
	Mode       string
	ClassName  string
	Optimize   bool
	Verbose    bool
	Stdout     io.Writer
	Stderr     io.Writer
}

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

	cfg := Config{
		SourceFile: flag.Arg(0),
		OutputDir:  outputDir,
		NoBuild:    noBuild,
		Compiler:   compiler,
		Mode:       mode,
		ClassName:  className,
		Optimize:   optimize,
		Verbose:    verbose,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
	}

	return execute(cfg)
}

// execute runs the generator with the given configuration.
// This is separated from run() for testability.
func execute(cfg Config) error {
	sourceDir := filepath.Dir(cfg.SourceFile)
	dirName := filepath.Base(sourceDir)
	if dirName == "." || dirName == "" {
		dirName = "main"
	}

	// Derive class name (can be overridden with --class-name)
	className := cfg.ClassName
	if className == "" {
		className = generator.DeriveClassName(dirName)
	}

	// Derive output paths
	tsFilename := generator.ToKebabCase(className) + ".ts"
	tsOutput := filepath.Join(cfg.OutputDir, tsFilename)
	goOutput := filepath.Join(sourceDir, "bindings_gen.go")
	wasmFile := filepath.Join(cfg.OutputDir, dirName+".wasm")
	wasmURL := dirName + ".wasm"

	if cfg.Verbose {
		//nolint:errcheck // debug output errors are not critical
		fmt.Fprintf(cfg.Stderr, "[DEBUG] Source file: %s\n", cfg.SourceFile)
		fmt.Fprintf(cfg.Stderr, "[DEBUG] Source dir: %s\n", sourceDir)     //nolint:errcheck
		fmt.Fprintf(cfg.Stderr, "[DEBUG] Output dir: %s\n", cfg.OutputDir) //nolint:errcheck
		fmt.Fprintf(cfg.Stderr, "[DEBUG] TS output: %s\n", tsOutput)       //nolint:errcheck
		fmt.Fprintf(cfg.Stderr, "[DEBUG] Go output: %s\n", goOutput)       //nolint:errcheck
		fmt.Fprintf(cfg.Stderr, "[DEBUG] WASM file: %s\n", wasmFile)       //nolint:errcheck
		fmt.Fprintf(cfg.Stderr, "[DEBUG] Mode: %s\n", cfg.Mode)            //nolint:errcheck
		fmt.Fprintf(cfg.Stderr, "[DEBUG] Class name: %s\n", className)     //nolint:errcheck
		fmt.Fprintf(cfg.Stderr, "[DEBUG] Compiler: %s\n", cfg.Compiler)    //nolint:errcheck
		fmt.Fprintf(cfg.Stderr, "[DEBUG] Optimize: %v\n", cfg.Optimize)    //nolint:errcheck
		fmt.Fprintf(cfg.Stderr, "[DEBUG] No build: %v\n", cfg.NoBuild)     //nolint:errcheck
	}

	// Check if source file exists
	if _, err := os.Stat(cfg.SourceFile); err != nil {
		return fmt.Errorf("source file not found: %s", cfg.SourceFile)
	}

	// Parse source file
	fmt.Fprintf(cfg.Stdout, "Parsing %s...\n", cfg.SourceFile) //nolint:errcheck
	parsed, err := parser.ParseSourceFile(cfg.SourceFile)
	if err != nil {
		return fmt.Errorf("parsing source file: %w", err)
	}

	fmt.Fprintf(cfg.Stdout, "Package: %s\n", parsed.Package)                           //nolint:errcheck
	fmt.Fprintf(cfg.Stdout, "Found %d exported function(s):\n", len(parsed.Functions)) //nolint:errcheck
	for _, fn := range parsed.Functions {
		fmt.Fprintf(cfg.Stdout, "  - %s\n", fn.Name) //nolint:errcheck
	}

	if cfg.Verbose {
		fmt.Fprintf(cfg.Stderr, "[DEBUG] Found %d types\n", len(parsed.Types)) //nolint:errcheck
		for _, fn := range parsed.Functions {
			fmt.Fprintf(cfg.Stderr, "[DEBUG] Function %s: %d params, %d returns\n", fn.Name, len(fn.Params), len(fn.Returns)) //nolint:errcheck
		}
	}

	if len(parsed.Functions) == 0 {
		return fmt.Errorf("no exported functions found in %s\n\n"+
			"Functions must be exported (start with uppercase letter) and have no receiver", cfg.SourceFile)
	}

	// Check for select {} in main (required for WASM to stay alive)
	if parsed.Package == "main" {
		hasSelect, err := parser.HasSelectInMain(cfg.SourceFile)
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
	if err := os.MkdirAll(cfg.OutputDir, 0750); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Generate Go bindings
	fmt.Fprintf(cfg.Stdout, "\nGenerating Go bindings...\n") //nolint:errcheck
	workerMode := cfg.Mode == "worker"
	bindingsCode := generator.GenerateGoBindings(parsed, workerMode)

	if err := os.WriteFile(goOutput, []byte(bindingsCode), 0600); err != nil {
		return fmt.Errorf("writing Go bindings: %w", err)
	}
	fmt.Fprintf(cfg.Stdout, "Generated %s\n", goOutput) //nolint:errcheck

	// Generate TypeScript client
	if cfg.Mode == "sync" {
		if cfg.Verbose {
			fmt.Fprintf(cfg.Stderr, "[DEBUG] Generating sync mode client\n") //nolint:errcheck
		}
		if err := generateSyncOutput(parsed, tsOutput, className); err != nil {
			return err
		}
	} else {
		if cfg.Verbose {
			fmt.Fprintf(cfg.Stderr, "[DEBUG] Generating worker mode client\n") //nolint:errcheck
		}
		if err := generateWorkerOutput(parsed, tsOutput, wasmURL, className); err != nil {
			return err
		}
	}

	// Stop here if --no-build
	if cfg.NoBuild {
		return nil
	}

	// Copy wasm_exec.js
	fmt.Fprintf(cfg.Stdout, "\nCopying wasm_exec.js...\n") //nolint:errcheck
	if err := copyWasmExec(cfg.Compiler, cfg.OutputDir); err != nil {
		return err
	}

	// Compile WASM
	fmt.Fprintf(cfg.Stdout, "\nCompiling WASM with %s...\n", cfg.Compiler) //nolint:errcheck
	if err := compileWasm(sourceDir, wasmFile, cfg.Compiler, cfg.Optimize); err != nil {
		return fmt.Errorf("compiling WASM: %w", err)
	}

	fmt.Fprintf(cfg.Stdout, "\nBuild complete!\n") //nolint:errcheck
	fmt.Fprintf(cfg.Stdout, "  %s\n", tsOutput)    //nolint:errcheck
	if cfg.Mode == "worker" {
		fmt.Fprintf(cfg.Stdout, "  %s\n", filepath.Join(cfg.OutputDir, "worker.js")) //nolint:errcheck
	}
	fmt.Fprintf(cfg.Stdout, "  %s\n", filepath.Join(cfg.OutputDir, "wasm_exec.js")) //nolint:errcheck
	fmt.Fprintf(cfg.Stdout, "  %s\n", wasmFile)                                     //nolint:errcheck

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
		exampleFunc := generator.LowerFirst(parsed.Functions[0].Name)
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
		exampleFunc := generator.LowerFirst(parsed.Functions[0].Name)
		fmt.Printf("  const result = await wasm.%s(...);\n", exampleFunc)
	}
	fmt.Printf("  wasm.terminate();\n")
	return nil
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
