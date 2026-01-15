package main

import (
	"syscall/js"
	"testing"
)

func TestGreet(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "World", want: "Hello, World!"},
		{name: "Go", want: "Hello, Go!"},
		{name: "WASM", want: "Hello, WASM!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := greet(js.Null(), []js.Value{
				js.ValueOf(tt.name),
			})
			jsResult := result.(js.Value)
			got := jsResult.String()
			if got != tt.want {
				t.Errorf("greet(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestCalculate(t *testing.T) {
	tests := []struct {
		a    int
		b    int
		op   string
		want int
	}{
		{a: 10, b: 5, op: "add", want: 15},
		{a: 10, b: 5, op: "sub", want: 5},
		{a: 10, b: 5, op: "mul", want: 50},
		{a: 10, b: 5, op: "div", want: 2},
		{a: 10, b: 0, op: "div", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			result := calculate(js.Null(), []js.Value{
				js.ValueOf(tt.a),
				js.ValueOf(tt.b),
				js.ValueOf(tt.op),
			})
			jsResult := result.(js.Value)
			got := jsResult.Int()
			if got != tt.want {
				t.Errorf("calculate(%d, %d, %q) = %v, want %v", tt.a, tt.b, tt.op, got, tt.want)
			}
		})
	}
}

func TestFormatUser(t *testing.T) {
	tests := []struct {
		name   string
		age    int
		active bool
	}{
		{name: "Alice", age: 30, active: true},
		{name: "Bob", age: 25, active: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatUser(js.Null(), []js.Value{
				js.ValueOf(tt.name),
				js.ValueOf(tt.age),
				js.ValueOf(tt.active),
			})
			jsResult := result.(js.Value)

			displayName := jsResult.Get("displayName").String()
			status := jsResult.Get("status").String()

			if tt.active && status != "active" {
				t.Errorf("expected status 'active', got %q", status)
			}
			if !tt.active && status != "inactive" {
				t.Errorf("expected status 'inactive', got %q", status)
			}
			if displayName == "" {
				t.Error("expected non-empty displayName")
			}
		})
	}
}

func TestSumNumbers(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{input: "1,2,3", want: 6},
		{input: "10, 20, 30, 40", want: 100},
		{input: "", want: 0},
		{input: "42", want: 42},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sumNumbers(js.Null(), []js.Value{
				js.ValueOf(tt.input),
			})
			jsResult := result.(js.Value)
			got := jsResult.Int()

			if got != tt.want {
				t.Errorf("sumNumbers(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email string
	}{
		{email: "user@example.com"},
		{email: "invalid"},
		{email: "missing@domain"},
		{email: "@nodomain.com"},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := validateEmail(js.Null(), []js.Value{
				js.ValueOf(tt.email),
			})
			jsResult := result.(js.Value)

			valid := jsResult.Get("valid").Bool()

			// Check error field for invalid emails
			if !valid {
				if jsResult.Get("error").IsUndefined() {
					t.Error("invalid email should have error message")
				} else {
					errMsg := jsResult.Get("error").String()
					if errMsg == "" {
						t.Error("error message should not be empty")
					}
				}
			}
		})
	}
}
