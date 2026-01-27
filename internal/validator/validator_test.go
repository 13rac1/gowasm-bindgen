package validator

import (
	"errors"
	"strings"
	"testing"

	"github.com/13rac1/gowasm-bindgen/internal/parser"
)

func TestValidateFunctions_Success(t *testing.T) {
	parsed := &parser.ParsedFile{
		Package: "wasm",
		Functions: []parser.GoFunction{
			{
				Name: "Greet",
				Params: []parser.GoParameter{
					{Name: "name", Type: parser.GoType{Name: "string", Kind: parser.KindPrimitive}},
				},
				Returns: []parser.GoType{
					{Name: "string", Kind: parser.KindPrimitive},
				},
			},
		},
		Types: map[string]*parser.GoType{},
	}

	err := ValidateFunctions(parsed)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateFunctions_AllPrimitives(t *testing.T) {
	parsed := &parser.ParsedFile{
		Package: "wasm",
		Functions: []parser.GoFunction{
			{
				Name: "ProcessAll",
				Params: []parser.GoParameter{
					{Name: "s", Type: parser.GoType{Name: "string", Kind: parser.KindPrimitive}},
					{Name: "i", Type: parser.GoType{Name: "int", Kind: parser.KindPrimitive}},
					{Name: "i64", Type: parser.GoType{Name: "int64", Kind: parser.KindPrimitive}},
					{Name: "f", Type: parser.GoType{Name: "float64", Kind: parser.KindPrimitive}},
					{Name: "b", Type: parser.GoType{Name: "bool", Kind: parser.KindPrimitive}},
				},
				Returns: []parser.GoType{
					{Name: "int", Kind: parser.KindPrimitive},
				},
			},
		},
		Types: map[string]*parser.GoType{},
	}

	err := ValidateFunctions(parsed)
	if err != nil {
		t.Errorf("expected no error for primitives, got: %v", err)
	}
}

func TestValidateFunctions_Slices(t *testing.T) {
	parsed := &parser.ParsedFile{
		Package: "wasm",
		Functions: []parser.GoFunction{
			{
				Name: "ProcessSlice",
				Params: []parser.GoParameter{
					{Name: "items", Type: parser.GoType{
						Name: "[]string",
						Kind: parser.KindSlice,
						Elem: &parser.GoType{Name: "string", Kind: parser.KindPrimitive},
					}},
				},
				Returns: []parser.GoType{
					{
						Name: "[]int",
						Kind: parser.KindSlice,
						Elem: &parser.GoType{Name: "int", Kind: parser.KindPrimitive},
					},
				},
			},
		},
		Types: map[string]*parser.GoType{},
	}

	err := ValidateFunctions(parsed)
	if err != nil {
		t.Errorf("expected no error for slices, got: %v", err)
	}
}

func TestValidateFunctions_StringMap(t *testing.T) {
	parsed := &parser.ParsedFile{
		Package: "wasm",
		Functions: []parser.GoFunction{
			{
				Name: "ProcessMap",
				Params: []parser.GoParameter{
					{Name: "data", Type: parser.GoType{
						Name:  "map[string]int",
						Kind:  parser.KindMap,
						Key:   &parser.GoType{Name: "string", Kind: parser.KindPrimitive},
						Value: &parser.GoType{Name: "int", Kind: parser.KindPrimitive},
					}},
				},
				Returns: []parser.GoType{
					{Name: "bool", Kind: parser.KindPrimitive},
				},
			},
		},
		Types: map[string]*parser.GoType{},
	}

	err := ValidateFunctions(parsed)
	if err != nil {
		t.Errorf("expected no error for string map, got: %v", err)
	}
}

func TestValidateFunctions_NonStringMapKey(t *testing.T) {
	parsed := &parser.ParsedFile{
		Package: "wasm",
		Functions: []parser.GoFunction{
			{
				Name: "ProcessMap",
				Params: []parser.GoParameter{
					{Name: "data", Type: parser.GoType{
						Name:  "map[int]string",
						Kind:  parser.KindMap,
						Key:   &parser.GoType{Name: "int", Kind: parser.KindPrimitive},
						Value: &parser.GoType{Name: "string", Kind: parser.KindPrimitive},
					}},
				},
				Returns: []parser.GoType{
					{Name: "bool", Kind: parser.KindPrimitive},
				},
			},
		},
		Types: map[string]*parser.GoType{},
	}

	err := ValidateFunctions(parsed)
	if err == nil {
		t.Fatal("expected error for non-string map key")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "only map[string]T is supported") {
		t.Errorf("expected error about map key type, got: %s", errStr)
	}
}

func TestValidateFunctions_ErrorReturn(t *testing.T) {
	parsed := &parser.ParsedFile{
		Package: "wasm",
		Functions: []parser.GoFunction{
			{
				Name: "MightFail",
				Params: []parser.GoParameter{
					{Name: "x", Type: parser.GoType{Name: "int", Kind: parser.KindPrimitive}},
				},
				Returns: []parser.GoType{
					{Name: "string", Kind: parser.KindPrimitive},
					{Name: "error", Kind: parser.KindError, IsError: true},
				},
			},
		},
		Types: map[string]*parser.GoType{},
	}

	err := ValidateFunctions(parsed)
	if err != nil {
		t.Errorf("expected no error for (T, error) return, got: %v", err)
	}
}

func TestValidateFunctions_ErrorNotLast(t *testing.T) {
	parsed := &parser.ParsedFile{
		Package: "wasm",
		Functions: []parser.GoFunction{
			{
				Name:   "BadReturn",
				Params: []parser.GoParameter{},
				Returns: []parser.GoType{
					{Name: "error", Kind: parser.KindError, IsError: true},
					{Name: "string", Kind: parser.KindPrimitive},
				},
			},
		},
		Types: map[string]*parser.GoType{},
	}

	err := ValidateFunctions(parsed)
	if err == nil {
		t.Fatal("expected error when error return is not last")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "error return type must be last") {
		t.Errorf("expected error about error position, got: %s", errStr)
	}
}

func TestValidateFunctions_Struct(t *testing.T) {
	parsed := &parser.ParsedFile{
		Package: "wasm",
		Functions: []parser.GoFunction{
			{
				Name:   "GetUser",
				Params: []parser.GoParameter{},
				Returns: []parser.GoType{
					{
						Name: "User",
						Kind: parser.KindStruct,
						Fields: []parser.GoField{
							{Name: "Name", Type: parser.GoType{Name: "string", Kind: parser.KindPrimitive}},
							{Name: "Age", Type: parser.GoType{Name: "int", Kind: parser.KindPrimitive}},
						},
					},
				},
			},
		},
		Types: map[string]*parser.GoType{},
	}

	err := ValidateFunctions(parsed)
	if err != nil {
		t.Errorf("expected no error for struct return, got: %v", err)
	}
}

func TestValidateFunctions_Pointer(t *testing.T) {
	parsed := &parser.ParsedFile{
		Package: "wasm",
		Functions: []parser.GoFunction{
			{
				Name: "GetOptional",
				Params: []parser.GoParameter{
					{Name: "id", Type: parser.GoType{Name: "int", Kind: parser.KindPrimitive}},
				},
				Returns: []parser.GoType{
					{
						Name: "*string",
						Kind: parser.KindPointer,
						Elem: &parser.GoType{Name: "string", Kind: parser.KindPrimitive},
					},
				},
			},
		},
		Types: map[string]*parser.GoType{},
	}

	err := ValidateFunctions(parsed)
	if err != nil {
		t.Errorf("expected no error for pointer type, got: %v", err)
	}
}

func TestValidateFunctions_UnsupportedTypes(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		wantErr  string
	}{
		{"channel", "chan", "unsupported type"},
		{"interface", "interface", "unsupported type"},
		{"external type", "time.Time", "unsupported type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := &parser.ParsedFile{
				Package: "wasm",
				Functions: []parser.GoFunction{
					{
						Name: "BadFunc",
						Params: []parser.GoParameter{
							{Name: "x", Type: parser.GoType{Name: tt.typeName, Kind: parser.KindUnsupported}},
						},
						Returns: []parser.GoType{},
					},
				},
				Types: map[string]*parser.GoType{},
			}

			err := ValidateFunctions(parsed)
			if err == nil {
				t.Fatalf("expected error for %s type", tt.name)
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got: %s", tt.wantErr, err.Error())
			}
		})
	}
}

