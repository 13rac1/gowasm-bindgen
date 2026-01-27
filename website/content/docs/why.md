---
title: "Why gowasm-bindgen?"
weight: 2
---

# Why gowasm-bindgen?

If you've tried using Go with WebAssembly, you've probably experienced the pain. This page explains what that pain is and how gowasm-bindgen solves it.

## What is Go WASM?

When you compile Go for the browser:

```bash
GOOS=js GOARCH=wasm go build -o main.wasm
```

You get a `.wasm` file—a **binary format**, not JavaScript. WebAssembly runs in the browser alongside JavaScript, but it can't directly interact with the DOM or JavaScript APIs.

To bridge Go and JavaScript, you need "glue code" on both sides.

## The Manual Approach (Without gowasm-bindgen)

Here's what you have to write **manually** for a simple `Greet` function:

### Go Side

```go
//go:build js && wasm

package main

import "syscall/js"

func main() {
    // Register function with JavaScript
    js.Global().Set("greet", js.FuncOf(greetWrapper))

    // Keep program alive
    select {}
}

func greetWrapper(this js.Value, args []js.Value) interface{} {
    // Manual argument validation
    if len(args) < 1 {
        return map[string]interface{}{"error": "missing name argument"}
    }

    // Manual type conversion (no compile-time safety!)
    name := args[0].String()  // What if args[0] isn't a string?

    // Call actual logic
    result := greet(name)

    return result
}

func greet(name string) string {
    return "Hello, " + name + "!"
}
```

### JavaScript Side

```javascript
// Load WASM (boilerplate)
const go = new Go();
const result = await WebAssembly.instantiateStreaming(
    fetch("main.wasm"),
    go.importObject
);
go.run(result.instance);

// Call the function
const greeting = window.greet("World");  // Works!
const oops = window.greet(123);          // No error... until runtime crash
const typo = window.greeet("World");     // No error... undefined!
```

### The Problems

1. **No type safety** - `args[0].String()` assumes the caller passed a string. If they pass a number, it crashes at runtime.

2. **No IDE support** - TypeScript has no idea `window.greet` exists. No autocomplete, no type checking.

3. **Boilerplate explosion** - Every function needs a wrapper with argument validation and type conversion.

4. **Easy to break** - Misspell `greet` as `greeet`? No error until you run it.

5. **Structs are painful** - Returning complex types requires manual `map[string]interface{}` construction.

## With gowasm-bindgen

Write normal Go:

```go
//go:build js && wasm

package main

// Greet returns a greeting for the given name.
func Greet(name string) string {
    return "Hello, " + name + "!"
}

func main() {
    select {}
}
```

Run gowasm-bindgen:

```bash
gowasm-bindgen wasm/main.go --output generated
```

Use typed TypeScript:

```typescript
import { GoWasm } from './generated/go-wasm';

const wasm = await GoWasm.init('./generated/worker.js');

const greeting = await wasm.greet("World");  // string
const oops = await wasm.greet(123);          // TypeScript Error!
const typo = await wasm.greeet("World");     // TypeScript Error!
```

### What Changed?

| Aspect | Manual | gowasm-bindgen |
|--------|--------|----------------|
| Go code | Wrappers + js.Value | Normal functions |
| Type safety | Runtime crashes | Compile-time errors |
| IDE support | None | Full autocomplete |
| Structs | Manual conversion | Automatic |
| Documentation | None | From Go comments |

## What Gets Generated

gowasm-bindgen creates three files:

### 1. `bindings_gen.go` - Go Glue Code

```go
//go:build js && wasm

package main

import "syscall/js"

func init() {
    js.Global().Set("greet", js.FuncOf(wasmGreet))
}

func wasmGreet(_ js.Value, args []js.Value) interface{} {
    name := args[0].String()
    result := Greet(name)
    return result
}
```

All the tedious `js.Value` handling, generated automatically.

### 2. `go-wasm.ts` - TypeScript Client

```typescript
export class GoWasm {
    static async init(workerUrl: string): Promise<GoWasm>;

    /** Greet returns a greeting for the given name. */
    greet(name: string): Promise<string>;
}
```

Full types, JSDoc from Go comments, IDE autocomplete.

### 3. `worker.js` - Web Worker (Optional)

Runs WASM in a background thread so long computations don't block the UI.

## The Bottom Line

**Without gowasm-bindgen:**
- You write boilerplate
- You debug runtime crashes
- You get no IDE help

**With gowasm-bindgen:**
- You write Go functions
- TypeScript catches errors at compile time
- Your IDE knows everything

---

Ready to try it? [Get Started]({{< relref "/docs/getting-started" >}})
