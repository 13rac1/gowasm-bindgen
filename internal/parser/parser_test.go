package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSourceFile(t *testing.T) {
	// Create a temp file with Go source
	src := `package wasm

import "fmt"

// User represents a user in the system
type User struct {
	Name   string ` + "`json:\"name\"`" + `
	Age    int    ` + "`json:\"age\"`" + `
	Active bool
}

// Greet returns a greeting message
func Greet(name string) string {
	return "Hello, " + name + "!"
}

// Calculate performs arithmetic
func Calculate(a int, b int, op string) int {
	return a + b
}

// FormatUser formats a user
func FormatUser(name string, age int, active bool) User {
	return User{Name: name, Age: age, Active: active}
}

// Divide divides two numbers and returns error if division by zero
func Divide(a float64, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return a / b, nil
}

// unexported function should be ignored
func helper() {}
`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(tmpFile, []byte(src), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	parsed, err := ParseSourceFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseSourceFile() error: %v", err)
	}

	if parsed.Package != "wasm" {
		t.Errorf("Package = %q, want %q", parsed.Package, "wasm")
	}

	// Should have 4 exported functions
	if len(parsed.Functions) != 4 {
		t.Errorf("got %d functions, want 4", len(parsed.Functions))
	}

	// Should have 1 exported type (User)
	if len(parsed.Types) != 1 {
		t.Errorf("got %d types, want 1", len(parsed.Types))
	}

	userType, ok := parsed.Types["User"]
	if !ok {
		t.Fatal("expected User type")
	}

	if userType.Kind != KindStruct {
		t.Errorf("User.Kind = %v, want KindStruct", userType.Kind)
	}

	if len(userType.Fields) != 3 {
		t.Errorf("User has %d fields, want 3", len(userType.Fields))
	}
}

func TestParseSourceFile_FunctionSignatures(t *testing.T) {
	src := `package main

// Add adds two numbers
func Add(a int, b int) int {
	return a + b
}

func NoParams() string {
	return "hello"
}

func MultiReturn(x int) (string, error) {
	return "", nil
}
`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "funcs.go")
	if err := os.WriteFile(tmpFile, []byte(src), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	parsed, err := ParseSourceFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseSourceFile() error: %v", err)
	}

	tests := []struct {
		name       string
		wantParams int
		wantReturn int
		wantDoc    string
	}{
		{"Add", 2, 1, "Add adds two numbers"},
		{"NoParams", 0, 1, ""},
		{"MultiReturn", 1, 2, ""},
	}

	funcMap := make(map[string]GoFunction)
	for _, fn := range parsed.Functions {
		funcMap[fn.Name] = fn
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := funcMap[tt.name]
			if !ok {
				t.Fatalf("function %s not found", tt.name)
			}

			if len(fn.Params) != tt.wantParams {
				t.Errorf("%s: got %d params, want %d", tt.name, len(fn.Params), tt.wantParams)
			}

			if len(fn.Returns) != tt.wantReturn {
				t.Errorf("%s: got %d returns, want %d", tt.name, len(fn.Returns), tt.wantReturn)
			}

			if fn.Doc != tt.wantDoc {
				t.Errorf("%s: got doc %q, want %q", tt.name, fn.Doc, tt.wantDoc)
			}
		})
	}
}

func TestParseSourceFile_ErrorReturn(t *testing.T) {
	src := `package main

func MightFail(x int) (string, error) {
	return "", nil
}
`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "err.go")
	if err := os.WriteFile(tmpFile, []byte(src), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	parsed, err := ParseSourceFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseSourceFile() error: %v", err)
	}

	if len(parsed.Functions) != 1 {
		t.Fatalf("got %d functions, want 1", len(parsed.Functions))
	}

	fn := parsed.Functions[0]
	if len(fn.Returns) != 2 {
		t.Fatalf("got %d returns, want 2", len(fn.Returns))
	}

	if !fn.Returns[1].IsError {
		t.Error("expected second return to be error")
	}

	if fn.Returns[1].Kind != KindError {
		t.Errorf("expected KindError, got %v", fn.Returns[1].Kind)
	}
}

