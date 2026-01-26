---
title: "JavaScript Sandbox Demo"
weight: 2
---

# JavaScript Sandbox Demo

Securely execute untrusted JavaScript using [Goja](https://github.com/dop251/goja)--a JavaScript interpreter written in Go--compiled to WebAssembly.

**Use case:** Run user-provided scripts in complete isolation, with no access to DOM, network, or browser APIs.

{{< rawhtml >}}
<style>
.demo-container {
  font-family: system-ui, -apple-system, sans-serif;
  max-width: 900px;
  margin: 2rem auto;
}
.code-editor {
  width: 100%;
  min-height: 300px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 14px;
  line-height: 1.5;
  padding: 1rem;
  border: 1px solid #444;
  border-radius: 8px;
  background: #1e1e1e;
  color: #d4d4d4;
  resize: vertical;
}
.controls {
  display: flex;
  gap: 1rem;
  align-items: center;
  margin: 1rem 0;
}
.controls button {
  padding: 0.75rem 1.5rem;
  background: #0066cc;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  font-size: 1rem;
}
.controls button:hover {
  background: #0055aa;
}
.controls button:disabled {
  background: #555;
  cursor: not-allowed;
}
.output-container {
  background: #1e1e1e;
  border: 1px solid #444;
  border-radius: 8px;
  padding: 1rem;
  margin-top: 1rem;
}
.output-section {
  margin-bottom: 1rem;
}
.output-section:last-child {
  margin-bottom: 0;
}
.output-label {
  font-size: 0.85rem;
  color: #888;
  margin-bottom: 0.5rem;
  font-weight: 600;
}
.output-content {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  color: #d4d4d4;
}
.output-content.logs {
  color: #9cdcfe;
}
.output-content.result {
  color: #4ec9b0;
}
.output-content.error {
  color: #f48771;
}
.status {
  padding: 0.75rem 1rem;
  margin-bottom: 1rem;
  border-radius: 6px;
  background: #3a3d41;
  color: #ccc;
}
.status.ready {
  background: #2d4a3e;
  color: #89d185;
}
.status.error {
  background: #5a2d2d;
  color: #f48771;
}
.info-box {
  background: #2d333b;
  border: 1px solid #444;
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 1rem;
}
.info-box h4 {
  margin: 0 0 0.5rem 0;
  color: #e6e6e6;
}
.info-box p {
  margin: 0;
  color: #aaa;
  font-size: 0.9rem;
}
</style>

<div class="demo-container">
  <div id="status" class="status">Loading WASM module (first load: ~3MB compressed)...</div>

  <div class="info-box">
    <h4>Sandbox Proof</h4>
    <p>The code below runs in Goja, not your browser's JS engine. The <code>__goja__</code> global only exists inside Goja, and <code>window</code>/<code>document</code> don't exist.</p>
  </div>

  <textarea id="code-input" class="code-editor">// Proof this runs in a sandbox, not browser JS!
console.log("Engine:", __goja__.engine);
console.log("Go Version:", __goja__.goVersion);
console.log("GOOS:", __goja__.goOS);
console.log("GOARCH:", __goja__.goArch);

// These don't exist in the sandbox (no DOM access)
console.log("window exists?", typeof window !== "undefined");
console.log("document exists?", typeof document !== "undefined");

// Run some JavaScript
function fibonacci(n) {
  if (n < 2) return n;
  return fibonacci(n - 1) + fibonacci(n - 2);
}

console.log("fibonacci(20) =", fibonacci(20));

// ES5+ features work
var arr = [1, 2, 3, 4, 5];
var doubled = arr.map(function(x) { return x * 2; });
console.log("doubled:", doubled.join(", "));

// Return a final value
"Executed in Goja sandbox!"</textarea>

  <div class="controls">
    <button id="run-btn" disabled>Run in Sandbox</button>
  </div>

  <div class="output-container">
    <div class="output-section">
      <div class="output-label">Console Output</div>
      <div id="logs-output" class="output-content logs">(run code to see output)</div>
    </div>
    <div class="output-section">
      <div class="output-label">Return Value</div>
      <div id="result-output" class="output-content result">-</div>
    </div>
    <div id="error-section" class="output-section" style="display: none;">
      <div class="output-label">Error</div>
      <div id="error-output" class="output-content error"></div>
    </div>
  </div>
</div>

<script src="demo.js"></script>
{{< /rawhtml >}}

## When To Use This

**Use Goja-in-WASM when:**
- Your backend is Go and you want the same interpreter client + server
- You need to integrate with other Go libraries in the sandbox
- You're already using gowasm-bindgen for other Go-to-WASM code

**Use [quickjs-emscripten](https://github.com/justjake/quickjs-emscripten) instead when:**
- Bundle size is critical (~500KB vs ~2.5MB)
- You don't need Go ecosystem integration
- You need better ES6+ support

## Alternatives

| Solution | Size (brotli) | Browser | Node | Notes |
|----------|---------------|---------|------|-------|
| **This demo (Goja)** | ~2.5 MB | Yes | Yes | Go ecosystem integration |
| [quickjs-emscripten](https://github.com/justjake/quickjs-emscripten) | ~500 KB | Yes | Yes | Smaller, C-based |
| [SandboxJS](https://github.com/nyariv/SandboxJS) | ~50 KB | Yes | Yes | JS-based parser |
| [vm2](https://github.com/patriksimek/vm2) | tiny | No | Yes | Node.js only |

For most browser sandboxing needs, **quickjs-emscripten is the better choice** due to its smaller size. This demo is valuable for Go-centric architectures where you want the same interpreter on client and server.

## How It Works

This demo uses **gowasm-bindgen** to generate TypeScript bindings for Go code that embeds the [Goja](https://github.com/dop251/goja) JavaScript interpreter:

```go
//go:build js && wasm

package main

import (
    "github.com/dop251/goja"
    "runtime"
)

// JSResult contains the output from running JavaScript code.
type JSResult struct {
    Result   string
    Logs     string
    ErrorMsg string
}

// RunJS executes JavaScript code in the Goja interpreter.
func RunJS(code string) JSResult {
    vm := goja.New()

    // Inject proof-of-goja globals
    vm.Set("__goja__", map[string]interface{}{
        "engine":    "goja",
        "goVersion": runtime.Version(),
        "goOS":      runtime.GOOS,
        "goArch":    runtime.GOARCH,
    })

    // Capture console.log output...
    // Run user code...

    return JSResult{Result: ..., Logs: ...}
}
```

### Generated TypeScript

gowasm-bindgen generates a typed client:

```typescript
export interface RunJSResult {
  result: string;
  logs: string;
  errorMsg: string;
}

export class GoGoja {
  static async init(workerUrl: string): Promise<GoGoja>;
  runJS(code: string): Promise<RunJSResult>;
}
```

## Why This Matters

This demo showcases:

1. **gowasm-bindgen works with standard Go** (not just TinyGo)
2. **Complex Go libraries** (Goja uses reflection heavily) compile to WASM
3. **Go ecosystem in browser** - same interpreter as your Go backend

### Bundle Size

| Format | Size |
|--------|------|
| Uncompressed | 17 MB |
| Gzip | 3.7 MB |
| Brotli | **2.5 MB** |

The WASM file is cached after first load.

## Source Code

- [Go implementation](https://github.com/13rac1/gowasm-bindgen/tree/main/examples/js-sandbox/goja/main.go)
- [Full example](https://github.com/13rac1/gowasm-bindgen/tree/main/examples/js-sandbox)
