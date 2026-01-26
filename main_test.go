package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/13rac1/gowasm-bindgen/internal/parser"
)

// Unit tests for functions in main.go

func TestCopyFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "copyfile-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create source file
	srcPath := filepath.Join(tmpDir, "source.txt")
	content := "test content for copy"
	if err := os.WriteFile(srcPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	// Copy file
	dstPath := filepath.Join(tmpDir, "dest.txt")
	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	// Verify destination content
	got, err := os.ReadFile(dstPath) //nolint:gosec // test file path
	if err != nil {
		t.Fatalf("failed to read dest file: %v", err)
	}
	if string(got) != content {
		t.Errorf("copyFile content mismatch: got %q, want %q", got, content)
	}
}

func TestCopyFile_SourceNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "copyfile-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	err = copyFile(filepath.Join(tmpDir, "nonexistent"), filepath.Join(tmpDir, "dest"))
	if err == nil {
		t.Error("expected error for nonexistent source file")
	}
}

func TestGenerateSyncOutput(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sync-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	parsed := &parser.ParsedFile{
		Package: "main",
		Functions: []parser.GoFunction{
			{Name: "Greet", Params: []parser.GoParameter{{Name: "name", Type: parser.GoType{Kind: parser.KindPrimitive, Name: "string"}}}, Returns: []parser.GoType{{Kind: parser.KindPrimitive, Name: "string"}}},
		},
	}

	output := filepath.Join(tmpDir, "test-client.ts")
	if err := generateSyncOutput(parsed, output, "TestClass"); err != nil {
		t.Fatalf("generateSyncOutput failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("output file was not created")
	}

	// Verify content
	content, _ := os.ReadFile(output) //nolint:gosec // test file path
	if !strings.Contains(string(content), "class TestClass") {
		t.Error("output should contain class definition")
	}
}

func TestGenerateWorkerOutput(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "worker-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	parsed := &parser.ParsedFile{
		Package: "main",
		Functions: []parser.GoFunction{
			{Name: "Greet", Params: []parser.GoParameter{{Name: "name", Type: parser.GoType{Kind: parser.KindPrimitive, Name: "string"}}}, Returns: []parser.GoType{{Kind: parser.KindPrimitive, Name: "string"}}},
		},
	}

	output := filepath.Join(tmpDir, "test-client.ts")
	if err := generateWorkerOutput(parsed, output, "test.wasm", "TestClass"); err != nil {
		t.Fatalf("generateWorkerOutput failed: %v", err)
	}

	// Verify client file was created
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("client file was not created")
	}

	// Verify worker file was created
	workerPath := filepath.Join(tmpDir, "worker.js")
	if _, err := os.Stat(workerPath); os.IsNotExist(err) {
		t.Error("worker file was not created")
	}

	// Verify client content
	content, _ := os.ReadFile(output) //nolint:gosec // test file path
	if !strings.Contains(string(content), "class TestClass") {
		t.Error("client should contain class definition")
	}

	// Verify worker content references wasm file
	workerContent, _ := os.ReadFile(workerPath) //nolint:gosec // test file path
	if !strings.Contains(string(workerContent), "test.wasm") {
		t.Error("worker should reference wasm file")
	}
}

func TestGetWasmExecPath_Go(t *testing.T) {
	path, err := getWasmExecPath("go")
	if err != nil {
		t.Fatalf("getWasmExecPath failed: %v", err)
	}
	if !strings.Contains(path, "wasm_exec.js") {
		t.Errorf("path should contain wasm_exec.js, got: %s", path)
	}
}

func TestGetWasmExecPath_TinyGo(t *testing.T) {
	// Skip if tinygo is not installed
	if _, err := exec.LookPath("tinygo"); err != nil {
		t.Skip("tinygo not installed")
	}

	path, err := getWasmExecPath("tinygo")
	if err != nil {
		t.Fatalf("getWasmExecPath failed: %v", err)
	}
	if !strings.Contains(path, "wasm_exec.js") {
		t.Errorf("path should contain wasm_exec.js, got: %s", path)
	}
}

func TestCLI_MissingSourceFile(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--no-build") //nolint:gosec // test command
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error for missing source file")
	}
	if !strings.Contains(string(output), "missing source file") {
		t.Errorf("expected 'missing source file' error, got: %s", output)
	}
}

