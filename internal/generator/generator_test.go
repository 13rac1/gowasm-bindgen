package generator

import (
	"strings"
	"testing"

	"github.com/13rac1/gowasm-bindgen/internal/extractor"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name string
		sigs []extractor.FunctionSignature
		want []string
	}{
		{
			name: "empty signatures",
			sigs: []extractor.FunctionSignature{},
			want: []string{
				"Auto-generated TypeScript declarations",
				"export {};",
			},
		},
		{
			name: "single function",
			sigs: []extractor.FunctionSignature{
				{
					Name: "hashData",
					Params: []extractor.Parameter{
						{Name: "data", Type: "string"},
					},
					Returns: extractor.ReturnType{Type: "string"},
				},
			},
			want: []string{
				"declare global",
				"interface Window",
				"hashData(data: string): string;",
				"export {};",
			},
		},
		{
			name: "multiple functions",
			sigs: []extractor.FunctionSignature{
				{
					Name:    "funcOne",
					Params:  []extractor.Parameter{{Name: "x", Type: "number"}},
					Returns: extractor.ReturnType{Type: "number"},
				},
				{
					Name:    "funcTwo",
					Params:  []extractor.Parameter{{Name: "s", Type: "string"}},
					Returns: extractor.ReturnType{Type: "boolean"},
				},
			},
			want: []string{
				"funcOne(x: number): number;",
				"funcTwo(s: string): boolean;",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Generate(tt.sigs)
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("Generate() missing %q in output:\n%s", want, got)
				}
			}
		})
	}
}

func TestGenerateHeader(t *testing.T) {
	got := GenerateHeader()
	want := []string{
		"Auto-generated TypeScript declarations",
		"gowasm-bindgen",
	}

	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("GenerateHeader() missing %q in output", w)
		}
	}
}

func TestGenerateFunction(t *testing.T) {
	tests := []struct {
		name string
		sig  extractor.FunctionSignature
		want []string
	}{
		{
			name: "simple function",
			sig: extractor.FunctionSignature{
				Name: "hashData",
				Params: []extractor.Parameter{
					{Name: "data", Type: "string"},
				},
				Returns: extractor.ReturnType{Type: "string"},
			},
			want: []string{
				"declare global",
				"interface Window",
				"hashData(data: string): string;",
			},
		},
		{
			name: "no parameters",
			sig: extractor.FunctionSignature{
				Name:    "getCurrentTime",
				Params:  []extractor.Parameter{},
				Returns: extractor.ReturnType{Type: "number"},
			},
			want: []string{
				"getCurrentTime(): number;",
			},
		},
		{
			name: "object return",
			sig: extractor.FunctionSignature{
				Name: "validate",
				Params: []extractor.Parameter{
					{Name: "input", Type: "string"},
				},
				Returns: extractor.ReturnType{
					Fields: []extractor.Field{
						{Name: "valid", Type: "boolean"},
						{Name: "hash", Type: "string"},
					},
				},
			},
			want: []string{
				"validate(input: string): {valid: boolean, hash: string};",
			},
		},
		{
			name: "with documentation",
			sig: extractor.FunctionSignature{
				Name:    "process",
				Params:  []extractor.Parameter{{Name: "data", Type: "string"}},
				Returns: extractor.ReturnType{Type: "string"},
				Doc:     "Process data and return result",
			},
			want: []string{
				"/**",
				"Process data and return result",
				"*/",
				"process(data: string): string;",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateFunction(tt.sig)
			for _, w := range tt.want {
				if !strings.Contains(got, w) {
					t.Errorf("GenerateFunction() missing %q in output:\n%s", w, got)
				}
			}
		})
	}
}

