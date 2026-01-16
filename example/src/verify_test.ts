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
import type { Main, FormatUserResult, ValidateEmailResult } from "../generated/client";

void test("Generated types compile and match WASM signatures", async () => {
  // Type-only test: verify the generated interfaces exist and have correct shapes

  // Verify FormatUserResult interface
  const user: FormatUserResult = {
    displayName: "Test User (30)",
    status: "active"
  };
  assert.strictEqual(user.displayName, "Test User (30)");
  assert.strictEqual(user.status, "active");

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
  const mockWasm: Pick<Main, 'greet' | 'calculate' | 'formatUser' | 'sumNumbers' | 'validateEmail' | 'divide' | 'terminate'> = {
    greet: async (name: string): Promise<string> => `Hello, ${name}!`,
    calculate: async (a: number, b: number, op: string): Promise<number> => {
      switch (op) {
        case "add": return a + b;
        case "sub": return a - b;
        case "mul": return a * b;
        case "div": return b === 0 ? 0 : a / b;
        default: return 0;
      }
    },
    formatUser: async (name: string, age: number, active: boolean): Promise<FormatUserResult> => ({
      displayName: `${name} (${age})`,
      status: active ? "active" : "inactive"
    }),
    sumNumbers: async (input: string): Promise<number> => {
      if (input === "") return 0;
      return input.split(',').map(s => parseInt(s.trim(), 10) || 0).reduce((a, b) => a + b, 0);
    },
    validateEmail: async (email: string): Promise<ValidateEmailResult> => ({
      valid: email.includes('@'),
      error: email.includes('@') ? '' : 'missing @ symbol'
    }),
    divide: async (a: number, b: number): Promise<number> => {
      if (b === 0) {
        throw new Error("division by zero");
      }
      return a / b;
    },
    terminate: (): void => {}
  };

  // Test the mock to verify types
  const greeting = await mockWasm.greet("TypeScript");
  assert.strictEqual(greeting, "Hello, TypeScript!");

  const sum = await mockWasm.calculate(10, 5, "add");
  assert.strictEqual(sum, 15);

  const subtraction = await mockWasm.calculate(10, 5, "sub");
  assert.strictEqual(subtraction, 5);

  const product = await mockWasm.calculate(10, 5, "mul");
  assert.strictEqual(product, 50);

  const division = await mockWasm.calculate(10, 5, "div");
  assert.strictEqual(division, 2);

  const divisionByZero = await mockWasm.calculate(10, 0, "div");
  assert.strictEqual(divisionByZero, 0);

  const formattedUser = await mockWasm.formatUser("Alice", 30, true);
  assert.strictEqual(formattedUser.displayName, "Alice (30)");
  assert.strictEqual(formattedUser.status, "active");

  const inactiveUser = await mockWasm.formatUser("Bob", 25, false);
  assert.strictEqual(inactiveUser.displayName, "Bob (25)");
  assert.strictEqual(inactiveUser.status, "inactive");

  const numbersSum = await mockWasm.sumNumbers("1,2,3");
  assert.strictEqual(numbersSum, 6);

  const spacedSum = await mockWasm.sumNumbers("10, 20, 30, 40");
  assert.strictEqual(spacedSum, 100);

  const emptySum = await mockWasm.sumNumbers("");
  assert.strictEqual(emptySum, 0);

  const validEmail = await mockWasm.validateEmail("user@example.com");
  assert.strictEqual(validEmail.valid, true);
  assert.strictEqual(validEmail.error, "");

  const invalidEmail = await mockWasm.validateEmail("invalid");
  assert.strictEqual(invalidEmail.valid, false);
  assert.strictEqual(typeof invalidEmail.error, "string");
  assert.ok(invalidEmail.error.length > 0);

  // Test divide function
  const divideResult = await mockWasm.divide(10, 2);
  assert.strictEqual(divideResult, 5);

  // Test divide throws on division by zero (automatic error throwing)
  await assert.rejects(
    mockWasm.divide(10, 0),
    { message: /division by zero/ }
  );
});
