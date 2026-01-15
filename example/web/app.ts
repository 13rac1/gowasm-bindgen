/// <reference path="../types.d.ts" />

// Load and initialize Go WASM
declare const Go: any;
const go = new Go();

WebAssembly.instantiateStreaming(fetch("../example.wasm"), go.importObject)
    .then((result) => {
        go.run(result.instance);
        document.getElementById("status")!.className = "ready";
        document.getElementById("status")!.textContent = "WASM loaded and ready!";
    })
    .catch((err) => {
        document.getElementById("status")!.className = "error";
        document.getElementById("status")!.textContent = "Failed to load WASM: " + err.message;
        console.error(err);
    });

// Type-safe function wrappers using generated types
function runGreet(): void {
    const name = (document.getElementById("greetName") as HTMLInputElement).value;
    const result: string = window.greet(name);
    document.getElementById("greetResult")!.textContent = JSON.stringify(result);
}

function runCalculate(): void {
    const a = parseInt((document.getElementById("calcA") as HTMLInputElement).value);
    const b = parseInt((document.getElementById("calcB") as HTMLInputElement).value);
    const op = (document.getElementById("calcOp") as HTMLSelectElement).value;
    const result: number = window.calculate(a, b, op);
    document.getElementById("calcResult")!.textContent = JSON.stringify(result);
}

function runFormatUser(): void {
    const name = (document.getElementById("userName") as HTMLInputElement).value;
    const age = parseInt((document.getElementById("userAge") as HTMLInputElement).value);
    const active = (document.getElementById("userActive") as HTMLInputElement).checked;
    const result: { displayName: string; status: string } = window.formatUser(name, age, active);
    document.getElementById("formatResult")!.textContent = JSON.stringify(result, null, 2);
}

function runSumNumbers(): void {
    const input = (document.getElementById("numbersInput") as HTMLInputElement).value;
    const result: number = window.sumNumbers(input);
    document.getElementById("sumResult")!.textContent = JSON.stringify(result);
}

function runValidateEmail(): void {
    const email = (document.getElementById("emailInput") as HTMLInputElement).value;
    const result: { valid: boolean; error: string } = window.validateEmail(email);
    document.getElementById("emailResult")!.textContent = JSON.stringify(result, null, 2);
}

// Expose functions to onclick handlers
(window as any).runGreet = runGreet;
(window as any).runCalculate = runCalculate;
(window as any).runFormatUser = runFormatUser;
(window as any).runSumNumbers = runSumNumbers;
(window as any).runValidateEmail = runValidateEmail;
