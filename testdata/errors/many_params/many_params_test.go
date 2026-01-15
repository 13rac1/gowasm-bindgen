package many_params

import (
	"syscall/js"
	"testing"
)

// processMany is a stub WASM function for testing
func processMany(this js.Value, args []js.Value) any {
	sum := 0.0
	for _, arg := range args {
		sum += arg.Float()
	}
	return sum
}

// TestProcessMany_ManyParams tests with 15 parameters to verify generateArgName fix
// This should trigger: "using fallback param names (arg0, arg1, ...)"
// Previously this would have generated invalid names for arg10-arg14
func TestProcessMany_ManyParams(t *testing.T) {
	result := processMany(js.Null(), []js.Value{
		js.ValueOf(1),
		js.ValueOf(2),
		js.ValueOf(3),
		js.ValueOf(4),
		js.ValueOf(5),
		js.ValueOf(6),
		js.ValueOf(7),
		js.ValueOf(8),
		js.ValueOf(9),
		js.ValueOf(10),
		js.ValueOf(11),
		js.ValueOf(12),
		js.ValueOf(13),
		js.ValueOf(14),
		js.ValueOf(15),
	})

	expected := 120.0 // 1+2+...+15 = 120
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
