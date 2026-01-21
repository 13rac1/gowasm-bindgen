---
title: "For Go Developers"
weight: 2
---

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

**With gowasm-bindgen, you write normal Go functions:**

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

## How It Works

gowasm-bindgen uses **source-based type inference** - it parses your Go source file's AST to extract:
- Exported function signatures (capitalized names, no receivers)
- Struct definitions with JSON tags
- Parameter names and types
- Return types including `(T, error)` patterns

No test files, annotations, or runtime analysis required. Just point it at your Go source file.

```bash
gowasm-bindgen main.go --output client.ts --go-output bindings_gen.go
```

## TinyGo vs Standard Go

| | Standard Go | TinyGo |
|---|-------------|--------|
| Binary size | ~2.4 MB | ~200 KB |
| Gzipped | ~600 KB | ~90 KB |
| Language support | Full | [Partial](https://tinygo.org/docs/reference/lang-support/) |
| Stdlib | Full | Partial |
| Reflection | Full | Limited |

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

```typescript
// Usage:
try {
    const result = await wasm.divide(10, 0);
} catch (e) {
    console.error(e.message);  // "division by zero"
}
```

### Typed Arrays

Go byte and numeric slices map to TypeScript typed arrays:

```go
// HashData processes binary data efficiently
func HashData(data []byte) []byte {
    hash := make([]byte, 4)
    for i, b := range data {
        hash[i%4] ^= b
    }
    return hash
}
```

**Performance note**: Byte arrays (`[]byte`) use `js.CopyBytesToGo()` and `js.CopyBytesToJS()` for efficient bulk copying (~10-100x faster for large arrays).

### Void Callbacks

Functions can accept callback parameters (void only):

```go
func ForEach(items []string, callback func(string, int)) {
    for i, item := range items {
        callback(item, i)
    }
}
```

```typescript
await wasm.forEach(["a", "b", "c"], (item, index) => {
    console.log(`${index}: ${item}`);
});
```

**Limitations:**
- Callbacks must have no return value (void)
- Callbacks are only valid during the Go function's execution
- No nested callbacks

## Type Mapping

| Go Type | TypeScript Type |
|---------|-----------------|
| `string` | `string` |
| `int`, `int64` | `number` |
| `float32`, `float64` | `number` |
| `bool` | `boolean` |
| `[]byte`, `[]uint8` | `Uint8Array` |
| `[]int32` | `Int32Array` |
| `[]float64` | `Float64Array` |
| `[]T` (other) | `T[]` |
| `map[string]T` | `{[key: string]: T}` |
| `func(T, U)` (void) | `(arg0: T, arg1: U) => void` |

## Generated Files

When you run gowasm-bindgen, it generates:

1. **TypeScript Client** (`client.ts`): Type-safe API for TypeScript
2. **Web Worker** (`worker.js`): Loads and runs WASM in background thread
3. **Go Bindings** (`bindings_gen.go`): Handles `js.Value` conversions automatically

**Important:** Add `bindings_gen.go` to your `.gitignore` - it's a build artifact.

## FAQ

### Do I need to write tests?

**No.** Tests are optional. gowasm-bindgen infers types directly from your function signatures. However, you should still write tests for correctness!

### Why source-based instead of annotations?

1. **Zero boilerplate**: No special annotations or tags needed
2. **Type safety**: Go compiler enforces correct types
3. **Normal Go code**: Write functions like you always do
4. **No learning curve**: If you know Go, you know how to use this
