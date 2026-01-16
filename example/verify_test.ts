import { test } from "node:test";
import assert from "node:assert";
import { readFile } from "node:fs/promises";

// Load Go WASM runtime
import "./wasm_exec.js";

void test("WASM functions with generated TypeScript types", async () => {
  // Load WASM binary
  const wasmCode = await readFile("./example.wasm");
  const go = new Go();
  const result = await WebAssembly.instantiate(wasmCode, go.importObject);
  void go.run(result.instance);

  // Test greet function - type-safe!
  const greeting: string = greet("TypeScript");
  assert.strictEqual(greeting, "Hello, TypeScript!");

  // Test calculate function - type-safe!
  const calcSum: number = calculate(10, 5, "add");
  assert.strictEqual(calcSum, 15);

  const product: number = calculate(10, 5, "mul");
  assert.strictEqual(product, 50);

  // Test formatUser function - type-safe!
  const user = formatUser("Alice", 30, true);
  assert.strictEqual(user.displayName, "Alice (30)");
  assert.strictEqual(user.status, "active");

  // Test sumNumbers function - type-safe!
  const sum: number = sumNumbers("1,2,3");
  assert.strictEqual(sum, 6);

  // Test validateEmail function - type-safe!
  const validResult = validateEmail("user@example.com");
  assert.strictEqual(validResult.valid, true);

  const invalidResult = validateEmail("invalid");
  assert.strictEqual(invalidResult.valid, false);
  assert.strictEqual(typeof invalidResult.error, "string");
});