func TestValidateFunctions_MultipleErrors(t *testing.T) {
	parsed := &parser.ParsedFile{
		Package: "wasm",
		Functions: []parser.GoFunction{
			{
				Name: "Func1",
				Params: []parser.GoParameter{
					{Name: "data", Type: parser.GoType{
						Name:  "map[int]string",
						Kind:  parser.KindMap,
						Key:   &parser.GoType{Name: "int", Kind: parser.KindPrimitive},
						Value: &parser.GoType{Name: "string", Kind: parser.KindPrimitive},
					}},
				},
				Returns: []parser.GoType{},
			},
			{
				Name: "Func2",
				Params: []parser.GoParameter{
					{Name: "data", Type: parser.GoType{
						Name:  "map[bool]int",
						Kind:  parser.KindMap,
						Key:   &parser.GoType{Name: "bool", Kind: parser.KindPrimitive},
						Value: &parser.GoType{Name: "int", Kind: parser.KindPrimitive},
					}},
				},
				Returns: []parser.GoType{},
			},
		},
		Types: map[string]*parser.GoType{},
	}

	err := ValidateFunctions(parsed)
	if err == nil {
		t.Fatal("expected errors")
	}

	var verr ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if len(verr.Errors) != 2 {
		t.Errorf("got %d errors, want 2: %v", len(verr.Errors), verr.Errors)
	}
}

