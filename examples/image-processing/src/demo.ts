/**
 * Image Processing Demo - Go WASM vs JavaScript Benchmark
 *
 * This demo compares image processing performance between Go compiled to
 * WebAssembly and pure JavaScript implementations.
 */

import { GoImage } from '../generated/go-image';

// Pure JavaScript implementations for comparison
const jsFilters = {
  blur(pixels: Uint8ClampedArray, width: number, height: number, radius: number): Uint8ClampedArray {
    const temp = new Uint8ClampedArray(pixels.length);
    const result = new Uint8ClampedArray(pixels.length);

    // 3 box blur passes approximate Gaussian
    boxBlurH(pixels, temp, width, height, radius);
    boxBlurV(temp, result, width, height, radius);
    boxBlurH(result, temp, width, height, radius);
    boxBlurV(temp, result, width, height, radius);
    boxBlurH(result, temp, width, height, radius);
    boxBlurV(temp, result, width, height, radius);

    return result;
  },
  sharpen(pixels: Uint8ClampedArray, width: number, height: number, strength: number): Uint8ClampedArray {
    const result = new Uint8ClampedArray(pixels.length);
    const center = 1 + 4 * strength;
    const edge = -strength;

    for (let y = 0; y < height; y++) {
      for (let x = 0; x < width; x++) {
        const idx = (y * width + x) * 4;

        const top = Math.max(0, y - 1);
        const bottom = Math.min(height - 1, y + 1);
        const left = Math.max(0, x - 1);
        const right = Math.min(width - 1, x + 1);

        const topIdx = (top * width + x) * 4;
        const bottomIdx = (bottom * width + x) * 4;
        const leftIdx = (y * width + left) * 4;
        const rightIdx = (y * width + right) * 4;

        for (let c = 0; c < 3; c++) {
          const val = center * pixels[idx + c]! +
            edge * pixels[topIdx + c]! +
            edge * pixels[bottomIdx + c]! +
            edge * pixels[leftIdx + c]! +
            edge * pixels[rightIdx + c]!;
          result[idx + c] = Math.max(0, Math.min(255, val));
        }
        result[idx + 3] = pixels[idx + 3]!;
      }
    }
    return result;
  }
};

// Optimized box blur with sliding window
function boxBlurH(src: Uint8ClampedArray, dst: Uint8ClampedArray, width: number, height: number, radius: number): void {
  const div = radius + radius + 1;

  for (let y = 0; y < height; y++) {
    const rowStart = y * width * 4;

    let rSum = 0, gSum = 0, bSum = 0, aSum = 0;
    for (let i = -radius; i <= radius; i++) {
      const px = Math.max(0, i);
      const idx = rowStart + px * 4;
      rSum += src[idx]!;
      gSum += src[idx + 1]!;
      bSum += src[idx + 2]!;
      aSum += src[idx + 3]!;
    }

    for (let x = 0; x < width; x++) {
      const idx = rowStart + x * 4;
      dst[idx] = (rSum / div) | 0;
      dst[idx + 1] = (gSum / div) | 0;
      dst[idx + 2] = (bSum / div) | 0;
      dst[idx + 3] = (aSum / div) | 0;

      const leftX = Math.max(0, x - radius);
      const rightX = Math.min(width - 1, x + radius + 1);

      const leftIdx = rowStart + leftX * 4;
      const rightIdx = rowStart + rightX * 4;

      rSum += src[rightIdx]! - src[leftIdx]!;
      gSum += src[rightIdx + 1]! - src[leftIdx + 1]!;
      bSum += src[rightIdx + 2]! - src[leftIdx + 2]!;
      aSum += src[rightIdx + 3]! - src[leftIdx + 3]!;
    }
  }
}

function boxBlurV(src: Uint8ClampedArray, dst: Uint8ClampedArray, width: number, height: number, radius: number): void {
  const div = radius + radius + 1;

  for (let x = 0; x < width; x++) {
    let rSum = 0, gSum = 0, bSum = 0, aSum = 0;
    for (let i = -radius; i <= radius; i++) {
      const py = Math.max(0, i);
      const idx = py * width * 4 + x * 4;
      rSum += src[idx]!;
      gSum += src[idx + 1]!;
      bSum += src[idx + 2]!;
      aSum += src[idx + 3]!;
    }

    for (let y = 0; y < height; y++) {
      const idx = y * width * 4 + x * 4;
      dst[idx] = (rSum / div) | 0;
      dst[idx + 1] = (gSum / div) | 0;
      dst[idx + 2] = (bSum / div) | 0;
      dst[idx + 3] = (aSum / div) | 0;

      const topY = Math.max(0, y - radius);
      const bottomY = Math.min(height - 1, y + radius + 1);

      const topIdx = topY * width * 4 + x * 4;
      const bottomIdx = bottomY * width * 4 + x * 4;

      rSum += src[bottomIdx]! - src[topIdx]!;
      gSum += src[bottomIdx + 1]! - src[topIdx + 1]!;
      bSum += src[bottomIdx + 2]! - src[topIdx + 2]!;
      aSum += src[bottomIdx + 3]! - src[topIdx + 3]!;
    }
  }
}

// DOM elements
const statusEl = document.getElementById('status') as HTMLElement;
const filterEl = document.getElementById('filter') as HTMLSelectElement;
const runBtn = document.getElementById('run-benchmark') as HTMLButtonElement;
const originalCanvas = document.getElementById('original') as HTMLCanvasElement;
const processedCanvas = document.getElementById('processed') as HTMLCanvasElement;

let wasm: GoImage | null = null;
let imageData: ImageData | null = null;

