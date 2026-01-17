# Limitations

gowasm-bindgen generates TypeScript declarations from Go source code. This document lists current limitations compared to Rust's [wasm-bindgen](https://rustwasm.github.io/docs/wasm-bindgen/) and potential future improvements.

## What's Been Resolved

✅ **No more awkward js.Value signatures** - Write normal Go functions
✅ **Direct type inference** - Types inferred from function signatures
✅ **Automatic bindings generation** - No manual `js.Global().Set()` calls
✅ **Automatic error throwing** - Go `(T, error)` returns automatically throw in TypeScript
✅ **Typed arrays** - Go `[]byte` maps to TypeScript `Uint8Array` with efficient bulk copy
✅ **Void callbacks** - Pass JavaScript functions as callback parameters (void only)

## Current Limitations

### Worker Mode by Default

Go WASM runs in a Web Worker by default, providing non-blocking async calls:

```typescript
// Default mode: non-blocking
import { Wasm } from './client';
const wasm = await Wasm.init('./worker.js');
const result = await wasm.heavyComputation(data);  // UI stays responsive!
```

**Want sync?** Use `--sync` flag to run on the main thread (blocks until complete):
```bash
gowasm-bindgen --output generated/client.ts --go-output go/bindings_gen.go --sync go/main.go
```

```typescript
// Sync mode: blocks main thread
import { Wasm } from './client';
const wasm = await Wasm.init('./example.wasm');
const result = wasm.heavyComputation(data);  // UI frozen!
```

### Typed Arrays

Go numeric slices map to TypeScript typed arrays:

| Go Type | TypeScript Type | Performance |
|---------|-----------------|-------------|
| `[]byte` / `[]uint8` | `Uint8Array` | Bulk copy (~10-100x faster) |
| `[]int8` | `Int8Array` | Element iteration |
| `[]int16` / `[]uint16` | `Int16Array` / `Uint16Array` | Element iteration |
| `[]int32` / `[]uint32` | `Int32Array` / `Uint32Array` | Element iteration |
| `[]float32` / `[]float64` | `Float32Array` / `Float64Array` | Element iteration |

```go
// Go function with typed arrays
func HashData(data []byte) []byte {
    hash := make([]byte, 4)
    for i, b := range data {
        hash[i%4] ^= b
    }
    return hash
}
```

```typescript
// TypeScript usage - native Uint8Array
const data = new Uint8Array([1, 2, 3, 4, 5]);
const hash = await wasm.hashData(data);  // Returns Uint8Array
```

**Performance note**: Byte arrays (`[]byte`) use `js.CopyBytesToGo()` and `js.CopyBytesToJS()` for efficient bulk copying. Other numeric types use element-by-element iteration since Go WASM doesn't provide bulk copy for non-byte types.

### Void Callbacks Only

Void callbacks (callbacks with no return value) are supported:

```go
// Go: void callback parameter
func ForEach(items []string, callback func(string, int)) {
    for i, item := range items {
        callback(item, i)
    }
}
```

```typescript
// TypeScript: callback type is inferred
await wasm.forEach(["a", "b", "c"], (item, index) => {
    console.log(`${index}: ${item}`);
});
```

**Limitations:**
- Callbacks must have no return value (void)
- Callbacks are called synchronously
- No nested callbacks (callback taking callback)
- If the TypeScript callback throws, it becomes a rejected Promise

**Not yet supported:**
```go
// Callbacks with return values - NOT supported
func Filter(items []string, predicate func(string) bool) []string

// Nested callbacks - NOT supported
func WithMiddleware(handler func(next func())) {}
```

### Class-Based but Not OOP

Functions are methods on a class instance, but not true object-oriented:

```typescript
// Current: methods on a WASM instance class
const wasm = await Wasm.init('./worker.js');
const user = await wasm.createUser("Alice", 30);
const name = await wasm.getUserName(user);

// Rust wasm-bindgen: exported structs become classes
const user = new User("Alice", 30);
const name = user.getName();
```

This is a fundamental difference in Go vs Rust WASM design. The class pattern in gowasm-bindgen is for API organization, not OOP.

### No JS→Go Imports

Can't get typed imports of JavaScript APIs in Go:

```rust
// Rust wasm-bindgen: typed DOM access
#[wasm_bindgen]
extern "C" {
    fn alert(s: &str);
}
```

```go
// Go: untyped js.Global() access
js.Global().Call("alert", "hello")
```

### Interface{} Becomes Any

Go functions that return `interface{}` become `any` in TypeScript:

```go
func GetValue() interface{} {
    return "hello"
}
// → TypeScript: getValue(): Promise<any>
```

**Mitigation:** Use concrete types whenever possible. Define structs for complex returns.

### No Integrated Build Toolchain

Rust has `wasm-pack` for a complete workflow. gowasm-bindgen is just the type generator:

| Rust | Go |
|------|-----|
| `wasm-pack build` | `tinygo build` + `gowasm-bindgen` + manual setup |

## Comparison with Rust wasm-bindgen

| Feature | Rust wasm-bindgen | gowasm-bindgen |
|---------|-------------------|----------------|
| Type source | `#[wasm_bindgen]` annotations | Inferred from source |
| Direction | Bidirectional (Rust↔JS) | Export only (Go→JS) |
| TypeScript generation | ✅ | ✅ |
| Primitives | ✅ | ✅ |
| Structs | ✅ Classes with methods | ✅ Interfaces |
| Typed arrays | ✅ | ✅ (bytes bulk copy, others element iteration) |
| Closures/callbacks | ✅ | ⚠️ Void callbacks only |
| Promises/async | ✅ | ✅ (default) |
| Error handling | ✅ Result<T,E> throws | ✅ (T, error) throws |
| JS imports | ✅ | ❌ |
| Build toolchain | ✅ wasm-pack | ❌ |
| Normal function syntax | ❌ Requires annotations | ✅ No annotations needed |

## Why Use gowasm-bindgen Anyway?

1. **You already know Go** — No need to learn Rust
2. **Existing Go code** — Share logic between backend and frontend
3. **Zero boilerplate** — Write normal Go functions, no annotations
4. **Type inference** — Types stay in sync automatically
5. **TinyGo binary size** — 90KB gzipped is competitive with Rust

## Future Roadmap

Potential improvements (contributions welcome):

- [x] Web Worker wrapper generation for async/Promise API (now default)
- [x] Class-based API instead of window globals
- [x] Source-based type inference (no tests required)
- [x] Automatic Go bindings generation
- [x] Automatic error throwing for `(T, error)` returns
- [x] Typed array detection and generation
- [x] Void callback support (callbacks with no return value)
- [ ] Callbacks with return values (`func(T) bool`, etc.)
- [ ] `wasm-pack`-style CLI for complete workflow

## References

- [Rust wasm-bindgen documentation](https://rustwasm.github.io/docs/wasm-bindgen/)
- [TinyGo WebAssembly](https://tinygo.org/docs/guides/webassembly/)
- [Go WebAssembly wiki](https://github.com/golang/go/wiki/WebAssembly)
