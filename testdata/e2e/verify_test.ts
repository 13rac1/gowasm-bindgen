import { assertEquals } from "https://deno.land/std@0.208.0/assert/mod.ts";

// Import types (this tests that .d.ts compiles)
/// <reference path="./test.d.ts" />

// Load Go WASM runtime
import "./wasm_exec.js";

declare const Go: any;

Deno.test("WASM functions with TypeScript types", async () => {
	// Load WASM
	const wasmCode = await Deno.readFile("./testdata/e2e/test.wasm");
	const go = new Go();
	const result = await WebAssembly.instantiate(wasmCode, go.importObject);
	go.run(result.instance);

	// Test greet function
	const greeting = (globalThis as any).greet("Deno");
	assertEquals(greeting, "Hello, Deno!");

	// Test add function
	const sum = (globalThis as any).add(10, 20);
	assertEquals(sum, 30);

	// Test getInfo function
	const info = (globalThis as any).getInfo("test");
	assertEquals(info.active, true);
});
