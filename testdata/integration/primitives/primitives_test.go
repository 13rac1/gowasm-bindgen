package primitives

import (
	"syscall/js"
	"testing"
)

// processPrimitives is a mock WASM function that processes primitive types
func processPrimitives(this js.Value, args []js.Value) interface{} {
	str := args[0].String()
	num := args[1].Int()
	flag := args[2].Bool()

	// Mock processing
	result := str + "-processed"
	if flag && num > 0 {
		return js.ValueOf(result)
	}
	return js.Null()
}

// TestPrimitives tests primitive type parameter extraction
func TestPrimitives(t *testing.T) {
	tests := []struct {
		name string
		str  string
		num  int
		flag bool
	}{
		{name: "test1", str: "hello", num: 42, flag: true},
		{name: "test2", str: "world", num: 100, flag: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processPrimitives(js.Null(), []js.Value{
				js.ValueOf(tt.str),
				js.ValueOf(tt.num),
				js.ValueOf(tt.flag),
			})

			// Type assertion pattern for return type inference
			jsResult := result.(js.Value)
			if jsResult.IsNull() {
				return
			}
			_ = jsResult.String()
		})
	}
}
