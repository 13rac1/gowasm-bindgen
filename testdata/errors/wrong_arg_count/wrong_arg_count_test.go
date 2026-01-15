package wrong_arg_count

import (
	"syscall/js"
	"testing"
)

func processData(this js.Value, args []js.Value) interface{} {
	return nil
}

func TestWrongArgCount(t *testing.T) {
	// Wrong: 4 arguments instead of 2
	result := processData(js.Null(), []js.Value{js.ValueOf("test")}, "extra", 123)
	_ = result
}
