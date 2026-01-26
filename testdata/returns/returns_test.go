package returns

import (
	"syscall/js"
	"testing"
)

// compute is a mock WASM function that returns an object
func compute(this js.Value, args []js.Value) interface{} {
	obj := make(map[string]interface{})
	obj["valid"] = true
	obj["hash"] = "abc123"
	return js.ValueOf(obj)
}

// validate is a mock WASM function that returns a union type
func validate(this js.Value, args []js.Value) interface{} {
	obj := make(map[string]interface{})
	if args[0].String() == "error" {
		obj["error"] = "validation failed"
	} else {
		obj["success"] = true
	}
	return js.ValueOf(obj)
}

// getString is a mock WASM function that returns a string
func getString(this js.Value, args []js.Value) interface{} {
	return js.ValueOf("result string")
}

// getNumber is a mock WASM function that returns a number
func getNumber(this js.Value, args []js.Value) interface{} {
	return js.ValueOf(42)
}

// TestReturnsObject tests extraction of object return type with fields
func TestReturnsObject(t *testing.T) {
	result := compute(js.Null(), []js.Value{js.ValueOf("data")})
	jsResult := result.(js.Value)

	valid := jsResult.Get("valid").Bool()
	if !valid {
		t.Error("expected valid to be true")
	}

	hash := jsResult.Get("hash").String()
	if hash != "abc123" {
		t.Errorf("expected hash to be abc123, got %s", hash)
	}
}

// TestReturnsUnion tests extraction of union return type (error|success)
func TestReturnsUnion(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantSuccess bool
		wantError   string
	}{
		{"success case", "data", true, ""},
		{"error case", "error", false, "validation failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validate(js.Null(), []js.Value{js.ValueOf(tt.input)})
			jsResult := result.(js.Value)

			if tt.wantError != "" {
				got := jsResult.Get("error").String()
				if got != tt.wantError {
					t.Errorf("error = %q, want %q", got, tt.wantError)
				}
			} else {
				if !jsResult.Get("error").IsUndefined() {
					t.Errorf("unexpected error: %s", jsResult.Get("error").String())
				}
				if got := jsResult.Get("success").Bool(); got != tt.wantSuccess {
					t.Errorf("success = %v, want %v", got, tt.wantSuccess)
				}
			}
		})
	}
}

// TestReturnsString tests extraction of string return type
func TestReturnsString(t *testing.T) {
	result := getString(js.Null(), []js.Value{js.ValueOf("input")})
	jsResult := result.(js.Value)

	str := jsResult.String()
	if str != "result string" {
		t.Errorf("expected result string, got %s", str)
	}
}

// TestReturnsNumber tests extraction of number return type
func TestReturnsNumber(t *testing.T) {
	result := getNumber(js.Null(), []js.Value{js.ValueOf("input")})
	jsResult := result.(js.Value)

	num := jsResult.Int()
	if num != 42 {
		t.Errorf("expected 42, got %d", num)
	}
}
