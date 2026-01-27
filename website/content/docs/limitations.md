---
title: "Limitations"
weight: 20
---

# Limitations

Current limitations compared to Rust's [wasm-bindgen](https://rustwasm.github.io/docs/wasm-bindgen/).

## What Works

- ✅ Normal Go function signatures (no `js.Value` boilerplate)
- ✅ Direct type inference from source code
- ✅ Automatic Go bindings generation
- ✅ Automatic error throwing for `(T, error)` returns
- ✅ Typed arrays (`[]byte` → `Uint8Array`)
- ✅ Void callbacks in both worker and sync modes
- ✅ Web Worker mode for non-blocking calls
- ✅ Node.js support via `BufferSource`

## Current Limitations

### Callbacks with Return Values

Only void callbacks are supported:

```go
// ✅ Supported - void callback
func ForEach(items []string, callback func(string, int)) { ... }

// ❌ Not supported - callback returns bool
func Filter(items []string, predicate func(string) bool) []string { ... }
```

### No JS→Go Imports

Can't get typed imports of JavaScript APIs in Go:

```go
// Go: untyped js.Global() access only
js.Global().Call("alert", "hello")
```

Compare to Rust wasm-bindgen which supports typed extern declarations.

### interface{} Becomes Any

Generic returns lose type information:

```go
func GetValue() interface{} { return "hello" }
// → TypeScript: getValue(): Promise<any>
```

**Mitigation**: Use concrete types whenever possible.

### Class-Based but Not OOP

Functions are methods on a class instance, not true object-oriented:

```typescript
// gowasm-bindgen: methods on WASM instance
const wasm = await GoWasm.init('./worker.js');
const user = await wasm.createUser("Alice", 30);

// Rust wasm-bindgen: exported structs become classes
const user = new User("Alice", 30);
```

### No Integrated Build Toolchain

gowasm-bindgen generates bindings only. You still need to:
- Run `tinygo build` or `go build` separately
- Copy `wasm_exec.js` manually
- Set up your own build pipeline

## Comparison with Rust wasm-bindgen

| Feature | Rust wasm-bindgen | gowasm-bindgen |
|---------|-------------------|----------------|
| Type source | `#[wasm_bindgen]` annotations | Inferred from source |
| Direction | Bidirectional (Rust↔JS) | Export only (Go→JS) |
| TypeScript generation | ✅ | ✅ |
| Structs | Classes with methods | Interfaces |
| Closures/callbacks | ✅ Full support | ⚠️ Void only |
| Error handling | ✅ `Result<T,E>` | ✅ `(T, error)` |
| JS imports | ✅ | ❌ |
| Build toolchain | ✅ wasm-pack | ❌ |
| Annotations required | Yes | **No** |

## Why Use gowasm-bindgen Anyway?

1. **You already know Go** — No need to learn Rust
2. **Existing Go code** — Share logic between backend and frontend
3. **Zero boilerplate** — Write normal Go functions
4. **Type inference** — Types stay in sync automatically
5. **TinyGo binary size** — 90KB gzipped is competitive with Rust

## Future Roadmap

- [ ] Callbacks with return values (`func(T) bool`)
- [ ] Integrated build command

Contributions welcome!
