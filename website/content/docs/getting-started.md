---
title: "Getting Started"
weight: 1
---

# Getting Started

Get type-safe Go in your browser in under 5 minutes.

## Prerequisites

- Go 1.21+
- [TinyGo](https://tinygo.org/getting-started/install/) (recommended) or standard Go
- Node.js 18+ (for TypeScript)

## Installation

```bash
go install github.com/13rac1/gowasm-bindgen@latest
```

## Quick Example

### 1. Write Go Functions

Create `wasm/main.go`:

```go
package main

// Greet returns a greeting message
func Greet(name string) string {
    return "Hello, " + name + "!"
}

// Add returns the sum of two numbers
func Add(a, b int) int {
    return a + b
}

func main() {
    // Keep the program running for WASM
    select {}
}
```

### 2. Build Everything

```bash
gowasm-bindgen wasm/main.go
```

This single command:
- Generates `generated/go-wasm.ts` - TypeScript client with full types
- Generates `generated/worker.js` - Web Worker for non-blocking calls
- Generates `wasm/bindings_gen.go` - Go WASM wrapper functions
- Copies `generated/wasm_exec.js` - Go runtime from TinyGo
- Compiles `generated/wasm.wasm` - Optimized WASM binary (~200KB)

### 3. Use in TypeScript

```typescript
import { GoWasm } from './generated/go-wasm';

const wasm = await GoWasm.init('./worker.js');

const greeting = await wasm.greet('World');
console.log(greeting);  // "Hello, World!"

const sum = await wasm.add(2, 3);
console.log(sum);  // 5

wasm.terminate();  // Clean up when done
```

### Build Options

```bash
# Generate only (skip WASM compilation)
gowasm-bindgen wasm/main.go --no-build

# Custom output directory
gowasm-bindgen wasm/main.go --output build/

# Use standard Go instead of TinyGo (larger binary, ~2.4MB)
gowasm-bindgen wasm/main.go --compiler go

# Synchronous mode (blocks main thread, no Web Worker)
gowasm-bindgen wasm/main.go --mode sync
```

## Project Structure

A typical project looks like:

```
my-project/
├── wasm/
│   ├── main.go           # Your Go code
│   └── bindings_gen.go   # Generated (gitignore)
├── src/
│   └── app.ts            # Your TypeScript code
├── generated/            # gowasm-bindgen output (gitignore)
│   ├── wasm.wasm         # Compiled WASM
│   ├── go-wasm.ts        # Generated TypeScript client
│   ├── worker.js         # Generated Web Worker
│   └── wasm_exec.js      # Go runtime
├── dist/                 # Final bundled output (gitignore)
│   └── app.js            # Bundled TypeScript
└── Makefile              # Build automation
```

## Example Makefile

```makefile
.PHONY: build generate clean

# Full build (default)
build:
	gowasm-bindgen wasm/main.go

# Generate only (no WASM compilation)
generate:
	gowasm-bindgen wasm/main.go --no-build

clean:
	rm -rf generated wasm/bindings_gen.go
```

## Next Steps

- [For Go Developers]({{< relref "/docs/for-go-devs" >}}) - Learn about supported types, error handling, and callbacks
- [For TypeScript Developers]({{< relref "/docs/for-typescript-devs" >}}) - Using the generated client API
- [Live Demo]({{< relref "/examples/image-processing" >}}) - See WASM performance in action
