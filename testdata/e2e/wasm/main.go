//go:build js && wasm

package main

import (
	"syscall/js"
)

// Simple string operation
func greet(this js.Value, args []js.Value) interface{} {
	name := args[0].String()
	return "Hello, " + name + "!"
}

// Numeric operation
func add(this js.Value, args []js.Value) interface{} {
	a := args[0].Int()
	b := args[1].Int()
	return a + b
}

// Object return
func getInfo(this js.Value, args []js.Value) interface{} {
	return map[string]interface{}{
		"name":    args[0].String(),
		"version": 1,
		"active":  true,
	}
}

func main() {
	js.Global().Set("greet", js.FuncOf(greet))
	js.Global().Set("add", js.FuncOf(add))
	js.Global().Set("getInfo", js.FuncOf(getInfo))

	// Keep the program running
	<-make(chan struct{})
}
