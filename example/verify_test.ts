import { test } from "node:test";
import assert from "node:assert";
import { readFile } from "node:fs/promises";

// Import generated types (tests that .d.ts compiles)
/// <reference path="./types.d.ts" />

// Load Go WASM runtime
import "./wasm_exec.js";

declare const Go: any;

test("WASM functions with generated TypeScript types", async () => {
	// Load WASM binary
	const wasmCode = await readFile("./example.wasm");
	const go = new Go();
	const result = await WebAssembly.instantiate(wasmCode, go.importObject);
	go.run(result.instance);

	// Test greet function
	const greeting = (globalThis as any).greet("TypeScript");
	assert.strictEqual(greeting, "Hello, TypeScript!");

	// Test calculate function
	const calcSum = (globalThis as any).calculate(10, 5, "add");
	assert.strictEqual(calcSum, 15);

	const product = (globalThis as any).calculate(10, 5, "mul");
	assert.strictEqual(product, 50);

	// Test formatUser function
	const user = (globalThis as any).formatUser("Alice", 30, true);
	assert.strictEqual(user.displayName, "Alice (30)");
	assert.strictEqual(user.status, "active");

	// Test sumNumbers function
	const sum = (globalThis as any).sumNumbers("1,2,3");
	assert.strictEqual(sum, 6);

	// Test validateEmail function
	const validResult = (globalThis as any).validateEmail("user@example.com");
	assert.strictEqual(validResult.valid, true);

	const invalidResult = (globalThis as any).validateEmail("invalid");
	assert.strictEqual(invalidResult.valid, false);
	assert.strictEqual(typeof invalidResult.error, "string");
});
