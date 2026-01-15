package multiple

import (
	"syscall/js"
	"testing"
)

// addNumbers is a mock WASM function that adds two numbers
func addNumbers(this js.Value, args []js.Value) interface{} {
	a := args[0].Int()
	b := args[1].Int()
	return js.ValueOf(a + b)
}

// greet is a mock WASM function that returns a greeting
func greet(this js.Value, args []js.Value) interface{} {
	name := args[0].String()
	return js.ValueOf("Hello, " + name)
}

// checkFlag is a mock WASM function that returns a boolean
func checkFlag(this js.Value, args []js.Value) interface{} {
	return js.ValueOf(true)
}

// TestAddNumbers tests number addition function
func TestAddNumbers(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{name: "add positive", a: 5, b: 3, want: 8},
		{name: "add negative", a: -2, b: 7, want: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addNumbers(js.Null(), []js.Value{
				js.ValueOf(tt.a),
				js.ValueOf(tt.b),
			})

			jsResult := result.(js.Value)
			got := jsResult.Int()
			if got != tt.want {
				t.Errorf("addNumbers(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestGreet tests string greeting function
func TestGreet(t *testing.T) {
	tests := []struct {
		name     string
		userName string
		want     string
	}{
		{name: "greet alice", userName: "Alice", want: "Hello, Alice"},
		{name: "greet bob", userName: "Bob", want: "Hello, Bob"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := greet(js.Null(), []js.Value{
				js.ValueOf(tt.userName),
			})

			jsResult := result.(js.Value)
			got := jsResult.String()
			if got != tt.want {
				t.Errorf("greet(%s) = %s, want %s", tt.userName, got, tt.want)
			}
		})
	}
}

// TestCheckFlag tests boolean return function
func TestCheckFlag(t *testing.T) {
	result := checkFlag(js.Null(), []js.Value{})
	jsResult := result.(js.Value)

	got := jsResult.Bool()
	if !got {
		t.Error("checkFlag() = false, want true")
	}
}
