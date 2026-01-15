package selector_func

import (
	"syscall/js"
	"testing"
)

type processor struct{}

func (p *processor) process(this js.Value, args []js.Value) interface{} {
	return nil
}

var pkg = &processor{}

func TestSelectorFunc(t *testing.T) {
	// Wrong: function is method call, not simple identifier
	result := pkg.process(js.Null(), []js.Value{js.ValueOf("test")})
	_ = result
}
