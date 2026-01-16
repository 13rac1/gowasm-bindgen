/**
 * Type-safe DOM element accessor.
 *
 * This pattern eliminates two common TypeScript escape hatches:
 * - Non-null assertions (!) - risky, element might not exist
 * - Type assertions (as) - risky, element might be wrong type
 *
 * @example
 * // ❌ Unsafe patterns this replaces:
 * const btn = document.getElementById("myBtn")!;  // might be null!
 * const input = document.getElementById("x") as HTMLInputElement;  // might be wrong type!
 *
 * // ✅ Safe pattern:
 * const btn = getElement("myBtn", HTMLButtonElement);  // throws if wrong
 *
 * @param id - The DOM element's id attribute
 * @param ctor - The expected element type (HTMLButtonElement, HTMLInputElement, etc.)
 * @returns The element, guaranteed to be the correct type
 * @throws {Error} If element not found or has wrong type
 */
function getElement<T extends HTMLElement>(id: string, ctor: new () => T): T {
  const el = document.getElementById(id);
  if (!(el instanceof ctor)) {
    throw new Error(`Element #${id} not found or wrong type`);
  }
  return el;
}

/**
 * Type-safe DOM element accessor that returns null if not found.
 *
 * Use this for optional elements or in error handling paths where
 * you don't want to throw if the element is missing.
 *
 * @param id - The DOM element's id attribute
 * @param ctor - The expected element type
 * @returns The element if found and correct type, null otherwise
 */
function getOptionalElement<T extends HTMLElement>(
  id: string,
  ctor: new () => T
): T | null {
  const el = document.getElementById(id);
  return el instanceof ctor ? el : null;
}

/**
 * Safely converts an unknown error to a displayable string.
 *
 * In strict TypeScript, catch variables are 'unknown' (not 'any').
 * This means you must check the type before accessing properties.
 *
 * @example
 * try {
 *   riskyOperation();
 * } catch (err: unknown) {
 *   // ❌ Can't do this - err might not have .message
 *   console.log(err.message);
 *
 *   // ✅ Use formatError to safely extract the message
 *   console.log(formatError(err));
 * }
 */
function formatError(err: unknown): string {
  if (err instanceof Error) {
    return err.message;
  }
  return String(err);
}

/**
 * Wraps a function with consistent error handling and display.
 *
 * This pattern keeps error handling DRY (Don't Repeat Yourself).
 * Instead of copy-pasting try/catch blocks, we centralize the logic.
 *
 * @example
 * // ❌ Without this helper - repetitive try/catch everywhere:
 * function runGreet(): void {
 *   try {
 *     // ... logic ...
 *   } catch (err: unknown) {
 *     const el = document.getElementById("greetResult");
 *     if (el) el.textContent = `Error: ${formatError(err)}`;
 *   }
 * }
 *
 * // ✅ With this helper - clean and consistent:
 * function runGreet(): void {
 *   withErrorHandling("greetResult", () => {
 *     // ... logic ...
 *   });
 * }
 *
 * @param resultElementId - ID of element to display errors in
 * @param fn - Function to execute with error handling
 */
function withErrorHandling(resultElementId: string, fn: () => void): void {
  try {
    fn();
  } catch (err: unknown) {
    const resultElement = getOptionalElement(resultElementId, HTMLElement);
    if (resultElement) {
      resultElement.textContent = `Error: ${formatError(err)}`;
    }
    console.error(`Error in ${resultElementId}:`, err);
  }
}

/**
 * Initialize the Go WASM module.
 *
 * This demonstrates the async/await pattern for WASM loading:
 * 1. Create Go runtime instance
 * 2. Fetch and compile WASM (async)
 * 3. Start the Go runtime (fire-and-forget)
 * 4. Set up UI after WASM is ready
 */
