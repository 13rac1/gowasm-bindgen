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

func TestDivide(t *testing.T) {
	tests := []struct {
		name    string
		a       int
		b       int
		want    int
		wantErr string
	}{
		{name: "normal", a: 10, b: 2, want: 5, wantErr: ""},
		{name: "division_by_zero", a: 10, b: 0, want: 0, wantErr: "division by zero"},
		{name: "negative", a: -10, b: 2, want: -5, wantErr: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Divide(tt.a, tt.b)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Divide(%d, %d) expected error %q, got nil", tt.a, tt.b, tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("Divide(%d, %d) error = %q, want %q", tt.a, tt.b, err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Divide(%d, %d) unexpected error: %v", tt.a, tt.b, err)
			}
			if got != tt.want {
				t.Errorf("Divide(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestHashData(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want []byte
	}{
		{name: "empty", data: []byte{}, want: []byte{0, 0, 0, 0}},
		{name: "single_byte", data: []byte{0xFF}, want: []byte{0xFF, 0, 0, 0}},
		{name: "four_bytes", data: []byte{1, 2, 3, 4}, want: []byte{1, 2, 3, 4}},
		{name: "xor_wraps", data: []byte{1, 2, 3, 4, 5, 6, 7, 8}, want: []byte{1 ^ 5, 2 ^ 6, 3 ^ 7, 4 ^ 8}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HashData(tt.data)
			if len(got) != len(tt.want) {
				t.Fatalf("HashData() returned %d bytes, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("HashData()[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestProcessNumbers(t *testing.T) {
	tests := []struct {
		name string
		nums []int32
		want []int32
	}{
		{name: "empty", nums: []int32{}, want: []int32{}},
		{name: "single", nums: []int32{5}, want: []int32{10}},
		{name: "multiple", nums: []int32{1, 2, 3, 4, 5}, want: []int32{2, 4, 6, 8, 10}},
		{name: "negative", nums: []int32{-5, -10}, want: []int32{-10, -20}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProcessNumbers(tt.nums)
			if len(got) != len(tt.want) {
				t.Fatalf("ProcessNumbers() returned %d elements, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ProcessNumbers()[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// Note: ForEach test removed because ForEach requires --sync mode
// (callbacks cannot be serialized across Web Worker boundary)
