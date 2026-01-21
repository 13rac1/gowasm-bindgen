---
title: "Image Processing Demo"
weight: 1
---

# Image Processing Demo

Process images in the browser using Go compiled to WebAssembly. This demo compares WASM performance against pure JavaScript.

{{< rawhtml >}}
<style>
.demo-container {
  font-family: system-ui, -apple-system, sans-serif;
  max-width: 900px;
  margin: 2rem auto;
}
.image-selector {
  margin-bottom: 1.5rem;
}
.thumbnails {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}
.thumb {
  width: 120px;
  height: 80px;
  object-fit: cover;
  cursor: pointer;
  border: 3px solid transparent;
  border-radius: 8px;
  transition: border-color 0.2s;
}
.thumb:hover {
  border-color: #666;
}
.thumb.selected {
  border-color: #0066cc;
}
.preview-area {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
  margin-bottom: 1.5rem;
}
@media (max-width: 600px) {
  .preview-area {
    grid-template-columns: 1fr;
  }
}
.canvas-container h4 {
  margin: 0 0 0.5rem 0;
  font-size: 0.9rem;
  color: #fff;
  font-weight: bold;
}
.canvas-container canvas {
  width: 100%;
  height: auto;
  border-radius: 8px;
  background: #f0f0f0;
}
.controls {
  display: flex;
  gap: 1rem;
  align-items: center;
  flex-wrap: wrap;
  margin-bottom: 1.5rem;
  padding: 1rem;
  background: #1e2124;
  border-radius: 8px;
}
.controls label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #fff;
}
.controls select, .controls input[type="range"] {
  padding: 0.5rem;
  border: 1px solid #ccc;
  border-radius: 4px;
}
.controls button {
  padding: 0.75rem 1.5rem;
  background: #0066cc;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
}
.controls button:hover {
  background: #0055aa;
}
.controls button:disabled {
  background: #ccc;
  cursor: not-allowed;
}
.results {
  padding: 1rem;
  border-radius: 8px;
  background: #1e2124;
}
.results table {
  width: 100%;
  border-collapse: collapse;
}
.results th, .results td {
  padding: 0.75rem;
  text-align: center;
  border-bottom: 1px solid #555;
  color: #fff;
}
.results th {
  font-weight: 600;
}
.speedup {
  font-size: 1.25rem;
  font-weight: bold;
  color: #66b3ff;
}
.status {
  padding: 0.5rem 1rem;
  margin-bottom: 1rem;
  border-radius: 4px;
  background: #fff3cd;
  color: #856404;
}
.status.ready {
  background: #d4edda;
  color: #155724;
}
.status.error {
  background: #f8d7da;
  color: #721c24;
}
</style>

<div class="demo-container">
  <div id="status" class="status">Loading WASM module...</div>

  <div class="image-selector">
    <h4>Select Image</h4>
    <div class="thumbnails">
      <img src="../../images/sea-ocean-snow-winter-cloud-lake-576858-pxhere.com.jpg" class="thumb selected" data-src="../../images/sea-ocean-snow-winter-cloud-lake-576858-pxhere.com.jpg" alt="Winter Lake" />
      <img src="../../images/cascade-ice-island-327394.jpg" class="thumb" data-src="../../images/cascade-ice-island-327394.jpg" alt="Landscape" />
    </div>
  </div>

  <div class="preview-area">
    <div class="canvas-container">
      <h4>Original</h4>
      <canvas id="original"></canvas>
    </div>
    <div class="canvas-container">
      <h4>Processed</h4>
      <canvas id="processed"></canvas>
    </div>
  </div>

  <div class="controls">
    <label>
      Filter:
      <select id="filter">
        <option value="blur">Gaussian Blur</option>
        <option value="sharpen">Sharpen</option>
      </select>
    </label>

    <label id="radius-control" style="display:flex">
      Radius: <input type="range" id="radius" min="1" max="20" value="5" />
      <span id="radius-value">5</span>
    </label>

    <label id="strength-control" style="display:none">
      Strength: <input type="range" id="strength" min="1" max="10" value="3" />
      <span id="strength-value">3</span>
    </label>

    <button id="run-benchmark" disabled>Run Benchmark</button>
  </div>

  <div class="results">
    <table>
      <tr>
        <th></th>
        <th>Go WASM</th>
        <th>JavaScript</th>
      </tr>
      <tr>
        <td>Time</td>
        <td><span id="wasm-time">-</span> ms</td>
        <td><span id="js-time">-</span> ms</td>
      </tr>
      <tr>
        <td>Speedup</td>
        <td colspan="2" class="speedup"><span id="speedup">-</span></td>
      </tr>
    </table>
  </div>
</div>

<script>
// Get base URL for assets
const baseUrl = document.currentScript ?
  new URL('.', document.currentScript.src).href :
  window.location.href.replace(/\/[^\/]*$/, '/');