func TestCLI_InvalidMode(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--mode", "invalid", "test/e2e/wasm/main.go") //nolint:gosec // test command
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error for invalid mode")
	}
	if !strings.Contains(string(output), "must be 'sync' or 'worker'") {
		t.Errorf("expected mode error, got: %s", output)
	}
}

func TestCLI_InvalidCompiler(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--compiler", "rustc", "test/e2e/wasm/main.go") //nolint:gosec // test command
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error for invalid compiler")
	}
	if !strings.Contains(string(output), "must be 'tinygo' or 'go'") {
		t.Errorf("expected compiler error, got: %s", output)
	}
}

func TestCLI_SourceFileNotFound(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--no-build", "nonexistent/main.go") //nolint:gosec // test command
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error for nonexistent source file")
	}
	if !strings.Contains(string(output), "source file not found") {
		t.Errorf("expected 'source file not found' error, got: %s", output)
	}
}

func TestCLI_NoBuildGeneratesFiles(t *testing.T) {
	// Create temp directory for output
	tmpDir, err := os.MkdirTemp("", "gowasm-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Run with --no-build to generate only
	cmd := exec.Command("go", "run", ".", "--no-build", "--output", tmpDir, "test/e2e/wasm/main.go") //nolint:gosec // test command
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

	// Verify TypeScript client was generated
	tsFile := filepath.Join(tmpDir, "go-wasm.ts")
	if _, err := os.Stat(tsFile); os.IsNotExist(err) {
		t.Errorf("TypeScript client not generated at %s", tsFile)
	}

	// Verify worker.js was generated (worker mode is default)
	workerFile := filepath.Join(tmpDir, "worker.js")
	if _, err := os.Stat(workerFile); os.IsNotExist(err) {
		t.Errorf("Worker not generated at %s", workerFile)
	}

	// Verify bindings_gen.go was generated
	bindingsFile := filepath.Join("test/e2e/wasm", "bindings_gen.go")
	if _, err := os.Stat(bindingsFile); os.IsNotExist(err) {
		t.Errorf("Go bindings not generated at %s", bindingsFile)
	}
}

func TestCLI_SyncModeNoWorker(t *testing.T) {
	// Create temp directory for output
	tmpDir, err := os.MkdirTemp("", "gowasm-test-sync-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Run with --no-build and --mode sync
	cmd := exec.Command("go", "run", ".", "--no-build", "--output", tmpDir, "--mode", "sync", "test/e2e/wasm/main.go") //nolint:gosec // test command
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

	// Verify TypeScript client was generated
	tsFile := filepath.Join(tmpDir, "go-wasm.ts")
	if _, err := os.Stat(tsFile); os.IsNotExist(err) {
		t.Errorf("TypeScript client not generated at %s", tsFile)
	}

	// Verify worker.js was NOT generated in sync mode
	workerFile := filepath.Join(tmpDir, "worker.js")
	if _, err := os.Stat(workerFile); !os.IsNotExist(err) {
		t.Errorf("Worker should not be generated in sync mode, but found at %s", workerFile)
	}
}

func TestCLI_CustomClassName(t *testing.T) {
	// Create temp directory for output
	tmpDir, err := os.MkdirTemp("", "gowasm-test-class-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Run with custom class name
	cmd := exec.Command("go", "run", ".", "--no-build", "--output", tmpDir, "--class-name", "MyCustomWasm", "test/e2e/wasm/main.go") //nolint:gosec // test command
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

	// Verify TypeScript client was generated with custom name (kebab-case)
	tsFile := filepath.Join(tmpDir, "my-custom-wasm.ts")
	if _, err := os.Stat(tsFile); os.IsNotExist(err) {
		t.Errorf("TypeScript client not generated at %s", tsFile)
	}

	// Verify content contains the custom class name
	content, err := os.ReadFile(tsFile) //nolint:gosec // test file path
	if err != nil {
		t.Fatalf("failed to read TypeScript file: %v", err)
	}
	if !strings.Contains(string(content), "class MyCustomWasm") {
		t.Errorf("TypeScript file should contain 'class MyCustomWasm', got:\n%s", content)
	}
}

func TestCLI_NoExportedFunctions(t *testing.T) {
	// Create temp directory with a Go file that has no exported functions
	tmpDir, err := os.MkdirTemp("", "gowasm-test-noexport-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a Go file with no exported functions
	goFile := filepath.Join(tmpDir, "main.go")
	content := `package main

func main() {
	select {}
}

func privateFunc() string {
	return "private"
}
`
	if err := os.WriteFile(goFile, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Run the CLI
	cmd := exec.Command("go", "run", ".", "--no-build", "--output", tmpDir, goFile) //nolint:gosec // test command
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error for no exported functions")
	}
	if !strings.Contains(string(output), "no exported functions") {
		t.Errorf("expected 'no exported functions' error, got: %s", output)
	}
}

func TestCLI_MissingSelectInMain(t *testing.T) {
	// Create temp directory with a Go file that lacks select {}
	tmpDir, err := os.MkdirTemp("", "gowasm-test-noselect-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a Go file without select {}
	goFile := filepath.Join(tmpDir, "main.go")
	content := `package main

func Greet(name string) string {
	return "Hello, " + name
}

func main() {
	// missing select {}
}
`
	if err := os.WriteFile(goFile, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Run the CLI
	cmd := exec.Command("go", "run", ".", "--no-build", "--output", tmpDir, goFile) //nolint:gosec // test command
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error for missing select {}")
	}
	if !strings.Contains(string(output), "select {}") {
		t.Errorf("expected 'select {}' error, got: %s", output)
	}
}

func TestCopyWasmExec_Go(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wasmexec-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if err := copyWasmExec("go", tmpDir); err != nil {
		t.Fatalf("copyWasmExec failed: %v", err)
	}

	// Verify wasm_exec.js was copied
	destPath := filepath.Join(tmpDir, "wasm_exec.js")
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("wasm_exec.js not copied to %s", destPath)
	}

	// Verify content is non-empty
	content, err := os.ReadFile(destPath) //nolint:gosec // test file path
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}
	if len(content) < 1000 {
		t.Errorf("wasm_exec.js seems too short: %d bytes", len(content))
	}
}

func TestCopyWasmExec_TinyGo(t *testing.T) {
	// Skip if tinygo is not installed
	if _, err := exec.LookPath("tinygo"); err != nil {
		t.Skip("tinygo not installed")
	}

	tmpDir, err := os.MkdirTemp("", "wasmexec-tinygo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if err := copyWasmExec("tinygo", tmpDir); err != nil {
		t.Fatalf("copyWasmExec failed: %v", err)
	}

	// Verify wasm_exec.js was copied
	destPath := filepath.Join(tmpDir, "wasm_exec.js")
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("wasm_exec.js not copied to %s", destPath)
	}
}

func TestCopyFile_DestDirNotExist(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "copyfile-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create source file
	srcPath := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(srcPath, []byte("test"), 0600); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	// Try to copy to non-existent directory
	err = copyFile(srcPath, filepath.Join(tmpDir, "nonexistent", "dest.txt"))
	if err == nil {
		t.Error("expected error for nonexistent destination directory")
	}
}

func TestGenerateSyncOutput_NoFunctions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sync-test-nofunc-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	parsed := &parser.ParsedFile{
		Package:   "main",
		Functions: []parser.GoFunction{}, // No functions
	}

	output := filepath.Join(tmpDir, "test-client.ts")
	if err := generateSyncOutput(parsed, output, "TestClass"); err != nil {
		t.Fatalf("generateSyncOutput failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("output file was not created")
	}
}

func TestGenerateWorkerOutput_NoFunctions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "worker-test-nofunc-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	parsed := &parser.ParsedFile{
		Package:   "main",
		Functions: []parser.GoFunction{}, // No functions
	}

	output := filepath.Join(tmpDir, "test-client.ts")
	if err := generateWorkerOutput(parsed, output, "test.wasm", "TestClass"); err != nil {
		t.Fatalf("generateWorkerOutput failed: %v", err)
	}

	// Verify files were created
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("client file was not created")
	}
	workerPath := filepath.Join(tmpDir, "worker.js")
	if _, err := os.Stat(workerPath); os.IsNotExist(err) {
		t.Error("worker file was not created")
	}
}

