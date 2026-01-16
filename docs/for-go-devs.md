# gowasm-bindgen for Go Developers

Full-stack Go with type-safe JavaScript interop. Write Go, compile to WASM, get TypeScript class APIs automatically.

## Why This Exists

Go WASM functions traditionally required awkward `js.Value` signatures:

```go
// Old way - awkward js.Value signatures
func myFunc(this js.Value, args []js.Value) interface{} {
    name := args[0].String()
    return js.ValueOf("Hello, " + name)
}
```

The `args` parameter is `[]js.Value` (untyped), and the return is `interface{}` (untyped). TypeScript sees these functions as `any`.

**With the new gowasm-bindgen, you write normal Go functions:**

```go
// New way - normal Go functions
func MyFunc(name string) string {
    return "Hello, " + name
}
```

gowasm-bindgen reads your Go source code, infers types from function signatures, and generates:
1. TypeScript client with proper types (`myFunc(name: string): Promise<string>`)
2. Go WASM bindings that handle the `js.Value` conversions automatically

Your package name becomes the TypeScript class name (e.g., `package main` → `Main` class).

## TinyGo vs Standard Go

| | Standard Go | TinyGo |
|---|-------------|--------|
| Binary size | ~2.4 MB | ~200 KB |
| Gzipped | ~600 KB | ~90 KB |
| Language support | Full | [Partial](https://tinygo.org/docs/reference/lang-support/) |
| Stdlib | Full | Partial |
| Reflection | Full | Limited |

**TinyGo Limitations**: TinyGo doesn't support all Go features. Notable gaps include:
- `reflect.Value.Call()` and `reflect.MakeFunc()`
- Some `encoding/json` edge cases
- `go:linkname` directives
- Three-index slicing (`a[1:2:3]`)

See the [TinyGo Language Support](https://tinygo.org/docs/reference/lang-support/) page for details.

**When to use Standard Go**: If your code uses unsupported features, or you need full `reflect` capabilities for JSON marshaling complex types, use standard Go and accept the larger binary.

```bash
# TinyGo (smaller binaries, some limitations)
tinygo build -o dist/example.wasm -target wasm -opt=z -no-debug -panic=trap ./go/

# Standard Go (larger but full compatibility)
GOOS=js GOARCH=wasm go build -o dist/example.wasm ./go/
```

## Writing Go Functions for WASM

### Basic Functions

Write normal Go functions with concrete types:

```go
package main

// Greet returns a greeting message
func Greet(name string) string {
    return "Hello, " + name + "!"
}

// Calculate performs arithmetic
func Calculate(a int, b int, op string) int {
    switch op {
    case "add":
        return a + b
    case "sub":
        return a - b
    default:
        return 0
    }
}
```

**Requirements:**
- Functions must be **exported** (start with uppercase letter)
- Functions must be **package-level** (no receivers)
- Use **concrete types** (avoid `interface{}` when possible)

### Struct Returns

Define structs with JSON tags for TypeScript interfaces:

```go
type User struct {
    DisplayName string `json:"displayName"`
    Status      string `json:"status"`
}

func FormatUser(name string, age int, active bool) User {
    status := "inactive"
    if active {
        status = "active"
    }
    return User{
        DisplayName: fmt.Sprintf("%s (%d)", name, age),
        Status:      status,
    }
}
```

This generates:

```typescript
interface User {
    displayName: string;
    status: string;
}

class Main {
    formatUser(name: string, age: number, active: boolean): Promise<User>;
}
```

### Error Returns

Functions that return `(T, error)` automatically throw in TypeScript:

```go
func Divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}
```

This generates:

```typescript
class Main {
    // Throws on error, returns number on success
    divide(a: number, b: number): Promise<number>;
}

// Usage:
try {
    const result = await wasm.divide(10, 0);
} catch (e) {
    console.error(e.message);  // "division by zero"
}
```

### Testing (Optional)

Tests are no longer required for type generation, but you should still write them:

```go
func TestGreet(t *testing.T) {
    got := Greet("World")
    want := "Hello, World!"
    if got != want {
        t.Errorf("Greet() = %v, want %v", got, want)
    }
}
```

These are normal unit tests - no `js.Value` required!

## Type Mapping

| Go Type | TypeScript Type |
|---------|-----------------|
| `string` | `string` |
| `int`, `int8`, `int16`, `int32`, `int64` | `number` |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64` | `number` |
| `float32`, `float64` | `number` |
| `bool` | `boolean` |
| `[]T` | `T[]` |
| `map[string]T` | `{[key: string]: T}` |
| Unknown | `any` |

## Generated Files

When you run gowasm-bindgen, it generates:

1. **TypeScript Client** (`client.ts`): Type-safe API for TypeScript
2. **Web Worker** (`worker.js`): Loads and runs WASM in background thread
3. **Go Bindings** (`bindings_gen.go`): Handles `js.Value` conversions automatically

The Go bindings file registers your functions and handles all the `js.Value` marshaling:

```go
// bindings_gen.go (generated)
package main

import "syscall/js"

func init() {
    js.Global().Set("greet", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
        // Automatic conversion from js.Value to Go types
        name := args[0].String()
        result := Greet(name)
        // Automatic conversion from Go types to js.Value
        return js.ValueOf(result)
    }))
}
```

**Important:** Add `bindings_gen.go` to your `.gitignore` - it's a build artifact.

## Workflow

1. **Write Go functions** with normal signatures:
   ```bash
   # go/main.go
   func Greet(name string) string { ... }
   ```

2. **Generate bindings**:
   ```bash
   gowasm-bindgen --output generated/client.ts --go-output go/bindings_gen.go go/main.go
   ```

3. **Build WASM** (include the generated bindings):
   ```bash
   tinygo build -o dist/example.wasm -target wasm ./go/
   ```

4. **Use in TypeScript**:
   ```typescript
   import { Main } from './client';
   const wasm = await Main.init('./worker.js');
   const result = await wasm.greet('World');
   ```

## Limitations

- **Exported functions only**: Only package-level exported functions are available
- **Concrete types**: `interface{}` returns become `any` in TypeScript
- **No function overloads**: Go doesn't support them either
- **Struct field tags**: Use JSON tags for TypeScript-friendly field names
- **No runtime validation**: Generated types don't validate at runtime

## Generated API

gowasm-bindgen generates a TypeScript class with your package name:

```typescript
// Generated client.ts from package main
export class Main {
  static async init(workerUrl: string): Promise<Main>;
  greet(name: string): Promise<string>;
  calculate(a: number, b: number, op: string): Promise<number>;
  terminate(): void;
}
```

Your TypeScript users import and use it:

```typescript
import { Main } from './generated/client';