</script>
<script src="wasm_exec.js"></script>
<script type="module">
// Import the generated WASM client
const wasmUrl = 'image.wasm';
let wasm = null;
let imageData = null;

const statusEl = document.getElementById('status');
const filterEl = document.getElementById('filter');
const runBtn = document.getElementById('run-benchmark');
const originalCanvas = document.getElementById('original');
const processedCanvas = document.getElementById('processed');

// Pure JavaScript implementations for comparison
const jsFilters = {
  blur(pixels, width, height, radius) {
    // Use same box blur approximation as Go for fair comparison
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
  sharpen(pixels, width, height, strength) {
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
          const val = center * pixels[idx + c] +
            edge * pixels[topIdx + c] +
            edge * pixels[bottomIdx + c] +
            edge * pixels[leftIdx + c] +
            edge * pixels[rightIdx + c];
          result[idx + c] = Math.max(0, Math.min(255, val));
        }
        result[idx + 3] = pixels[idx + 3];
      }
    }
    return result;
  }
};

// Optimized box blur with sliding window - O(n) regardless of radius
function boxBlurH(src, dst, width, height, radius) {
  const div = radius + radius + 1;

  for (let y = 0; y < height; y++) {
    const rowStart = y * width * 4;

    // Initialize accumulator with left edge pixels
    let rSum = 0, gSum = 0, bSum = 0, aSum = 0;
    for (let i = -radius; i <= radius; i++) {
      const px = Math.max(0, i);
      const idx = rowStart + px * 4;
      rSum += src[idx];
      gSum += src[idx + 1];
      bSum += src[idx + 2];
      aSum += src[idx + 3];
    }

    for (let x = 0; x < width; x++) {
      const idx = rowStart + x * 4;
      dst[idx] = (rSum / div) | 0;
      dst[idx + 1] = (gSum / div) | 0;
      dst[idx + 2] = (bSum / div) | 0;
      dst[idx + 3] = (aSum / div) | 0;

      // Slide window
      const leftX = Math.max(0, x - radius);
      const rightX = Math.min(width - 1, x + radius + 1);

      const leftIdx = rowStart + leftX * 4;
      const rightIdx = rowStart + rightX * 4;

      rSum += src[rightIdx] - src[leftIdx];
      gSum += src[rightIdx + 1] - src[leftIdx + 1];
      bSum += src[rightIdx + 2] - src[leftIdx + 2];
      aSum += src[rightIdx + 3] - src[leftIdx + 3];
    }
  }
}

function boxBlurV(src, dst, width, height, radius) {
  const div = radius + radius + 1;

  for (let x = 0; x < width; x++) {
    // Initialize accumulator with top edge pixels
    let rSum = 0, gSum = 0, bSum = 0, aSum = 0;
    for (let i = -radius; i <= radius; i++) {
      const py = Math.max(0, i);
      const idx = py * width * 4 + x * 4;
      rSum += src[idx];
      gSum += src[idx + 1];
      bSum += src[idx + 2];
      aSum += src[idx + 3];
    }

    for (let y = 0; y < height; y++) {
      const idx = y * width * 4 + x * 4;
      dst[idx] = (rSum / div) | 0;
      dst[idx + 1] = (gSum / div) | 0;
      dst[idx + 2] = (bSum / div) | 0;
      dst[idx + 3] = (aSum / div) | 0;

      // Slide window
      const topY = Math.max(0, y - radius);
      const bottomY = Math.min(height - 1, y + radius + 1);

      const topIdx = topY * width * 4 + x * 4;
      const bottomIdx = bottomY * width * 4 + x * 4;

      rSum += src[bottomIdx] - src[topIdx];
      gSum += src[bottomIdx + 1] - src[topIdx + 1];
      bSum += src[bottomIdx + 2] - src[topIdx + 2];
      aSum += src[bottomIdx + 3] - src[topIdx + 3];
    }
  }
}

// Initialize WASM
async function initWasm() {
  try {
    const go = new Go();
    const result = await WebAssembly.instantiateStreaming(fetch(wasmUrl), go.importObject);
    go.run(result.instance);

    // Check that functions are available
    if (typeof gaussianBlur === 'undefined') {
      throw new Error('WASM functions not exported');
    }

    statusEl.textContent = 'WASM loaded! Select an image and run benchmark.';
    statusEl.className = 'status ready';
    runBtn.disabled = false;
    wasm = { gaussianBlur, sharpen };
  } catch (err) {
    statusEl.textContent = 'Error loading WASM: ' + err.message;
    statusEl.className = 'status error';
    console.error(err);
  }
}

// Load image
async function loadImage(src) {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.crossOrigin = 'anonymous';
    img.onload = () => resolve(img);
    img.onerror = reject;
    img.src = src;
  });
}

