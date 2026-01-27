# gowasm-bindgen

<p align="center">
  <img src="website/static/images/gowasm-bindgen-logo.png" alt="gowasm-bindgen logo" width="400">
</p>

[![CI](https://github.com/13rac1/gowasm-bindgen/actions/workflows/ci.yml/badge.svg)](https://github.com/13rac1/gowasm-bindgen/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/13rac1/gowasm-bindgen/graph/badge.svg)](https://codecov.io/gh/13rac1/gowasm-bindgen)
[![Go Report Card](https://goreportcard.com/badge/github.com/13rac1/gowasm-bindgen)](https://goreportcard.com/report/github.com/13rac1/gowasm-bindgen)
[![Go Reference](https://pkg.go.dev/badge/github.com/13rac1/gowasm-bindgen.svg)](https://pkg.go.dev/github.com/13rac1/gowasm-bindgen)
[![GitHub release](https://img.shields.io/github/v/release/13rac1/gowasm-bindgen)](https://github.com/13rac1/gowasm-bindgen/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Type-safe Go in the browser.**

**[Try the Image Processing Demo](https://13rac1.github.io/gowasm-bindgen/examples/image-processing/)** | **[Try the JS Sandbox](https://13rac1.github.io/gowasm-bindgen/examples/js-sandbox/)**

Generate TypeScript declarations and Go WASM bindings from your Go source code. Ship 90KB gzipped binaries with TinyGo.

## The Problem

Go WASM functions are invisible to TypeScript:

```typescript
// TypeScript has no idea what this returns or accepts
const result = window.myGoFunction(???, ???);  // any
```

Standard Go WASM binaries are huge (~2.4MB), and WASM runs synchronously on the main thread, blocking your UI.

[Learn more about the problem →](https://13rac1.github.io/gowasm-bindgen/docs/why/)

## The Solution

**gowasm-bindgen** generates bindings from your normal Go functions:

```go
// wasm/main.go - Write normal Go functions
func Greet(name string) string {
    return "Hello, " + name + "!"
}
```

Generates TypeScript with full type safety:

```typescript
// generated/go-wasm.ts
export class GoWasm {
  static async init(workerUrl: string): Promise<GoWasm>;
  greet(name: string): Promise<string>;
  terminate(): void;
}
```

With TinyGo, your WASM binary drops from 2.4MB to **200KB (90KB gzipped)**, and Web Workers keep your UI responsive.

## Quick Start

```bash
# Install
go install github.com/13rac1/gowasm-bindgen@latest

# Full build: generate bindings, copy runtime, compile WASM
gowasm-bindgen wasm/main.go
# Creates in generated/: go-wasm.ts, worker.js, wasm_exec.js, wasm.wasm
# Creates in wasm/: bindings_gen.go

# Generate only (skip WASM compilation)
gowasm-bindgen wasm/main.go --no-build

# Build with standard Go (larger binary, ~2.4MB)
gowasm-bindgen wasm/main.go --compiler go
```

## Usage

```typescript
import { GoWasm } from './generated/go-wasm';

// Worker mode (default) - non-blocking
const wasm = await GoWasm.init('./worker.js');
const greeting = await wasm.greet('World');
wasm.terminate();

// Sync mode (--mode sync)
const wasm = await GoWasm.init('./wasm.wasm');
const greeting = wasm.greet('World');  // no await needed
```

See the [CLI Reference](https://13rac1.github.io/gowasm-bindgen/docs/cli-reference/) for all options.

## Get Started

Choose your path:

- **[For TypeScript Developers](https://13rac1.github.io/gowasm-bindgen/docs/for-typescript-devs/)** — Run Go in your frontend in 10 minutes
- **[For Go Developers](https://13rac1.github.io/gowasm-bindgen/docs/for-go-devs/)** — Full-stack Go with proper JS types

## See It Working

The [examples/simple/](examples/simple/) directory has a complete demo:

```bash
cd examples/simple
make all    # Build WASM, generate types, verify, compile TypeScript
make serve  # Open http://localhost:8080
```

## How It Works

1. Parse your Go source file
2. Find exported functions (capitalized names with no receiver)
3. Extract parameter names and types from function signatures
4. Extract return types, supporting structs with JSON tags and (T, error) patterns
5. Generate TypeScript client with a type-safe API
6. Generate `bindings_gen.go` with WASM wrapper functions that handle type conversions
7. Generate `worker.js` for async/non-blocking calls (default) or sync mode with `--mode sync`
8. Copy `wasm_exec.js` runtime from your Go/TinyGo installation
9. Compile WASM binary with TinyGo (or standard Go with `--compiler go`)

No annotations. No build plugins. Just normal Go code.

## Requirements

- Go 1.21+
- [TinyGo](https://tinygo.org/getting-started/install/) (recommended for small binaries) or standard Go
- Node.js 18+ (for TypeScript verification and example demo)
- Write normal exported Go functions - the tool generates the WASM wrappers for you

**Note**: TinyGo produces much smaller binaries but has [language limitations](https://tinygo.org/docs/reference/lang-support/). Use standard Go if you need full reflection or unsupported features.

## Supported Types

Primitives, slices, maps, structs (with JSON tags), errors, and pointers. See [Type Mapping](https://13rac1.github.io/gowasm-bindgen/docs/type-mapping/) for the full conversion table.

## Limitations

See [Limitations](https://13rac1.github.io/gowasm-bindgen/docs/limitations/) for a comparison with Rust's wasm-bindgen and current gaps. Highlights:

- Worker mode is default (async Promise-based), use `--mode sync` for synchronous blocking calls
- Void callbacks supported (fire-and-forget), no return value callbacks
- Typed arrays for byte slices, element-by-element for other numeric slices
- Class-based API (methods on class instances)

## Related Projects

- **[gowebapi/webapi](https://github.com/gowebapi/webapi)** — Go bindings for browser APIs (DOM, Fetch, Canvas, etc.). Use it alongside gowasm-bindgen: gowebapi/webapi lets your Go code *call* browser APIs, while gowasm-bindgen lets JavaScript *call* your Go functions.

## License

MIT

## Thanks

Logo assets sourced from:
- [WebAssembly](https://webassembly.org/) logo
- [Go](https://go.dev/) gopher
- [TypeScript](https://www.typescriptlang.org/) logo
- Cruise Ship by [FreePNGimg.com](https://freepngimg.com/)
