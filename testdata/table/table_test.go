package table

import (
	"syscall/js"
	"testing"
)

// processData is a mock WASM function for testing
func processData(this js.Value, args []js.Value) interface{} {
	return js.ValueOf("mock result")
}

// TestWithTableDriven tests parameter extraction from table-driven tests
func TestWithTableDriven(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		count   int
		enabled bool
		want    string
	}{
		{name: "test1", input: "hello", count: 5, enabled: true, want: "result"},
		{name: "test2", input: "world", count: 10, enabled: false, want: "output"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processData(js.Null(), []js.Value{
				js.ValueOf(tt.input),
				js.ValueOf(tt.count),
				js.ValueOf(tt.enabled),
			})

			if got := result.(js.Value).String(); got != tt.want {
				t.Errorf("processData() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestWithoutTable tests fallback to arg0, arg1 when no table struct
func TestWithoutTable(t *testing.T) {
	result := processData(js.Null(), []js.Value{
		js.ValueOf("data"),
		js.ValueOf(42),
	})

	if result == nil {
		t.Error("processData() returned nil")
	}
}