func TestParseSourceFile_JSONTags(t *testing.T) {
	src := `package main

type Data struct {
	FirstName string ` + "`json:\"first_name\"`" + `
	LastName  string ` + "`json:\"last_name,omitempty\"`" + `
	NoTag     int
}

func GetData() Data {
	return Data{}
}
`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tags.go")
	if err := os.WriteFile(tmpFile, []byte(src), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	parsed, err := ParseSourceFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseSourceFile() error: %v", err)
	}

	dataType, ok := parsed.Types["Data"]
	if !ok {
		t.Fatal("expected Data type")
	}

	expectedTags := map[string]string{
		"FirstName": "first_name",
		"LastName":  "last_name",
		"NoTag":     "",
	}

	for _, field := range dataType.Fields {
		want := expectedTags[field.Name]
		if field.JSONTag != want {
			t.Errorf("field %s: JSONTag = %q, want %q", field.Name, field.JSONTag, want)
		}
	}
}

func TestParseSourceFile_SlicesAndMaps(t *testing.T) {
	src := `package main

func ProcessSlice(items []string) []int {
	return nil
}

func ProcessMap(data map[string]int) map[string]bool {
	return nil
}
`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "collections.go")
	if err := os.WriteFile(tmpFile, []byte(src), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	parsed, err := ParseSourceFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseSourceFile() error: %v", err)
	}

	funcMap := make(map[string]GoFunction)
	for _, fn := range parsed.Functions {
		funcMap[fn.Name] = fn
	}

	// Check ProcessSlice
	sliceFn := funcMap["ProcessSlice"]
	if sliceFn.Params[0].Type.Kind != KindSlice {
		t.Errorf("expected slice param, got %v", sliceFn.Params[0].Type.Kind)
	}
	if sliceFn.Returns[0].Kind != KindSlice {
		t.Errorf("expected slice return, got %v", sliceFn.Returns[0].Kind)
	}

	// Check ProcessMap
	mapFn := funcMap["ProcessMap"]
	if mapFn.Params[0].Type.Kind != KindMap {
		t.Errorf("expected map param, got %v", mapFn.Params[0].Type.Kind)
	}
	if mapFn.Returns[0].Kind != KindMap {
		t.Errorf("expected map return, got %v", mapFn.Returns[0].Kind)
	}
}

func TestParseSourceFile_MethodsIgnored(t *testing.T) {
	src := `package main

type Service struct{}

// Method should be ignored
func (s *Service) DoSomething() string {
	return ""
}

// Function should be included
func CreateService() *Service {
	return &Service{}
}
`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "methods.go")
	if err := os.WriteFile(tmpFile, []byte(src), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	parsed, err := ParseSourceFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseSourceFile() error: %v", err)
	}

	if len(parsed.Functions) != 1 {
		t.Errorf("got %d functions, want 1 (methods should be ignored)", len(parsed.Functions))
	}

	if parsed.Functions[0].Name != "CreateService" {
		t.Errorf("expected CreateService, got %s", parsed.Functions[0].Name)
	}
}

func TestParseSourceFile_CallbackParams(t *testing.T) {
	src := `package main

// ForEach iterates over items and calls callback for each
func ForEach(items []string, callback func(string, int)) {
}

// OnComplete calls callback when done
func OnComplete(callback func()) {
}

// WithNamedParams has named callback params
func WithNamedParams(callback func(item string, index int)) {
}
`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "callbacks.go")
	if err := os.WriteFile(tmpFile, []byte(src), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	parsed, err := ParseSourceFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseSourceFile() error: %v", err)
	}

	funcMap := make(map[string]GoFunction)
	for _, fn := range parsed.Functions {
		funcMap[fn.Name] = fn
	}

	// Check ForEach callback parameter
	forEachFn := funcMap["ForEach"]
	if len(forEachFn.Params) != 2 {
		t.Fatalf("ForEach: got %d params, want 2", len(forEachFn.Params))
	}
	cbParam := forEachFn.Params[1]
	if cbParam.Type.Kind != KindFunction {
		t.Errorf("ForEach callback: got kind %v, want KindFunction", cbParam.Type.Kind)
	}
	if !cbParam.Type.IsVoid {
		t.Error("ForEach callback: expected IsVoid=true")
	}
	if len(cbParam.Type.CallbackParams) != 2 {
		t.Errorf("ForEach callback: got %d params, want 2", len(cbParam.Type.CallbackParams))
	}
	if cbParam.Type.CallbackParams[0].Name != "string" {
		t.Errorf("ForEach callback param 0: got %q, want string", cbParam.Type.CallbackParams[0].Name)
	}
	if cbParam.Type.CallbackParams[1].Name != "int" {
		t.Errorf("ForEach callback param 1: got %q, want int", cbParam.Type.CallbackParams[1].Name)
	}

	// Check OnComplete callback (no params)
	onCompleteFn := funcMap["OnComplete"]
	cbParam = onCompleteFn.Params[0]
	if cbParam.Type.Kind != KindFunction {
		t.Errorf("OnComplete callback: got kind %v, want KindFunction", cbParam.Type.Kind)
	}
	if len(cbParam.Type.CallbackParams) != 0 {
		t.Errorf("OnComplete callback: got %d params, want 0", len(cbParam.Type.CallbackParams))
	}

	// Check WithNamedParams (named params should work same as unnamed)
	withNamedFn := funcMap["WithNamedParams"]
	cbParam = withNamedFn.Params[0]
	if len(cbParam.Type.CallbackParams) != 2 {
		t.Errorf("WithNamedParams callback: got %d params, want 2", len(cbParam.Type.CallbackParams))
	}
}

