---
title: "gowasm-bindgen"
type: docs
---

# gowasm-bindgen

[![CI](https://github.com/13rac1/gowasm-bindgen/actions/workflows/ci.yml/badge.svg)](https://github.com/13rac1/gowasm-bindgen/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/13rac1/gowasm-bindgen/graph/badge.svg)](https://codecov.io/gh/13rac1/gowasm-bindgen)
[![Go Report Card](https://goreportcard.com/badge/github.com/13rac1/gowasm-bindgen)](https://goreportcard.com/report/github.com/13rac1/gowasm-bindgen)
[![Go Reference](https://pkg.go.dev/badge/github.com/13rac1/gowasm-bindgen.svg)](https://pkg.go.dev/github.com/13rac1/gowasm-bindgen)
[![GitHub release](https://img.shields.io/github/v/release/13rac1/gowasm-bindgen)](https://github.com/13rac1/gowasm-bindgen/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Type-safe Go in the browser.** Generate TypeScript bindings from Go source code.

## The Problem

Go compiles to WebAssembly (a binary format), not JavaScript. Bridging Go and JS requires tedious glue code:

```typescript
// TypeScript has no idea what this returns or accepts
const result = window.myGoFunction(???, ???);  // any
```

No type safety. No IDE support. Runtime crashes instead of compile errors.

[Learn more about the problem →]({{< relref "/docs/why" >}})

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

{{< rawhtml >}}
<div class="feature-grid">
  <div class="feature-card">
    <h3>Zero Boilerplate</h3>
    <p>Write normal Go functions. No annotations, decorators, or special syntax needed.</p>
  </div>
  <div class="feature-card">
    <h3>Full Type Inference</h3>
    <p>Types are inferred from Go function signatures. Structs, slices, maps—all handled.</p>
  </div>
  <div class="feature-card">
    <h3>Worker Mode</h3>
    <p>Non-blocking async calls via Web Workers. Keep your UI responsive by default.</p>
  </div>
  <div class="feature-card">
    <h3>Sync Mode</h3>
    <p>Direct synchronous calls when you need them. Simple flag to switch modes.</p>
  </div>
  <div class="feature-card">
    <h3>TinyGo Support</h3>
    <p>Ship 90KB gzipped binaries. Perfect for performance-critical applications.</p>
  </div>
  <div class="feature-card">
    <h3>Standard Go</h3>
    <p>Works with regular Go compiler too. 2-3MB gzipped for full stdlib access.</p>
  </div>
</div>
{{< /rawhtml >}}

## Quick Start

```bash
# Install
go install github.com/13rac1/gowasm-bindgen@latest

# Generate TypeScript client and Go bindings
gowasm-bindgen wasm/main.go --output generated
```

{{< button href="/docs/getting-started" >}}Get Started{{< /button >}}
{{< button href="/examples" >}}Live Demos{{< /button >}}