const wasm = await Main.init('./worker.js');
const result = await wasm.greet('World');
wasm.terminate();
```

## Project Structure

```
your-project/
├── go/                   # All Go code
│   ├── main.go           # Your WASM implementation
│   ├── main_test.go      # Optional unit tests
│   └── bindings_gen.go   # Generated Go bindings (gitignored)
├── src/                  # TypeScript source
│   └── app.ts            # TypeScript frontend
├── public/               # Static assets
│   └── index.html
├── generated/            # Generated TS/JS (gitignored)
│   ├── client.ts         # Generated TypeScript class API
│   └── worker.js         # Generated Web Worker wrapper
└── dist/                 # Build output (gitignored)
    └── example.wasm      # Compiled WASM
```

## Complete Example

See the [example/](../example/) directory for a working demo with:
- 5 WASM functions with different parameter/return types
- Normal Go functions (no `js.Value` signatures)
- Unit tests
- TinyGo build with size optimizations
- TypeScript web demo
- TypeScript verification tests

## FAQ

### Do I need to write tests?

**No.** Tests are optional. gowasm-bindgen infers types directly from your function signatures:

```go
// No test needed - type inference is automatic
func Greet(name string) string {
    return "Hello, " + name
}
// → TypeScript: greet(name: string): Promise<string>
```

However, you should still write tests for correctness!

### Why source-based instead of annotations?

1. **Zero boilerplate**: No special annotations or tags needed
2. **Type safety**: Go compiler enforces correct types
3. **Normal Go code**: Write functions like you always do
4. **No learning curve**: If you know Go, you know how to use this