func TestGenerateSyncOutput_WritesUsageExample(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sync-test-usage-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test with multiple functions to cover all usage example branches
	parsed := &parser.ParsedFile{
		Package: "main",
		Functions: []parser.GoFunction{
			{Name: "Greet", Params: []parser.GoParameter{{Name: "name", Type: parser.GoType{Kind: parser.KindPrimitive, Name: "string"}}}, Returns: []parser.GoType{{Kind: parser.KindPrimitive, Name: "string"}}},
			{Name: "Calculate", Params: []parser.GoParameter{{Name: "a", Type: parser.GoType{Kind: parser.KindPrimitive, Name: "int"}}}, Returns: []parser.GoType{{Kind: parser.KindPrimitive, Name: "int"}}},
		},
	}

	output := filepath.Join(tmpDir, "test-client.ts")
	if err := generateSyncOutput(parsed, output, "TestClass"); err != nil {
		t.Fatalf("generateSyncOutput failed: %v", err)
	}

	// Verify file content includes both functions
	content, err := os.ReadFile(output) //nolint:gosec // test file path
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if !strings.Contains(string(content), "greet") || !strings.Contains(string(content), "calculate") {
		t.Error("output should contain both methods")
	}
}

func TestGenerateWorkerOutput_WritesUsageExample(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "worker-test-usage-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test with multiple functions
	parsed := &parser.ParsedFile{
		Package: "main",
		Functions: []parser.GoFunction{
			{Name: "Process", Params: []parser.GoParameter{{Name: "data", Type: parser.GoType{Kind: parser.KindPrimitive, Name: "string"}}}, Returns: []parser.GoType{{Kind: parser.KindPrimitive, Name: "string"}}},
		},
	}

	output := filepath.Join(tmpDir, "test-client.ts")
	if err := generateWorkerOutput(parsed, output, "custom.wasm", "CustomClass"); err != nil {
		t.Fatalf("generateWorkerOutput failed: %v", err)
	}

	// Verify worker references custom wasm
	workerContent, err := os.ReadFile(filepath.Join(tmpDir, "worker.js")) //nolint:gosec // test file path
	if err != nil {
		t.Fatalf("failed to read worker file: %v", err)
	}
	if !strings.Contains(string(workerContent), "custom.wasm") {
		t.Error("worker should reference custom.wasm")
	}
}

