package parser

import (
	"os"
	"path/filepath"
	"strings"
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
	if err := os.WriteFile(tmpFile, []byte(src), 0600); err != nil {
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
	if err := os.WriteFile(tmpFile, []byte(src), 0600); err != nil {
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
	if err := os.WriteFile(tmpFile, []byte(src), 0600); err != nil {
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
	if err := os.WriteFile(tmpFile, []byte(src), 0600); err != nil {
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
	if err := os.WriteFile(tmpFile, []byte(src), 0600); err != nil {
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
	if err := os.WriteFile(tmpFile, []byte(src), 0600); err != nil {
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
	if err := os.WriteFile(tmpFile, []byte(src), 0600); err != nil {
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

func TestIsByteSlice(t *testing.T) {
	tests := []struct {
		name     string
		goType   GoType
		expected bool
	}{
		{"byte slice", GoType{Kind: KindSlice, Elem: &GoType{Name: "byte", Kind: KindPrimitive}}, true},
		{"uint8 slice", GoType{Kind: KindSlice, Elem: &GoType{Name: "uint8", Kind: KindPrimitive}}, true},
		{"int slice", GoType{Kind: KindSlice, Elem: &GoType{Name: "int", Kind: KindPrimitive}}, false},
		{"string slice", GoType{Kind: KindSlice, Elem: &GoType{Name: "string", Kind: KindPrimitive}}, false},
		{"not a slice", GoType{Kind: KindPrimitive, Name: "int"}, false},
		{"nil elem", GoType{Kind: KindSlice, Elem: nil}, false},
		{"struct elem", GoType{Kind: KindSlice, Elem: &GoType{Name: "User", Kind: KindStruct}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isByteSlice(tt.goType)
			if result != tt.expected {
				t.Errorf("isByteSlice(%+v) = %v, want %v", tt.goType, result, tt.expected)
			}
		})
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
		{"string map", GoType{Name: "map[string]int", Kind: KindMap, Key: &GoType{Name: "string", Kind: KindPrimitive}, Value: &GoType{Name: "int", Kind: KindPrimitive}}, "{[key: string]: number}"},
		{"error", GoType{Name: "error", Kind: KindError, IsError: true}, "string"},
		// Callbacks
		{"void callback no params", GoType{Kind: KindFunction, IsVoid: true, CallbackParams: []GoType{}}, "() => void"},
		{"void callback one param", GoType{Kind: KindFunction, IsVoid: true, CallbackParams: []GoType{{Name: "string", Kind: KindPrimitive}}}, "(arg0: string) => void"},
		{"void callback two params", GoType{Kind: KindFunction, IsVoid: true, CallbackParams: []GoType{
			{Name: "string", Kind: KindPrimitive},
			{Name: "int", Kind: KindPrimitive},
		}}, "(arg0: string, arg1: number) => void"},
		// Struct inline interface
		{"struct with fields", GoType{
			Kind: KindStruct,
			Name: "User",
			Fields: []GoField{
				{Name: "Name", JSONTag: "name", Type: GoType{Name: "string", Kind: KindPrimitive}},
				{Name: "Age", JSONTag: "", Type: GoType{Name: "int", Kind: KindPrimitive}},
			},
		}, "{name: string, Age: number}"},
		{"empty struct", GoType{Kind: KindStruct, Fields: []GoField{}}, "any"},
		// Pointer
		{"pointer to string", GoType{Kind: KindPointer, Elem: &GoType{Name: "string", Kind: KindPrimitive}}, "string"},
		{"pointer nil elem", GoType{Kind: KindPointer, Elem: nil}, "any"},
		// Map with non-string key
		{"map[int]string", GoType{Kind: KindMap, Key: &GoType{Name: "int", Kind: KindPrimitive}, Value: &GoType{Name: "string", Kind: KindPrimitive}}, "Record<number, string>"},
		// Unknown kind
		{"unknown kind", GoType{Kind: 999}, "any"},
		// Slice with nil elem
		{"slice nil elem", GoType{Kind: KindSlice, Elem: nil}, "any[]"},
		// Map with nil key/value
		{"map nil parts", GoType{Kind: KindMap, Key: nil, Value: nil}, "any"},
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

func TestGoTypeToJSExtraction(t *testing.T) {
	tests := []struct {
		name       string
		goType     GoType
		argExpr    string
		workerMode bool
		contains   []string // Strings that should appear in output
	}{
		// Primitive types
		{"string", GoType{Name: "string", Kind: KindPrimitive}, "args[0]", false, []string{"args[0].String()"}},
		{"int", GoType{Name: "int", Kind: KindPrimitive}, "args[0]", false, []string{"args[0].Int()"}},
		{"int64", GoType{Name: "int64", Kind: KindPrimitive}, "args[0]", false, []string{"int64(args[0].Float())"}},
		{"int32", GoType{Name: "int32", Kind: KindPrimitive}, "args[0]", false, []string{"int32(args[0].Int())"}},
		{"int16", GoType{Name: "int16", Kind: KindPrimitive}, "args[0]", false, []string{"int16(args[0].Int())"}},
		{"int8", GoType{Name: "int8", Kind: KindPrimitive}, "args[0]", false, []string{"int8(args[0].Int())"}},
		{"uint", GoType{Name: "uint", Kind: KindPrimitive}, "args[0]", false, []string{"uint(args[0].Int())"}},
		{"uint64", GoType{Name: "uint64", Kind: KindPrimitive}, "args[0]", false, []string{"uint64(args[0].Float())"}},
		{"uint32", GoType{Name: "uint32", Kind: KindPrimitive}, "args[0]", false, []string{"uint32(args[0].Int())"}},
		{"uint16", GoType{Name: "uint16", Kind: KindPrimitive}, "args[0]", false, []string{"uint16(args[0].Int())"}},
		{"uint8", GoType{Name: "uint8", Kind: KindPrimitive}, "args[0]", false, []string{"uint8(args[0].Int())"}},
		{"float64", GoType{Name: "float64", Kind: KindPrimitive}, "args[0]", false, []string{"args[0].Float()"}},
		{"float32", GoType{Name: "float32", Kind: KindPrimitive}, "args[0]", false, []string{"float32(args[0].Float())"}},
		{"bool", GoType{Name: "bool", Kind: KindPrimitive}, "args[0]", false, []string{"args[0].Bool()"}},
		{"unknown primitive", GoType{Name: "unknown", Kind: KindPrimitive}, "args[0]", false, []string{"args[0]"}},

		// Byte slice (bulk copy)
		{"byte slice", GoType{Kind: KindSlice, Elem: &GoType{Name: "byte", Kind: KindPrimitive}}, "args[0]", false,
			[]string{"js.CopyBytesToGo", "args[0].Length()", "make([]byte, length)"}},

		// Non-byte slice (element by element)
		{"int slice", GoType{Kind: KindSlice, Elem: &GoType{Name: "int", Kind: KindPrimitive}}, "args[0]", false,
			[]string{"make([]int, length)", "args[0].Index(i)", ".Int()"}},
		{"string slice", GoType{Kind: KindSlice, Elem: &GoType{Name: "string", Kind: KindPrimitive}}, "args[0]", false,
			[]string{"make([]string, length)", "args[0].Index(i)", ".String()"}},
		{"nil elem slice", GoType{Kind: KindSlice, Elem: nil}, "args[0]", false, []string{"nil"}},

		// Map extraction
		{"map[string]int", GoType{
			Kind:  KindMap,
			Key:   &GoType{Name: "string", Kind: KindPrimitive},
			Value: &GoType{Name: "int", Kind: KindPrimitive},
		}, "args[0]", false,
			[]string{"make(map[string]int)", "Object", "keys", ".Get(key)", ".Int()"}},
		{"map nil parts", GoType{Kind: KindMap, Key: nil, Value: nil}, "args[0]", false, []string{"nil"}},
		{"map[int]string unsupported", GoType{
			Kind:  KindMap,
			Key:   &GoType{Name: "int", Kind: KindPrimitive},
			Value: &GoType{Name: "string", Kind: KindPrimitive},
		}, "args[0]", false, []string{"nil"}},

		// Struct extraction
		{"struct", GoType{
			Kind: KindStruct,
			Name: "User",
			Fields: []GoField{
				{Name: "Name", JSONTag: "name", Type: GoType{Name: "string", Kind: KindPrimitive}},
				{Name: "Age", JSONTag: "", Type: GoType{Name: "int", Kind: KindPrimitive}},
			},
		}, "args[0]", false,
			[]string{"User{", "Name: ", ".Get(\"name\")", ".String()", "Age: ", ".Get(\"Age\")", ".Int()"}},

		// Pointer extraction
		{"pointer to int", GoType{Kind: KindPointer, Elem: &GoType{Name: "int", Kind: KindPrimitive}}, "args[0]", false,
			[]string{"args[0].Int()"}},
		{"pointer nil elem", GoType{Kind: KindPointer, Elem: nil}, "args[0]", false, []string{"args[0]"}},

		// Callback (sync mode)
		{"callback sync mode", GoType{
			Kind:   KindFunction,
			IsVoid: true,
			CallbackParams: []GoType{
				{Name: "string", Kind: KindPrimitive},
				{Name: "int", Kind: KindPrimitive},
			},
		}, "args[0]", false,
			[]string{"func(arg0 string, arg1 int)", ".Invoke(arg0, arg1)"}},

		// Callback (worker mode)
		{"callback worker mode", GoType{
			Kind:   KindFunction,
			IsVoid: true,
			CallbackParams: []GoType{
				{Name: "string", Kind: KindPrimitive},
			},
		}, "args[0]", true,
			[]string{"func(arg0 string)", "invokeCallback", "cbArgs.Call(\"push\"", ".Int()"}},

		// Unknown kind
		{"unknown kind", GoType{Kind: 999}, "args[0]", false, []string{"args[0]"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GoTypeToJSExtraction(tt.goType, tt.argExpr, tt.workerMode)
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("GoTypeToJSExtraction() = %q, should contain %q", result, substr)
				}
			}
		})
	}
}

func TestGoTypeToJSReturn(t *testing.T) {
	tests := []struct {
		name      string
		goType    GoType
		valueExpr string
		contains  []string // Strings that should appear in output
	}{
		// Primitives - returned directly
		{"string", GoType{Name: "string", Kind: KindPrimitive}, "result", []string{"result"}},
		{"int", GoType{Name: "int", Kind: KindPrimitive}, "result", []string{"result"}},
		{"bool", GoType{Name: "bool", Kind: KindPrimitive}, "result", []string{"result"}},

		// Byte slice (bulk copy to Uint8Array)
		{"byte slice", GoType{Kind: KindSlice, Elem: &GoType{Name: "byte", Kind: KindPrimitive}}, "result",
			[]string{"Uint8Array", "js.CopyBytesToJS", "result"}},
		{"uint8 slice", GoType{Kind: KindSlice, Elem: &GoType{Name: "uint8", Kind: KindPrimitive}}, "result",
			[]string{"Uint8Array", "js.CopyBytesToJS"}},

		// Typed arrays
		{"int8 slice", GoType{Kind: KindSlice, Elem: &GoType{Name: "int8", Kind: KindPrimitive}}, "result",
			[]string{"Int8Array", "SetIndex"}},
		{"int16 slice", GoType{Kind: KindSlice, Elem: &GoType{Name: "int16", Kind: KindPrimitive}}, "result",
			[]string{"Int16Array", "SetIndex"}},
		{"int32 slice", GoType{Kind: KindSlice, Elem: &GoType{Name: "int32", Kind: KindPrimitive}}, "result",
			[]string{"Int32Array", "SetIndex"}},
		{"uint16 slice", GoType{Kind: KindSlice, Elem: &GoType{Name: "uint16", Kind: KindPrimitive}}, "result",
			[]string{"Uint16Array", "SetIndex"}},
		{"uint32 slice", GoType{Kind: KindSlice, Elem: &GoType{Name: "uint32", Kind: KindPrimitive}}, "result",
			[]string{"Uint32Array", "SetIndex"}},
		{"float32 slice", GoType{Kind: KindSlice, Elem: &GoType{Name: "float32", Kind: KindPrimitive}}, "result",
			[]string{"Float32Array", "SetIndex"}},
		{"float64 slice", GoType{Kind: KindSlice, Elem: &GoType{Name: "float64", Kind: KindPrimitive}}, "result",
			[]string{"Float64Array", "SetIndex"}},

		// Non-typed array slices (returned directly)
		{"int slice", GoType{Kind: KindSlice, Elem: &GoType{Name: "int", Kind: KindPrimitive}}, "result",
			[]string{"result"}},
		{"string slice", GoType{Kind: KindSlice, Elem: &GoType{Name: "string", Kind: KindPrimitive}}, "result",
			[]string{"result"}},
		{"nil elem slice", GoType{Kind: KindSlice, Elem: nil}, "result", []string{"nil"}},

		// Struct slices (element conversion)
		{"struct slice", GoType{Kind: KindSlice, Elem: &GoType{
			Kind: KindStruct,
			Name: "User",
			Fields: []GoField{
				{Name: "Name", JSONTag: "name", Type: GoType{Name: "string", Kind: KindPrimitive}},
			},
		}}, "result",
			[]string{"[]interface{}", "for i, v := range result", "map[string]interface{}"}},

		// Map return
		{"map", GoType{Kind: KindMap, Key: &GoType{Name: "string"}, Value: &GoType{Name: "int"}}, "result",
			[]string{"map[string]interface{}(result)"}},

		// Struct return
		{"struct", GoType{
			Kind: KindStruct,
			Name: "User",
			Fields: []GoField{
				{Name: "Name", JSONTag: "name", Type: GoType{Name: "string", Kind: KindPrimitive}},
				{Name: "Age", JSONTag: "", Type: GoType{Name: "int", Kind: KindPrimitive}},
			},
		}, "result",
			[]string{"map[string]interface{}{", "\"name\": result.Name", "\"age\": result.Age"}},

		// Pointer return
		{"pointer to int", GoType{Kind: KindPointer, Elem: &GoType{Name: "int", Kind: KindPrimitive}}, "result",
			[]string{"result"}},
		{"pointer nil elem", GoType{Kind: KindPointer, Elem: nil}, "result", []string{"result"}},

		// Error return
		{"error", GoType{Kind: KindError}, "err", []string{"err.Error()"}},

		// Unknown kind
		{"unknown kind", GoType{Kind: 999}, "result", []string{"result"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GoTypeToJSReturn(tt.goType, tt.valueExpr)
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("GoTypeToJSReturn() = %q, should contain %q", result, substr)
				}
			}
		})
	}
}

func TestCallbackWrapperCode(t *testing.T) {
	tests := []struct {
		name     string
		goType   GoType
		argExpr  string
		contains []string
	}{
		{"no params", GoType{
			Kind:           KindFunction,
			CallbackParams: []GoType{},
		}, "cb", []string{"func()", "cb.Invoke()"}},

		{"one param", GoType{
			Kind: KindFunction,
			CallbackParams: []GoType{
				{Name: "string", Kind: KindPrimitive},
			},
		}, "cb", []string{"func(arg0 string)", "cb.Invoke(arg0)"}},

		{"two params", GoType{
			Kind: KindFunction,
			CallbackParams: []GoType{
				{Name: "string", Kind: KindPrimitive},
				{Name: "int", Kind: KindPrimitive},
			},
		}, "cb", []string{"func(arg0 string, arg1 int)", "cb.Invoke(arg0, arg1)"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := callbackWrapperCode(tt.goType, tt.argExpr)
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("callbackWrapperCode() = %q, should contain %q", result, substr)
				}
			}
		})
	}
}

func TestWorkerCallbackCode(t *testing.T) {
	tests := []struct {
		name     string
		goType   GoType
		argExpr  string
		contains []string
	}{
		{"no params", GoType{
			Kind:           KindFunction,
			CallbackParams: []GoType{},
		}, "cb", []string{"func()", "invokeCallback", "cb.Int()"}},

		{"one param", GoType{
			Kind: KindFunction,
			CallbackParams: []GoType{
				{Name: "string", Kind: KindPrimitive},
			},
		}, "cb", []string{"func(arg0 string)", "cbArgs.Call(\"push\", arg0)", "invokeCallback"}},

		{"two params", GoType{
			Kind: KindFunction,
			CallbackParams: []GoType{
				{Name: "int", Kind: KindPrimitive},
				{Name: "bool", Kind: KindPrimitive},
			},
		}, "cb", []string{"func(arg0 int, arg1 bool)", "cbArgs.Call(\"push\", arg0)", "cbArgs.Call(\"push\", arg1)"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := workerCallbackCode(tt.goType, tt.argExpr)
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("workerCallbackCode() = %q, should contain %q", result, substr)
				}
			}
		})
	}
}

func TestHasSelectInMain(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected bool
	}{
		{
			name: "main with select {}",
			src: `package main

func main() {
	// setup code
	select {}
}
`,
			expected: true,
		},
		{
			name: "main without select",
			src: `package main

func main() {
	println("hello")
}
`,
			expected: false,
		},
		{
			name: "no main function",
			src: `package main

func other() {
	select {}
}
`,
			expected: false,
		},
		{
			name: "select in nested if",
			src: `package main

func main() {
	if true {
		select {}
	}
}
`,
			expected: true,
		},
		{
			name: "select in for loop",
			src: `package main

func main() {
	for {
		select {}
	}
}
`,
			expected: true,
		},
		{
			name: "select with cases (not empty)",
			src: `package main

func main() {
	ch := make(chan int)
	select {
	case <-ch:
	}
}
`,
			expected: false,
		},
		{
			name: "method named main ignored",
			src: `package main

type S struct{}

func (s S) main() {
	select {}
}

func main() {
	println("no select")
}
`,
			expected: false,
		},
		{
			name: "select in switch case",
			src: `package main

func main() {
	switch x := 1; x {
	case 1:
		select {}
	}
}
`,
			expected: true,
		},
		{
			name: "select in if-else branch",
			src: `package main

func main() {
	if false {
		println("nope")
	} else {
		select {}
	}
}
`,
			expected: true,
		},
		{
			name: "select in range loop",
			src: `package main

func main() {
	for range []int{1} {
		select {}
	}
}
`,
			expected: true,
		},
		{
			name: "select in type switch",
			src: `package main

func main() {
	var x interface{} = 1
	switch x.(type) {
	case int:
		select {}
	}
}
`,
			expected: true,
		},
		{
			name: "nested block with select",
			src: `package main

func main() {
	{
		{
			select {}
		}
	}
}
`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "main.go")
			if err := os.WriteFile(tmpFile, []byte(tt.src), 0600); err != nil {
				t.Fatalf("failed to write temp file: %v", err)
			}

			result, err := HasSelectInMain(tmpFile)
			if err != nil {
				t.Fatalf("HasSelectInMain() error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("HasSelectInMain() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseSourceFile_RecursiveStruct(t *testing.T) {
	// Self-referential struct should be parsed without infinite loop
	src := `package main

type Node struct {
	Value int
	Next  *Node
}

type Tree struct {
	Left  *Tree
	Right *Tree
	Data  string
}

func GetNode() *Node {
	return nil
}

func GetTree() *Tree {
	return nil
}
`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "recursive.go")
	if err := os.WriteFile(tmpFile, []byte(src), 0600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	parsed, err := ParseSourceFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseSourceFile() error: %v", err)
	}

	// Should have 2 types
	if len(parsed.Types) != 2 {
		t.Errorf("got %d types, want 2", len(parsed.Types))
	}

	// Verify Node struct
	nodeType, ok := parsed.Types["Node"]
	if !ok {
		t.Fatal("expected Node type")
	}
	if nodeType.Kind != KindStruct {
		t.Errorf("Node.Kind = %v, want KindStruct", nodeType.Kind)
	}
	if len(nodeType.Fields) != 2 {
		t.Errorf("Node has %d fields, want 2", len(nodeType.Fields))
	}

	// Verify Tree struct
	treeType, ok := parsed.Types["Tree"]
	if !ok {
		t.Fatal("expected Tree type")
	}
	if treeType.Kind != KindStruct {
		t.Errorf("Tree.Kind = %v, want KindStruct", treeType.Kind)
	}
	if len(treeType.Fields) != 3 {
		t.Errorf("Tree has %d fields, want 3", len(treeType.Fields))
	}

	// Should have 2 functions
	if len(parsed.Functions) != 2 {
		t.Errorf("got %d functions, want 2", len(parsed.Functions))
	}
}

func TestParseSourceFile_AnonymousField(t *testing.T) {
	// Anonymous/embedded fields should be tracked for validator to reject
	src := `package main

type Meta struct {
	CreatedAt string
}

type User struct {
	Name string
	Meta
}

func GetUser() User {
	return User{}
}
`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "anon.go")
	if err := os.WriteFile(tmpFile, []byte(src), 0600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	parsed, err := ParseSourceFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseSourceFile() error: %v", err)
	}

	userType, ok := parsed.Types["User"]
	if !ok {
		t.Fatal("expected User type")
	}

	// User should have 2 fields: Name and the anonymous Meta
	if len(userType.Fields) != 2 {
		t.Fatalf("User has %d fields, want 2", len(userType.Fields))
	}

	// Find the anonymous field (empty name)
	var hasAnonymous bool
	for _, field := range userType.Fields {
		if field.Name == "" {
			hasAnonymous = true
			break
		}
	}

	if !hasAnonymous {
		t.Error("expected anonymous field with empty name for validator to detect")
	}
}
