package generator

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	goparser "github.com/13rac1/gowasm-bindgen/internal/parser"
)

func TestGenerateGoBindings(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		workerMode bool
		checks     []func(*testing.T, string)
	}{
		{
			name:   "empty package",
			source: `package main`,
			checks: []func(*testing.T, string){
				checkBuildConstraint,
				checkGeneratedComment,
				checkPackage("main"),
				checkImportSyscallJS,
				checkInitFunction,
			},
		},
		{
			name: "simple string function",
			source: `package main
func Greet(name string) string { return "Hello, " + name }`,
			checks: []func(*testing.T, string){
				checkBuildConstraint,
				checkPackage("main"),
				checkFunctionRegistered("greet", "wasmGreet"),
				checkWrapperSignature("wasmGreet"),
				checkContains(`name := args[0].String()`),
				checkContains(`result := Greet(name)`),
				checkContains(`return result`),
			},
		},
		{
			name: "multiple functions",
			source: `package main
func Add(a, b int) int { return a + b }
func Sub(a, b int) int { return a - b }`,
			checks: []func(*testing.T, string){
				checkFunctionRegistered("add", "wasmAdd"),
				checkFunctionRegistered("sub", "wasmSub"),
				checkWrapperSignature("wasmAdd"),
				checkWrapperSignature("wasmSub"),
			},
		},
		{
			name: "int parameters",
			source: `package main
func Add(a int, b int) int { return a + b }`,
			checks: []func(*testing.T, string){
				checkContains(`a := args[0].Int()`),
				checkContains(`b := args[1].Int()`),
			},
		},
		{
			name: "float64 parameter",
			source: `package main
func Scale(x float64) float64 { return x * 2 }`,
			checks: []func(*testing.T, string){
				checkContains(`x := args[0].Float()`),
			},
		},
		{
			name: "bool parameter",
			source: `package main
func Toggle(flag bool) bool { return !flag }`,
			checks: []func(*testing.T, string){
				checkContains(`flag := args[0].Bool()`),
			},
		},
		{
			name: "error return",
			source: `package main
import "errors"
func MightFail(x int) (string, error) {
	if x < 0 { return "", errors.New("negative") }
	return "ok", nil
}`,
			checks: []func(*testing.T, string){
				checkContains(`result, err := MightFail(x)`),
				checkContains(`if err != nil {`),
				checkContains(`"__error": err.Error()`),
				checkContains(`return result`),
			},
		},
		{
			name: "error only return",
			source: `package main
func Validate(x int) error { return nil }`,
			checks: []func(*testing.T, string){
				checkContains(`err := Validate(x)`),
				checkContains(`if err != nil {`),
				checkContains(`return nil`),
			},
		},
		{
			name: "void function",
			source: `package main
func DoSomething() {}`,
			checks: []func(*testing.T, string){
				checkContains(`DoSomething()`),
				checkContains(`return nil`),
			},
		},
		{
			name: "struct return",
			source: `package main
type User struct {
	Name string ` + "`json:\"name\"`" + `
	Age  int    ` + "`json:\"age\"`" + `
}
func GetUser() User { return User{} }`,
			checks: []func(*testing.T, string){
				checkContains(`return map[string]interface{}{`),
				checkContains(`"name": result.Name`),
				checkContains(`"age": result.Age`),
			},
		},
		{
			name: "byte slice parameter",
			source: `package main
func ProcessBytes(data []byte) int { return len(data) }`,
			checks: []func(*testing.T, string){
				checkContains(`js.CopyBytesToGo`),
				checkContains(`make([]byte, length)`),
			},
		},
		{
			name: "byte slice return",
			source: `package main
func GetBytes() []byte { return []byte{1, 2, 3} }`,
			checks: []func(*testing.T, string){
				checkContains(`js.CopyBytesToJS`),
				checkContains(`Uint8Array`),
			},
		},
		{
			name: "int32 slice return (typed array)",
			source: `package main
func GetNumbers() []int32 { return []int32{1, 2, 3} }`,
			checks: []func(*testing.T, string){
				checkContains(`Int32Array`),
				checkContains(`SetIndex`),
			},
		},
		{
			name: "string slice parameter",
			source: `package main
func JoinStrings(items []string) string { return "" }`,
			checks: []func(*testing.T, string){
				checkContains(`make([]string, length)`),
				checkContains(`args[0].Index(i)`),
				checkContains(`.String()`),
			},
		},
		{
			name: "map parameter",
			source: `package main
func ProcessMap(data map[string]int) int { return 0 }`,
			checks: []func(*testing.T, string){
				checkContains(`make(map[string]int)`),
				checkContains(`Object`),
				checkContains(`keys`),
			},
		},
		{
			name:       "callback sync mode",
			workerMode: false,
			source: `package main
func ForEach(items []string, cb func(string, int)) {}`,
			checks: []func(*testing.T, string){
				checkContains(`.Invoke(`),
			},
		},
		{
			name:       "callback worker mode",
			workerMode: true,
			source: `package main
func ForEach(items []string, cb func(string, int)) {}`,
			checks: []func(*testing.T, string){
				checkContains(`invokeCallback`),
				checkContains(`cbArgs.Call("push"`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := mustParse(t, tt.source)
			output := GenerateGoBindings(parsed, tt.workerMode)

			for _, check := range tt.checks {
				check(t, output)
			}

			assertValidGoSyntax(t, output)
		})
	}
}

