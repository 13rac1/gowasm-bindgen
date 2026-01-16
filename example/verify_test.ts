/**
 * Type verification tests for generated WASM bindings.
 *
 * This file verifies that the generated TypeScript types correctly match
 * the actual WASM function signatures. It runs in Node.js, not the browser.
 *
 * ## Node.js vs Browser WASM Loading
 *
 * | Aspect | Browser | Node.js (this file) |
 * |--------|---------|---------------------|
 * | Get binary | `fetch("app.wasm")` | `readFile("app.wasm")` |
 * | Instantiate | `instantiateStreaming()` | `instantiate()` |
 * | Global access | `window.greet()` | `greet()` (global) |
 *
 * ### Why `instantiate()` instead of `instantiateStreaming()`?
 *
 * `instantiateStreaming()` takes a Response object (from fetch) and compiles
 * the WASM while it's still downloading—an optimization for browsers.
 *
 * Node.js doesn't have fetch() by default, so we use `readFile()` to get the
 * bytes, then `instantiate()` which takes a BufferSource directly.
 *
 * ### Why no `window.` prefix?
 *
 * In browsers, WASM functions are attached to `window` (the global object).
 * In Node.js, they're attached to `globalThis`. The generated types declare
 * both `Window` interface extensions and `var` declarations, so both work:
 *
 * ```typescript
 * // Browser
 * window.greet("World")
 *
 * // Node.js
 * greet("World")  // Works because of: declare var greet: ...
 * ```
 */
import { test } from "node:test";
import assert from "node:assert";
import { readFile } from "node:fs/promises";

// Load Go WASM runtime - this adds the Go class to globalThis
import "./wasm_exec.js";

void test("WASM functions with generated TypeScript types", async () => {
  // In Node.js, we read the file directly from disk.
  // In browsers, you'd use: fetch("./example.wasm")
  const wasmCode = await readFile("./example.wasm");

  const go = new Go();

  // WebAssembly.instantiate() takes a BufferSource (our file bytes).
  // In browsers, prefer instantiateStreaming() with fetch() for better perf.
  const result = await WebAssembly.instantiate(wasmCode, go.importObject);

  // Fire-and-forget: go.run() never resolves (Go program runs indefinitely).
  // The 'void' operator marks this as intentionally not awaited.
  void go.run(result.instance);

  // ========================================================================
  // Type verification tests
  // If these compile, the generated types match the actual WASM signatures.
  // ========================================================================

  // greet(name: string): string
  const greeting: string = greet("TypeScript");
  assert.strictEqual(greeting, "Hello, TypeScript!");

  // calculate(a: number, b: number, op: string): number
  const calcSum: number = calculate(10, 5, "add");
  assert.strictEqual(calcSum, 15);

  const product: number = calculate(10, 5, "mul");
  assert.strictEqual(product, 50);

  // formatUser(name: string, age: number, active: boolean): FormatUserResult
  // The return type is a named interface, not an inline object type
  const user = formatUser("Alice", 30, true);
  assert.strictEqual(user.displayName, "Alice (30)");
  assert.strictEqual(user.status, "active");

  // sumNumbers(input: string): number
  const sum: number = sumNumbers("1,2,3");
  assert.strictEqual(sum, 6);

  // validateEmail(email: string): ValidateEmailResult
  const validResult = validateEmail("user@example.com");
  assert.strictEqual(validResult.valid, true);

  const invalidResult = validateEmail("invalid");
  assert.strictEqual(invalidResult.valid, false);
  assert.strictEqual(typeof invalidResult.error, "string");
});
