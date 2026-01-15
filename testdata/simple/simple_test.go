package simple

import (
	"syscall/js"
	"testing"
)

func hashData(this js.Value, args []js.Value) interface{} {
	return args[0].String() + "-hashed"
}

func TestHashData(t *testing.T) {
	result := hashData(js.Null(), []js.Value{
		js.ValueOf("hello"),
	})
	if result != "hello-hashed" {
		t.Fail()
	}
}