// Helper functions

func checkBuildConstraint(t *testing.T, output string) {
	t.Helper()
	if !strings.HasPrefix(output, "//go:build js && wasm\n") {
		t.Error("missing or incorrect build constraint")
	}
}

func checkGeneratedComment(t *testing.T, output string) {
	t.Helper()
	if !strings.Contains(output, "Code generated by gowasm-bindgen. DO NOT EDIT.") {
		t.Error("missing generated comment")
	}
}

func checkPackage(name string) func(*testing.T, string) {
	return func(t *testing.T, output string) {
		t.Helper()
		if !strings.Contains(output, "package "+name+"\n") {
			t.Errorf("missing or incorrect package declaration, want %q", name)
		}
	}
}

func checkImportSyscallJS(t *testing.T, output string) {
	t.Helper()
	if !strings.Contains(output, `import "syscall/js"`) {
		t.Error("missing syscall/js import")
	}
}

func checkInitFunction(t *testing.T, output string) {
	t.Helper()
	if !strings.Contains(output, "func init() {") {
		t.Error("missing init function")
	}
}

func checkFunctionRegistered(jsName, wrapperName string) func(*testing.T, string) {
	return func(t *testing.T, output string) {
		t.Helper()
		want := `js.Global().Set("` + jsName + `", js.FuncOf(` + wrapperName + `))`
		if !strings.Contains(output, want) {
			t.Errorf("function %q not registered correctly, want %q", jsName, want)
		}
	}
}

func checkWrapperSignature(name string) func(*testing.T, string) {
	return func(t *testing.T, output string) {
		t.Helper()
		want := "func " + name + "(_ js.Value, args []js.Value) interface{}"
		if !strings.Contains(output, want) {
			t.Errorf("wrapper %q has incorrect signature", name)
		}
	}
}

func checkContains(substr string) func(*testing.T, string) {
	return func(t *testing.T, output string) {
		t.Helper()
		if !strings.Contains(output, substr) {
			t.Errorf("output missing expected content: %q", substr)
		}
	}
}

func assertValidGoSyntax(t *testing.T, code string) {
	t.Helper()
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, "bindings_gen.go", code, parser.AllErrors)
	if err != nil {
		t.Errorf("generated code has syntax errors: %v\n\nCode:\n%s", err, code)
	}
}

func mustParse(t *testing.T, source string) *goparser.ParsedFile {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(tmpFile, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	parsed, err := goparser.ParseSourceFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
