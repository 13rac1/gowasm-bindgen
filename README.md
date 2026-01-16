# gowasm-bindgen

**Type-safe Go in the browser.**

Generate TypeScript declarations from your Go tests. Ship 90KB gzipped binaries with TinyGo.

## The Problem

Go WASM functions are invisible to TypeScript:

```typescript
// TypeScript has no idea what this returns or accepts
const result = window.myGoFunction(???, ???);  // any
```

And standard Go WASM binaries are huge (~2.4MB).

## The Solution

**gowasm-bindgen** extracts types from your existing Go tests:

```go
// Your Go test
func TestGreet(t *testing.T) {
    tests := []struct {
        name string  // ← parameter name
        want string
    }{
        {name: "World", want: "Hello, World!"},
    }
    for _, tt := range tests {
        result := greet(js.Null(), []js.Value{
            js.ValueOf(tt.name),  // ← type: string
        })
        jsResult := result.(js.Value)
        got := jsResult.String()  // ← return type: string
        // ...
    }
}
```

Generates `types.d.ts` and `wasm_exec.d.ts`:

```typescript
// types.d.ts - your function types
declare global {
  interface Window {
    greet(name: string): string;
  }
  var greet: (name: string) => string;
}

// wasm_exec.d.ts - Go runtime types (also generated)
declare class Go {
  constructor();
  importObject: WebAssembly.Imports;
  run(instance: WebAssembly.Instance): Promise<number>;
}
```

With TinyGo, your WASM binary drops from 2.4MB to **200KB (90KB gzipped)**.

## Quick Start

```bash
# Install
go install github.com/13rac1/gowasm-bindgen/cmd/gowasm-bindgen@latest

# Generate types from your tests
gowasm-bindgen --tests "wasm/*_test.go" --output types.d.ts
# Creates: types.d.ts (your functions) + wasm_exec.d.ts (Go runtime)
```

## Get Started

Choose your path:

- **[For TypeScript Developers](docs/for-typescript-devs.md)** — Run Go in your frontend in 10 minutes
- **[For Go Developers](docs/for-go-devs.md)** — Full-stack Go with proper JS types

## See It Working

The [example/](example/) directory has a complete demo:

```bash
cd example
make all    # Build WASM, generate types, verify, compile TypeScript
make serve  # Open http://localhost:8080/web/
```

## How It Works

1. Parse your Go test files
2. Find WASM function calls (`funcName(js.Null(), []js.Value{...})`)
3. Extract parameter names from struct fields or variable names
4. Extract parameter types from `js.ValueOf()` calls
5. Infer return types from result accessors (`.String()`, `.Int()`, `.Bool()`, `.Get()`)
6. Generate `types.d.ts` with proper TypeScript declarations
7. Generate `wasm_exec.d.ts` with Go runtime types

No annotations. No build plugins. Just tests.

## Requirements

- Go 1.21+
- [TinyGo](https://tinygo.org/getting-started/install/) (recommended for small binaries) or standard Go
- Your WASM functions must use the standard signature: `func(js.Value, []js.Value) interface{}`

**Note**: TinyGo produces much smaller binaries but has [language limitations](https://tinygo.org/docs/reference/lang-support/). Use standard Go if you need full reflection or unsupported features.

## License

MIT
