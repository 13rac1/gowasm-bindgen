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

<script src="demo.js"></script>
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

- [Go implementation](https://github.com/13rac1/gowasm-bindgen/tree/main/examples/image-processing/image/main.go)
- [Full example](https://github.com/13rac1/gowasm-bindgen/tree/main/examples/image-processing)
