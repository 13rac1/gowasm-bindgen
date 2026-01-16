// Type-safe DOM element accessor - eliminates all `!` and `as` assertions
function getElement<T extends HTMLElement>(id: string, ctor: new () => T): T {
  const el = document.getElementById(id);
  if (!(el instanceof ctor)) {
    throw new Error(`Element #${id} not found or wrong type`);
  }
  return el;
}

// Format error for display
function formatError(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}

// Load and initialize Go WASM
const go = new Go();

WebAssembly.instantiateStreaming(fetch("../example.wasm"), go.importObject)
  .then((result): void => {
    void go.run(result.instance);

    const statusElement = getElement("status", HTMLElement);
    statusElement.className = "ready";
    statusElement.textContent = "WASM loaded and ready!";

    // Attach event listeners after WASM is loaded
    getElement("greetBtn", HTMLButtonElement).addEventListener("click", runGreet);
    getElement("calcBtn", HTMLButtonElement).addEventListener("click", runCalculate);
    getElement("formatBtn", HTMLButtonElement).addEventListener("click", runFormatUser);
    getElement("sumBtn", HTMLButtonElement).addEventListener("click", runSumNumbers);
    getElement("emailBtn", HTMLButtonElement).addEventListener("click", runValidateEmail);
  })
  .catch((err: unknown): void => {
    const statusElement = document.getElementById("status");
    if (statusElement) {
      statusElement.className = "error";
      statusElement.textContent = `Failed to load WASM: ${formatError(err)}`;
    }
    console.error(err);
  });

// Type-safe function wrappers using generated types
function runGreet(): void {
  try {
    const nameInput = getElement("greetName", HTMLInputElement);
    const resultElement = getElement("greetResult", HTMLElement);

    const result: string = window.greet(nameInput.value);
    resultElement.textContent = JSON.stringify(result);
  } catch (err: unknown) {
    const resultElement = document.getElementById("greetResult");
    if (resultElement) {
      resultElement.textContent = `Error: ${formatError(err)}`;
    }
  }
}

function runCalculate(): void {
  try {
    const aInput = getElement("calcA", HTMLInputElement);
    const bInput = getElement("calcB", HTMLInputElement);
    const opSelect = getElement("calcOp", HTMLSelectElement);
    const resultElement = getElement("calcResult", HTMLElement);

    const a = parseInt(aInput.value, 10);
    const b = parseInt(bInput.value, 10);
    if (Number.isNaN(a) || Number.isNaN(b)) {
      throw new Error("Please enter valid numbers");
    }

    const result: number = window.calculate(a, b, opSelect.value);
    resultElement.textContent = JSON.stringify(result);
  } catch (err: unknown) {
    const resultElement = document.getElementById("calcResult");
    if (resultElement) {
      resultElement.textContent = `Error: ${formatError(err)}`;
    }
  }
}

function runFormatUser(): void {
  try {
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
  } catch (err: unknown) {
    const resultElement = document.getElementById("formatResult");
    if (resultElement) {
      resultElement.textContent = `Error: ${formatError(err)}`;
    }
  }
}

function runSumNumbers(): void {
  try {
    const numbersInput = getElement("numbersInput", HTMLInputElement);
    const resultElement = getElement("sumResult", HTMLElement);

    const result: number = window.sumNumbers(numbersInput.value);
    resultElement.textContent = JSON.stringify(result);
  } catch (err: unknown) {
    const resultElement = document.getElementById("sumResult");
    if (resultElement) {
      resultElement.textContent = `Error: ${formatError(err)}`;
    }
  }
}

function runValidateEmail(): void {
  try {
    const emailInput = getElement("emailInput", HTMLInputElement);
    const resultElement = getElement("emailResult", HTMLElement);

    const result = window.validateEmail(emailInput.value);
    resultElement.textContent = JSON.stringify(result, null, 2);
  } catch (err: unknown) {
    const resultElement = document.getElementById("emailResult");
    if (resultElement) {
      resultElement.textContent = `Error: ${formatError(err)}`;
    }
  }
}
