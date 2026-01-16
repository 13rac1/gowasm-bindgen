package main

import (
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
			got := Greet(tt.name)
			if got != tt.want {
				t.Errorf("Greet(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestCalculate(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		op   string
		want int
	}{
		{name: "add", a: 10, b: 5, op: "add", want: 15},
		{name: "subtract", a: 10, b: 5, op: "sub", want: 5},
		{name: "multiply", a: 10, b: 5, op: "mul", want: 50},
		{name: "divide", a: 10, b: 5, op: "div", want: 2},
		{name: "divide_by_zero", a: 10, b: 0, op: "div", want: 0},
		{name: "unknown_op", a: 10, b: 5, op: "unknown", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Calculate(tt.a, tt.b, tt.op)
			if got != tt.want {
				t.Errorf("Calculate(%d, %d, %q) = %v, want %v", tt.a, tt.b, tt.op, got, tt.want)
			}
		})
	}
}

func TestFormatUser(t *testing.T) {
	tests := []struct {
		name           string
		userName       string
		age            int
		active         bool
		wantStatus     string
		wantNamePrefix string
	}{
		{
			name:           "active_user",
			userName:       "Alice",
			age:            30,
			active:         true,
			wantStatus:     "active",
			wantNamePrefix: "Alice (30)",
		},
		{
			name:           "inactive_user",
			userName:       "Bob",
			age:            25,
			active:         false,
			wantStatus:     "inactive",
			wantNamePrefix: "Bob (25)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatUser(tt.userName, tt.age, tt.active)
			if got.Status != tt.wantStatus {
				t.Errorf("FormatUser().Status = %v, want %v", got.Status, tt.wantStatus)
			}
			if got.DisplayName != tt.wantNamePrefix {
				t.Errorf("FormatUser().DisplayName = %v, want %v", got.DisplayName, tt.wantNamePrefix)
			}
		})
	}
}

func TestSumNumbers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{name: "simple", input: "1,2,3", want: 6},
		{name: "with_spaces", input: "10, 20, 30, 40", want: 100},
		{name: "empty", input: "", want: 0},
		{name: "single", input: "42", want: 42},
		{name: "with_invalid", input: "1,abc,3", want: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SumNumbers(tt.input)
			if got != tt.want {
				t.Errorf("SumNumbers(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		wantValid bool
		wantError string
	}{
		{
			name:      "valid",
			email:     "user@example.com",
			wantValid: true,
			wantError: "",
		},
		{
			name:      "missing_at",
			email:     "invalid",
			wantValid: false,
			wantError: "missing @ symbol",
		},
		{
			name:      "missing_domain",
			email:     "missing@domain",
			wantValid: false,
			wantError: "domain must contain a dot",
		},
		{
			name:      "empty_local",
			email:     "@nodomain.com",
			wantValid: false,
			wantError: "invalid email format",
		},
		{
			name:      "empty_domain",
			email:     "user@",
			wantValid: false,
			wantError: "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateEmail(tt.email)
			if got.Valid != tt.wantValid {
				t.Errorf("ValidateEmail(%q).Valid = %v, want %v", tt.email, got.Valid, tt.wantValid)
			}
			if !tt.wantValid && got.Error != tt.wantError {
				t.Errorf("ValidateEmail(%q).Error = %v, want %v", tt.email, got.Error, tt.wantError)
			}
		})
	}
}
