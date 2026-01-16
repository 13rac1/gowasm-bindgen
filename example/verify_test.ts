/**
 * Type verification tests for generated WASM bindings with class-based API.
 *
 * This file verifies that the generated TypeScript class correctly matches
 * the actual WASM function signatures. It runs in Node.js, not the browser.
 *
 * ## Class-Based API Pattern
 *
 * The new API uses a class instance instead of global functions:
 *
 * ```typescript
 * import { Main } from './client';
 * const wasm = await Main.init('./worker.js');
 * const result = await wasm.greet('World');
 * wasm.terminate();
 * ```
 *
 * ## Testing in Node.js
 *
 * Node.js doesn't have Web Workers, so we can't test the worker mode directly.
 * Instead, we'll verify the TypeScript types compile correctly and test with
 * the synchronous mode by mocking the API.
 *
 * Note: This is a limitation of testing worker-based WASM in Node.js.
 * The actual worker functionality should be tested in a browser environment.
 */
import { test } from "node:test";
import assert from "node:assert";
import type { Main, FormatUserResult, ValidateEmailResult } from "./client";

void test("Generated types compile and match WASM signatures", async () => {
  // Type-only test: verify the generated interfaces exist and have correct shapes

  // Verify FormatUserResult interface
  const formatUserResult: FormatUserResult = {
    displayName: "Test User (30)",
    status: "active"
  };
  assert.strictEqual(formatUserResult.displayName, "Test User (30)");
  assert.strictEqual(formatUserResult.status, "active");

  // Verify ValidateEmailResult interface
  const validEmailResult: ValidateEmailResult = {
    valid: true,
    error: ""
  };
  assert.strictEqual(validEmailResult.valid, true);

  const invalidEmailResult: ValidateEmailResult = {
    valid: false,
    error: "Invalid email format"
  };
  assert.strictEqual(invalidEmailResult.valid, false);
  assert.strictEqual(typeof invalidEmailResult.error, "string");

  // Verify Main class type structure
  // This is a compile-time check - if these type annotations compile,
  // the generated class has the correct method signatures
  const mockWasm: Pick<Main, 'greet' | 'calculate' | 'formatUser' | 'sumNumbers' | 'validateEmail' | 'terminate'> = {
    greet: async (name: string): Promise<string> => `Hello, ${name}!`,
    calculate: async (a: number, b: number, _op: string): Promise<number> => a + b,
    formatUser: async (name: string, age: number, active: boolean): Promise<FormatUserResult> => ({
      displayName: `${name} (${age})`,
      status: active ? "active" : "inactive"
    }),
    sumNumbers: async (input: string): Promise<number> => {
      return input.split(',').map(Number).reduce((a, b) => a + b, 0);
    },
    validateEmail: async (email: string): Promise<ValidateEmailResult> => ({
      valid: email.includes('@'),
      error: email.includes('@') ? '' : 'Invalid email'
    }),
    terminate: (): void => {}
  };

  // Test the mock to verify types
  const greeting = await mockWasm.greet("TypeScript");
  assert.strictEqual(greeting, "Hello, TypeScript!");

  const sum = await mockWasm.calculate(10, 5, "add");
  assert.strictEqual(sum, 15);

  const user = await mockWasm.formatUser("Alice", 30, true);
  assert.strictEqual(user.displayName, "Alice (30)");
  assert.strictEqual(user.status, "active");

  const numbersSum = await mockWasm.sumNumbers("1,2,3");
  assert.strictEqual(numbersSum, 6);

  const validEmail = await mockWasm.validateEmail("user@example.com");
  assert.strictEqual(validEmail.valid, true);

  const invalidEmail = await mockWasm.validateEmail("invalid");
  assert.strictEqual(invalidEmail.valid, false);
});
