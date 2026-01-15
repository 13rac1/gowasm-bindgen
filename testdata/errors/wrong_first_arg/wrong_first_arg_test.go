package wrong_first_arg

import (
	"syscall/js"
	"testing"
)

func processData(this js.Value, args []js.Value) interface{} {
	return nil
}

func TestWrongFirstArg(t *testing.T) {
	// Wrong: first argument is nil, not js.Null()
	result := processData(nil, []js.Value{js.ValueOf("test")})
	_ = result
}
