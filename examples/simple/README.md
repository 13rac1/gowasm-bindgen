# gowasm-bindgen Example

Go WASM functions traditionally required awkward `js.Value` signatures. With gowasm-bindgen, you write normal Go functions with standard types, and the tool generates TypeScript bindings automatically.

## Directory Structure

```
examples/simple/
├── wasm/                # Go source code
│   ├── main.go          # Normal Go functions (no js.Value!)
│   ├── main_test.go     # Unit tests
│   └── bindings_gen.go  # Generated WASM bindings (gitignored)
│
├── src/                 # TypeScript source
│   ├── app.ts           # Browser demo app
│   └── verify_test.ts   # Type verification tests
│
├── public/              # Static assets
│   └── index.html
│
├── generated/           # gowasm-bindgen output (gitignored)
│   ├── go-wasm.ts       # TypeScript class API
│   ├── worker.js        # Web Worker loader
│   ├── wasm.wasm        # Compiled WASM binary
│   └── wasm_exec.js     # Go runtime
│
└── dist/                # Final bundled output (gitignored)
```

## Quick Start

```bash
# Build everything (WASM binary + TypeScript types)
make all

# This runs:
# 1. setup     - Copies wasm_exec.js from TinyGo
# 2. generate  - Creates go-wasm.ts, worker.js, bindings_gen.go
# 3. build     - Compiles Go to WASM with TinyGo
# 4. typecheck - Validates TypeScript types
# 5. verify    - Runs TypeScript tests
# 6. dist      - Bundles everything for serving
```

### Using Standard Go (Alternative)

Standard Go produces larger binaries (~2.4MB vs ~200KB) but has full language support:

```bash
make setup-go   # Copy wasm_exec.js from Go installation
make build-go   # Build with standard Go
```

## Worker Mode (Default - Non-blocking)

By default, gowasm-bindgen generates a Web Worker for non-blocking async calls:

```bash
make generate  # Creates generated/go-wasm.ts + generated/worker.js
```

```typescript
import { GoWasm } from './generated/go-wasm';

const wasm = await GoWasm.init('./worker.js');

// Non-blocking! UI stays responsive
const greeting = await wasm.greet('World');
const result = await wasm.calculate(5, 3, 'add');

wasm.terminate();  // Clean up when done
```

### Sync Mode (Opt-in - Blocking)

Use `--mode sync` flag for synchronous calls that block the main thread:

```bash
make generate-sync  # Creates generated/go-wasm.ts only (no worker.js)
```

```typescript
import { GoWasm } from './generated/go-wasm';

const wasm = await GoWasm.init('./wasm.wasm');  // async load
const greeting = wasm.greet('World');  // sync call, no await
```

## Generated Output

### TypeScript Client (`generated/go-wasm.ts`)

```typescript
export interface FormatUserResult {
  displayName: string;
  status: string;
}

export class GoWasm {
  static async init(workerUrl: string): Promise<GoWasm>;
  greet(name: string): Promise<string>;
  calculate(a: number, b: number, op: string): Promise<number>;
  formatUser(name: string, age: number, active: boolean): Promise<FormatUserResult>;
  terminate(): void;
}
```

### Go Bindings (`wasm/bindings_gen.go`)

Handles `js.Value` conversions automatically:

```go
func init() {
    js.Global().Set("greet", js.FuncOf(wasmGreet))
    // ...
}

func wasmGreet(_ js.Value, args []js.Value) interface{} {
    name := args[0].String()
    return Greet(name)
}
```

## Try the Web Demo

```bash
make serve  # Build and start server at http://localhost:8080
```

## How It Works

1. **Write normal Go functions** in `wasm/main.go`:
   ```go
   func Greet(name string) string {
       return "Hello, " + name + "!"
   }
   ```

2. **Run gowasm-bindgen** (generates TypeScript client, worker.js, and compiles WASM):
   ```bash
   gowasm-bindgen wasm/main.go
   ```

3. Or build with custom options:
   ```bash
   # Generate only (skip WASM compilation)
   gowasm-bindgen wasm/main.go --no-build

   # Compile WASM separately
   tinygo build -o generated/wasm.wasm -target wasm ./wasm/
   ```

## Requirements

- [TinyGo](https://tinygo.org/getting-started/install/) (recommended)
- Go 1.21+ (alternative, larger binaries)
- Node.js 18+
