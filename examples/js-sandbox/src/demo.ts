/**
 * JavaScript Sandbox Demo
 *
 * Securely execute untrusted JavaScript using Goja (Go's JavaScript interpreter)
 * compiled to WebAssembly.
 */

import { GoGoja, RunJSResult } from '../generated/go-goja';

const statusEl = document.getElementById('status') as HTMLElement;
const runBtn = document.getElementById('run-btn') as HTMLButtonElement;
const codeInput = document.getElementById('code-input') as HTMLTextAreaElement;
const logsOutput = document.getElementById('logs-output') as HTMLElement;
const resultOutput = document.getElementById('result-output') as HTMLElement;
const errorSection = document.getElementById('error-section') as HTMLElement;
const errorOutput = document.getElementById('error-output') as HTMLElement;

let wasm: GoGoja | null = null;

// Initialize WASM
async function init(): Promise<void> {
  try {
    wasm = await GoGoja.init('worker.js');

    statusEl.textContent = 'Sandbox ready! Enter JavaScript code and click Run.';
    statusEl.className = 'status ready';
    runBtn.disabled = false;
  } catch (err) {
    statusEl.textContent = 'Error loading WASM: ' + (err instanceof Error ? err.message : String(err));
    statusEl.className = 'status error';
    console.error(err);
  }
}

// Run code in Goja
async function runCode(): Promise<void> {
  if (!wasm) return;

  const code = codeInput.value;

  runBtn.disabled = true;
  runBtn.textContent = 'Running...';

  try {
    const output: RunJSResult = await wasm.runJS(code);

    // Display logs
    logsOutput.textContent = output.logs || '(no console output)';

    // Display result
    resultOutput.textContent = output.result || '(undefined)';

    // Display error if any
    if (output.errorMsg) {
      errorSection.style.display = 'block';
      errorOutput.textContent = output.errorMsg;
    } else {
      errorSection.style.display = 'none';
    }
  } catch (err) {
    errorSection.style.display = 'block';
    errorOutput.textContent = err instanceof Error ? err.message : String(err);
    logsOutput.textContent = '(execution failed)';
    resultOutput.textContent = '-';
  }

  runBtn.disabled = false;
  runBtn.textContent = 'Run in Sandbox';
}

// Event listeners
runBtn.addEventListener('click', () => void runCode());

// Ctrl+Enter to run
codeInput.addEventListener('keydown', (e: KeyboardEvent) => {
  if (e.ctrlKey && e.key === 'Enter') {
    e.preventDefault();
    if (!runBtn.disabled) {
      void runCode();
    }
  }
});

// Initialize
void init();