async function initWasm(): Promise<void> {
  try {
    const go = new Go();

    // WebAssembly.instantiateStreaming is a browser optimization that
    // compiles the WASM while it's still downloading. This is faster
    // than fetch().then(instantiate) for large WASM files.
    const result = await WebAssembly.instantiateStreaming(
      fetch("../example.wasm"),
      go.importObject
    );

    // Fire-and-forget: go.run() returns a Promise that never resolves
    // because the Go program runs indefinitely. The 'void' operator
    // explicitly marks this as intentional (satisfies ESLint's
    // @typescript-eslint/no-floating-promises rule).
    void go.run(result.instance);

    // Update UI to show WASM is ready
    const statusElement = getElement("status", HTMLElement);
    statusElement.className = "ready";
    statusElement.textContent = "WASM loaded and ready!";

    // Attach event listeners after WASM is loaded
    // This ensures Go functions exist before we try to call them
    getElement("greetBtn", HTMLButtonElement).addEventListener("click", runGreet);
    getElement("calcBtn", HTMLButtonElement).addEventListener("click", runCalculate);
    getElement("formatBtn", HTMLButtonElement).addEventListener("click", runFormatUser);
    getElement("sumBtn", HTMLButtonElement).addEventListener("click", runSumNumbers);
    getElement("emailBtn", HTMLButtonElement).addEventListener("click", runValidateEmail);
  } catch (err: unknown) {
    // Handle initialization errors (network failure, invalid WASM, etc.)
    const statusElement = getOptionalElement("status", HTMLElement);
    if (statusElement) {
      statusElement.className = "error";
      statusElement.textContent = `Failed to load WASM: ${formatError(err)}`;
    }
    console.error("WASM initialization failed:", err);
  }
}

// Start initialization. The 'void' marks this as intentionally not awaited
// at the module level. Errors are handled inside initWasm().
void initWasm();

// ============================================================================
// Event Handlers - Each uses withErrorHandling for consistent error display
// ============================================================================

function runGreet(): void {
  withErrorHandling("greetResult", () => {
    const nameInput = getElement("greetName", HTMLInputElement);
    const resultElement = getElement("greetResult", HTMLElement);

    const result = window.greet(nameInput.value);
    resultElement.textContent = JSON.stringify(result);
  });
}

function runCalculate(): void {
  withErrorHandling("calcResult", () => {
    const aInput = getElement("calcA", HTMLInputElement);
    const bInput = getElement("calcB", HTMLInputElement);
    const opSelect = getElement("calcOp", HTMLSelectElement);
    const resultElement = getElement("calcResult", HTMLElement);

    const a = parseInt(aInput.value, 10);
    const b = parseInt(bInput.value, 10);
    if (Number.isNaN(a) || Number.isNaN(b)) {
      throw new Error("Please enter valid numbers");
    }

    const result = window.calculate(a, b, opSelect.value);
    resultElement.textContent = JSON.stringify(result);
  });
}

function runFormatUser(): void {
  withErrorHandling("formatResult", () => {
    const nameInput = getElement("userName", HTMLInputElement);
    const ageInput = getElement("userAge", HTMLInputElement);
    const activeInput = getElement("userActive", HTMLInputElement);
    const resultElement = getElement("formatResult", HTMLElement);

    const age = parseInt(ageInput.value, 10);
    if (Number.isNaN(age)) {
      throw new Error("Please enter a valid age");
    }

    const result = window.formatUser(nameInput.value, age, activeInput.checked);
    resultElement.textContent = JSON.stringify(result, null, 2);
  });
}

function runSumNumbers(): void {
  withErrorHandling("sumResult", () => {
    const numbersInput = getElement("numbersInput", HTMLInputElement);
    const resultElement = getElement("sumResult", HTMLElement);

    const result = window.sumNumbers(numbersInput.value);
    resultElement.textContent = JSON.stringify(result);
  });
}

function runValidateEmail(): void {
  withErrorHandling("emailResult", () => {
    const emailInput = getElement("emailInput", HTMLInputElement);
    const resultElement = getElement("emailResult", HTMLElement);

    const result = window.validateEmail(emailInput.value);
    resultElement.textContent = JSON.stringify(result, null, 2);
  });
}
