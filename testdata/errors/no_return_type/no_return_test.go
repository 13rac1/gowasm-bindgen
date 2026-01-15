package no_return_type

import (
	"syscall/js"
	"testing"
)

// compute is a stub WASM function for testing
func compute(this js.Value, args []js.Value) any {
	return map[string]any{"value": 42}
}

// TestCompute_NoReturnType tests without result type inference
// This should trigger: "return type inferred as 'any'"
// AND: "using fallback param names (arg0, arg1, ...)"
func TestCompute_NoReturnType(t *testing.T) {
	// No result.Get("field").Type() calls - can't infer return type
	result := compute(js.Null(), []js.Value{
		js.ValueOf(123),
	})

	// Just checking non-nil - no type assertions that help inference
	if result == nil {
		t.Error("expected non-nil result")
	}
}