func TestGoTypeToTS(t *testing.T) {
	tests := []struct {
		name     string
		goType   GoType
		expected string
	}{
		{"string", GoType{Name: "string", Kind: KindPrimitive}, "string"},
		{"int", GoType{Name: "int", Kind: KindPrimitive}, "number"},
		{"int64", GoType{Name: "int64", Kind: KindPrimitive}, "number"},
		{"float64", GoType{Name: "float64", Kind: KindPrimitive}, "number"},
		{"bool", GoType{Name: "bool", Kind: KindPrimitive}, "boolean"},
		// Typed arrays
		{"byte slice", GoType{Name: "[]byte", Kind: KindSlice, Elem: &GoType{Name: "byte", Kind: KindPrimitive}}, "Uint8Array"},
		{"uint8 slice", GoType{Name: "[]uint8", Kind: KindSlice, Elem: &GoType{Name: "uint8", Kind: KindPrimitive}}, "Uint8Array"},
		{"int8 slice", GoType{Name: "[]int8", Kind: KindSlice, Elem: &GoType{Name: "int8", Kind: KindPrimitive}}, "Int8Array"},
		{"int16 slice", GoType{Name: "[]int16", Kind: KindSlice, Elem: &GoType{Name: "int16", Kind: KindPrimitive}}, "Int16Array"},
		{"int32 slice", GoType{Name: "[]int32", Kind: KindSlice, Elem: &GoType{Name: "int32", Kind: KindPrimitive}}, "Int32Array"},
		{"uint16 slice", GoType{Name: "[]uint16", Kind: KindSlice, Elem: &GoType{Name: "uint16", Kind: KindPrimitive}}, "Uint16Array"},
		{"uint32 slice", GoType{Name: "[]uint32", Kind: KindSlice, Elem: &GoType{Name: "uint32", Kind: KindPrimitive}}, "Uint32Array"},
		{"float32 slice", GoType{Name: "[]float32", Kind: KindSlice, Elem: &GoType{Name: "float32", Kind: KindPrimitive}}, "Float32Array"},
		{"float64 slice", GoType{Name: "[]float64", Kind: KindSlice, Elem: &GoType{Name: "float64", Kind: KindPrimitive}}, "Float64Array"},
		// Non-typed arrays (no bulk copy available)
		{"int slice", GoType{Name: "[]int", Kind: KindSlice, Elem: &GoType{Name: "int", Kind: KindPrimitive}}, "number[]"},
		{"string slice", GoType{Name: "[]string", Kind: KindSlice, Elem: &GoType{Name: "string", Kind: KindPrimitive}}, "string[]"},
		{"string map", GoType{Name: "map[string]int", Kind: KindMap, Key: &GoType{Name: "string"}, Value: &GoType{Name: "int", Kind: KindPrimitive}}, "{[key: string]: number}"},
		{"error", GoType{Name: "error", Kind: KindError, IsError: true}, "string"},
		// Callbacks
		{"void callback no params", GoType{Kind: KindFunction, IsVoid: true, CallbackParams: []GoType{}}, "() => void"},
		{"void callback one param", GoType{Kind: KindFunction, IsVoid: true, CallbackParams: []GoType{{Name: "string", Kind: KindPrimitive}}}, "(arg0: string) => void"},
		{"void callback two params", GoType{Kind: KindFunction, IsVoid: true, CallbackParams: []GoType{
			{Name: "string", Kind: KindPrimitive},
			{Name: "int", Kind: KindPrimitive},
		}}, "(arg0: string, arg1: number) => void"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GoTypeToTS(tt.goType)
			if result != tt.expected {
				t.Errorf("GoTypeToTS(%+v) = %q, want %q", tt.goType, result, tt.expected)
			}
		})
	}
}