// Load image
async function loadImage(src: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.crossOrigin = 'anonymous';
    img.onload = () => resolve(img);
    img.onerror = reject;
    img.src = src;
  });
}

async function selectImage(src: string): Promise<void> {
  const img = await loadImage(src);

  originalCanvas.width = img.width;
  originalCanvas.height = img.height;
  const origCtx = originalCanvas.getContext('2d');
  if (!origCtx) return;
  origCtx.drawImage(img, 0, 0);

  imageData = origCtx.getImageData(0, 0, img.width, img.height);

  processedCanvas.width = img.width;
  processedCanvas.height = img.height;

  const wasmTimeEl = document.getElementById('wasm-time');
  const jsTimeEl = document.getElementById('js-time');
  const speedupEl = document.getElementById('speedup');
  if (wasmTimeEl) wasmTimeEl.textContent = '-';
  if (jsTimeEl) jsTimeEl.textContent = '-';
  if (speedupEl) speedupEl.textContent = '-';
}

// Run benchmark
async function runBenchmark(): Promise<void> {
  if (!imageData || !wasm) return;

  const filter = filterEl.value;
  const pixels = new Uint8Array(imageData.data.buffer.slice(0));

  runBtn.disabled = true;
  runBtn.textContent = 'Running...';

  await new Promise(r => setTimeout(r, 10));

  // WASM benchmark (async worker mode)
  let wasmResult: Uint8Array;
  const wasmStart = performance.now();
  if (filter === 'blur') {
    const radiusEl = document.getElementById('radius') as HTMLInputElement;
    const radius = parseInt(radiusEl.value);
    wasmResult = await wasm.gaussianBlur(pixels, imageData.width, imageData.height, radius);
  } else {
    const strengthEl = document.getElementById('strength') as HTMLInputElement;
    const strength = parseInt(strengthEl.value);
    wasmResult = await wasm.sharpen(pixels, imageData.width, imageData.height, strength);
  }
  const wasmTime = performance.now() - wasmStart;

  // JS benchmark
  let jsResult: Uint8ClampedArray;
  const jsStart = performance.now();
  if (filter === 'blur') {
    const radiusEl = document.getElementById('radius') as HTMLInputElement;
    const radius = parseInt(radiusEl.value);
    jsResult = jsFilters.blur(imageData.data, imageData.width, imageData.height, radius);
  } else {
    const strengthEl = document.getElementById('strength') as HTMLInputElement;
    const strength = parseInt(strengthEl.value);
    jsResult = jsFilters.sharpen(imageData.data, imageData.width, imageData.height, strength);
  }
  const jsTime = performance.now() - jsStart;

  // Display results
  const wasmTimeEl = document.getElementById('wasm-time');
  const jsTimeEl = document.getElementById('js-time');
  const speedupEl = document.getElementById('speedup');
  if (wasmTimeEl) wasmTimeEl.textContent = wasmTime.toFixed(2);
  if (jsTimeEl) jsTimeEl.textContent = jsTime.toFixed(2);

  const speedup = jsTime / wasmTime;
  if (speedupEl) {
    speedupEl.textContent =
      speedup > 1 ? `${speedup.toFixed(1)}x faster (WASM wins!)` :
      speedup < 1 ? `${(1/speedup).toFixed(1)}x faster (JS wins)` :
      'Same speed';
  }

  // Draw processed image
  const processedCtx = processedCanvas.getContext('2d');
  if (processedCtx) {
    const newImageData = new ImageData(
      new Uint8ClampedArray(wasmResult),
      imageData.width,
      imageData.height
    );
    processedCtx.putImageData(newImageData, 0, 0);
  }

  runBtn.disabled = false;
  runBtn.textContent = 'Run Benchmark';
}

// Initialize
async function init(): Promise<void> {
  try {
    wasm = await GoImage.init('worker.js');

    statusEl.textContent = 'WASM loaded! Select an image and run benchmark.';
    statusEl.className = 'status ready';
    runBtn.disabled = false;
  } catch (err) {
    statusEl.textContent = 'Error loading WASM: ' + (err instanceof Error ? err.message : String(err));
    statusEl.className = 'status error';
    console.error(err);
  }
}

// Event listeners
filterEl.addEventListener('change', () => {
  const filter = filterEl.value;
  const radiusControl = document.getElementById('radius-control');
  const strengthControl = document.getElementById('strength-control');
  if (radiusControl) radiusControl.style.display = filter === 'blur' ? 'flex' : 'none';
  if (strengthControl) strengthControl.style.display = filter === 'sharpen' ? 'flex' : 'none';
});

const radiusInput = document.getElementById('radius') as HTMLInputElement;
radiusInput.addEventListener('input', () => {
  const radiusValue = document.getElementById('radius-value');
  if (radiusValue) radiusValue.textContent = radiusInput.value;
});

const strengthInput = document.getElementById('strength') as HTMLInputElement;
strengthInput.addEventListener('input', () => {
  const strengthValue = document.getElementById('strength-value');
  if (strengthValue) strengthValue.textContent = strengthInput.value;
});

runBtn.addEventListener('click', () => void runBenchmark());

document.querySelectorAll('.thumb').forEach(thumb => {
  thumb.addEventListener('click', () => {
    document.querySelectorAll('.thumb').forEach(t => t.classList.remove('selected'));
    thumb.classList.add('selected');
    const src = (thumb as HTMLImageElement).dataset['src'];
    if (src) void selectImage(src);
  });
});

// Initialize
void init();
void selectImage('../../images/sea-ocean-snow-winter-cloud-lake-576858-pxhere.com.jpg');
