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

Create `main.go`:

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

### 2. Generate Bindings

```bash
gowasm-bindgen main.go --ts-output client.ts --go-output bindings_gen.go
```

This creates:
- `client.ts` - TypeScript client with full types
- `worker.js` - Web Worker for non-blocking calls
- `bindings_gen.go` - Go WASM wrapper functions

### 3. Build WASM

With TinyGo (smaller binary, ~90KB gzipped):
```bash
tinygo build -o example.wasm -target wasm .
```

Or with standard Go (~600KB gzipped):
```bash
GOOS=js GOARCH=wasm go build -o example.wasm .
```

### 4. Copy Runtime

```bash
# TinyGo
cp "$(tinygo env TINYGOROOT)/targets/wasm_exec.js" .

# Standard Go
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .
```

### 5. Use in TypeScript

```typescript
import { GoMain } from './client';

const wasm = await GoMain.init('./worker.js');

const greeting = await wasm.greet('World');
console.log(greeting);  // "Hello, World!"

const sum = await wasm.add(2, 3);
console.log(sum);  // 5

wasm.terminate();  // Clean up when done
```

## Project Structure

A typical project looks like:

```
my-project/
├── go/
│   ├── main.go           # Your Go code
│   └── bindings_gen.go   # Generated (gitignore)
├── src/
│   └── app.ts            # Your TypeScript code
├── dist/
│   ├── example.wasm      # Compiled WASM
│   ├── client.ts         # Generated TypeScript
│   ├── worker.js         # Generated Web Worker
│   └── wasm_exec.js      # Go runtime
└── Makefile              # Build automation
```

## Example Makefile

```makefile
.PHONY: build generate

generate:
	gowasm-bindgen go/main.go \
		--ts-output dist/client.ts \
		--go-output go/bindings_gen.go \
		--wasm-url example.wasm

build: generate
	cp "$$(tinygo env TINYGOROOT)/targets/wasm_exec.js" dist/
	tinygo build -o dist/example.wasm -target wasm -opt=z -no-debug ./go/
```

## Next Steps

- [For Go Developers]({{< relref "/docs/for-go-devs" >}}) - Learn about supported types, error handling, and callbacks
- [For TypeScript Developers]({{< relref "/docs/for-typescript-devs" >}}) - Using the generated client API
- [Live Demo]({{< relref "/examples/image-processing" >}}) - See WASM performance in action
