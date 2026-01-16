# Limitations

gowasm-bindgen generates TypeScript declarations from Go source code. This document lists current limitations compared to Rust's [wasm-bindgen](https://rustwasm.github.io/docs/wasm-bindgen/) and potential future improvements.

## What's Been Resolved

✅ **No more awkward js.Value signatures** - Write normal Go functions
✅ **Direct type inference** - Types inferred from function signatures
✅ **Automatic bindings generation** - No manual `js.Global().Set()` calls
✅ **Better error handling** - Return structs with error fields

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

### No Typed Arrays

Array parameters require string serialization:

```typescript
// Current: manual serialization
const sum = await wasm.sumNumbers("1,2,3,4,5");

// Ideal: native arrays
const sum = await wasm.sumNumbers([1, 2, 3, 4, 5]);
```

Go supports typed arrays via `js.CopyBytesToGo()`, but gowasm-bindgen doesn't detect this pattern yet.

### No Callbacks

Can't pass JavaScript functions to Go:

```typescript
// Not supported
await wasm.forEach(items, (item) => console.log(item));
```

Go does support callbacks via `js.FuncOf()`, but detection and type generation is complex.

### No Automatic Error Throwing

Go functions that return `(T, error)` are supported, but errors aren't automatically thrown:

```go
// Go code with error return
func ValidateEmail(email string) (EmailResult, error) {
    if !strings.Contains(email, "@") {
        return EmailResult{}, errors.New("invalid email")
    }
    return EmailResult{Valid: true}, nil
}
```

```typescript
// TypeScript: error returned as second value
const [result, err] = await wasm.validateEmail("bad");
if (err) {
  console.error(err);
}

// Rust wasm-bindgen: Result<T,E> becomes try/catch
try {
  const result = validateEmail("bad");
} catch (e) {
  console.error(e);
}
```

**Workaround:** Return a struct with error field instead:

```go
type EmailResult struct {
    Valid bool   `json:"valid"`
    Error string `json:"error"`
}

func ValidateEmail(email string) EmailResult {
    if !strings.Contains(email, "@") {
        return EmailResult{Valid: false, Error: "invalid email"}
    }
    return EmailResult{Valid: true}
}
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
| Typed arrays | ✅ | ❌ |
| Closures/callbacks | ✅ | ❌ |
| Promises/async | ✅ | ✅ (default) |
| Error handling | ✅ Result<T,E> throws | ⚠️ Manual (struct fields) |
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
- [ ] Typed array detection and generation
- [ ] Automatic error throwing for `(T, error)` returns
- [ ] `wasm-pack`-style CLI for complete workflow
- [ ] Callback/closure support via `js.FuncOf()` detection

## References

- [Rust wasm-bindgen documentation](https://rustwasm.github.io/docs/wasm-bindgen/)
- [TinyGo WebAssembly](https://tinygo.org/docs/guides/webassembly/)
- [Go WebAssembly wiki](https://github.com/golang/go/wiki/WebAssembly)
