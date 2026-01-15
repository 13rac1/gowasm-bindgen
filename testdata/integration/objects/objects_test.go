package objects

import (
	"syscall/js"
	"testing"
)

// getObject is a mock WASM function that returns an object
func getObject(this js.Value, args []js.Value) interface{} {
	obj := make(map[string]interface{})
	obj["valid"] = true
	obj["hash"] = "abc123"
	obj["count"] = 42
	return js.ValueOf(obj)
}

// TestObjectReturn tests extraction of object return type
func TestObjectReturn(t *testing.T) {
	result := getObject(js.Null(), []js.Value{js.ValueOf("data")})
	jsResult := result.(js.Value)

	valid := jsResult.Get("valid").Bool()
	if !valid {
		t.Error("expected valid to be true")
	}

	hash := jsResult.Get("hash").String()
	if hash != "abc123" {
		t.Errorf("expected hash abc123, got %s", hash)
	}

	count := jsResult.Get("count").Int()
	if count != 42 {
		t.Errorf("expected count 42, got %d", count)
	}
}
