package unions

import (
	"syscall/js"
	"testing"
)

// validate is a mock WASM function that returns a union type
func validate(this js.Value, args []js.Value) interface{} {
	input := args[0].String()

	obj := make(map[string]interface{})
	if input == "invalid" {
		obj["error"] = "validation failed"
	} else {
		obj["success"] = true
	}
	return js.ValueOf(obj)
}

// TestUnionReturn tests extraction of union return type
func TestUnionReturn(t *testing.T) {
	result := validate(js.Null(), []js.Value{js.ValueOf("valid-data")})
	jsResult := result.(js.Value)

	if !jsResult.Get("error").IsUndefined() {
		errMsg := jsResult.Get("error").String()
		t.Errorf("unexpected error: %s", errMsg)
	} else {
		success := jsResult.Get("success").Bool()
		if !success {
			t.Error("expected success to be true")
		}
	}
}
