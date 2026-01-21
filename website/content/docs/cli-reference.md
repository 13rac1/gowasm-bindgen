---
title: "CLI Reference"
weight: 10
---

# CLI Reference

## Usage

```
gowasm-bindgen <source.go> --output <file> [options]
```

## Required Flags

| Flag | Description |
|------|-------------|
| `--output FILE` | TypeScript client output path |

## Optional Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--go-output FILE` | (none) | Go bindings output path |
| `--sync` | false | Generate synchronous API (default: worker mode) |
| `--wasm-path PATH` | `module.wasm` | WASM file path in worker.js |
| `--verbose` | false | Enable debug output to stderr |

## Examples

### Worker Mode (Default)

Generates async, non-blocking API using Web Workers:

```bash
gowasm-bindgen main.go --output client.ts --go-output bindings_gen.go
```

Creates:
- `client.ts` - TypeScript client with Promise-based API
- `worker.js` - Web Worker entry point
- `bindings_gen.go` - Go WASM wrapper functions

### Sync Mode

Generates synchronous API that runs on main thread:

```bash
gowasm-bindgen main.go --output client.ts --go-output bindings_gen.go --sync
```

Creates:
- `client.ts` - TypeScript client with synchronous methods
- `bindings_gen.go` - Go WASM wrapper functions

No `worker.js` is generated in sync mode.

### Custom WASM Path

For monorepos or CDN deployment:

```bash
gowasm-bindgen main.go --output client.ts --wasm-path dist/app.wasm
```

The generated `worker.js` will load WASM from `dist/app.wasm` instead of `module.wasm`.

### Debug Output

Troubleshoot generation issues:

```bash
gowasm-bindgen main.go --output client.ts --verbose
```

Outputs to stderr:
```
[DEBUG] Source file: main.go
[DEBUG] Package: main
[DEBUG] Found 5 exported function(s):
  - Greet
  - Add
  - FormatUser
  ...
```

## Output Files

### client.ts

The generated TypeScript client exports a class named after your Go package:

```typescript
// Worker mode
export class Main {
  static async init(workerUrl: string): Promise<Main>;
  greet(name: string): Promise<string>;
  terminate(): void;
}

// Sync mode
export class Main {
  static async init(wasmSource: string | BufferSource): Promise<Main>;
  greet(name: string): string;  // No Promise
}
```

### worker.js

Web Worker script that loads and runs WASM (worker mode only):

```javascript
importScripts('./wasm_exec.js');
const go = new Go();
fetch('./module.wasm')  // or custom --wasm-path
  .then(response => response.arrayBuffer())
  .then(bytes => WebAssembly.instantiate(bytes, go.importObject))
  .then(result => { go.run(result.instance); });
// ... message handling
```

### bindings_gen.go

Go WASM wrapper functions that register your exports:

```go
//go:build js && wasm
package main

import "syscall/js"

func init() {
    js.Global().Set("greet", js.FuncOf(wasmGreet))
    // ...
}
```

## Build Workflow

Typical build sequence:

```bash
# 1. Generate bindings
gowasm-bindgen main.go --output dist/client.ts --go-output bindings_gen.go

# 2. Copy wasm_exec.js (TinyGo)
cp "$(tinygo env TINYGOROOT)/targets/wasm_exec.js" dist/

# 3. Build WASM
tinygo build -o dist/main.wasm -target wasm .

# 4. Build TypeScript
npx tsc
```

Or use a Makefile - see [Getting Started]({{< relref "/docs/getting-started" >}}).