func TestValidateFunctions_VoidCallback(t *testing.T) {
	// Valid: void callback with parameters
	parsed := &parser.ParsedFile{
		Package: "wasm",
		Functions: []parser.GoFunction{
			{
				Name: "ForEach",
				Params: []parser.GoParameter{
					{Name: "items", Type: parser.GoType{
						Name: "[]string",
						Kind: parser.KindSlice,
						Elem: &parser.GoType{Name: "string", Kind: parser.KindPrimitive},
					}},
					{Name: "callback", Type: parser.GoType{
						Name:   "func",
						Kind:   parser.KindFunction,
						IsVoid: true,
						CallbackParams: []parser.GoType{
							{Name: "string", Kind: parser.KindPrimitive},
							{Name: "int", Kind: parser.KindPrimitive},
						},
					}},
				},
				Returns: []parser.GoType{},
			},
		},
		Types: map[string]*parser.GoType{},
	}

	err := ValidateFunctions(parsed)
	if err != nil {
		t.Errorf("expected no error for void callback, got: %v", err)
	}
}

func TestValidateFunctions_VoidCallbackNoParams(t *testing.T) {
	// Valid: void callback with no parameters
	parsed := &parser.ParsedFile{
		Package: "wasm",
		Functions: []parser.GoFunction{
			{
				Name: "OnComplete",
				Params: []parser.GoParameter{
					{Name: "callback", Type: parser.GoType{
						Name:           "func",
						Kind:           parser.KindFunction,
						IsVoid:         true,
						CallbackParams: []parser.GoType{},
					}},
				},
				Returns: []parser.GoType{},
			},
		},
		Types: map[string]*parser.GoType{},
	}

	err := ValidateFunctions(parsed)
	if err != nil {
		t.Errorf("expected no error for void callback with no params, got: %v", err)
	}
}

// TestUnsupportedCallbackPatterns verifies the validator correctly rejects
// callback patterns that gowasm-bindgen doesn't support.
func TestUnsupportedCallbackPatterns(t *testing.T) {
	tests := []struct {
		name string
		fn   parser.GoFunction
		want string // expected error substring
	}{
		{
			name: "callback with return value",
			fn: parser.GoFunction{
				Name: "Filter",
				Params: []parser.GoParameter{{
					Name: "cb",
					Type: parser.GoType{
						Kind:           parser.KindFunction,
						CallbackParams: []parser.GoType{{Name: "string", Kind: parser.KindPrimitive}},
						IsVoid:         false, // has return value
					},
				}},
			},
			want: "only void callbacks are supported",
		},
		{
			name: "nested callback",
			fn: parser.GoFunction{
				Name: "WithCallback",
				Params: []parser.GoParameter{{
					Name: "cb",
					Type: parser.GoType{
						Kind:   parser.KindFunction,
						IsVoid: true,
						CallbackParams: []parser.GoType{{
							Kind:   parser.KindFunction, // nested callback
							IsVoid: true,
						}},
					},
				}},
			},
			want: "nested callback",
		},
		{
			name: "function return type",
			fn: parser.GoFunction{
				Name:   "GetHandler",
				Params: []parser.GoParameter{},
				Returns: []parser.GoType{{
					Kind:   parser.KindFunction,
					IsVoid: true,
				}},
			},
			want: "functions are only supported as callback parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := &parser.ParsedFile{
				Package:   "wasm",
				Functions: []parser.GoFunction{tt.fn},
				Types:     map[string]*parser.GoType{},
			}
			err := ValidateFunctions(parsed)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("expected error containing %q, got %q", tt.want, err.Error())
			}
		})
	}
}

func TestValidateFunctions_AnonymousFields(t *testing.T) {
	// Struct with anonymous/embedded field should be rejected
	parsed := &parser.ParsedFile{
		Package: "wasm",
		Functions: []parser.GoFunction{
			{
				Name:   "GetUser",
				Params: []parser.GoParameter{},
				Returns: []parser.GoType{{
					Name: "User",
					Kind: parser.KindStruct,
					Fields: []parser.GoField{
						{Name: "Name", Type: parser.GoType{Name: "string", Kind: parser.KindPrimitive}},
						{Name: "", Type: parser.GoType{Name: "Meta", Kind: parser.KindStruct}}, // anonymous field
					},
				}},
			},
		},
		Types: map[string]*parser.GoType{},
	}

	err := ValidateFunctions(parsed)
	if err == nil {
		t.Fatal("expected error for anonymous field, got nil")
	}
	if !strings.Contains(err.Error(), "anonymous/embedded field") {
		t.Errorf("expected error about anonymous field, got: %v", err)
	}
}
