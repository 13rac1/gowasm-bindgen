package wrong_slice_type

import (
	"syscall/js"
	"testing"
)

func processData(this js.Value, args []js.Value) interface{} {
	return nil
}

func TestWrongSliceType(t *testing.T) {
	// Wrong: second argument is []interface{}, not []js.Value{}
	result := processData(js.Null(), []interface{}{js.ValueOf("test")})
	_ = result
}
