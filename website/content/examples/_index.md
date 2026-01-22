---
title: "Examples"
weight: 5
bookFlatSection: true
---

# Examples

See gowasm-bindgen in action with these interactive demos.

## Image Processing

[**Live Demo →**]({{< relref "/examples/image-processing" >}})

Process images in the browser using Go compiled to WebAssembly. Compare performance between WASM and pure JavaScript implementations.

Features:
- Grayscale, brightness, contrast, sepia, invert filters
- Side-by-side WASM vs JavaScript performance comparison
- Real-time benchmarking with millisecond timing

## JavaScript Sandbox

[**Live Demo →**]({{< relref "/examples/js-sandbox" >}})

Execute untrusted JavaScript securely using Goja (a Go-based JS interpreter) compiled to WebAssembly.

Features:
- Complete isolation from browser APIs
- Same interpreter as your Go backend
- Demonstrates gowasm-bindgen with standard Go (not TinyGo)