func TestGetWasmExecPath_InvalidCompiler(t *testing.T) {
	// This exercises error handling paths
	_, err := getWasmExecPath("invalid-compiler")
	// Should fallback to go compiler (the else branch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecute_FullGeneration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "execute-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Use existing e2e test source file
	cfg := Config{
		SourceFile: "test/e2e/wasm/main.go",
		OutputDir:  tmpDir,
		NoBuild:    true, // Skip wasm compilation
		Compiler:   "go",
		Mode:       "worker",
		ClassName:  "", // Test auto-derivation
		Optimize:   false,
		Verbose:    false,
		Stdout:     io.Discard, // Suppress output
		Stderr:     io.Discard,
	}

	if err := execute(cfg); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	// Verify files were created
	tsFile := filepath.Join(tmpDir, "go-wasm.ts")
	if _, err := os.Stat(tsFile); os.IsNotExist(err) {
		t.Errorf("TypeScript client not generated at %s", tsFile)
	}
	workerFile := filepath.Join(tmpDir, "worker.js")
	if _, err := os.Stat(workerFile); os.IsNotExist(err) {
		t.Errorf("Worker not generated at %s", workerFile)
	}
}

func TestExecute_SyncMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "execute-sync-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	cfg := Config{
		SourceFile: "test/e2e/wasm/main.go",
		OutputDir:  tmpDir,
		NoBuild:    true,
		Compiler:   "go",
		Mode:       "sync",
		ClassName:  "CustomClass",
		Optimize:   false,
		Verbose:    false,
		Stdout:     io.Discard,
		Stderr:     io.Discard,
	}

	if err := execute(cfg); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	// Verify TypeScript file was created with custom class name
	tsFile := filepath.Join(tmpDir, "custom-class.ts")
	if _, err := os.Stat(tsFile); os.IsNotExist(err) {
		t.Errorf("TypeScript client not generated at %s", tsFile)
	}

	// Verify no worker file in sync mode
	workerFile := filepath.Join(tmpDir, "worker.js")
	if _, err := os.Stat(workerFile); !os.IsNotExist(err) {
		t.Errorf("Worker should not be generated in sync mode")
	}
}

