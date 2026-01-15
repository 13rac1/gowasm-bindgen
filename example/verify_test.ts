import { assertEquals } from "https://deno.land/std@0.208.0/assert/mod.ts";

// Import generated types (tests that .d.ts compiles)
/// <reference path="./types.d.ts" />

// Load Go WASM runtime
import "./wasm_exec.js";

declare const Go: any;

Deno.test("WASM functions with generated TypeScript types", async () => {
	// Load WASM binary
	const wasmCode = await Deno.readFile("./example.wasm");
	const go = new Go();
	const result = await WebAssembly.instantiate(wasmCode, go.importObject);
	go.run(result.instance);

	// Test greet function
	const greeting = (globalThis as any).greet("TypeScript");
	assertEquals(greeting, "Hello, TypeScript!");

	// Test calculate function
	const calcSum = (globalThis as any).calculate(10, 5, "add");
	assertEquals(calcSum, 15);

	const product = (globalThis as any).calculate(10, 5, "mul");
	assertEquals(product, 50);

	// Test formatUser function
	const user = (globalThis as any).formatUser("Alice", 30, true);
	assertEquals(user.displayName, "Alice (30)");
	assertEquals(user.status, "active");

	// Test sumNumbers function
	const sum = (globalThis as any).sumNumbers("1,2,3");
	assertEquals(sum, 6);

	// Test validateEmail function
	const validResult = (globalThis as any).validateEmail("user@example.com");
	assertEquals(validResult.valid, true);

	const invalidResult = (globalThis as any).validateEmail("invalid");
	assertEquals(invalidResult.valid, false);
	assertEquals(typeof invalidResult.error, "string");
});
