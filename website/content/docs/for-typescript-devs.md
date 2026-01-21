---
title: "For TypeScript Developers"
weight: 3
---

# gowasm-bindgen for TypeScript Developers

You know TypeScript. You're curious about Go. Here's how to run Go code in your browser with full type safety.

## Why Go in the Browser?

- **Performance**: Go compiles to WebAssembly, running at near-native speed
- **Shared logic**: Use the same code on your backend and frontend
- **Type safety**: With gowasm-bindgen, your Go functions get proper TypeScript types
- **Non-blocking**: Web Worker mode keeps your UI responsive

## What You Need to Know

Your Go teammate writes normal Go functions with standard types. gowasm-bindgen reads the source code and generates TypeScript bindings automatically. You import the generated `client.ts` class. That's it.

**The files you care about:**

| File | What it is |
|------|------------|
| `example.wasm` | The compiled Go code (runs in Web Worker) |
| `worker.js` | Generated Web Worker script (loads and runs WASM) |
| `client.ts` | Generated TypeScript class (your API with full types) |

## Using the Generated Class API

### 1. Import and initialize

```typescript
import { Main } from './generated/client';

// Initialize with the Web Worker URL
const wasm = await Main.init('./worker.js');
```

### 2. Call Go functions with full type checking

```typescript
// TypeScript knows: greet(name: string): Promise<string>
const greeting = await wasm.greet("World");
console.log(greeting);  // "Hello, World!"

// TypeScript knows: calculate(a: number, b: number, op: string): Promise<number>
const sum = await wasm.calculate(5, 3, "add");
console.log(sum);  // 8

// TypeScript knows the return type
const user = await wasm.formatUser("Alice", 30, true);
console.log(user.displayName);  // "Alice (30)"
```

### 3. Handle errors with try/catch

Go functions that return `(T, error)` automatically throw in TypeScript:

```typescript
try {
  const result = await wasm.divide(10, 0);
} catch (e) {
  console.error(e.message);  // "division by zero"
}
```

### 4. Clean up when done

```typescript
wasm.terminate();
```

### 5. TypeScript catches your mistakes

```typescript
// Error: Argument of type 'number' is not assignable to parameter of type 'string'
await wasm.greet(42);

// Error: Expected 3 arguments, but got 2
await wasm.calculate(5, 3);
```

## Troubleshooting

### "Cannot find module './client'"

Make sure you've generated the TypeScript client:

```bash
gowasm-bindgen --output generated/client.ts --go-output go/bindings_gen.go go/main.go
```

### Worker fails to load

- Make sure `worker.js` and `example.wasm` are in the correct location
- Check that your bundler copies these files to your output directory
- Verify the worker URL path is correct (relative to your HTML page)

### Want synchronous calls?

Use the `--sync` flag:

```bash
gowasm-bindgen --output generated/client.ts --go-output go/bindings_gen.go --sync go/main.go
```

```typescript
// Sync mode - no await needed
const wasm = await Main.init('./example.wasm');
const greeting = wasm.greet('World');  // sync call
```

### Using in Node.js

The sync mode `init()` accepts either a URL string or a `BufferSource`:

```typescript
import { readFileSync } from 'fs';
import { Main } from './generated/client.js';
import './wasm_exec.js';

const wasmBytes = readFileSync('./dist/example.wasm');
const wasm = await Main.init(wasmBytes);

const result = wasm.greet('Node.js');
```

## Common Gotchas

### 1. Worker Mode Is Async

By default, all calls return Promises:

```typescript
// ✅ Correct - await the Promise
const result = await wasm.greet("World");

// ❌ Wrong - forgot await
const result = wasm.greet("World");
console.log(result);  // Promise { <pending> }
```

### 2. Always Await init()

Module initialization is always async, even in sync mode:

```typescript
// ❌ Wrong
const wasm = Main.init('./worker.js');
await wasm.greet("World");  // Error: wasm is a Promise!

// ✅ Correct
const wasm = await Main.init('./worker.js');
await wasm.greet("World");
```

### 3. Binary Data Uses Typed Arrays

Go `[]byte` functions accept and return `Uint8Array`:

```typescript
const data = new Uint8Array([1, 2, 3, 4, 5]);
const hash = await wasm.hashData(data);  // Returns Uint8Array
```

| Go Type | TypeScript Type |
|---------|-----------------|
| `[]byte` | `Uint8Array` |
| `[]int32` | `Int32Array` |
| `[]float64` | `Float64Array` |

### 4. Void Callbacks

Callbacks work in both worker and sync modes:

```typescript
await wasm.forEach(["a", "b", "c"], (item, index) => {
    console.log(`${index}: ${item}`);
});
```

**Limitations:**
- Callbacks must have no return value (void)
- In worker mode, callback errors are logged but cannot propagate to Go

## Next Steps

- [Live Demo]({{< relref "/examples/image-processing" >}}) - See WASM performance in action
- [For Go Developers]({{< relref "/docs/for-go-devs" >}}) - Understand how the Go side works