func TestExecute_VerboseMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "execute-verbose-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	var stderr strings.Builder
	cfg := Config{
		SourceFile: "test/e2e/wasm/main.go",
		OutputDir:  tmpDir,
		NoBuild:    true,
		Compiler:   "go",
		Mode:       "worker",
		ClassName:  "",
		Optimize:   false,
		Verbose:    true,
		Stdout:     io.Discard,
		Stderr:     &stderr,
	}

	if err := execute(cfg); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	// Verify debug output was written
	debugOutput := stderr.String()
	if !strings.Contains(debugOutput, "[DEBUG]") {
		t.Errorf("verbose mode should produce debug output, got: %s", debugOutput)
	}
}

func TestExecute_SourceNotFound(t *testing.T) {
	cfg := Config{
		SourceFile: "nonexistent/main.go",
		OutputDir:  "out",
		NoBuild:    true,
		Compiler:   "go",
		Mode:       "worker",
		Stdout:     io.Discard,
		Stderr:     io.Discard,
	}

	err := execute(cfg)
	if err == nil {
		t.Fatal("expected error for nonexistent source file")
	}
	if !strings.Contains(err.Error(), "source file not found") {
		t.Errorf("expected 'source file not found' error, got: %v", err)
	}
}

func TestExecute_NoExportedFunctions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "execute-noexport-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a Go file with no exported functions
	goFile := filepath.Join(tmpDir, "main.go")
	content := `package main

func main() { select {} }

func privateFunc() string { return "private" }
`
	if err := os.WriteFile(goFile, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg := Config{
		SourceFile: goFile,
		OutputDir:  tmpDir,
		NoBuild:    true,
		Compiler:   "go",
		Mode:       "worker",
		Stdout:     io.Discard,
		Stderr:     io.Discard,
	}

	err = execute(cfg)
	if err == nil {
		t.Fatal("expected error for no exported functions")
	}
	if !strings.Contains(err.Error(), "no exported functions") {
		t.Errorf("expected 'no exported functions' error, got: %v", err)
	}
}

func TestExecute_MissingSelect(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "execute-noselect-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a Go file without select {}
	goFile := filepath.Join(tmpDir, "main.go")
	content := `package main

func Greet(name string) string { return "Hello, " + name }

func main() { /* missing select {} */ }
`
	if err := os.WriteFile(goFile, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg := Config{
		SourceFile: goFile,
		OutputDir:  tmpDir,
		NoBuild:    true,
		Compiler:   "go",
		Mode:       "worker",
		Stdout:     io.Discard,
		Stderr:     io.Discard,
	}

	err = execute(cfg)
	if err == nil {
		t.Fatal("expected error for missing select {}")
	}
	if !strings.Contains(err.Error(), "select {}") {
		t.Errorf("expected 'select {}' error, got: %v", err)
	}
}

func TestExecute_DirNameMain(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "execute-dirname-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create Go file in temp dir
	goFile := filepath.Join(tmpDir, "main.go")
	content := `package main

func Greet(name string) string { return "Hello, " + name }

func main() { select {} }
`
	if err := os.WriteFile(goFile, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	outDir := filepath.Join(tmpDir, "out")

	// Change to tmpDir and use relative path to trigger dirName == "." case
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	cfg := Config{
		SourceFile: "main.go", // Relative path -> dir = "." -> dirName = "main"
		OutputDir:  outDir,
		NoBuild:    true,
		Compiler:   "go",
		Mode:       "worker",
		ClassName:  "", // Auto-derive from "main"
		Stdout:     io.Discard,
		Stderr:     io.Discard,
	}

	if err := execute(cfg); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	// Should derive class name "GoMain" -> "go-main.ts"
	tsFile := filepath.Join(outDir, "go-main.ts")
	if _, err := os.Stat(tsFile); os.IsNotExist(err) {
		t.Errorf("TypeScript client not generated at %s", tsFile)
	}
}
