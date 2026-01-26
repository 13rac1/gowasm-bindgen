import { test } from "node:test";
import assert from "node:assert";
import { readFile } from "node:fs/promises";

// Load Go WASM runtime
import "./generated/wasm_exec.js";

// Import generated client
import { GoWasm } from "./generated/go-wasm.js";

void test("WASM functions with TypeScript types", async () => {
  // Load WASM using generated client
  const wasmCode = await readFile("./test/e2e/generated/wasm.wasm");
  const wasm = await GoWasm.init(wasmCode);

  // Test greet function - type-safe!
  const greeting: string = wasm.greet("Node");
  assert.strictEqual(greeting, "Hello, Node!");

  // Test add function - type-safe!
  const sum: number = wasm.add(10, 20);
  assert.strictEqual(sum, 30);

  // Test getInfo function - type-safe!
  const info = wasm.getInfo("test");
  assert.strictEqual(info.name, "test");
  assert.strictEqual(info.version, 1);
  assert.strictEqual(info.active, true);

  // Test panic recovery - should throw error, not crash WASM
  assert.throws(
    () => wasm.triggerPanic(),
    {
      name: "Error",
      message: "panic: intentional panic for testing",
    }
  );

  // Verify WASM still works after panic
  const afterPanic: string = wasm.greet("AfterPanic");
  assert.strictEqual(afterPanic, "Hello, AfterPanic!");
});
