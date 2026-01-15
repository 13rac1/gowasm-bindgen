package arrays

import (
	"syscall/js"
	"testing"
)

// processArrays is a mock WASM function that processes array types
func processArrays(this js.Value, args []js.Value) interface{} {
	// Extract string array
	stringsLen := args[0].Length()
	strings := make([]string, stringsLen)
	for i := 0; i < stringsLen; i++ {
		strings[i] = args[0].Index(i).String()
	}

	// Extract number array
	numbersLen := args[1].Length()
	numbers := make([]int, numbersLen)
	for i := 0; i < numbersLen; i++ {
		numbers[i] = args[1].Index(i).Int()
	}

	// Mock processing
	return js.ValueOf("processed")
}

// TestArrays tests array type parameter extraction
func TestArrays(t *testing.T) {
	tests := []struct {
		name    string
		strings []string
		numbers []int
	}{
		{name: "test1", strings: []string{"a", "b", "c"}, numbers: []int{1, 2, 3}},
		{name: "test2", strings: []string{"x", "y"}, numbers: []int{10, 20}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processArrays(js.Null(), []js.Value{
				js.ValueOf(tt.strings),
				js.ValueOf(tt.numbers),
			})

			jsResult := result.(js.Value)
			got := jsResult.String()
			if got != "processed" {
				t.Errorf("expected 'processed', got %s", got)
			}
		})
	}
}
