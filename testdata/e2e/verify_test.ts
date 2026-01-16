import { test } from "node:test";
import assert from "node:assert";
import { readFile } from "node:fs/promises";

// Import types (this tests that .d.ts compiles)
/// <reference path="./test.d.ts" />

// Load Go WASM runtime
import "./wasm_exec.js";

declare const Go: any;

test("WASM functions with TypeScript types", async () => {
	// Load WASM
	const wasmCode = await readFile("./testdata/e2e/test.wasm");
	const go = new Go();
	const result = await WebAssembly.instantiate(wasmCode, go.importObject);
	go.run(result.instance);

	// Test greet function
	const greeting = (globalThis as any).greet("Node");
	assert.strictEqual(greeting, "Hello, Node!");

	// Test add function
	const sum = (globalThis as any).add(10, 20);
	assert.strictEqual(sum, 30);

	// Test getInfo function
	const info = (globalThis as any).getInfo("test");
	assert.strictEqual(info.active, true);
});
