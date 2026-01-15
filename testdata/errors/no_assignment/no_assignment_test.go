package no_assignment

import (
	"syscall/js"
	"testing"
)

func processData(this js.Value, args []js.Value) interface{} {
	return nil
}

func TestNoAssignment(t *testing.T) {
	// Wrong: call result is not assigned to a variable
	processData(js.Null(), []js.Value{js.ValueOf("test")})
}
