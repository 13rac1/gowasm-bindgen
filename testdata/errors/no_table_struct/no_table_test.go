package no_table_struct

import (
	"syscall/js"
	"testing"
)

// processData is a stub WASM function for testing
func processData(this js.Value, args []js.Value) any {
	return args[0].String() + args[1].String()
}

// TestProcessData_NoTableStruct tests without table-driven tests
// This should trigger: "using fallback param names (arg0, arg1, ...)"
func TestProcessData_NoTableStruct(t *testing.T) {
	// Direct call without table struct - no way to infer param names
	result := processData(js.Null(), []js.Value{
		js.ValueOf("hello"),
		js.ValueOf("world"),
	})

	if result != "helloworld" {
		t.Errorf("expected helloworld, got %v", result)
	}
}
