# gowasm-bindgen for Go Developers

Full-stack Go with type-safe JavaScript interop. Write Go, compile to WASM, get TypeScript declarations automatically.

## Why This Exists

Go WASM functions have a generic signature that erases type information:

```go
func myFunc(this js.Value, args []js.Value) interface{}
```

The `args` parameter is `[]js.Value` (untyped), and the return is `interface{}` (untyped). TypeScript sees these functions as `any`.

**gowasm-bindgen extracts types from your tests** to generate accurate `.d.ts` files:
- `js.ValueOf(stringVar)` reveals the parameter is a string
- `result.Int()` reveals the return type is a number
- Table struct fields like `username string` provide parameter names

If you have tests, your types are correct by definition.

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
tinygo build -o app.wasm -target wasm -opt=z -no-debug -panic=trap ./wasm/

# Standard Go (larger but full compatibility)
GOOS=js GOARCH=wasm go build -o app.wasm ./wasm/
```

## Writing Tests That Generate Good Types

### Table-Driven Tests (Recommended)

The struct field names become parameter names:

```go
func TestGreet(t *testing.T) {
    tests := []struct {
        name string  // Parameter name: "name"
        want string
    }{
        {name: "World", want: "Hello, World!"},
        {name: "Go", want: "Hello, Go!"},
    }
    for _, tt := range tests {
        result := greet(js.Null(), []js.Value{
            js.ValueOf(tt.name),
        })
        jsResult := result.(js.Value)
        if got := jsResult.String(); got != tt.want {
            t.Errorf("got %v, want %v", got, tt.want)
        }
    }
}
```

### Return Type Inference

gowasm-bindgen infers return types from how you access the result:

```go
jsResult := result.(js.Value)

jsResult.String()           // → string
jsResult.Int()              // → number
jsResult.Float()            // → number
jsResult.Bool()             // → boolean
jsResult.Get("field")       // → object with fields
jsResult.Index(0)           // → array
```

For object returns, access the fields in your test:

```go
jsResult := result.(js.Value)
displayName := jsResult.Get("displayName").String()  // → { displayName: string, ... }
status := jsResult.Get("status").String()            // → { ..., status: string }
```

### Parameter Names from Variables

If not using table-driven tests, variable names are used:

```go
func TestGreet(t *testing.T) {
    userName := "Alice"
    result := greet(js.Null(), []js.Value{
        js.ValueOf(userName),  // Parameter name: "userName"
    })
}
```

Literals fall back to `arg0`, `arg1`, etc:

```go
js.ValueOf("hello")  // Parameter name: "arg0"
js.ValueOf(42)       // Parameter name: "arg1"
```

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

## WASM Call Pattern

Tests must call functions with the standard Go WASM signature:

```go
func myFunc(this js.Value, args []js.Value) interface{}
```

And pass arguments using `js.ValueOf()`:

```go
result := myFunc(js.Null(), []js.Value{
    js.ValueOf(param1),
    js.ValueOf(param2),
})
```

## Error Handling

gowasm-bindgen validates patterns and fails fast with clear errors:

```
error: found 2 malformed WASM call pattern(s):
  wasm/broken_test.go:15: function has 3 arguments, expected exactly 2 (badFunc)
  wasm/broken_test.go:25: first argument is not js.Null() (anotherFunc)

Expected pattern:
  result := funcName(js.Null(), []js.Value{js.ValueOf(arg), ...})
```

### Detected Errors

| Error | Cause |
|-------|-------|
| `function has N arguments, expected exactly 2` | Wrong number of arguments |
| `first argument is not js.Null()` | Missing or wrong first argument |
| `second argument is not []js.Value{...}` | Args not passed as slice literal |
| `function is method/selector` | Using `pkg.Func()` instead of `Func()` |
| `call is not assigned to a variable` | Missing `result :=` assignment |
| `return type inferred as 'any'` | No result accessor methods found |

## Limitations

- **Only supports js.Value-based WASM patterns** (standard Go WASM signature)
- **Return type inference is heuristic-based** and may default to `any` for complex types
- **No support for function overloads** (Go doesn't support them either)
- **Complex nested types** (deeply nested objects, unions) may need manual refinement
- **No runtime validation** of generated types against actual WASM behavior

## Project Structure

```
your-project/
├── wasm/
│   ├── main.go           # Your WASM implementation
│   └── main_test.go      # Tests (parsed by gowasm-bindgen)
├── web/
│   └── app.ts            # TypeScript frontend
├── types.d.ts            # Generated (your function types)
├── wasm_exec.d.ts        # Generated (Go runtime types)
├── example.wasm          # Compiled WASM
└── wasm_exec.js          # From TinyGo or Go installation
```

## Complete Example

See the [example/](../example/) directory for a working demo with:
- 5 WASM functions with different parameter/return types
- Table-driven tests
- TinyGo build with size optimizations
- TypeScript web demo
- TypeScript verification tests

## FAQ

### Do I need my implementation in a `_test.go` file?

**No.** Your implementation lives in normal `.go` files. Only the test file needs to follow the pattern:

```go
// main.go - your implementation
func Greet(this js.Value, args []js.Value) interface{} {
    name := args[0].String()
    return js.ValueOf("Hello, " + name)
}

// main_test.go - parsed by gowasm-bindgen
func TestGreet(t *testing.T) {
    result := Greet(js.Null(), []js.Value{js.ValueOf("World")})
    // ...
}
```

### Why tests instead of annotations?

1. **Zero maintenance**: Types stay in sync with actual usage
2. **No build step changes**: Works with existing test infrastructure
3. **Self-documenting**: Tests show how functions are called
4. **Correctness by construction**: If tests pass, types are accurate
