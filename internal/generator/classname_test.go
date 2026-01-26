package generator

import "testing"

func TestDeriveClassName(t *testing.T) {
	tests := []struct {
		dirName string
		want    string
	}{
		{"image-wasm", "GoImageWasm"},
		{"go-calculator", "GoCalculator"},
		{"Go-Calculator", "GoCalculator"},
		{"GO-WASM", "GoWasm"},
		{"wasm", "GoWasm"},
		{"simple", "GoSimple"},
		{"my_package", "GoMyPackage"},
		{".", "GoMain"},
		{"", "GoMain"},
	}

	for _, tt := range tests {
		t.Run(tt.dirName, func(t *testing.T) {
			got := DeriveClassName(tt.dirName)
			if got != tt.want {
				t.Errorf("DeriveClassName(%q) = %q, want %q", tt.dirName, got, tt.want)
			}
		})
	}
}

func TestToTitleCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello-world", "HelloWorld"},
		{"my_package", "MyPackage"},
		{"simple", "Simple"},
		{"UPPER", "Upper"},
		{"mixed-Case_test", "MixedCaseTest"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toTitleCase(tt.input)
			if got != tt.want {
				t.Errorf("toTitleCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
