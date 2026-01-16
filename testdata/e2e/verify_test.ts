import { test } from "node:test";
import assert from "node:assert";
import { readFile } from "node:fs/promises";

// Load Go WASM runtime
import "./wasm_exec.js";

void test("WASM functions with TypeScript types", async () => {
  // Load WASM
  const wasmCode = await readFile("./testdata/e2e/test.wasm");
  const go = new Go();
  const result = await WebAssembly.instantiate(wasmCode, go.importObject);
  void go.run(result.instance);

  // Test greet function - type-safe!
  const greeting: string = greet("Node");
  assert.strictEqual(greeting, "Hello, Node!");

  // Test add function - type-safe!
  const sum: number = add(10, 20);
  assert.strictEqual(sum, 30);

  // Test getInfo function - type-safe!
  const info = getInfo("test");
  assert.strictEqual(info.active, true);
});
