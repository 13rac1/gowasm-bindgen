# Limitations

gowasm-bindgen generates TypeScript declarations from Go tests. This document lists current limitations compared to Rust's [wasm-bindgen](https://rustwasm.github.io/docs/wasm-bindgen/) and potential future improvements.

## Current Limitations

### Synchronous by Default

Go WASM functions block the main thread. Long-running operations freeze the UI.

```typescript
// Default mode: blocks until complete
const result = window.heavyComputation(data);  // UI frozen!
```

**Solution:** Use `--worker` flag to generate a Web Worker wrapper with Promise-based API:
```bash
gowasm-bindgen --tests "wasm/*_test.go" --output types.d.ts --worker
```

```typescript
// Worker mode: non-blocking
import { init, heavyComputation } from './client';
await init('./worker.js');
const result = await heavyComputation(data);  // UI stays responsive!
```

### No Typed Arrays

Array parameters require string serialization:

```typescript
// Current: manual serialization
const sum = window.sumNumbers("1,2,3,4,5");

// Ideal: native arrays
const sum = window.sumNumbers([1, 2, 3, 4, 5]);
```

Go supports typed arrays via `js.CopyBytesToGo()`, but gowasm-bindgen doesn't detect this pattern yet.

### No Callbacks

Can't pass JavaScript functions to Go:

```typescript
// Not supported
window.forEach(items, (item) => console.log(item));
```

Go does support callbacks via `js.FuncOf()`, but detection and type generation is complex.

### No Error Mapping

Errors are returned as values, not thrown:

```typescript
// Current: check result manually
const result = window.validateEmail("bad");
if (!result.valid) {
  console.error(result.error);
}

// Rust wasm-bindgen: Result<T,E> becomes try/catch
try {
  const result = validateEmail("bad");
} catch (e) {
  console.error(e);
}
```

### No Class/Method Support

Go WASM exposes functions on `window`, not classes:

```typescript
// Current: standalone functions
const user = window.createUser("Alice", 30);
const name = window.getUserName(user);

// Rust wasm-bindgen: exported structs become classes
const user = new User("Alice", 30);
const name = user.getName();
```

This is a fundamental difference in Go vs Rust WASM design.

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

### Heuristic-Based Inference

Types are inferred from test code. If tests don't exercise all return paths, types may be incomplete:

```go
// If your test only checks the valid case...
result := validateEmail(js.Null(), []js.Value{js.ValueOf("user@example.com")})
jsResult := result.(js.Value)
valid := jsResult.Get("valid").Bool()  // infers: { valid: boolean }

// ...but doesn't check the error field on invalid input,
// the generated type might miss the 'error' field
```

**Mitigation:** Write comprehensive tests that access all fields of returned objects.

### No Integrated Build Toolchain

Rust has `wasm-pack` for a complete workflow. gowasm-bindgen is just the type generator:

| Rust | Go |
|------|-----|
| `wasm-pack build` | `tinygo build` + `gowasm-bindgen` + manual setup |

## Comparison with Rust wasm-bindgen

| Feature | Rust wasm-bindgen | gowasm-bindgen |
|---------|-------------------|----------------|
| Type source | `#[wasm_bindgen]` annotations | Inferred from tests |
| Direction | Bidirectional (Rust↔JS) | Export only (Go→JS) |
| TypeScript generation | ✅ | ✅ |
| Primitives | ✅ | ✅ |
| Objects | ✅ Classes with methods | ✅ Plain objects |
| Typed arrays | ✅ | ❌ |
| Closures/callbacks | ✅ | ❌ |
| Promises/async | ✅ | ✅ (`--worker`) |
| Error handling | ✅ Result<T,E> | ❌ |
| JS imports | ✅ | ❌ |
| Build toolchain | ✅ wasm-pack | ❌ |

## Why Use gowasm-bindgen Anyway?

1. **You already know Go** — No need to learn Rust
2. **Existing Go code** — Share logic between backend and frontend
3. **Test-driven types** — If your tests pass, your types are correct
4. **No annotations** — Types stay in sync automatically
5. **TinyGo binary size** — 90KB gzipped is competitive with Rust

## Future Roadmap

Potential improvements (contributions welcome):

- [x] Web Worker wrapper generation for async/Promise API (`--worker` flag)
- [ ] Typed array detection and generation
- [ ] Error/Result pattern detection
- [ ] `wasm-pack`-style CLI for complete workflow
- [ ] Callback/closure support via `js.FuncOf()` detection

## References

- [Rust wasm-bindgen documentation](https://rustwasm.github.io/docs/wasm-bindgen/)
- [TinyGo WebAssembly](https://tinygo.org/docs/guides/webassembly/)
- [Go WebAssembly wiki](https://github.com/golang/go/wiki/WebAssembly)
