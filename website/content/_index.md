---
title: "gowasm-bindgen"
type: docs
---

# gowasm-bindgen

**Type-safe Go in the browser.** Generate TypeScript bindings from Go source code.

## The Problem

Go WASM functions are invisible to TypeScript:

```typescript
// TypeScript has no idea what this returns or accepts
const result = window.myGoFunction(???, ???);  // any
```

## The Solution

Write normal Go functions:

```go
func Greet(name string) string {
    return "Hello, " + name + "!"
}
```

Get typed TypeScript APIs automatically:

```typescript
// Full type safety - greet(name: string): Promise<string>
const greeting = await wasm.greet("World");
```

## Features

- **Zero boilerplate** - Write normal Go functions, no annotations needed
- **Full type inference** - Types inferred from Go function signatures
- **Worker mode** - Non-blocking async calls via Web Workers (default)
- **Sync mode** - Direct synchronous calls when needed
- **TinyGo support** - Ship 90KB gzipped binaries

## Quick Start

```bash
# Install
go install github.com/13rac1/gowasm-bindgen/cmd/gowasm-bindgen@latest

# Generate TypeScript client and Go bindings
gowasm-bindgen main.go --output client.ts --go-output bindings_gen.go
```

{{< button href="/docs/getting-started" >}}Get Started{{< /button >}}
{{< button href="/examples" >}}Live Demos{{< /button >}}
