package wrong_second_arg

import (
	"syscall/js"
	"testing"
)

func processData(this js.Value, args []js.Value) interface{} {
	return nil
}

func TestWrongSecondArg(t *testing.T) {
	// Wrong: second argument is a variable, not []js.Value{...} literal
	args := []js.Value{js.ValueOf("test")}
	result := processData(js.Null(), args)
	_ = result
}
