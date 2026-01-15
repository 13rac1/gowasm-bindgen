// Load and initialize Go WASM
const go = new Go();

WebAssembly.instantiateStreaming(fetch("../example.wasm"), go.importObject)
    .then((result) => {
        go.run(result.instance);
        document.getElementById("status").className = "ready";
        document.getElementById("status").textContent = "WASM loaded and ready!";
    })
    .catch((err) => {
        document.getElementById("status").className = "error";
        document.getElementById("status").textContent = "Failed to load WASM: " + err.message;
        console.error(err);
    });

// Function wrappers
function runGreet() {
    const name = document.getElementById("greetName").value;
    const result = window.greet(name);
    document.getElementById("greetResult").textContent = JSON.stringify(result);
}

function runCalculate() {
    const a = parseInt(document.getElementById("calcA").value);
    const b = parseInt(document.getElementById("calcB").value);
    const op = document.getElementById("calcOp").value;
    const result = window.calculate(a, b, op);
    document.getElementById("calcResult").textContent = JSON.stringify(result);
}

function runFormatUser() {
    const name = document.getElementById("userName").value;
    const age = parseInt(document.getElementById("userAge").value);
    const active = document.getElementById("userActive").checked;
    const result = window.formatUser(name, age, active);
    document.getElementById("formatResult").textContent = JSON.stringify(result, null, 2);
}

function runSumNumbers() {
    const input = document.getElementById("numbersInput").value;
    const result = window.sumNumbers(input);
    document.getElementById("sumResult").textContent = JSON.stringify(result);
}

function runValidateEmail() {
    const email = document.getElementById("emailInput").value;
    const result = window.validateEmail(email);
    document.getElementById("emailResult").textContent = JSON.stringify(result, null, 2);
}
