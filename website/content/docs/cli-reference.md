---
title: "CLI Reference"
weight: 10
---

# CLI Reference

## Usage

```
gowasm-bindgen <source.go> [options]
```

By default, gowasm-bindgen generates bindings, copies the runtime, and compiles WASM in one step.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output DIR` | `generated` | Output directory for all artifacts |
| `--no-build` | false | Skip WASM compilation (generate only) |
| `--compiler NAME` | `tinygo` | Compiler: `tinygo` or `go` |
| `-m, --mode MODE` | `worker` | Generation mode: `sync` or `worker` |
| `-c, --class-name NAME` | `Go<DirName>` | TypeScript class name |
| `--optimize` | true | Enable size optimizations (tinygo only) |
| `-v, --verbose` | false | Enable debug output to stderr |

## Examples

### Full Build (Default)

Generates bindings, copies runtime, and compiles WASM:

```bash
gowasm-bindgen wasm/main.go
```

Creates:
- `generated/go-wasm.ts` - TypeScript client (kebab-case from class name)
- `generated/worker.js` - Web Worker entry point
- `generated/wasm_exec.js` - Go runtime (copied from TinyGo)
- `generated/wasm.wasm` - Compiled WASM binary
- `wasm/bindings_gen.go` - Go WASM wrapper functions

### Generate Only

Skip WASM compilation (useful for CI or custom build pipelines):

```bash
gowasm-bindgen wasm/main.go --no-build
```

### Custom Output Directory

```bash
gowasm-bindgen wasm/main.go --output build/
```

### Standard Go Compiler

For larger binary with full Go compatibility:

```bash
gowasm-bindgen wasm/main.go --compiler go
```

### Sync Mode

Generates synchronous API that runs on main thread (blocks UI):

```bash
gowasm-bindgen wasm/main.go --mode sync
```

Creates:
- `generated/go-wasm.ts` - TypeScript client with synchronous methods
- `wasm/bindings_gen.go` - Go WASM wrapper functions

No `worker.js` is generated in sync mode.

### Custom Class Name

The default class name is derived from the directory: `Go` + TitleCase(dirname).

| Directory | Default Class Name | TypeScript File |
|-----------|-------------------|-----------------|
| `wasm/` | `GoWasm` | `go-wasm.ts` |
| `image-wasm/` | `GoImageWasm` | `go-image-wasm.ts` |
| `go/` | `Go` | `go.ts` |

Override with `--class-name`:

```bash
gowasm-bindgen wasm/main.go --class-name ImageProcessor
# Creates: generated/image-processor.ts with class ImageProcessor
```

### Debug Output

Troubleshoot generation issues:

```bash
gowasm-bindgen main.go --verbose
```

## Output Files

### TypeScript Client

The generated TypeScript client exports a class:

```typescript
// Worker mode (from wasm/ directory)
export class GoWasm {
  static async init(workerUrl: string): Promise<GoWasm>;
  greet(name: string): Promise<string>;
  terminate(): void;
}

// Sync mode
export class GoWasm {
  static async init(wasmSource: string | BufferSource): Promise<GoWasm>;
  greet(name: string): string;  // No Promise
}
```

### worker.js

Web Worker script that loads and runs WASM (worker mode only).

### bindings_gen.go

Go WASM wrapper functions with `//go:build js && wasm` tag:

```go
//go:build js && wasm
package main

import "syscall/js"

func init() {
    js.Global().Set("greet", js.FuncOf(wasmGreet))
    // ...
}
```

### wasm_exec.js

Go runtime copied from your TinyGo or Go installation.

## Build Workflow

With the default settings, a single command does everything:

```bash
gowasm-bindgen wasm/main.go
```

For more control, use `--no-build` and run compilation separately:

```bash
# Generate bindings only
gowasm-bindgen wasm/main.go --no-build

# Compile with custom flags
tinygo build -o generated/wasm.wasm -target wasm -opt=2 ./wasm/
```