async function selectImage(src) {
  const img = await loadImage(src);

  // Draw original
  originalCanvas.width = img.width;
  originalCanvas.height = img.height;
  const origCtx = originalCanvas.getContext('2d');
  origCtx.drawImage(img, 0, 0);

  // Store image data
  imageData = origCtx.getImageData(0, 0, img.width, img.height);

  // Setup processed canvas
  processedCanvas.width = img.width;
  processedCanvas.height = img.height;

  // Clear results
  document.getElementById('wasm-time').textContent = '-';
  document.getElementById('js-time').textContent = '-';
  document.getElementById('speedup').textContent = '-';
}

// Run benchmark
async function runBenchmark() {
  if (!imageData || !wasm) return;

  const filter = filterEl.value;
  const pixels = new Uint8Array(imageData.data.buffer.slice(0));

  runBtn.disabled = true;
  runBtn.textContent = 'Running...';

  // Small delay to update UI
  await new Promise(r => setTimeout(r, 10));

  // WASM benchmark
  let wasmResult;
  const wasmStart = performance.now();
  if (filter === 'blur') {
    const radius = parseInt(document.getElementById('radius').value);
    wasmResult = wasm.gaussianBlur(pixels, imageData.width, imageData.height, radius);
  } else if (filter === 'sharpen') {
    const strength = parseInt(document.getElementById('strength').value);
    wasmResult = wasm.sharpen(pixels, imageData.width, imageData.height, strength);
  }
  const wasmTime = performance.now() - wasmStart;

  // JS benchmark
  let jsResult;
  const jsStart = performance.now();
  if (filter === 'blur') {
    const radius = parseInt(document.getElementById('radius').value);
    jsResult = jsFilters.blur(imageData.data, imageData.width, imageData.height, radius);
  } else if (filter === 'sharpen') {
    const strength = parseInt(document.getElementById('strength').value);
    jsResult = jsFilters.sharpen(imageData.data, imageData.width, imageData.height, strength);
  }
  const jsTime = performance.now() - jsStart;

  // Display results
  document.getElementById('wasm-time').textContent = wasmTime.toFixed(2);
  document.getElementById('js-time').textContent = jsTime.toFixed(2);

  const speedup = jsTime / wasmTime;
  document.getElementById('speedup').textContent =
    speedup > 1 ? `${speedup.toFixed(1)}x faster (WASM wins!)` :
    speedup < 1 ? `${(1/speedup).toFixed(1)}x faster (JS wins)` :
    'Same speed';

  // Draw processed image (using WASM result)
  const processedCtx = processedCanvas.getContext('2d');
  const newImageData = new ImageData(
    new Uint8ClampedArray(wasmResult),
    imageData.width,
    imageData.height
  );
  processedCtx.putImageData(newImageData, 0, 0);

  runBtn.disabled = false;
  runBtn.textContent = 'Run Benchmark';
}

// Event listeners
filterEl.addEventListener('change', () => {
  const filter = filterEl.value;
  document.getElementById('radius-control').style.display = filter === 'blur' ? 'flex' : 'none';
  document.getElementById('strength-control').style.display = filter === 'sharpen' ? 'flex' : 'none';
});

document.getElementById('radius').addEventListener('input', (e) => {
  document.getElementById('radius-value').textContent = e.target.value;
});

document.getElementById('strength').addEventListener('input', (e) => {
  document.getElementById('strength-value').textContent = e.target.value;
});

runBtn.addEventListener('click', runBenchmark);

document.querySelectorAll('.thumb').forEach(thumb => {
  thumb.addEventListener('click', () => {
    document.querySelectorAll('.thumb').forEach(t => t.classList.remove('selected'));
    thumb.classList.add('selected');
    selectImage(thumb.dataset.src);
  });
});

// Initialize
initWasm();
selectImage('../../images/sea-ocean-snow-winter-cloud-lake-576858-pxhere.com.jpg');
</script>
{{< /rawhtml >}}

## How It Works

This demo uses gowasm-bindgen to generate TypeScript bindings for Go image processing functions compiled with TinyGo:

```go
// Go code compiled to WASM with TinyGo
func Sharpen(pixels []byte, width, height, strength int) []byte {
    result := make([]byte, len(pixels))
    center := 1 + 4*strength
    edge := -strength

    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            // 5-point convolution kernel
            // ...
        }
    }
    return result
}
```

The same algorithms are implemented in JavaScript for comparison. Modern JS engines are highly optimized, so performance is roughly comparable for these compute-intensive tasks.

## Why Go WASM?

The real value of Go-to-WASM isn't raw speed - modern JavaScript engines are excellent. The benefits are:

- **Code reuse** - Share validation, parsing, or business logic between server and browser
- **Type safety** - Go's type system catches errors at compile time
- **Existing libraries** - Use Go packages directly in the browser
- **Predictable performance** - No JIT warmup or garbage collection pauses

## Source Code

- [Go implementation](https://github.com/13rac1/gowasm-bindgen/tree/main/examples/image-processing/go/main.go)
- [Full example](https://github.com/13rac1/gowasm-bindgen/tree/main/examples/image-processing)