func TestGenerateJSDoc(t *testing.T) {
	tests := []struct {
		name string
		sig  extractor.FunctionSignature
		want string
	}{
		{
			name: "no doc no examples",
			sig:  extractor.FunctionSignature{},
			want: "",
		},
		{
			name: "doc only",
			sig: extractor.FunctionSignature{
				Doc: "This function does something",
			},
			want: "/**\n * This function does something\n */",
		},
		{
			name: "examples only",
			sig: extractor.FunctionSignature{
				Name: "test",
				Examples: []extractor.Example{
					{Name: "basic case", Args: []string{`"hello"`}},
				},
			},
			want: "/**\n * @example\n * // basic case\n * test(\"hello\")\n */",
		},
		{
			name: "doc and examples",
			sig: extractor.FunctionSignature{
				Name: "compute",
				Doc:  "Computes a value",
				Examples: []extractor.Example{
					{Name: "simple", Args: []string{"42"}},
				},
			},
			want: "/**\n * Computes a value\n *\n * @example\n * // simple\n * compute(42)\n */",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateJSDoc(tt.sig)
			if got != tt.want {
				t.Errorf("GenerateJSDoc() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateParams(t *testing.T) {
	tests := []struct {
		name   string
		params []extractor.Parameter
		want   string
	}{
		{
			name:   "no parameters",
			params: []extractor.Parameter{},
			want:   "",
		},
		{
			name: "single parameter",
			params: []extractor.Parameter{
				{Name: "data", Type: "string"},
			},
			want: "data: string",
		},
		{
			name: "multiple parameters",
			params: []extractor.Parameter{
				{Name: "input", Type: "string"},
				{Name: "count", Type: "number"},
				{Name: "enabled", Type: "boolean"},
			},
			want: "input: string, count: number, enabled: boolean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateParams(tt.params)
			if got != tt.want {
				t.Errorf("GenerateParams() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateReturnType_Primitive(t *testing.T) {
	tests := []struct {
		name string
		ret  extractor.ReturnType
		want string
	}{
		{
			name: "string",
			ret:  extractor.ReturnType{Type: "string"},
			want: "string",
		},
		{
			name: "number",
			ret:  extractor.ReturnType{Type: "number"},
			want: "number",
		},
		{
			name: "boolean",
			ret:  extractor.ReturnType{Type: "boolean"},
			want: "boolean",
		},
		{
			name: "void",
			ret:  extractor.ReturnType{},
			want: "void",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateReturnType(tt.ret)
			if got != tt.want {
				t.Errorf("GenerateReturnType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateReturnType_Object(t *testing.T) {
	tests := []struct {
		name string
		ret  extractor.ReturnType
		want string
	}{
		{
			name: "single field",
			ret: extractor.ReturnType{
				Fields: []extractor.Field{
					{Name: "result", Type: "string"},
				},
			},
			want: "{result: string}",
		},
		{
			name: "multiple fields",
			ret: extractor.ReturnType{
				Fields: []extractor.Field{
					{Name: "valid", Type: "boolean"},
					{Name: "hash", Type: "string"},
					{Name: "count", Type: "number"},
				},
			},
			want: "{valid: boolean, hash: string, count: number}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateReturnType(tt.ret)
			if got != tt.want {
				t.Errorf("GenerateReturnType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateReturnType_Union(t *testing.T) {
	tests := []struct {
		name string
		ret  extractor.ReturnType
		want string
	}{
		{
			name: "two variants",
			ret: extractor.ReturnType{
				IsUnion: true,
				Fields: []extractor.Field{
					{Name: "success", Type: "boolean"},
					{Name: "error", Type: "string"},
				},
			},
			want: "{success: boolean} | {error: string}",
		},
		{
			name: "three variants",
			ret: extractor.ReturnType{
				IsUnion: true,
				Fields: []extractor.Field{
					{Name: "data", Type: "string"},
					{Name: "error", Type: "string"},
					{Name: "loading", Type: "boolean"},
				},
			},
			want: "{data: string} | {error: string} | {loading: boolean}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateReturnType(tt.ret)
			if got != tt.want {
				t.Errorf("GenerateReturnType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatExample(t *testing.T) {
	tests := []struct {
		name string
		sig  extractor.FunctionSignature
		ex   extractor.Example
		want string
	}{
		{
			name: "simple example",
			sig:  extractor.FunctionSignature{Name: "hashData"},
			ex: extractor.Example{
				Name: "basic hash",
				Args: []string{`"hello"`},
			},
			want: " * // basic hash\n * hashData(\"hello\")",
		},
		{
			name: "multiple args",
			sig:  extractor.FunctionSignature{Name: "compute"},
			ex: extractor.Example{
				Name: "with numbers",
				Args: []string{"42", "true", `"test"`},
			},
			want: " * // with numbers\n * compute(42, true, \"test\")",
		},
		{
			name: "no name",
			sig:  extractor.FunctionSignature{Name: "process"},
			ex: extractor.Example{
				Args: []string{`"data"`},
			},
			want: " * process(\"data\")",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatExample(tt.sig, tt.ex)
			if got != tt.want {
				t.Errorf("FormatExample() = %q, want %q", got, tt.want)
			}
		})
	}
}
